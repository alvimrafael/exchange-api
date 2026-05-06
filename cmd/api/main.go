package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/alvimrafael/exchange-api/internal/cache"
	"github.com/alvimrafael/exchange-api/internal/handler"
	"github.com/alvimrafael/exchange-api/internal/middleware"
	"github.com/alvimrafael/exchange-api/internal/provider"
	"github.com/alvimrafael/exchange-api/internal/repository"
	"github.com/alvimrafael/exchange-api/internal/service"
	"github.com/alvimrafael/exchange-api/internal/webhook"
	"github.com/alvimrafael/exchange-api/pkg/config"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("aviso: .env não encontrado, usando variáveis do sistema")
	}

	cfg := config.Load()

	if cfg.ExchangeAPIKey == "" {
		log.Fatal("EXCHANGE_API_KEY não configurada")
	}

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatal("postgres: erro ao inicializar cliente: ", err)
	}
	defer db.Close()

	if err := db.PingContext(context.Background()); err != nil {
		log.Fatal("postgres: falha na conexão: ", err)
	}
	log.Println("✓ postgres conectado")

	rateRepo := repository.NewRateRepository(db)
	webhookRepo := repository.NewWebhookRepository(db)

	opts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Fatal("redis: URL inválida: ", err)
	}
	rdb := redis.NewClient(opts)
	defer rdb.Close()

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatal("redis: falha na conexão: ", err)
	}
	log.Println("✓ redis conectado")

	rateCache := cache.NewRedisCache(rdb)
	ttl := time.Duration(cfg.CacheTTL) * time.Second

	exchangeProvider := provider.NewExchangeRateAPI(cfg.ExchangeAPIKey)
	rateSvc := service.NewRateService(exchangeProvider, rateCache, rateRepo, ttl)
	rateHandler := handler.NewRateHandler(rateSvc)

	webhookInterval := time.Duration(cfg.WebhookIntervalSecs) * time.Second
	webhookWorker := webhook.NewWorker(webhookRepo, rateRepo, webhookInterval)
	go webhookWorker.Start(context.Background())

	webhookHandler := handler.NewWebhookHandler(webhookRepo)
	rateLimiter := middleware.NewIPRateLimiter(cfg.RateLimitRPS, cfg.RateLimitBurst)

	r := gin.Default()
	r.Use(middleware.RateLimit(rateLimiter))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":   "ok",
			"postgres": "up",
			"redis":    "up",
		})
	})

	r.GET("/rates", rateHandler.GetRate)
	r.GET("/rates/history", rateHandler.GetHistory)

	r.POST("/webhooks", webhookHandler.Create)
	r.GET("/webhooks", webhookHandler.List)
	r.DELETE("/webhooks/:id", webhookHandler.Delete)

	log.Println("servidor na porta", cfg.Port)
	r.Run(":" + cfg.Port)
}
