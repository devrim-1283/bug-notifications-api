package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port              int
	RedisURL          string
	DatabaseURL       string
	Sites             []string // allowed site domains
	RateLimitRPS      int
	WorkerConcurrency int
	TLSCertFile       string
	TLSKeyFile        string
	TrustedProxies    []*net.IPNet
	ImageAPIURL       string
	ImageAPIKey       string
	PortalDomain      string
	TurnstileSiteKey  string
	TurnstileSecretKey string
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	cfg := &Config{
		Port:              8080,
		RateLimitRPS:      10,
		WorkerConcurrency: 10,
	}

	if p := os.Getenv("PORT"); p != "" {
		port, err := strconv.Atoi(p)
		if err != nil {
			return nil, fmt.Errorf("invalid PORT: %w", err)
		}
		cfg.Port = port
	}

	cfg.RedisURL = os.Getenv("REDIS_URL")
	if cfg.RedisURL == "" {
		cfg.RedisURL = "redis://localhost:6379"
	}

	cfg.DatabaseURL = os.Getenv("DATABASE_URL")
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	// ALLOWED_SITES format: "example.com,other.com,shop.example.com"
	allowedSites := os.Getenv("ALLOWED_SITES")
	if allowedSites == "" {
		return nil, fmt.Errorf("ALLOWED_SITES is required")
	}
	for _, domain := range strings.Split(allowedSites, ",") {
		domain = strings.ToLower(strings.TrimSpace(domain))
		if domain == "" {
			continue
		}
		cfg.Sites = append(cfg.Sites, domain)
	}

	if r := os.Getenv("RATE_LIMIT_RPS"); r != "" {
		rps, err := strconv.Atoi(r)
		if err != nil {
			return nil, fmt.Errorf("invalid RATE_LIMIT_RPS: %w", err)
		}
		cfg.RateLimitRPS = rps
	}

	if w := os.Getenv("WORKER_CONCURRENCY"); w != "" {
		wc, err := strconv.Atoi(w)
		if err != nil {
			return nil, fmt.Errorf("invalid WORKER_CONCURRENCY: %w", err)
		}
		cfg.WorkerConcurrency = wc
	}

	cfg.ImageAPIURL = os.Getenv("IMAGE_API_URL")
	cfg.ImageAPIKey = os.Getenv("IMAGE_API_KEY")

	cfg.PortalDomain = strings.ToLower(strings.TrimSpace(os.Getenv("PORTAL_DOMAIN")))
	cfg.TurnstileSiteKey = os.Getenv("TURNSTILE_SITE_KEY")
	cfg.TurnstileSecretKey = os.Getenv("TURNSTILE_SECRET_KEY")

	cfg.TLSCertFile = os.Getenv("TLS_CERT_FILE")
	cfg.TLSKeyFile = os.Getenv("TLS_KEY_FILE")
	if (cfg.TLSCertFile == "") != (cfg.TLSKeyFile == "") {
		return nil, fmt.Errorf("TLS_CERT_FILE and TLS_KEY_FILE must both be set or both be empty")
	}

	// TRUSTED_PROXIES format: "10.0.0.0/8,172.16.0.0/12,192.168.1.1"
	if tp := os.Getenv("TRUSTED_PROXIES"); tp != "" {
		for _, entry := range strings.Split(tp, ",") {
			entry = strings.TrimSpace(entry)
			if entry == "" {
				continue
			}
			// Single IP without CIDR notation — add /32 (IPv4) or /128 (IPv6)
			if !strings.Contains(entry, "/") {
				if strings.Contains(entry, ":") {
					entry += "/128"
				} else {
					entry += "/32"
				}
			}
			_, network, err := net.ParseCIDR(entry)
			if err != nil {
				return nil, fmt.Errorf("invalid TRUSTED_PROXIES entry %q: %w", entry, err)
			}
			cfg.TrustedProxies = append(cfg.TrustedProxies, network)
		}
	}

	return cfg, nil
}

// AllowedDomains returns a list of all registered site domains.
func (c *Config) AllowedDomains() []string {
	return c.Sites
}

// AllowedOrigins returns full origin URLs (https://domain) for CORS.
// Includes the portal domain if configured.
func (c *Config) AllowedOrigins() []string {
	seen := make(map[string]bool)
	var origins []string
	for _, d := range c.Sites {
		o := "https://" + d
		if !seen[o] {
			origins = append(origins, o)
			seen[o] = true
		}
	}
	if c.PortalDomain != "" {
		o := "https://" + c.PortalDomain
		if !seen[o] {
			origins = append(origins, o)
			seen[o] = true
		}
	}
	return origins
}

// FindSiteByDomain returns the domain if it's in the allowed sites list, or empty string.
func (c *Config) FindSiteByDomain(domain string) string {
	domain = strings.ToLower(domain)
	for _, d := range c.Sites {
		if d == domain {
			return d
		}
	}
	return ""
}

// TLSEnabled returns true if TLS certificate and key files are configured.
func (c *Config) TLSEnabled() bool {
	return c.TLSCertFile != "" && c.TLSKeyFile != ""
}

// IsPortal returns true if the given domain matches the portal domain.
func (c *Config) IsPortal(domain string) bool {
	return c.PortalDomain != "" && strings.ToLower(domain) == c.PortalDomain
}

// ReportableDomains returns all registered site domains (portal is not in Sites).
func (c *Config) ReportableDomains() []string {
	return c.Sites
}
