APP=metrics

DB_URI=postgres://user:pass@localhost:5432/metrics?sslmode=disable

BUILD_VERSION ?= 0.1.0
BUILD_DATE := $(shell date +%Y-%m-%dT%H:%M:%S)
BUILD_COMMIT := $(shell git rev-parse --short HEAD)

LDFLAGS := -X github.com/ilyinon/go-musthave-metrics/internal/buildinfo.Version=$(BUILD_VERSION) \
           -X github.com/ilyinon/go-musthave-metrics/internal/buildinfo.Date=$(BUILD_DATE) \
           -X github.com/ilyinon/go-musthave-metrics/internal/buildinfo.Commit=$(BUILD_COMMIT)

.PHONY: build run test up down migrate migrate-down

build:
	go build -ldflags "$(LDFLAGS)" -o bin/server ./cmd/server/main.go
	go build -ldflags "$(LDFLAGS)" -o bin/agent ./cmd/agent/main.go
	go build -o bin/staticlint ./cmd/staticlint

run_server:
	go run ./cmd/server/main.go

run_agent:
	go run ./cmd/agent/main.go

run_static:
	./bin/staticlint ./... 

test:
	go test ./... -cover

up:
	docker-compose up -d

down:
	docker-compose down

logs:
	docker-compose logs -f

db:
	docker exec -it metrics-db psql -U user -d metrics

migrate:
	goose -dir migrations postgres "$(DB_URI)" up

migrate-down:
	goose -dir migrations postgres "$(DB_URI)" down
