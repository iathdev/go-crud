# Phase 0 — Init Project

## Mục tiêu

Thiết lập nền tảng kỹ thuật cho toàn bộ dự án: project skeleton, infrastructure layer, shared kernel, và conventions. Sau phase này, các module nghiệp vụ (auth, vocabulary, ...) có thể được phát triển song song mà không cần thay đổi foundation.

## Scope

### 1. Project Skeleton

- Entry point `cmd/api/main.go` với graceful shutdown (SIGINT/SIGTERM, 5s timeout)
- Module `learning-go`, Go 1.25+
- Makefile: `run`, `build`, `docker-up`, `docker-down`, `migrate-up`, `migrate-down`, `migrate-reset`
- Docker Compose: PostgreSQL 15+ và Redis 7+
- `.env` config via Viper

### 2. Infrastructure Layer (`internal/infrastructure/`)

| Package | Chức năng |
|---|---|
| `config/` | Load env vars từ `.env` qua Viper. Tách config struct theo domain (auth, db, redis, log, circuit breaker, observability) |
| `database/` | GORM PostgreSQL connection + custom GORM logger (slow query detection >200ms) |
| `redis/` | Redis client factory |
| `di/` | Manual DI container — `NewApp()` khởi tạo persistence, observability, modules, router, trả về `(server, cleanup, error)` |
| `logging/` | Zap structured logging: console + daily file rotation (lumberjack) + async OTLP channel |
| `circuitbreaker/` | gobreaker v2 wrapper + `BreakerRegistry`. In-memory per-process. Config qua `CB_*` env vars |
| `sentry/` | Sentry error tracking integration |
| `tracing/` | OpenTelemetry OTLP gRPC tracer setup |

### 3. Shared Kernel (`internal/shared/`)

| Package | Chức năng |
|---|---|
| `error/` | `AppError` với typed codes: BAD_REQUEST, UNAUTHORIZED, FORBIDDEN, NOT_FOUND, CONFLICT, UNPROCESSABLE_ENTITY, INTERNAL_SERVER_ERROR, SERVICE_UNAVAILABLE |
| `response/` | Unified API response: `Success()`, `SuccessWithMetadata()`, `HandleError()`. Tự động map `AppError.Code()` → HTTP status + i18n translation |
| `logger/` | Logger interface + global logger. Tất cả log phải có module prefix `[AUTH]`, `[VOCABULARY]`, ... |
| `ctxlog/` | Context-based field storage cho structured logging (request_id, trace_id, user_id) |
| `i18n/` | Translation engine: 5 ngôn ngữ (en, vi, zh, th, id). JSON files trong `resources/i18n/<lang>/`. Fallback: English → raw key |
| `middleware/` | SecurityHeaders, CORS, RequestID (UUID v7), OTEL, RequestLogger, Language detection, Recovery (panic → Sentry), RateLimit, Auth (JWT) |
| `dto/` | PaginationRequest, PaginationMeta, PaginatedResponse |

### 4. HTTP Server (`internal/server/`)

- `router.go` — Middleware chain: SecurityHeaders → CORS → RequestID → OTEL → RequestLogger → Language → Recovery
- Route groups: public (rate-limited) và protected (`/api/*` — JWT auth)
- `server.go` — HTTP server wrapper với configurable timeouts
- Health check `GET /health` — ping PostgreSQL

### 5. Hexagonal Architecture Convention

Mỗi module tuân theo cấu trúc:

```
internal/<module>/
├── domain/          # Entities, value objects, domain errors. Zero external dependencies.
├── application/
│   ├── port/        # inbound.go (driving) + outbound.go (driven)
│   ├── dto/         # Request/Response DTOs với Gin binding tags
│   └── usecase/     # CQRS: *_command.go + *_query.go
├── adapter/
│   ├── handler/     # HTTP handlers (Gin)
│   ├── repository/  # DB models + GORM implementations. Entity ↔ Model mapping.
│   └── security/    # Module-specific security adapters
└── module.go        # NewModule() wiring + RegisterRoutes(public, protected)
```

**Error flow:** Repo (raw error / nil) → Use case (wrap thành AppError) → Handler (`response.HandleError`) → HTTP status + i18n

**Module registration:** DI container tạo module → Router gọi `module.RegisterRoutes(public, api)`

### 6. Conventions được thiết lập

- **UUID v7** cho tất cả entity IDs và request IDs (time-ordered, B-tree friendly)
- **Domain entities vs DB models** tách biệt — repository chịu trách nhiệm mapping
- **i18n keys** thay vì English strings trong AppError messages
- **4xx errors không log**, 5xx errors log khi cần context debug
- **Migrations** dạng SQL: `migrations/NNNNNN_description.{up,down}.sql`

## Kết quả

Sau phase này, để thêm một module mới chỉ cần:

1. Tạo `internal/<module>/` theo convention ở trên
2. Wire trong `infrastructure/di/container.go`
3. Register routes trong `server/router.go`
