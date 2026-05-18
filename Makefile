.PHONY: env-check run swagger build test tidy deps docker-up docker-down docker-logs
env-check:
	@test -f .env || (echo "ERRO: .env não encontrado. Copie .env.example" && exit 1)

run: env-check
	go run cmd/api/main.go

swagger:
	~/go/bin/swag init -g cmd/api/main.go -o docs

build: swagger
	go build -o bin/api cmd/api/main.go

test:
	go test ./...

tidy:
	go mod tidy

deps:
	go get github.com/lib/pq
	go get github.com/redis/go-redis/v9
	go get github.com/gin-gonic/gin
	go get github.com/joho/godotenv
	go mod tidy

docker-up:
	docker compose --env-file .env.docker up --build

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f api