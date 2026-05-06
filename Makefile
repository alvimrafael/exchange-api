.PHONY: env-check
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