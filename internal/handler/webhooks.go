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

type createWebhookRequest struct {
	URL       string  `json:"url"       binding:"required,url"`
	From      string  `json:"from"      binding:"required"`
	To        string  `json:"to"        binding:"required"`
	Threshold float64 `json:"threshold" binding:"required,gt=0"`
	Direction string  `json:"direction" binding:"required,oneof=above below"`
}

func (h *WebhookHandler) Create(c *gin.Context) {
	var req createWebhookRequest
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
