APP=metrics

DB_URI=postgres://user:pass@localhost:5432/metrics?sslmode=disable

.PHONY: build run test up down migrate migrate-down

build:
	go build -o bin/server ./cmd/server/main.go
	go build -o bin/agent ./cmd/agent/main.go
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
