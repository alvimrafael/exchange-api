package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"

	"github.com/alvimrafael/exchange-api/pkg/config"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

func main() {
	// 1. Carrega .env PRIMEIRO, antes de qualquer config.Load()
	if err := godotenv.Load(); err != nil {
		log.Println("aviso: .env não encontrado, usando variáveis do sistema")
		// não é fatal — em produção as envs vêm do sistema, não do arquivo
	}

	// 2. Lê configuração (agora as envs já estão carregadas)
	cfg := config.Load()

	// 3. Conecta no PostgreSQL
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatal("postgres: erro ao inicializar cliente: ", err)
	}
	defer db.Close()

	if err := db.PingContext(context.Background()); err != nil {
		log.Fatal("postgres: falha na conexão: ", err)
	}
	log.Println("✓ postgres conectado")

	// 4. Conecta no Redis
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

	// 5. Monta o servidor — SEMPRE por último
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":   "ok",
			"postgres": "up",
			"redis":    "up",
		})
	})

	log.Println("servidor na porta", cfg.Port)
	r.Run(":" + cfg.Port)
}
