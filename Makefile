.PHONY: up down build logs restart clean dev

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

# Development mode (frontend only, assumes backend and DB are running locally)
dev:
	cd frontend && npm run dev

# Local backend run
dev-backend:
	cd backend && go run ./cmd/server

# Database shell
db-shell:
	docker-compose exec db psql -U reqmango -d reqmango
