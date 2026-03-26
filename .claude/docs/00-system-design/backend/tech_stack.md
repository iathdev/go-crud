# Danh sách Công nghệ và Thư viện (Tech Stack)

Tài liệu này liệt kê các công nghệ, công cụ và thư viện quan trọng được sử dụng trong dự án.

## 1. Công nghệ Cốt lõi (Core Technologies)

| Công nghệ | Phiên bản | Mô tả |
| :--- | :--- | :--- |
| **Go (Golang)** | 1.25+ | Ngôn ngữ lập trình chính, hiệu năng cao, concurrency tốt. |
| **PostgreSQL** | 15+ | Hệ quản trị cơ sở dữ liệu quan hệ (RDBMS) chính. |
| **Redis** | 7+ | In-memory data store cho refresh token storage. |
| **Docker** | Latest | Containerization platform để đóng gói và chạy ứng dụng. |
| **Docker Compose** | Latest | Công cụ định nghĩa và chạy multi-container Docker applications. |

## 2. Thư viện Go Quan trọng (Key Go Packages)

### Web Framework
- **[Gin Gonic](https://github.com/gin-gonic/gin)** (`github.com/gin-gonic/gin`):
  - Web framework high-performance cho Go.
  - HTTP routing, middleware chain, JSON binding/validation.
  - Validation powered by `go-playground/validator/v10`.

### Database & ORM
- **[GORM](https://gorm.io/)** (`gorm.io/gorm`):
  - ORM phổ biến nhất cho Go.
  - Entity ↔ Model mapping, soft delete (`gorm.DeletedAt`), auto timestamps.
- **[GORM Postgres Driver](https://github.com/go-gorm/postgres)** (`gorm.io/driver/postgres`):
  - Driver kết nối PostgreSQL.
- **[Redis Go](https://github.com/redis/go-redis)** (`github.com/redis/go-redis/v9`):
  - Redis client cho refresh token storage (hashed tokens, pipeline operations).

### Configuration
- **[Viper](https://github.com/spf13/viper)** (`github.com/spf13/viper`):
  - Configuration management — load từ `.env` file.

### Authentication & Security
- **[JWT Go](https://github.com/golang-jwt/jwt)** (`github.com/golang-jwt/jwt/v5`):
  - Tạo và xác thực JWT tokens (HS256).
  - Access token (configurable expiry) + refresh token rotation.
- **[Go Crypto](https://pkg.go.dev/golang.org/x/crypto)** (`golang.org/x/crypto`):
  - `bcrypt` password hashing — triển khai qua `PasswordServicePort` adapter (không ở domain layer).

### Observability
- **[Uber Zap](https://github.com/uber-go/zap)** (`go.uber.org/zap`):
  - Structured logging (JSON format). Multiple channels: console, daily file rotation, async OTLP.
- **[OpenTelemetry](https://opentelemetry.io/)** (`go.opentelemetry.io/otel`):
  - Distributed tracing via OTLP gRPC exporter.
  - Auto-inject trace_id/span_id vào logs.
- **[Sentry Go](https://github.com/getsentry/sentry-go)** (`github.com/getsentry/sentry-go`):
  - Error tracking và panic capture.
- **[Lumberjack](https://github.com/natefinch/lumberjack)** (`gopkg.in/natefinch/lumberjack.v2`):
  - Log file rotation (configurable max size, backups, age).

### Resilience
- **[GoBreaker](https://github.com/sony/gobreaker)** (`github.com/sony/gobreaker`):
  - Circuit breaker pattern cho database/external service calls.
  - Registry-based: mỗi service có config riêng.

### Utilities
- **[Google UUID](https://github.com/google/uuid)** (`github.com/google/uuid` v1.6.0):
  - UUID v7 (time-ordered) cho entity IDs và request IDs.
  - Tốt cho DB indexing (B-tree friendly), chứa timestamp.
- **[golang-migrate](https://github.com/golang-migrate/migrate)**:
  - Database migration CLI tool.

## 3. Kiến trúc (Architecture)

- **CQRS**: Command/Query split cho Vocabulary và Learning modules.
- **Ports split**: `inbound.go` (driving — handlers gọi usecases) và `outbound.go` (driven — usecases gọi repositories/services).
- **Domain Services**: Business logic không thuộc entity cụ thể (SM-2 scoring algorithm).
- **Dependency Injection**: Manual constructor injection qua DI container. Constructors trả về port interfaces.
- **i18n**: 5 ngôn ngữ (en, vi, th, zh, id). Response messages translate qua i18n keys.

## 4. Công cụ Phát triển (Development Tools)

- **Makefile**: Tự động hóa (run, build, docker-up, migrate).
- **Postman / cURL**: Test API.
- **Docker Compose**: Local development (PostgreSQL, Redis).
