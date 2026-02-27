package db

import (
	"context"
	"fmt"

	"github.com/devrimsoft/bug-notifications-api/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// InsertReport inserts a bug report into the database.
// Uses ON CONFLICT to ensure idempotency via event_id.
func (r *Repository) InsertReport(ctx context.Context, msg *model.QueueMessage) error {
	query := `
		INSERT INTO bug_reports (id, site_id, title, description, category, page_url, contact_type, contact_value, first_name, last_name, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, 'new', $11::timestamptz)
		ON CONFLICT (id) DO NOTHING
	`
	_, err := r.pool.Exec(ctx, query,
		msg.EventID,
		msg.SiteID,
		msg.Title,
		msg.Description,
		string(msg.Category),
		msg.PageURL,
		msg.ContactType,
		msg.ContactValue,
		msg.FirstName,
		msg.LastName,
		msg.ReceivedAt,
	)
	if err != nil {
		return fmt.Errorf("insert report: %w", err)
	}
	return nil
}

// GetReport retrieves a single bug report by ID.
func (r *Repository) GetReport(ctx context.Context, id string) (*model.BugReport, error) {
	query := `
		SELECT id, site_id, title, description, category, page_url, contact_type, contact_value, first_name, last_name, status, created_at
		FROM bug_reports WHERE id = $1
	`
	row := r.pool.QueryRow(ctx, query, id)

	var report model.BugReport
	err := row.Scan(
		&report.ID, &report.SiteID, &report.Title, &report.Description,
		&report.Category, &report.PageURL, &report.ContactType, &report.ContactValue,
		&report.FirstName, &report.LastName, &report.Status, &report.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get report: %w", err)
	}
	return &report, nil
}
