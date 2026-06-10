# golang-api-server — Project Instructions

## Architecture

Dependency direction flows strictly inward:

```
handler / grpc  →  service  →  repository  →  database (GORM)
                     ↑
              adapter (external HTTP clients)
```

| Package | Role |
|---------|------|
| `internal/handler/` | Gin HTTP handlers — bind request, call service, return JSON |
| `internal/grpc/` | gRPC handlers — implement protobuf server interfaces |
| `internal/service/` | Business logic. Defines its own dependency interfaces here |
| `internal/repository/` | GORM implementations of service interfaces |
| `internal/adapter/` | HTTP clients for external APIs (exchange rate) |
| `internal/server/` | Composition root — `deps.go` wires everything, `server.go` registers routes |
| `internal/middleware/` | JWT auth (`AuthMiddleware`), role check (`RequireRole`) |
| `internal/model/` | Domain models and request/response DTOs |
| `internal/domain/` | Shared sentinel errors (`ErrNotFound`, `ErrUnauthorized`, …) |
| `internal/config/` | Env-based config loaded via `godotenv` |
| `internal/cache/` | `Cache` interface + Redis implementation |
| `internal/database/` | `TransactionManager` for context-propagated DB transactions |
| `internal/consumer/` | Kafka consumer lifecycle |
| `pkg/jwt/` | JWT generation and validation |
| `pkg/password/` | Bcrypt helpers |
| `proto/` | Protobuf definitions and generated Go code |
| `migrations/` | SQL files for golang-migrate (`NNN_name.up.sql` / `.down.sql`) |

## Tech Stack

- **HTTP**: `github.com/gin-gonic/gin`
- **ORM**: `gorm.io/gorm` + `gorm.io/driver/postgres`
- **gRPC**: `google.golang.org/grpc` + `google.golang.org/protobuf`
- **Messaging**: `github.com/segmentio/kafka-go`
- **Cache**: `github.com/redis/go-redis/v9`
- **Auth**: `github.com/golang-jwt/jwt/v5` + `golang.org/x/crypto`
- **Migrations**: `github.com/golang-migrate/migrate/v4`
- **Testing**: `github.com/stretchr/testify`
- **Module**: `github.com/golang-api-server`

## Conventions

### Errors
- Always wrap: `fmt.Errorf("context: %w", err)` — never bare `return err`
- Compare with `errors.Is(err, domain.ErrNotFound)` — never `==`
- Service layer returns domain sentinel errors; handlers map them to HTTP status codes
- Never discard errors with `_`

### Context
- `ctx context.Context` is always the first parameter
- Propagate through the full call chain

### Interfaces
- Service layer defines its own interfaces (in `service/service.go` or alongside the service file)
- Inject concrete types only at `server/deps.go`

### Testing
- Table-driven tests: `t.Run(tc.name, ...)`
- Mocks live in `mock_test.go` within the same package
- Run: `make test` (race detector on)

### Naming
- Package names: short, lowercase, no underscores
- Error messages: lowercase, no trailing punctuation

## Adding a New Feature

1. **Model** — add domain struct + DTOs in `internal/model/`
2. **Migration** — `migrations/NNN_name.up.sql` + `.down.sql`
3. **Service interface** — define in `internal/service/service.go`
4. **Repository** — implement interface in `internal/repository/`
5. **Service** — implement business logic in `internal/service/`
6. **Handler** — Gin handler in `internal/handler/`
7. **Wire** — add to `Dependencies` in `server/deps.go`; register route in `server/server.go`
8. **gRPC** *(optional)* — add proto, `make proto`, implement in `internal/grpc/`, register in `server.go`

## Domain Error → HTTP Status

| Sentinel | Status |
|----------|--------|
| `domain.ErrNotFound` | 404 |
| `domain.ErrUnauthorized` | 401 |
| `domain.ErrForbidden` | 403 |
| `domain.ErrConflict` | 409 |
| other | 500 |

## Make Commands

```bash
make run            # Build and run
make test           # Tests with race detector
make test-cover     # Coverage report
make lint           # golangci-lint
make migrate-up     # Apply migrations
make migrate-down   # Rollback last migration
make docker-up      # Start all services (PostgreSQL, Kafka, app)
make docker-down    # Stop containers
make proto          # Regenerate protobuf code
```

## Ports

| Service | Port |
|---------|------|
| HTTP API | `:8080` |
| gRPC | `:9090` |

## Agents

A project-scoped agent is available at `.claude/agents/golang-api-server.md` with deeper implementation guidance. The following global agents are recommended for this repo:

| Agent | When to use |
|-------|-------------|
| `go-reviewer` | After writing or modifying any `.go` file |
| `go-build-resolver` | When `go build` or `go vet` fails |
| `tdd-guide` | When adding new features or fixing bugs |
| `security-reviewer` | Before committing auth, JWT, or input-handling code |
| `database-reviewer` | When writing SQL migrations or GORM queries |
