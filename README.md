# Waste Space API

REST API for waste management services built with Go, Gin, PostgreSQL, and Dragonfly.

## Quick Start

```bash
cp .env.example .env
docker-compose up -d
go run cmd/api/main.go
```

## Requirements

- Go 1.25+
- Docker & Docker Compose
- PostgreSQL 17
- Dragonfly (Redis-compatible cache)

## Development

```bash
# Install dependencies
go mod download

# Start infrastructure
docker-compose up -d

# Run migrations
goose -dir migrations postgres "postgresql://waste_space:waste_space@localhost:5432/waste_space?sslmode=disable" up

# Generate Swagger docs
swag init -g cmd/api/main.go -o docs

# Run application
go run cmd/api/main.go
```

## API Documentation

Swagger UI: http://localhost:8080/swagger/index.html

## Project Structure

```
cmd/api/              - Application entry point
internal/
  app/                - App initialization
  config/             - Configuration
  controller/v1/      - HTTP handlers
  service/            - Business logic
  storage/
    repository/       - Database operations
    cache/            - Redis cache
  model/              - Domain models
  dto/                - Request/response DTOs
  middleware/         - HTTP middleware
pkg/
  auth/               - JWT authentication
  db/                 - Database clients
  errors/             - Custom errors
migrations/           - Database migrations
```

## Environment Variables

See `.env.example` for all available configuration options.

## Stack

- **Framework**: Gin
- **ORM**: GORM
- **Database**: PostgreSQL
- **Cache**: Dragonfly
- **Migrations**: Goose
- **Docs**: Swagger
- **Auth**: JWT
