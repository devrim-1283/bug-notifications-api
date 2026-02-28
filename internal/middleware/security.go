package middleware

import (
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/devrimsoft/bug-notifications-api/internal/config"
)

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

// BrowserOnly is a casual-abuse filter, NOT a security boundary.
// A determined attacker can spoof all checked headers. Real security comes from
// CORS enforcement and Turnstile bot verification.
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
// Allows origins from the portal domain and its subdomains.
func CORSMiddleware(cfg *config.Config) func(http.Handler) http.Handler {
	// Build allowed domains list: portal domain + all site domains
	var domains []string
	if cfg.PortalDomain != "" {
		domains = append(domains, cfg.PortalDomain)
	}
	domains = append(domains, cfg.Sites...)

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
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
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
