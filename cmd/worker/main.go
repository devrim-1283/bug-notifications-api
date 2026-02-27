package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/devrimsoft/bug-notifications-api/internal/config"
	"github.com/devrimsoft/bug-notifications-api/internal/db"
	"github.com/devrimsoft/bug-notifications-api/internal/queue"
	"github.com/devrimsoft/bug-notifications-api/internal/worker"
	"github.com/redis/go-redis/v9"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

	cfg, err := config.Load()
	if err != nil {
		slog.Error("config load failed", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Redis
	opts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		slog.Error("invalid redis url", "error", err)
		os.Exit(1)
	}
	rdb := redis.NewClient(opts)
	defer rdb.Close()

	pingCtx, pingCancel := context.WithTimeout(ctx, 5*time.Second)
	if err := rdb.Ping(pingCtx).Err(); err != nil {
		pingCancel()
		slog.Error("redis ping failed", "error", err)
		os.Exit(1)
	}
	pingCancel()

	// PostgreSQL
	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Run migrations
	if err := db.Migrate(ctx, pool); err != nil {
		slog.Error("migration failed", "error", err)
		os.Exit(1)
	}

	repo := db.NewRepository(pool)
	consumer := queue.NewConsumer(rdb)

	// Start workers
	var wg sync.WaitGroup
	slog.Info("starting workers", "concurrency", cfg.WorkerConcurrency)

	for i := 0; i < cfg.WorkerConcurrency; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			slog.Info("worker started", "worker_id", id)
			w := worker.New(consumer, repo)
			w.Run(ctx)
			slog.Info("worker stopped", "worker_id", id)
		}(i)
	}

	// Wait for shutdown signal
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)
	<-done

	slog.Info("shutdown signal received, stopping workers...")
	cancel()
	wg.Wait()
	slog.Info("all workers stopped")
}
