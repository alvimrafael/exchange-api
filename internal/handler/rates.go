package handler

import (
	"errors"
	"net/http"

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
