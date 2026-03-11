# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Go REST API using **Modular Monolith** with **Hexagonal Architecture (Ports and Adapters)** per module. Built with Gin, GORM, and PostgreSQL. Module name: `learning-go`. Requires **Go 1.25+**.

## Commands

```bash
make docker-up          # Start PostgreSQL + Redis via Docker Compose
make docker-down        # Stop Docker Compose services
make run                # Run the API server (go run cmd/api/main.go)
make build              # Build binary to bin/api
make migrate-up         # Run all pending migrations (requires golang-migrate CLI)
make migrate-down       # Roll back 1 migration (use make migrate-down-N for N steps)
make migrate-reset      # Drop all tables
go test ./...           # Run all tests
go test ./internal/auth/domain/...         # Run tests for a specific package
go test -run TestName ./internal/...       # Run a single test by name
```

Requires a `.env` file (copy from `.env.example`). The Makefile reads `.env` for DB connection vars. Migration commands require the [`golang-migrate`](https://github.com/golang-migrate/migrate) CLI.

## Architecture

**Modular Monolith** — each subdomain is a self-contained module with its own Hexagonal Architecture. Modules communicate through exported interfaces, never by importing each other's internals.

**Request flow:** HTTP Router -> Module.RegisterRoutes() -> Handler -> Input Port (interface) -> Use Case -> Output Port (interface) -> Repository -> Database

### Module Structure

Each module (`auth/`, `vocabulary/`) follows the same internal layout:

```
internal/<module>/
├── docs/                # Module-specific documentation (requirement, API contract, plan)
├── domain/              # Entities and domain errors. Zero external dependencies.
├── application/
│   ├── port/            # Input (driving) and output (driven) port interfaces
│   ├── dto/             # Data transfer objects for the module
│   └── usecase/         # Use case implementations
├── adapter/
│   ├── handler/         # HTTP handlers (Gin)
│   ├── repository/      # Repository implementations (Postgres, Redis)
│   └── security/        # Module-specific security (e.g., JWT in auth)
└── module.go            # Module wiring + RegisterRoutes(public, protected)
```

### Layer Map

- **`cmd/api/main.go`** — Entry point, calls DI container
- **`internal/auth/`** — Auth subdomain module (SSO login, refresh, logout, profile, onboarding)
  - `domain/` — User value object (from Prep), UserProfile entity
  - `application/port/` — AuthUseCasePort, UserProfileRepositoryPort, PrepUserServicePort, TokenServicePort, RefreshTokenStorePort
  - `application/usecase/` — SSO login, token pair generation, profile management
  - `adapter/handler/` — HTTP handlers for auth endpoints
  - `adapter/repository/` — Postgres user profile repo, Redis refresh token store
  - `adapter/security/` — JWT token service implementation
  - `adapter/external/` — Prep User Service HTTP adapter + dev stub
  - `module.go` — Wires all auth internals, exposes RegisterRoutes
- **`internal/vocabulary/`** — Vocabulary subdomain module (CRUD vocabularies, folders, topics, grammar points, OCR, import)
  - `domain/` — Vocabulary entity (with Example value object), Folder entity, Topic entity, GrammarPoint entity
  - `application/port/` — VocabularyCommandPort, VocabularyQueryPort, FolderCommandPort, FolderQueryPort, TopicQueryPort, OCRCommandPort, ImportCommandPort, VocabularyRepositoryPort, FolderRepositoryPort, TopicRepositoryPort, GrammarPointRepositoryPort
  - `application/usecase/` — Vocabulary CRUD, Folder CRUD, Topic query, OCR scan, bulk import
  - `adapter/handler/` — HTTP handlers for vocabulary, folder, topic, OCR, import endpoints
  - `adapter/repository/` — Postgres vocabulary repo, folder repo, topic repo, grammar point repo
  - `module.go` — Wires all vocabulary internals, exposes RegisterRoutes
- **`internal/shared/`** — Shared kernel
  - `error/` — AppError with typed codes (NOT_FOUND, INVALID_INPUT, etc.)
  - `logger/` — Logger interface and global logger
  - `ctxlog/` — Context-based field storage for structured logging
  - `i18n/` — Internationalization engine (en, vi, zh, th, id)
  - `middleware/` — HTTP middleware (auth JWT, CORS, rate limiting, i18n, recovery, security headers, request ID, logging)
  - `response/` — Unified API response formatting with i18n translation
  - `dto/` — Shared DTOs (PaginationRequest, PaginationMeta, PaginatedResponse)
- **`internal/server/`** — HTTP server and router
  - `router.go` — Registers global middleware, creates route groups, calls module.RegisterRoutes()
  - `server.go` — HTTP server wrapper with timeouts
- **`internal/infrastructure/`** — Cross-cutting infrastructure
  - `di/` — Manual dependency injection (creates modules, wires persistence & observability)
  - `config/` — Viper-based config from `.env`
  - `database/` — Postgres connection setup
  - `circuitbreaker/` — gobreaker v2 wrapper with registry pattern
  - `logging/` — Zap logger factory, multi-logger, async logger
  - `redis/` — Redis client factory
  - `sentry/` — Sentry error tracking
  - `tracing/` — OpenTelemetry setup

### Key Patterns

- **Module boundary**: Each module is self-contained. Module A must NOT import internal packages of Module B. Cross-module communication goes through exported ports/interfaces.
- **Module registration**: Each module exposes `NewModule(deps...) *Module` and `RegisterRoutes(public, protected *gin.RouterGroup)`. The DI container creates modules; the router calls RegisterRoutes.
- **Domain entities vs DB models**: Domain entities live in `<module>/domain/`, DB models in `<module>/adapter/repository/`. Repositories handle mapping between them.
- **CQRS**: Vocabulary and learning modules split use cases into `*Command` and `*Query` types with corresponding port interfaces. Follow this pattern for new modules.
- **Circuit breaker**: Infrastructure-level gobreaker v2 wrapper lives in `infrastructure/circuitbreaker/` with a `BreakerRegistry` pattern. The breaker converts `gobreaker.ErrOpenState`/`ErrTooManyRequests` into domain `ErrServiceUnavailable`. Only `nil` and `ErrNotFound` count as success. Settings configurable via `CB_*` env vars. Currently not wired to repository adapters but available for use (e.g., wrapping external service calls).
- **Error handling**: `AppError` in `shared/error/` carries a typed `Code`. Handlers switch on `AppError.Code()` to map to HTTP status codes. All error messages are i18n-translated at the response layer.
- **i18n**: Language detected from `lang` query param > `X-Lang` header > `Accept-Language` header. Translation files are JSON in `resources/i18n/<lang>/<domain>.json`. Falls back to English, then to the raw key.
- **Middleware chain**: SecurityHeaders -> CORS -> RequestID -> OTEL -> RequestLogger -> Language -> Recovery. Rate limiting on public routes. JWT auth on `/api/*` routes.
- **Config**: All config via environment variables loaded from `.env` using Viper.
- **DI**: Manual constructor injection in `infrastructure/di/`. No framework. Creates modules with their dependencies and returns a cleanup function for graceful shutdown.
- **Testing**: Unit tests are colocated with source files (`*_test.go`). Uses table-driven tests with `t.Run()` subtests. No mock framework — tests focus on domain logic.
- **Migrations**: SQL files in `migrations/` named `NNNNNN_description.{up,down}.sql` (6-digit zero-padded prefix).

### Adding a New Module

1. Create `internal/<module>/` with the standard layout (domain, application/{port,dto,usecase}, adapter/{handler,repository}, module.go)
2. In `module.go`: wire internal dependencies in `NewModule()`, register routes in `RegisterRoutes()`
3. In `infrastructure/di/container.go`: create the module and pass to `server.NewRouter()`
4. In `server/router.go`: add module parameter and call `module.RegisterRoutes(public, api)`

## API Routes

- `POST /register`, `POST /login`, `POST /refresh` — Public (rate-limited: 5 req/sec, burst 10)
- `/api/*` — Protected by JWT auth middleware (vocabulary CRUD, folders, learning, review, logout)
- `GET /health` — Health check

## Documentation

Additional design docs in `docs/`: `architecture.md`, `tech_stack.md`, `requirement.md`.

## Rules

- **Coding Style**: See [`docs/coding_style.md`](docs/coding_style.md)
- **Planning Rules**: See [`docs/planning_rules.md`](docs/planning_rules.md)
