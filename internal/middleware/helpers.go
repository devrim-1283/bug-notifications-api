package middleware

import (
	"context"
	"encoding/json"
	"net/http"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func withValue(ctx context.Context, key contextKey, val string) context.Context {
	return context.WithValue(ctx, key, val)
}

// GetSiteID retrieves the site_id from request context.
func GetSiteID(ctx context.Context) string {
	if v, ok := ctx.Value(SiteIDKey).(string); ok {
		return v
	}
	return ""
}
