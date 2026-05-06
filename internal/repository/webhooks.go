package repository

import (
	"context"
	"database/sql"
	"time"
)

type Webhook struct {
	ID              int        `json:"id"`
	URL             string     `json:"url"`
	FromCurrency    string     `json:"from"`
	ToCurrency      string     `json:"to"`
	Threshold       float64    `json:"threshold"`
	Direction       string     `json:"direction"`
	LastTriggeredAt *time.Time `json:"last_triggered_at"`
	CreatedAt       time.Time  `json:"created_at"`
}

type WebhookRepository struct {
	db *sql.DB
}

func NewWebhookRepository(db *sql.DB) *WebhookRepository {
	return &WebhookRepository{db: db}
}

func (r *WebhookRepository) Save(ctx context.Context, url, from, to, direction string, threshold float64) (*Webhook, error) {
	var w Webhook
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO webhooks (url, from_currency, to_currency, threshold, direction)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, url, from_currency, to_currency, threshold, direction, last_triggered_at, created_at`,
		url, from, to, threshold, direction,
	).Scan(&w.ID, &w.URL, &w.FromCurrency, &w.ToCurrency, &w.Threshold, &w.Direction, &w.LastTriggeredAt, &w.CreatedAt)
	return &w, err
}

func (r *WebhookRepository) List(ctx context.Context) ([]Webhook, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, url, from_currency, to_currency, threshold, direction, last_triggered_at, created_at
		 FROM webhooks ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hooks []Webhook
	for rows.Next() {
		var w Webhook
		if err := rows.Scan(&w.ID, &w.URL, &w.FromCurrency, &w.ToCurrency, &w.Threshold, &w.Direction, &w.LastTriggeredAt, &w.CreatedAt); err != nil {
			return nil, err
		}
		hooks = append(hooks, w)
	}
	return hooks, rows.Err()
}

func (r *WebhookRepository) Delete(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM webhooks WHERE id = $1`, id)
	return err
}

func (r *WebhookRepository) UpdateLastTriggered(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, `UPDATE webhooks SET last_triggered_at = NOW() WHERE id = $1`, id)
	return err
}
