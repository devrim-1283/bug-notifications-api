package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/devrimsoft/bug-notifications-api/internal/middleware"
	"github.com/devrimsoft/bug-notifications-api/internal/model"
	"github.com/devrimsoft/bug-notifications-api/internal/queue"
	"github.com/devrimsoft/bug-notifications-api/internal/validate"
	"github.com/google/uuid"
)

type Handler struct {
	producer *queue.Producer
}

func NewHandler(producer *queue.Producer) *Handler {
	return &Handler{producer: producer}
}

// CreateReport handles POST /v1/reports
func (h *Handler) CreateReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, model.ErrorResponse{
			Error: "method not allowed",
			Code:  "METHOD_NOT_ALLOWED",
		})
		return
	}

	var req model.ReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, model.ErrorResponse{
			Error: "invalid JSON body",
			Code:  "INVALID_JSON",
		})
		return
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

	slog.Info("report queued", "event_id", eventID, "site_id", req.SiteID)

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
