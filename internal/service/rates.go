package service

import (
	"context"
	"strings"

	"github.com/alvimrafael/exchange-api/internal/provider"
)

type RateService struct {
	provider provider.ExchangeProvider
}

func NewRateService(p provider.ExchangeProvider) *RateService {
	return &RateService{provider: p}
}

type RateResult struct {
	From string  `json:"from"`
	To   string  `json:"to"`
	Rate float64 `json:"rate"`
}

func (s *RateService) GetRate(ctx context.Context, from, to string) (*RateResult, error) {
	from = strings.ToUpper(from)
	to = strings.ToUpper(to)

	rate, err := s.provider.GetRate(ctx, from, to)
	if err != nil {
		return nil, err
	}

	return &RateResult{From: from, To: to, Rate: rate}, nil
}
