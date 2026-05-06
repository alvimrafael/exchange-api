package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

var ErrCurrencyNotFound = errors.New("currency not found")

type ExchangeProvider interface {
	GetRate(ctx context.Context, from, to string) (float64, error)
}

type ExchangeRateAPI struct {
	apiKey string
	client *http.Client
}

func NewExchangeRateAPI(apiKey string) *ExchangeRateAPI {
	return &ExchangeRateAPI{
		apiKey: apiKey,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

type exchangeRateResponse struct {
	Result          string             `json:"result"`
	ErrorType       string             `json:"error-type"`
	ConversionRates map[string]float64 `json:"conversion_rates"`
}

func (e *ExchangeRateAPI) GetRate(ctx context.Context, from, to string) (float64, error) {
	url := fmt.Sprintf("https://v6.exchangerate-api.com/v6/%s/latest/%s", e.apiKey, strings.ToUpper(from))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}

	resp, err := e.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var result exchangeRateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}

	if result.Result != "success" {
		if result.ErrorType == "unsupported-code" {
			return 0, fmt.Errorf("%w: %s", ErrCurrencyNotFound, from)
		}
		return 0, fmt.Errorf("provider error: %s", result.ErrorType)
	}

	rate, ok := result.ConversionRates[strings.ToUpper(to)]
	if !ok {
		return 0, fmt.Errorf("%w: %s", ErrCurrencyNotFound, to)
	}

	return rate, nil
}
