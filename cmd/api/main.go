package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/devrimsoft/bug-notifications-api/internal/api"
	"github.com/devrimsoft/bug-notifications-api/internal/config"
	"github.com/devrimsoft/bug-notifications-api/internal/middleware"
	"github.com/devrimsoft/bug-notifications-api/internal/queue"
	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

	cfg, err := config.Load()
	if err != nil {
		slog.Error("config load failed", "error", err)
		os.Exit(1)
	}

	// Redis
	opts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		slog.Error("invalid redis url", "error", err)
		os.Exit(1)
	}
	rdb := redis.NewClient(opts)
	defer rdb.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	if err := rdb.Ping(ctx).Err(); err != nil {
		cancel()
		slog.Error("redis ping failed", "error", err)
		os.Exit(1)
	}
	cancel()

	producer := queue.NewProducer(rdb)
	handler := api.NewHandler(producer)

	// Router
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.SecureHeaders())
	r.Use(middleware.RequireHTTPS())
	r.Use(middleware.CORSMiddleware(cfg))
	r.Use(middleware.BodyLimit(256 * 1024)) // 256KB
	r.Use(middleware.RateLimit(rdb, cfg.RateLimitRPS, cfg.TrustedProxies))

	// Health check (no auth required)
	r.Get("/health", handler.HealthCheck)

	// Protected routes
	r.Route("/v1", func(r chi.Router) {
		r.Use(middleware.BrowserOnly())
		r.Use(middleware.APIKeyAuth(cfg))
		r.Post("/reports", handler.CreateReport)
	})

	addr := fmt.Sprintf(":%d", cfg.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	go func() {
		if cfg.TLSEnabled() {
			slog.Info("api server starting with TLS", "addr", addr, "sites", len(cfg.Sites))
			if err := srv.ListenAndServeTLS(cfg.TLSCertFile, cfg.TLSKeyFile); err != nil && err != http.ErrServerClosed {
				slog.Error("server error", "error", err)
				os.Exit(1)
			}
		} else {
			slog.Info("api server starting (TLS via reverse proxy expected)", "addr", addr, "sites", len(cfg.Sites))
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				slog.Error("server error", "error", err)
				os.Exit(1)
			}
		}
	}()

	<-done
	slog.Info("shutting down...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown error", "error", err)
	}

	slog.Info("server stopped")
}
