package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

// SiteConfig holds the allowed domain and its API key.
type SiteConfig struct {
	Domain string
	APIKey string
}

type Config struct {
	Port              int
	RedisURL          string
	DatabaseURL       string
	Sites             []SiteConfig
	RateLimitRPS      int
	WorkerConcurrency int
	TLSCertFile       string
	TLSKeyFile        string
	TrustedProxies    []*net.IPNet
	ImageAPIURL        string
	ImageAPIKey        string
	PortalDomain       string
	TurnstileSiteKey   string
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

	// SITE_KEYS format: "example.com:secret123,other.com:key456"
	siteKeys := os.Getenv("SITE_KEYS")
	if siteKeys == "" {
		return nil, fmt.Errorf("SITE_KEYS is required")
	}
	for _, pair := range strings.Split(siteKeys, ",") {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		parts := strings.SplitN(pair, ":", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return nil, fmt.Errorf("invalid SITE_KEYS entry: %q (expected domain:key)", pair)
		}
		cfg.Sites = append(cfg.Sites, SiteConfig{
			Domain: strings.ToLower(strings.TrimSpace(parts[0])),
			APIKey: strings.TrimSpace(parts[1]),
		})
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

// AllowedDomains returns a list of all registered domains.
func (c *Config) AllowedDomains() []string {
	domains := make([]string, len(c.Sites))
	for i, s := range c.Sites {
		domains[i] = s.Domain
	}
	return domains
}

// AllowedOrigins returns full origin URLs (https://domain) for CORS.
func (c *Config) AllowedOrigins() []string {
	origins := make([]string, len(c.Sites))
	for i, s := range c.Sites {
		origins[i] = "https://" + s.Domain
	}
	return origins
}

// FindSiteByAPIKey returns the SiteConfig matching the given API key, or nil.
func (c *Config) FindSiteByAPIKey(apiKey string) *SiteConfig {
	for i := range c.Sites {
		if c.Sites[i].APIKey == apiKey {
			return &c.Sites[i]
		}
	}
	return nil
}

// TLSEnabled returns true if TLS certificate and key files are configured.
func (c *Config) TLSEnabled() bool {
	return c.TLSCertFile != "" && c.TLSKeyFile != ""
}

// FindSiteByDomain returns the SiteConfig matching the given domain, or nil.
func (c *Config) FindSiteByDomain(domain string) *SiteConfig {
	domain = strings.ToLower(domain)
	for i := range c.Sites {
		if c.Sites[i].Domain == domain {
			return &c.Sites[i]
		}
	}
	return nil
}

// IsPortal returns true if the given domain matches the portal domain.
func (c *Config) IsPortal(domain string) bool {
	return c.PortalDomain != "" && strings.ToLower(domain) == c.PortalDomain
}

// ReportableDomains returns all registered domains except the portal domain.
func (c *Config) ReportableDomains() []string {
	domains := make([]string, 0)
	for _, s := range c.Sites {
		if !c.IsPortal(s.Domain) {
			domains = append(domains, s.Domain)
		}
	}
	return domains
}

// PortalSite returns the SiteConfig for the portal domain, or nil.
func (c *Config) PortalSite() *SiteConfig {
	if c.PortalDomain == "" {
		return nil
	}
	return c.FindSiteByDomain(c.PortalDomain)
}
