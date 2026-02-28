package api

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/devrimsoft/bug-notifications-api/internal/config"
	"github.com/devrimsoft/bug-notifications-api/internal/middleware"
	"github.com/devrimsoft/bug-notifications-api/internal/model"
	"github.com/devrimsoft/bug-notifications-api/internal/queue"
	"github.com/devrimsoft/bug-notifications-api/internal/validate"
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

	// Override site_id from the authenticated context (don't trust client)
	siteID := middleware.GetSiteID(r.Context())
	if siteID != "" {
		req.SiteID = siteID
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
