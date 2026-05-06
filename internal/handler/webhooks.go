package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/alvimrafael/exchange-api/internal/repository"
	"github.com/gin-gonic/gin"
)

type WebhookHandler struct {
	repo *repository.WebhookRepository
}

func NewWebhookHandler(repo *repository.WebhookRepository) *WebhookHandler {
	return &WebhookHandler{repo: repo}
}

// CreateWebhookRequest is the body for registering a webhook.
type CreateWebhookRequest struct {
	URL       string  `json:"url"       binding:"required,url"          example:"https://example.com/notify"`
	From      string  `json:"from"      binding:"required"               example:"USD"`
	To        string  `json:"to"        binding:"required"               example:"BRL"`
	Threshold float64 `json:"threshold" binding:"required,gt=0"          example:"5.80"`
	Direction string  `json:"direction" binding:"required,oneof=above below" example:"above"`
}

// Create godoc
// @Summary     Register a webhook
// @Description Registers a URL to be notified when an exchange rate crosses a threshold
// @Tags        webhooks
// @Accept      json
// @Produce     json
// @Param       body  body      CreateWebhookRequest  true  "Webhook configuration"
// @Success     201   {object}  repository.Webhook
// @Failure     400   {object}  ErrorResponse
// @Failure     500   {object}  ErrorResponse
// @Router      /webhooks [post]
func (h *WebhookHandler) Create(c *gin.Context) {
	var req CreateWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	w, err := h.repo.Save(c.Request.Context(),
		req.URL,
		strings.ToUpper(req.From),
		strings.ToUpper(req.To),
		req.Direction,
		req.Threshold,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao registrar webhook"})
		return
	}

	c.JSON(http.StatusCreated, w)
}

// List godoc
// @Summary     List webhooks
// @Description Returns all registered webhooks
// @Tags        webhooks
// @Produce     json
// @Success     200  {array}   repository.Webhook
// @Failure     500  {object}  ErrorResponse
// @Router      /webhooks [get]
func (h *WebhookHandler) List(c *gin.Context) {
	hooks, err := h.repo.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao listar webhooks"})
		return
	}

	if hooks == nil {
		hooks = []repository.Webhook{}
	}
	c.JSON(http.StatusOK, hooks)
}

// Delete godoc
// @Summary     Delete a webhook
// @Description Removes a registered webhook by ID
// @Tags        webhooks
// @Param       id   path  int  true  "Webhook ID"
// @Success     204
// @Failure     400  {object}  ErrorResponse
// @Failure     500  {object}  ErrorResponse
// @Router      /webhooks/{id} [delete]
func (h *WebhookHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}

	if err := h.repo.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao deletar webhook"})
		return
	}

	c.Status(http.StatusNoContent)
}
