.PHONY: all tidy fmt run test

all: tidy fmt lint test run-it

tidy:
	go mod tidy

fmt:
	gofmt -s -w .

run:
	go run -race ./cmd/app/main.go --config ./config/config.local.yaml

test: tidy fmt lint
	go test -race $(shell go list ./... | grep -v '/tests') -v

start-dep:
	mkdir -p ./postgres_data
	docker compose up -d postgres
	docker compose up -d jaeger

stop-dep:
	docker compose down postgres

docker-build:
	DOCKER_BUILDKIT=1 docker buildx build -f ./docker/Dockerfile -t kart-challenge .

docker-start:
	ENVIRONMENT=docker docker compose up -d postgres
	ENVIRONMENT=docker docker compose up -d --build kart-challenge

docker-stop:
	docker compose stop kart-challenge postgres

lint:
	golangci-lint run ./...

run-it:
	docker compose down
	ENVIRONMENT=test docker compose up -d jaeger
	ENVIRONMENT=test docker compose up -d postgres
	until docker exec postgres pg_isready -U user; do sleep 1; done
	ENVIRONMENT=test docker compose up -d --build kart-challenge
	# Wait for kart-challenge to be healthy
	until [ "$$(docker inspect --format='{{json .State.Health.Status}}' kart-challenge)" = "\"healthy\"" ]; do \
		echo "Waiting for kart-challenge to be healthy..."; \
		sleep 2; \
	done
	go test -v ./tests/e2e -args -config=../../config/config.test.yaml

help:
	@echo "Available commands:"
	@echo "  make all           - Run tidy, fmt, lint, unit tests, and end-to-end tests"
	@echo "  make tidy          - Tidy go modules (update go.mod/go.sum)"
	@echo "  make fmt           - Format Go code using gofmt"
	@echo "  make lint          - Lint Go code with golangci-lint"
	@echo "  make run           - Run the application locally with race detector"
	@echo "  make test          - Run unit tests (excluding /tests)"
	@echo "  make start-dep     - Start PostgreSQL dependency (with local volume)"
	@echo "  make stop-dep      - Stop PostgreSQL dependency"
	@echo "  make docker-build  - Build Docker image with BuildKit"
	@echo "  make docker-start  - Start PostgreSQL and kart-challenge in Docker"
	@echo "  make docker-stop   - Stop PostgreSQL and kart-challenge containers"
	@echo "  make run-it        - Run end-to-end tests with test Docker setup"
	@echo "  make help          - Show this help message"
