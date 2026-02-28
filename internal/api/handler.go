package api

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/devrimsoft/bug-notifications-api/internal/config"
	"github.com/devrimsoft/bug-notifications-api/internal/middleware"
	"github.com/devrimsoft/bug-notifications-api/internal/model"
	"github.com/devrimsoft/bug-notifications-api/internal/queue"
	"github.com/devrimsoft/bug-notifications-api/internal/validate"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

const MaxImages = 5

type Handler struct {
	producer *queue.Producer
	cfg      *config.Config
}

func NewHandler(producer *queue.Producer, cfg *config.Config) *Handler {
	return &Handler{producer: producer, cfg: cfg}
}

// CreateReport handles POST /v1/reports
// Accepts application/json (no images) or multipart/form-data (with optional images).
func (h *Handler) CreateReport(w http.ResponseWriter, r *http.Request) {
	var req model.ReportRequest

	ct := r.Header.Get("Content-Type")

	// Validated images ready for upload: [{data, filename}, ...]
	type pendingImage struct {
		data     []byte
		filename string
	}
	var pendingImages []pendingImage

	if strings.HasPrefix(ct, "multipart/form-data") {
		// 5 images * 5MB + 1MB form overhead
		if err := r.ParseMultipartForm(MaxImages*MaxImageSize + 1024*1024); err != nil {
			writeJSON(w, http.StatusBadRequest, model.ErrorResponse{
				Error: "invalid multipart form",
				Code:  "INVALID_FORM",
			})
			return
		}

		req.SiteID = r.FormValue("site_id")
		req.ReportType = model.ReportType(r.FormValue("report_type"))
		req.Title = r.FormValue("title")
		req.Description = r.FormValue("description")
		req.Category = model.Category(r.FormValue("category"))

		if v := r.FormValue("page_url"); v != "" {
			req.PageURL = &v
		}
		if v := r.FormValue("contact_type"); v != "" {
			req.ContactType = &v
		}
		if v := r.FormValue("contact_value"); v != "" {
			req.ContactValue = &v
		}
		if v := r.FormValue("first_name"); v != "" {
			req.FirstName = &v
		}
		if v := r.FormValue("last_name"); v != "" {
			req.LastName = &v
		}

		// Multiple images: field name "images" (multiple files)
		if r.MultipartForm != nil && r.MultipartForm.File != nil {
			files := r.MultipartForm.File["images"]
			if len(files) > MaxImages {
				writeJSON(w, http.StatusBadRequest, model.ErrorResponse{
					Error: fmt.Sprintf("maximum %d images allowed", MaxImages),
					Code:  "TOO_MANY_IMAGES",
				})
				return
			}

			for i, fh := range files {
				data, _, err := validateImage(fh)
				if err != nil {
					writeJSON(w, http.StatusBadRequest, model.ErrorResponse{
						Error: fmt.Sprintf("image[%d]: %s", i, err.Error()),
						Code:  "INVALID_IMAGE",
					})
					return
				}
				pendingImages = append(pendingImages, pendingImage{data: data, filename: fh.Filename})
			}
		}
	} else {
		// JSON body (no images)
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, model.ErrorResponse{
				Error: "invalid JSON body",
				Code:  "INVALID_JSON",
			})
			return
		}
	}

	// Determine site_id: portal mode lets client choose; normal mode overrides from auth context
	siteID := middleware.GetSiteID(r.Context())
	if siteID != "" {
		if h.cfg.IsPortal(siteID) {
			// Portal mode: validate that client-supplied site_id is a registered reportable domain
			if req.SiteID == "" {
				writeJSON(w, http.StatusBadRequest, model.ErrorResponse{
					Error: "site_id is required",
					Code:  "MISSING_SITE_ID",
				})
				return
			}
			target := h.cfg.FindSiteByDomain(req.SiteID)
			if target == nil || h.cfg.IsPortal(target.Domain) {
				writeJSON(w, http.StatusBadRequest, model.ErrorResponse{
					Error: "invalid site_id",
					Code:  "INVALID_SITE_ID",
				})
				return
			}
		} else {
			// Normal mode: override site_id from authenticated context (don't trust client)
			req.SiteID = siteID
		}
	}

	// Validate
	if errs := validate.ReportRequest(&req); len(errs) > 0 {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]any{
			"error":  "validation failed",
			"code":   "VALIDATION_ERROR",
			"fields": errs,
		})
		return
	}

	// Turnstile verification — required when configured
	if h.cfg.TurnstileSecretKey == "" {
		writeJSON(w, http.StatusServiceUnavailable, model.ErrorResponse{
			Error: "turnstile is not configured",
			Code:  "TURNSTILE_NOT_CONFIGURED",
		})
		return
	}
	{
		var turnstileToken string
		if strings.HasPrefix(ct, "multipart/form-data") {
			turnstileToken = r.FormValue("cf-turnstile-response")
		} else {
			turnstileToken = r.Header.Get("X-Turnstile-Token")
		}
		if err := verifyTurnstile(h.cfg.TurnstileSecretKey, turnstileToken, r.RemoteAddr); err != nil {
			slog.Warn("turnstile verification failed", "error", err, "remote_addr", r.RemoteAddr)
			writeJSON(w, http.StatusForbidden, model.ErrorResponse{
				Error: "bot verification failed",
				Code:  "TURNSTILE_FAILED",
			})
			return
		}
	}

	// Upload images to R2
	if len(pendingImages) > 0 {
		if h.cfg.ImageAPIURL == "" || h.cfg.ImageAPIKey == "" {
			slog.Error("image upload attempted but IMAGE_API_URL/IMAGE_API_KEY not configured")
			writeJSON(w, http.StatusServiceUnavailable, model.ErrorResponse{
				Error: "image upload is not configured",
				Code:  "IMAGE_NOT_CONFIGURED",
			})
			return
		}

		for i, img := range pendingImages {
			imageURL, err := uploadToR2(h.cfg.ImageAPIURL, h.cfg.ImageAPIKey, img.data, img.filename)
			if err != nil {
				slog.Error("r2 image upload failed", "error", err, "index", i, "filename", img.filename)
				writeJSON(w, http.StatusBadGateway, model.ErrorResponse{
					Error: fmt.Sprintf("image[%d] upload failed", i),
					Code:  "IMAGE_UPLOAD_FAILED",
				})
				return
			}
			req.ImageURLs = append(req.ImageURLs, imageURL)
		}
	}

	// Build queue message
	eventID := uuid.New().String()
	msg := &model.QueueMessage{
		EventID:      eventID,
		SiteID:       req.SiteID,
		ReportType:   req.ReportType,
		Title:        strings.TrimSpace(req.Title),
		Description:  strings.TrimSpace(req.Description),
		Category:     req.Category,
		PageURL:      req.PageURL,
		ContactType:  req.ContactType,
		ContactValue: req.ContactValue,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		ImageURLs:    req.ImageURLs,
		ReceivedAt:   time.Now().UTC().Format(time.RFC3339),
		RetryCount:   0,
	}

	// Enqueue
	if err := h.producer.Enqueue(r.Context(), msg); err != nil {
		slog.Error("enqueue failed", "error", err)
		writeJSON(w, http.StatusServiceUnavailable, model.ErrorResponse{
			Error: "service temporarily unavailable",
			Code:  "QUEUE_ERROR",
		})
		return
	}

	slog.Info("report queued", "event_id", eventID, "site_id", req.SiteID, "images", len(req.ImageURLs))

	writeJSON(w, http.StatusAccepted, model.ReportResponse{
		EventID: eventID,
		Queued:  true,
	})
}

// ListSites handles GET /v1/sites — returns reportable domains (portal excluded).
func (h *Handler) ListSites(w http.ResponseWriter, r *http.Request) {
	domains := h.cfg.ReportableDomains()
	if domains == nil {
		domains = []string{}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"sites": domains,
	})
}

// MountFrontend sets up SPA serving from the embedded dist/ directory.
// Static assets are served with cache headers; all other paths get index.html with config injected.
func (h *Handler) MountFrontend(r chi.Router, embeddedFS embed.FS) {
	// Sub-filesystem rooted at frontend/dist
	distFS, err := fs.Sub(embeddedFS, "frontend/dist")
	if err != nil {
		slog.Error("failed to open embedded frontend/dist", "error", err)
		return
	}

	// Pre-read index.html and inject config once at startup
	indexBytes, err := fs.ReadFile(distFS, "index.html")
	if err != nil {
		slog.Error("failed to read embedded index.html", "error", err)
		return
	}

	portalSite := h.cfg.PortalSite()
	apiKey := ""
	if portalSite != nil {
		apiKey = portalSite.APIKey
	}

	reportableDomains := h.cfg.ReportableDomains()

	configJSON, _ := json.Marshal(map[string]any{
		"apiKey":           apiKey,
		"turnstileSiteKey": h.cfg.TurnstileSiteKey,
		"sites":            reportableDomains,
		"portalDomain":     h.cfg.PortalDomain,
	})

	injectedHTML := strings.Replace(string(indexBytes), "__APP_CONFIG_JSON__", string(configJSON), 1)

	serveIndex := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' https://challenges.cloudflare.com; style-src 'self' 'unsafe-inline'; img-src 'self' data: blob:; connect-src 'self'; frame-src https://challenges.cloudflare.com; frame-ancestors 'none'; font-src 'self' data:")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(injectedHTML))
	}

	fileServer := http.FileServer(http.FS(distFS))

	// Catch-all: serve static files if they exist, otherwise serve index.html (SPA fallback)
	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		// Clean the path
		urlPath := r.URL.Path
		if urlPath == "/" || urlPath == "" {
			serveIndex(w, r)
			return
		}

		// Strip leading slash for fs lookup
		filePath := strings.TrimPrefix(urlPath, "/")

		// Check if the file exists in the embedded filesystem
		if info, err := fs.Stat(distFS, filePath); err == nil && !info.IsDir() {
			// Serve static asset with appropriate caching
			ext := path.Ext(filePath)
			if ext == ".js" || ext == ".css" || ext == ".woff2" || ext == ".woff" || ext == ".ttf" {
				w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			}
			fileServer.ServeHTTP(w, r)
			return
		}

		// SPA fallback: serve index.html
		serveIndex(w, r)
	})
}

// HealthCheck handles GET /health
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
