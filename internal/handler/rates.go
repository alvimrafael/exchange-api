package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/alvimrafael/exchange-api/internal/provider"
	"github.com/alvimrafael/exchange-api/internal/service"
	"github.com/gin-gonic/gin"
)

type RateHandler struct {
	svc *service.RateService
}

func NewRateHandler(svc *service.RateService) *RateHandler {
	return &RateHandler{svc: svc}
}

// GetRate godoc
// @Summary     Get exchange rate
// @Description Returns the current exchange rate between two currencies. Served from Redis cache when available.
// @Tags        rates
// @Produce     json
// @Param       from  query     string  true  "Source currency code (e.g. USD)"
// @Param       to    query     string  true  "Target currency code (e.g. BRL)"
// @Success     200   {object}  service.RateResult
// @Failure     400   {object}  ErrorResponse
// @Failure     500   {object}  ErrorResponse
// @Router      /rates [get]
func (h *RateHandler) GetRate(c *gin.Context) {
	from := c.Query("from")
	to := c.Query("to")

	if from == "" || to == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query params 'from' and 'to' are required"})
		return
	}

	result, err := h.svc.GetRate(c.Request.Context(), from, to)
	if err != nil {
		if errors.Is(err, provider.ErrCurrencyNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch exchange rate"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetHistory godoc
// @Summary     Get rate history
// @Description Returns historical exchange rate records for a currency pair stored in the database
// @Tags        rates
// @Produce     json
// @Param       from  query     string  true   "Source currency code (e.g. USD)"
// @Param       to    query     string  true   "Target currency code (e.g. BRL)"
// @Param       days  query     int     false  "Number of past days to include (default: 7)"
// @Success     200   {array}   repository.RateRecord
// @Failure     500   {object}  ErrorResponse
// @Router      /rates/history [get]
func (h *RateHandler) GetHistory(c *gin.Context) {
	from := strings.ToUpper(c.Query("from"))
	to := strings.ToUpper(c.Query("to"))
	days := 7

	if d := c.Query("days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 {
			days = parsed
		}
	}

	records, err := h.svc.GetHistory(c.Request.Context(), from, to, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao buscar histórico"})
		return
	}

	c.JSON(http.StatusOK, records)
}
