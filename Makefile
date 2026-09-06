.PHONY: up down build logs restart clean dev dev-backend db-shell lint lint-fix test test-backend test-frontend ci coverage tools test-tools

# ======== Docker ========

# Start all services
up:
	docker-compose up -d

# Stop all services
down:
	docker-compose down

# Rebuild and restart
build:
	docker-compose build --no-cache
	docker-compose up -d

# View logs
logs:
	docker-compose logs -f

# Restart services
restart:
	docker-compose restart

# Clean volumes (reset DB)
clean:
	docker-compose down -v

# ======== Development ========

# Frontend dev server
dev:
	cd frontend && npm run dev

# Local backend run
dev-backend:
	cd backend && go run ./cmd/server

# Database shell
db-shell:
	docker-compose exec db psql -U reqmango -d reqmango

# ======== Lint ========

lint:
	cd backend && golangci-lint run ./...
	cd frontend && npx eslint src/ --ext .ts,.vue

lint-fix:
	cd backend && golangci-lint run --fix ./...
	cd frontend && npx eslint src/ --ext .ts,.vue --fix

# ======== Test ========

test-backend:
	cd backend && go test -race -coverprofile=coverage.out ./internal/...

test-frontend:
	cd frontend && npx vitest run

test: test-backend test-frontend

coverage:
	cd backend && go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: backend/coverage.html"

# ======== Tools (MCP + CLI) ========

tools:
	cd sdk && go build -o ../bin/reqmango ./cmd/reqmango
	cd sdk && go build -o ../bin/reqmango-mcp ./cmd/reqmango-mcp

test-tools:
	cd sdk && go test ./...

# ======== CI ========

ci: lint test
	cd backend && go build ./cmd/server
	cd frontend && npx vite build
	@echo "CI: all checks passed"
