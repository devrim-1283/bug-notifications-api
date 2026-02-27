package worker

import (
	"context"
	"log/slog"

	"github.com/devrimsoft/bug-notifications-api/internal/db"
	"github.com/devrimsoft/bug-notifications-api/internal/queue"
)

type Worker struct {
	consumer *queue.Consumer
	repo     *db.Repository
}

func New(consumer *queue.Consumer, repo *db.Repository) *Worker {
	return &Worker{
		consumer: consumer,
		repo:     repo,
	}
}

// Run starts the worker loop. Blocks until context is cancelled.
func (w *Worker) Run(ctx context.Context) {
	slog.Info("worker started")

	for {
		select {
		case <-ctx.Done():
			slog.Info("worker stopping")
			return
		default:
		}

		msg, err := w.consumer.Dequeue(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return // context cancelled
			}
			slog.Error("dequeue failed", "error", err)
			continue
		}
		if msg == nil {
			continue // timeout, no message
		}

		slog.Info("processing report", "event_id", msg.EventID, "site_id", msg.SiteID, "retry", msg.RetryCount)

		if err := w.repo.InsertReport(ctx, msg); err != nil {
			slog.Error("insert failed, requeuing", "event_id", msg.EventID, "error", err, "retry", msg.RetryCount)
			if reqErr := w.consumer.Requeue(ctx, msg); reqErr != nil {
				slog.Error("requeue failed", "event_id", msg.EventID, "error", reqErr)
			}
			continue
		}

		slog.Info("report saved", "event_id", msg.EventID)
	}
}
