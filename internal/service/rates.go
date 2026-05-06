package service

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/alvimrafael/exchange-api/internal/cache"
	"github.com/alvimrafael/exchange-api/internal/provider"
	"github.com/alvimrafael/exchange-api/internal/repository"
)

type RateService struct {
	provider provider.ExchangeProvider
	cache    cache.CacheProvider
	repo     *repository.RateRepository
	ttl      time.Duration
}

func NewRateService(p provider.ExchangeProvider, c cache.CacheProvider, repo *repository.RateRepository, ttl time.Duration) *RateService {
	return &RateService{provider: p, cache: c, repo: repo, ttl: ttl}
}

type RateResult struct {
	From   string  `json:"from"`
	To     string  `json:"to"`
	Rate   float64 `json:"rate"`
	Cached bool    `json:"cached"`
}

func (s *RateService) GetRate(ctx context.Context, from, to string) (*RateResult, error) {
	from = strings.ToUpper(from)
	to = strings.ToUpper(to)

	// 1. tenta o cache primeiro
	key := cache.CacheKey(from, to)

	if val, err := s.cache.Get(ctx, key); err == nil && val != "" {
		rate, err := strconv.ParseFloat(val, 64)
		if err == nil {
			return &RateResult{From: from, To: to, Rate: rate, Cached: true}, nil
		}
	}

	// 2. cache miss - chama a API externa
	rate, err := s.provider.GetRate(ctx, from, to)
	if err != nil {
		return nil, err
	}

	// 3. salva no cache para as próximas requisições
	if err := s.cache.Set(ctx, key, fmt.Sprintf("%f", rate), s.ttl); err != nil {
		// não é fatal - loga mas continua
		_ = err
	}

	result := &RateResult{From: from, To: to, Rate: rate, Cached: false}
	go func() {
		if err := s.repo.Save(context.Background(), result.From, result.To, result.Rate, result.Cached); err != nil {
			log.Printf("repository: erro ao salvar cotação: %v", err)
		}
	}()
	return result, nil
}

func (s *RateService) GetHistory(ctx context.Context, from, to string, days int) ([]repository.RateRecord, error) {
	return s.repo.History(ctx, from, to, days)
}
