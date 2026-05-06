package webhook

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/alvimrafael/exchange-api/internal/repository"
)

type Worker struct {
	webhookRepo *repository.WebhookRepository
	rateRepo    *repository.RateRepository
	interval    time.Duration
	client      *http.Client
}

func NewWorker(webhookRepo *repository.WebhookRepository, rateRepo *repository.RateRepository, interval time.Duration) *Worker {
	return &Worker{
		webhookRepo: webhookRepo,
		rateRepo:    rateRepo,
		interval:    interval,
		client:      &http.Client{Timeout: 10 * time.Second},
	}
}

type payload struct {
	WebhookID   int       `json:"webhook_id"`
	From        string    `json:"from"`
	To          string    `json:"to"`
	Rate        float64   `json:"rate"`
	Threshold   float64   `json:"threshold"`
	Direction   string    `json:"direction"`
	TriggeredAt time.Time `json:"triggered_at"`
}

func (w *Worker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	log.Printf("webhook worker iniciado (intervalo: %s)", w.interval)
	for {
		select {
		case <-ticker.C:
			w.check(ctx)
		case <-ctx.Done():
			log.Println("webhook worker encerrado")
			return
		}
	}
}

func (w *Worker) check(ctx context.Context) {
	hooks, err := w.webhookRepo.List(ctx)
	if err != nil {
		log.Printf("webhook worker: erro ao carregar webhooks: %v", err)
		return
	}

	for _, hook := range hooks {
		rec, err := w.rateRepo.GetLatest(ctx, hook.FromCurrency, hook.ToCurrency)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				continue // ainda não há cotação registrada para esse par
			}
			log.Printf("webhook worker: erro ao buscar cotação %s/%s: %v", hook.FromCurrency, hook.ToCurrency, err)
			continue
		}

		triggered := (hook.Direction == "above" && rec.Rate > hook.Threshold) ||
			(hook.Direction == "below" && rec.Rate < hook.Threshold)

		if !triggered {
			continue
		}

		if err := w.fire(hook, rec.Rate); err != nil {
			log.Printf("webhook worker: erro ao disparar webhook %d: %v", hook.ID, err)
			continue
		}

		if err := w.webhookRepo.UpdateLastTriggered(ctx, hook.ID); err != nil {
			log.Printf("webhook worker: erro ao atualizar last_triggered_at %d: %v", hook.ID, err)
		}
	}
}

func (w *Worker) fire(hook repository.Webhook, rate float64) error {
	p := payload{
		WebhookID:   hook.ID,
		From:        hook.FromCurrency,
		To:          hook.ToCurrency,
		Rate:        rate,
		Threshold:   hook.Threshold,
		Direction:   hook.Direction,
		TriggeredAt: time.Now().UTC(),
	}

	body, err := json.Marshal(p)
	if err != nil {
		return err
	}

	resp, err := w.client.Post(hook.URL, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	log.Printf("webhook worker: disparado webhook %d → %s (status %d)", hook.ID, hook.URL, resp.StatusCode)
	return nil
}
