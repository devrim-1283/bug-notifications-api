package middleware

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

// tokenBucketScript implements an atomic token bucket rate limiter in Redis.
//
// KEYS[1]: rate limit key ("rl:{ip}")
// ARGV[1]: current time in milliseconds
// ARGV[2]: refill rate (tokens per second)
// ARGV[3]: burst size (max tokens)
// ARGV[4]: key TTL in seconds
//
// Returns 1 if allowed, 0 if rate limited.
var tokenBucketScript = redis.NewScript(`
local key    = KEYS[1]
local now_ms = tonumber(ARGV[1])
local rate   = tonumber(ARGV[2])
local burst  = tonumber(ARGV[3])
local ttl    = tonumber(ARGV[4])

local data    = redis.call('HMGET', key, 't', 'ts')
local tokens  = tonumber(data[1])
local last_ms = tonumber(data[2])

if tokens == nil then
    tokens  = burst
    last_ms = now_ms
end

local elapsed_s = math.max(0, (now_ms - last_ms)) / 1000
tokens = math.min(burst, tokens + elapsed_s * rate)

if tokens < 1 then
    redis.call('HSET', key, 't', tostring(tokens), 'ts', tostring(now_ms))
    redis.call('EXPIRE', key, ttl)
    return 0
end

tokens = tokens - 1
redis.call('HSET', key, 't', tostring(tokens), 'ts', tostring(now_ms))
redis.call('EXPIRE', key, ttl)
return 1
`)

// RateLimit implements a distributed token bucket rate limiter backed by Redis.
// All API instances share the same counters, preventing bypass via load balancing.
func RateLimit(rdb *redis.Client, rps int, trustedProxies []*net.IPNet) func(http.Handler) http.Handler {
	rate := float64(rps)
	burst := float64(rps * 2)
	ttl := 300 // key TTL in seconds (expire after 5 min of inactivity)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := realIP(r, trustedProxies)
			key := "rl:" + ip
			nowMs := time.Now().UnixMilli()

			ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
			defer cancel()

			result, err := tokenBucketScript.Run(ctx, rdb, []string{key},
				nowMs, rate, burst, ttl,
			).Int()

			if err != nil {
				// Fail-open: allow request on Redis error but log it
				slog.Error("rate limiter redis error", "error", err, "ip", ip)
				next.ServeHTTP(w, r)
				return
			}

			if result == 0 {
				w.Header().Set("Retry-After", "1")
				writeJSON(w, http.StatusTooManyRequests, map[string]string{
					"error": "rate limit exceeded",
					"code":  "RATE_LIMITED",
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// realIP extracts the client IP address from the request.
// X-Forwarded-For and X-Real-IP headers are only trusted when the immediate
// connection (RemoteAddr) comes from a configured trusted proxy. This prevents
// spoofing when the application is directly exposed to the internet.
func realIP(r *http.Request, trustedProxies []*net.IPNet) string {
	remoteIP, _, _ := net.SplitHostPort(r.RemoteAddr)

	if isTrustedProxy(remoteIP, trustedProxies) {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			parts := splitFirst(xff, ",")
			return trimSpace(parts)
		}
		if xri := r.Header.Get("X-Real-IP"); xri != "" {
			return trimSpace(xri)
		}
	}

	return remoteIP
}

// isTrustedProxy checks whether the given IP belongs to a trusted proxy network.
func isTrustedProxy(ip string, trustedProxies []*net.IPNet) bool {
	if len(trustedProxies) == 0 {
		return false
	}
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return false
	}
	for _, network := range trustedProxies {
		if network.Contains(parsed) {
			return true
		}
	}
	return false
}

func splitFirst(s, sep string) string {
	for i := 0; i < len(s); i++ {
		if s[i] == sep[0] {
			return s[:i]
		}
	}
	return s
}

func trimSpace(s string) string {
	start := 0
	for start < len(s) && s[start] == ' ' {
		start++
	}
	end := len(s)
	for end > start && s[end-1] == ' ' {
		end--
	}
	return s[start:end]
}
