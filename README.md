# Go API Server

A RESTful and gRPC API server built with Go, Gin, GORM, Kafka, and PostgreSQL.

## Project Structure

```
cmd/
  api/          HTTP + gRPC server entry point
  migrate/      Database migration CLI tool
internal/
  adapter/      Infrastructure adapters (exchange rate client)
  config/       Environment-based configuration
  consumer/     Kafka consumer lifecycle
  database/     Transaction manager
  domain/       Shared error types
  grpc/         gRPC handler implementations
  handler/      HTTP handler implementations
  middleware/   Auth and role middleware
  model/        Domain models and request/response DTOs
  repository/   Database access layer (GORM)
  server/       Server wiring and dependency injection
  service/      Business logic layer
pkg/
  kafka/        Kafka producer and consumer helpers
  jwt/          JWT token generation and validation
  password/     Bcrypt password hashing
proto/          Protobuf definitions and generated code
migrations/     SQL migration files (golang-migrate)
```

## Prerequisites

- Go 1.24+
- PostgreSQL 16+
- Docker & Docker Compose (for containerized setup)

## Quick Start

### Docker (recommended)

```bash
# Start all services (PostgreSQL, Kafka, Zookeeper, app)
make docker-up

# Run database migrations
docker compose run --rm migrate up
```

The API server starts on `:8080` (HTTP) and `:9090` (gRPC).

### Kubernetes (local)

Requires a local cluster (Docker Desktop, minikube, etc.) with `kubectl` configured.

```bash
# 1. Build the image into the local Docker daemon
make k8s-build

# 2. Deploy all services, migrations, and the app
make k8s-up

# 3. Wait for pods to be ready
kubectl get pods -n golang-api-server -w
```

The API is exposed via NodePort:
- HTTP → `http://localhost:30080`
- gRPC → `localhost:30090`

Before deploying to a real cluster, update `k8s/secret.yaml` with a strong `JWT_SECRET` and push the image to a registry (update `image:` in `k8s/app.yaml` and `k8s/migrate.yaml` accordingly).

To tear everything down:
```bash
make k8s-down
```

### Local Development

1. Start PostgreSQL and Kafka locally
2. Copy environment config:

   ```bash
   cp .env.example .env
   ```

3. Run migrations:

   ```bash
   make migrate-up
   ```

4. Start the server:

   ```bash
   make run
   ```

## Environment Variables

Copy `.env.example` to `.env` and adjust values for your environment:

```bash
cp .env.example .env
```

### Application

| Variable | Default | Description |
| -------- | ------- | ----------- |
| `APP_ENV` | `development` | Runtime environment (`development`, `test`, `production`) |
| `SERVER_ADDRESS` | `:8080` | HTTP server listen address |
| `SERVER_TIMEOUT` | `30` | HTTP request timeout in seconds |
| `ALLOWED_ORIGINS` | `*` | CORS allowed origins (comma-separated) |

### Database

| Variable | Default | Description |
| -------- | ------- | ----------- |
| `DATABASE_URL` | `postgres://postgres:postgres@localhost:5432/users_db?sslmode=disable` | PostgreSQL connection string |
| `DB_MAX_CONNS` | `25` | Maximum number of open DB connections |
| `DB_MIN_CONNS` | `5` | Minimum number of idle DB connections |

### JWT

| Variable | Default | Description |
| -------- | ------- | ----------- |
| `JWT_SECRET` | `change-me-in-production` | Secret key for signing tokens — **must be overridden in production** |
| `JWT_ACCESS_EXPIRY` | `15` | Access token expiry in minutes |
| `JWT_REFRESH_EXPIRY` | `7` | Refresh token expiry in days |

### gRPC

| Variable | Default | Description |
| -------- | ------- | ----------- |
| `GRPC_ADDRESS` | `:9090` | gRPC server listen address |

### Kafka

| Variable | Default | Description |
| -------- | ------- | ----------- |
| `KAFKA_BROKERS` | `localhost:9092` | Comma-separated list of Kafka broker addresses |
| `KAFKA_TOPIC` | `conversions` | Topic for exchange rate conversion events |
| `KAFKA_GROUP_ID` | `api-server` | Consumer group ID |

### Redis

| Variable | Default | Description |
| -------- | ------- | ----------- |
| `REDIS_ADDR` | `localhost:6379` | Redis server address |
| `REDIS_PASSWORD` | _(empty)_ | Redis password (leave empty if not set) |
| `REDIS_DB` | `0` | Redis database index |
| `CACHE_TTL` | `600` | Cache entry TTL in seconds |

### Logging

| Variable | Default | Description |
| -------- | ------- | ----------- |
| `LOG_LEVEL` | `info` | Log level (`debug`, `info`, `warn`, `error`) |
| `LOG_FORMAT` | `json` | Log format (`json` or `text`) |
| `LOG_KAFKA_TOPIC` | `app-logs` | Kafka topic for structured log streaming |
| `LOG_KAFKA_GROUP_ID` | `log-consumer` | Consumer group for log streaming |

### Rate Limiting

| Variable | Default | Description |
| -------- | ------- | ----------- |
| `AUTH_RATE_LIMIT` | `10` | Max requests per minute on auth endpoints |
| `API_RATE_LIMIT` | `60` | Max requests per minute on all other endpoints |

> **Production note:** `JWT_SECRET` must be set to a strong random value in any non-development environment. The server will refuse to start if it detects the default value outside of `development`/`test`.

## Database Migrations

Migrations are managed with [golang-migrate](https://github.com/golang-migrate/migrate) and run as a separate step from the application.

```bash
make migrate-up        # Apply all pending migrations
make migrate-down      # Rollback the last migration
make migrate-version   # Show current migration version
```

SQL files are in `migrations/`. Follow the naming convention:

```
migrations/
  001_create_users.up.sql
  001_create_users.down.sql
  002_add_orders.up.sql
  002_add_orders.down.sql
```

**Important:** The server does not auto-migrate on startup. Always run `make migrate-up` before starting the application.

## Make Commands

| Command                | Description                    |
| ---------------------- | ------------------------------ |
| `make build`           | Compile to `bin/server`        |
| `make run`             | Build and run the server       |
| `make test`            | Run tests with race detection  |
| `make test-cover`      | Run tests with coverage report |
| `make lint`            | Run golangci-lint              |
| `make migrate-up`      | Apply pending migrations       |
| `make migrate-down`    | Rollback last migration        |
| `make migrate-version` | Show current version           |
| `make docker-up`       | Start all services with Docker |
| `make docker-down`     | Stop and remove containers     |
| `make proto`           | Regenerate protobuf code       |

## API Endpoints

### Auth

| Method | Path                    | Description          |
| ------ | ----------------------- | -------------------- |
| POST   | `/api/v1/auth/register` | Register a new user  |
| POST   | `/api/v1/auth/login`    | Login and get tokens |
| POST   | `/api/v1/auth/refresh`  | Refresh access token |
| POST   | `/api/v1/auth/logout`   | Logout (protected)   |

### Users (admin)

| Method | Path                      | Description                          |
| ------ | ------------------------- | ------------------------------------ |
| GET    | `/api/v1/me`              | Get current user profile (protected) |
| GET    | `/api/v1/admin/users`     | List users (admin)                   |
| DELETE | `/api/v1/admin/users/:id` | Delete user (admin)                  |

### Exchange

| Method | Path                       | Description                  |
| ------ | -------------------------- | ---------------------------- |
| POST   | `/api/v1/exchange/convert` | Convert currency (protected) |

### Events

| Method | Path                     | Description               |
| ------ | ------------------------ | ------------------------- |
| POST   | `/api/v1/events/publish` | Publish event (protected) |

## Architecture

The project follows clean architecture principles:

- **service** — Business logic. Defines interfaces for what it needs (`UserRepository`). No framework dependencies.
- **repository** — Data access. Implements service interfaces using GORM.
- **handler** — HTTP adapters. Calls service layer only.
- **grpc** — gRPC adapters. Calls service layer only.
- **adapter** — External service clients (exchange rate API).
- **consumer** — Kafka consumer lifecycle.
- **database** — Transaction management via context propagation.
- **domain** — Shared error types used across layers.
- **server** — Composition root. Wires all dependencies together.

Dependency direction flows inward: `handler → service → repository → database`.
