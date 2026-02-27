package middleware

import (
	"crypto/subtle"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/devrimsoft/bug-notifications-api/internal/config"
)

type contextKey string

const SiteIDKey contextKey = "site_id"

// SecureHeaders adds defense-in-depth security headers to every response.
func SecureHeaders() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
			next.ServeHTTP(w, r)
		})
	}
}

// RequireHTTPS enforces HTTPS on all requests. It checks for direct TLS or
// the X-Forwarded-Proto header set by a trusted reverse proxy. Non-HTTPS
// requests are rejected with 403. HSTS header is added to all responses.
func RequireHTTPS() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Allow loopback requests (container health checks)
			if isLoopback(r.RemoteAddr) {
				next.ServeHTTP(w, r)
				return
			}

			isHTTPS := r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https")

			if !isHTTPS {
				slog.Warn("rejected non-HTTPS request",
					"remote_addr", r.RemoteAddr,
					"proto", r.Header.Get("X-Forwarded-Proto"),
				)
				writeJSON(w, http.StatusForbidden, map[string]string{
					"error": "HTTPS required",
					"code":  "HTTPS_REQUIRED",
				})
				return
			}

			// HSTS: tell browsers to always use HTTPS (1 year, include subdomains)
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			next.ServeHTTP(w, r)
		})
	}
}

// APIKeyAuth validates the X-API-Key header and ensures it matches a registered site.
// It also cross-checks that the Origin/Referer domain matches the site associated with the key.
func APIKeyAuth(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiKey := r.Header.Get("X-API-Key")
			if apiKey == "" {
				writeJSON(w, http.StatusUnauthorized, map[string]string{
					"error": "missing X-API-Key header",
					"code":  "MISSING_API_KEY",
				})
				return
			}

			site := cfg.FindSiteByAPIKey(apiKey)
			if site == nil {
				writeJSON(w, http.StatusForbidden, map[string]string{
					"error": "invalid API key",
					"code":  "INVALID_API_KEY",
				})
				return
			}

			// Cross-check: the request origin must match the site domain for this key
			originDomain := extractOriginDomain(r)
			if originDomain == "" {
				writeJSON(w, http.StatusForbidden, map[string]string{
					"error": "request origin could not be determined",
					"code":  "MISSING_ORIGIN",
				})
				return
			}

			if !domainMatch(originDomain, site.Domain) {
				slog.Warn("API key / origin mismatch",
					"api_key_domain", site.Domain,
					"origin_domain", originDomain,
				)
				writeJSON(w, http.StatusForbidden, map[string]string{
					"error": "API key does not match request origin",
					"code":  "ORIGIN_MISMATCH",
				})
				return
			}

			// Store site_id (domain) in context
			ctx := r.Context()
			ctx = withValue(ctx, SiteIDKey, site.Domain)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// BrowserOnly is a casual-abuse filter, NOT a security boundary.
// A determined attacker can spoof all checked headers. Real security comes from
// API key + origin cross-check in APIKeyAuth and CORS enforcement.
// This middleware only raises the bar for lazy/accidental misuse.
func BrowserOnly() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Browsers always send Origin for CORS POST requests
			origin := r.Header.Get("Origin")
			referer := r.Header.Get("Referer")
			if origin == "" && referer == "" {
				writeJSON(w, http.StatusForbidden, map[string]string{
					"error": "browser origin required",
					"code":  "NO_BROWSER_ORIGIN",
				})
				return
			}

			// Sec-Fetch-Site is set by browsers and cannot be overridden by JS.
			// It's the strongest signal we have — more reliable than User-Agent.
			secFetchSite := r.Header.Get("Sec-Fetch-Site")
			if secFetchSite == "" {
				writeJSON(w, http.StatusForbidden, map[string]string{
					"error": "missing browser security headers",
					"code":  "MISSING_SEC_HEADERS",
				})
				return
			}

			// Only allow cross-site or same-origin requests (not "none" which means direct navigation)
			if secFetchSite != "cross-site" && secFetchSite != "same-origin" && secFetchSite != "same-site" {
				writeJSON(w, http.StatusForbidden, map[string]string{
					"error": "invalid request context",
					"code":  "INVALID_SEC_FETCH",
				})
				return
			}

			// Sec-Fetch-Mode must be "cors" for cross-origin API calls from browsers
			secFetchMode := r.Header.Get("Sec-Fetch-Mode")
			if secFetchMode != "cors" && secFetchMode != "same-origin" {
				writeJSON(w, http.StatusForbidden, map[string]string{
					"error": "invalid fetch mode",
					"code":  "INVALID_SEC_FETCH_MODE",
				})
				return
			}

			// Sec-Fetch-Dest must be "empty" for fetch/XHR calls (not "document", "image", etc.)
			secFetchDest := r.Header.Get("Sec-Fetch-Dest")
			if secFetchDest != "empty" {
				writeJSON(w, http.StatusForbidden, map[string]string{
					"error": "invalid fetch destination",
					"code":  "INVALID_SEC_FETCH_DEST",
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// CORSMiddleware handles CORS preflight and response headers.
// Allows origins from registered site domains and their subdomains.
// e.g. if "example.com" is registered, "staging.example.com" is also allowed.
func CORSMiddleware(cfg *config.Config) func(http.Handler) http.Handler {
	domains := cfg.AllowedDomains()

	isAllowed := func(origin string) bool {
		u, err := url.Parse(origin)
		if err != nil || u.Scheme != "https" {
			return false
		}
		host := strings.ToLower(u.Hostname())
		for _, d := range domains {
			if host == d || strings.HasSuffix(host, "."+d) {
				return true
			}
		}
		return false
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			if origin != "" && isAllowed(origin) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-API-Key")
				w.Header().Set("Access-Control-Max-Age", "86400")
				w.Header().Set("Vary", "Origin")
			}

			if r.Method == http.MethodOptions {
				if origin == "" || !isAllowed(origin) {
					w.WriteHeader(http.StatusForbidden)
					return
				}
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// BodyLimit restricts the maximum request body size.
func BodyLimit(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}

// --- helpers ---

func isLoopback(remoteAddr string) bool {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		host = remoteAddr
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func extractOriginDomain(r *http.Request) string {
	origin := r.Header.Get("Origin")
	if origin != "" {
		if u, err := url.Parse(origin); err == nil && u.Hostname() != "" {
			return strings.ToLower(u.Hostname())
		}
	}
	referer := r.Header.Get("Referer")
	if referer != "" {
		if u, err := url.Parse(referer); err == nil && u.Hostname() != "" {
			return strings.ToLower(u.Hostname())
		}
	}
	return ""
}

// domainMatch checks if got matches want exactly or is a subdomain of want.
// e.g. domainMatch("staging.example.com", "example.com") returns true.
func domainMatch(got, want string) bool {
	if subtle.ConstantTimeCompare([]byte(got), []byte(want)) == 1 {
		return true
	}
	// Subdomain check: got must end with ".want"
	suffix := "." + want
	return len(got) > len(suffix) && strings.HasSuffix(got, suffix)
}
