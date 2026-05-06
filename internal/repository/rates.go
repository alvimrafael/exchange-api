package repository

import (
	"context"
	"database/sql"
	"time"
)

type RateRecord struct {
	ID           int       `json:"id"`
	FromCurrency string    `json:"from"`
	ToCurrency   string    `json:"to"`
	Rate         float64   `json:"rate"`
	Cached       bool      `json:"cached"`
	QueriedAt    time.Time `json:"queried_at"`
}

type RateRepository struct {
	db *sql.DB
}

func NewRateRepository(db *sql.DB) *RateRepository {
	return &RateRepository{db: db}
}

func (r *RateRepository) Save(ctx context.Context, from, to string, rate float64, cached bool) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO rates (from_currency, to_currency, rate, cached) VALUES ($1, $2, $3, $4)`,
		from, to, rate, cached,
	)
	return err
}

func (r *RateRepository) History(ctx context.Context, from, to string, days int) ([]RateRecord, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, from_currency, to_currency, rate, cached, queried_at
		 FROM rates
		 WHERE from_currency = $1 AND to_currency = $2
		   AND queried_at >= NOW() - ($3 || ' days')::INTERVAL
		 ORDER BY queried_at DESC`,
		from, to, days,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []RateRecord
	for rows.Next() {
		var rec RateRecord
		if err := rows.Scan(&rec.ID, &rec.FromCurrency, &rec.ToCurrency, &rec.Rate, &rec.Cached, &rec.QueriedAt); err != nil {
			return nil, err
		}
		records = append(records, rec)
	}
	return records, rows.Err()
}

func (r *RateRepository) GetLatest(ctx context.Context, from, to string) (*RateRecord, error) {
	var rec RateRecord
	err := r.db.QueryRowContext(ctx,
		`SELECT id, from_currency, to_currency, rate, cached, queried_at
		 FROM rates
		 WHERE from_currency = $1 AND to_currency = $2
		 ORDER BY queried_at DESC LIMIT 1`,
		from, to,
	).Scan(&rec.ID, &rec.FromCurrency, &rec.ToCurrency, &rec.Rate, &rec.Cached, &rec.QueriedAt)
	if err != nil {
		return nil, err
	}
	return &rec, nil
}
