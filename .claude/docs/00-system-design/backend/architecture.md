# Kiến trúc Dự án

## 1. Cấu trúc Thư mục

```
myapp/
├── cmd/
│   └── api/
│       └── main.go                    # Entry point, khởi tạo DI container
│
├── internal/
│   ├── auth/                          # Auth module
│   │   ├── domain/                    # Entities (User)
│   │   ├── application/
│   │   │   ├── port/
│   │   │   │   ├── inbound.go         # AuthUseCasePort
│   │   │   │   └── outbound.go        # UserRepositoryPort, TokenServicePort, PasswordServicePort, RefreshTokenStorePort
│   │   │   ├── dto/                   # Request/Response DTOs
│   │   │   └── usecase/               # AuthUseCase
│   │   ├── adapter/
│   │   │   ├── handler/               # HTTP handlers (Gin)
│   │   │   ├── repository/            # Postgres (User), Redis (RefreshToken)
│   │   │   └── security/              # JWTService, BcryptPasswordService
│   │   └── module.go                  # Module wiring + RegisterRoutes
│   │
│   ├── vocabulary/                    # Vocabulary module
│   │   ├── domain/                    # Entities (Vocabulary, Folder)
│   │   ├── application/
│   │   │   ├── port/
│   │   │   │   ├── inbound.go         # VocabularyCommand/QueryPort, FolderCommand/QueryPort
│   │   │   │   └── outbound.go        # VocabularyRepositoryPort, FolderRepositoryPort
│   │   │   ├── dto/
│   │   │   └── usecase/               # CQRS: vocabulary_command, vocabulary_query, folder_command, folder_query
│   │   ├── adapter/
│   │   │   ├── handler/
│   │   │   └── repository/
│   │   └── module.go
│   │
│   ├── learning/                      # Learning module
│   │   ├── domain/
│   │   │   ├── learning_card.go       # Entity + LearningMode constants
│   │   │   └── service/
│   │   │       └── scoring.go         # SM-2 domain service (review intervals, mastery reset)
│   │   ├── application/
│   │   │   ├── port/
│   │   │   │   ├── inbound.go         # LearningCommand/QueryPort, ReviewCommand/QueryPort
│   │   │   │   └── outbound.go        # LearningCardRepositoryPort
│   │   │   ├── dto/
│   │   │   └── usecase/               # CQRS: learning_command, learning_query, review_command, review_query
│   │   ├── adapter/
│   │   │   ├── handler/
│   │   │   └── repository/
│   │   └── module.go
│   │
│   ├── shared/                        # Shared kernel
│   │   ├── error/                     # AppError (codes: NOT_FOUND, INVALID_INPUT, etc.)
│   │   ├── logger/                    # Logger interface + Field constructors
│   │   ├── ctxlog/                    # Context-aware log fields (request_id, trace_id)
│   │   ├── i18n/                      # Translation engine (5 languages)
│   │   ├── middleware/                # Auth, CORS, i18n, Logger, RateLimit, Recovery, RequestID, Security
│   │   ├── response/                  # APIResponse helpers (Success, BadRequest, ValidationBadRequest, etc.)
│   │   └── dto/                       # PaginationRequest/PaginatedResponse
│   │
│   ├── server/                        # HTTP server + router
│   │   ├── router.go                  # Route registration + health check
│   │   └── server.go
│   │
│   └── infrastructure/                # Cross-cutting infrastructure
│       ├── di/                        # Container (NewApp), persistence init, observability init
│       ├── config/                    # Viper config (auth, db, redis, log, circuitbreaker, observability)
│       ├── database/                  # GORM postgres connection + custom GORM logger
│       ├── circuitbreaker/            # gobreaker wrapper + registry
│       ├── logging/                   # Zap adapter (console, daily file, async OTLP)
│       ├── redis/                     # Redis client init
│       ├── sentry/                    # Sentry error tracking
│       └── tracing/                   # OpenTelemetry OTLP tracer
│
├── resources/
│   └── i18n/                          # Translation files (en, vi, th, zh, id)
│
├── go.mod
├── go.sum
├── Makefile
└── CLAUDE.md
```

```
HTTP Request
    ↓
[Server] router.go → module.RegisterRoutes()
    ↓
[Middleware] RequestID → Language → Auth → RateLimit → Logger → Recovery
    ↓
[Adapter] handler/
    ↓  calls interface
[Application] port/inbound.go (input port)
    ↓  implemented by
[Application] usecase/
    ↓  calls interface
[Application] port/outbound.go (output port)
    ↓  implemented by
[Adapter] repository/ | security/
    ↓
Database / Redis / External Services
```

## 2. Các Lớp (Layers)

### Domain Layer (`<module>/domain/`)
Đây là lớp trong cùng, chứa các quy tắc nghiệp vụ cốt lõi.
- **Entities**: Các đối tượng có định danh (Identity): `User`, `Vocabulary`, `Folder`, `LearningCard`.
- **Domain Constants**: `LearningMode`, `MemoryState`, mode weights.
- **Domain Services**: Logic nghiệp vụ không thuộc entity cụ thể (ví dụ: `learning/domain/service/scoring.go` — SM-2 algorithm).
- **Entity Errors**: Lỗi đặc thù của entity (ví dụ: `ErrHanziRequired`, `ErrFolderNameRequired`).
- **UUID v7**: Tất cả entity IDs dùng `uuid.Must(uuid.NewV7())` — time-ordered, tốt cho DB indexing.
- **Đặc điểm**: Không phụ thuộc vào bất kỳ lớp nào khác bên ngoài. Không import framework, ORM, hay crypto libraries.

### Application Layer (`<module>/application/`)
Lớp này điều phối các hoạt động của ứng dụng.
- **Inbound Ports (`port/inbound.go`)**: Interfaces cho handlers gọi usecases. CQRS split: Command ports (write) và Query ports (read).
- **Outbound Ports (`port/outbound.go`)**: Interfaces cho usecases gọi repositories và services (PasswordServicePort, TokenServicePort, etc.).
- **Use Cases (`usecase/`)**: Triển khai các Inbound Ports. Vocabulary và Learning modules dùng CQRS (tách command/query files riêng).
- **DTOs (`dto/`)**: Data Transfer Objects với Gin binding tags cho validation (`required`, `email`, `min`, `max`).
- **Đặc điểm**: Chỉ phụ thuộc vào Domain Layer.

### Adapter Layer (`<module>/adapter/`)
Chứa các implementations cụ thể để kết nối Core với thế giới bên ngoài.
- **Handler (Driving)**: Nhận request từ bên ngoài. Bind JSON/Query → validate → gọi usecase qua inbound port. Validation errors trả field-level details qua `ValidationBadRequest()`.
- **Repository (Driven)**: GORM repositories implement outbound ports. Entity ↔ Model tách biệt với `toEntity()`/`fromEntity()`. Timestamps sync back sau Create/Save.
- **Security (Driven)**: JWT token service, bcrypt password service — implement outbound ports.
- **Đặc điểm**: Phụ thuộc vào Application Layer (implement các Ports).

### Infrastructure Layer (`internal/infrastructure/`)
Cung cấp các công cụ và cấu hình để chạy ứng dụng.
- **Config**: Load biến môi trường qua Viper từ `.env`.
- **Database**: GORM postgres connection + custom GORM logger (slow query detection >200ms).
- **DI**: `container.go` → `initPersistence()` → module factories. Manual constructor injection.
- **Observability**: Zap structured logging + OTEL tracing + Sentry error tracking.
- **Resilience**: Circuit breaker (gobreaker) registry cho persistence layer. gobreaker chạy in-memory per-process, không hỗ trợ distributed — chấp nhận được vì mỗi instance tự protect.

### Deployment & Scaling
- Dự án sẽ **horizontal scaling** với **Kubernetes replicas** và **load balancer**.
- **Rate limiter**: Cần chuyển sang **distributed (Redis-backed)** để đảm bảo rate limit chính xác across instances. In-memory rate limit sẽ bị bypass khi có N instances (mỗi instance cho phép riêng → user thực tế được N × limit).
- **Circuit breaker**: In-memory per-instance (gobreaker) vẫn ok — mỗi instance tự bảo vệ, không cần share state.
- **Stateful components**: Khi thêm component in-memory mới, luôn cân nhắc multi-instance scenario.

### Shared Kernel (`internal/shared/`)
Code dùng chung giữa các modules.
- **AppError**: Error codes layered: entity errors → `ErrInvalidInput`/`ErrNotFound` (usecase) → HTTP status + i18n key (handler).
- **Response**: `Success()`, `SuccessWithMetadata()` (pagination), `ValidationBadRequest()` (field-level validation errors).
- **Middleware**: Auth (JWT), CORS, i18n (language detection), Request Logger, Rate Limiting, Recovery (panic → Sentry), RequestID (UUID v7), Security headers.

---

## 3. Vòng đời của một API Request (Request Lifecycle)

**Ví dụ: Tạo từ vựng mới (Create Vocabulary)**

1.  **Client Request**:
    - Client gửi HTTP POST request tới `/api/vocabularies` với JSON body (hanzi, pinyin, meaning).
    - Request đi kèm Header `Authorization: Bearer <token>`.

2.  **Infrastructure (Server)**:
    - `http.Server` nhận request.
    - Request đi qua **Router** (`gin.Engine`) và middleware chain.

3.  **Middleware Chain**:
    - `RequestIDMiddleware` — gán UUID v7 request ID, propagate qua context.
    - `LanguageMiddleware` — detect ngôn ngữ từ query/header.
    - `AuthMiddleware` — validate JWT token, set `user_id` vào context.
    - `RateLimitMiddleware` — kiểm tra rate limit.
    - `RequestLoggerMiddleware` — log request/response (skip sensitive paths).
    - `RecoveryMiddleware` — catch panic, report Sentry.

4.  **Adapter Layer (Handler)**:
    - **Handler** (`VocabularyHandler.CreateVocabulary`) nhận request.
    - **Binding/Validation**: `ShouldBindJSON(&req)` validate DTO.
    - Nếu dữ liệu sai → `ValidationBadRequest(c, err)` trả field-level details (`{"hanzi": "required"}`).
    - Nếu dữ liệu đúng → Gọi xuống Application Layer thông qua Inbound Port.

5.  **Application Layer (Use Case)**:
    - **Use Case** (`VocabularyCommand.CreateVocabulary`) nhận DTO.
    - **Business Logic**:
        - Chuyển đổi DTO sang Domain Entity (`domain.NewVocabulary(...)` — validate, generate UUID v7).
        - Entity validate trả entity errors → usecase map sang `AppError`.
    - Gọi xuống Persistence Layer thông qua Outbound Port `VocabularyRepositoryPort`.

6.  **Adapter Layer (Repository)**:
    - **Repository** (`VocabularyRepository`) nhận Entity.
    - **Mapping**: `fromVocabEntity()` chuyển Domain Entity sang DB Model.
    - **Database Execution**: GORM INSERT vào PostgreSQL (qua circuit breaker).
    - **Timestamp sync**: GORM-managed CreatedAt/UpdatedAt sync back vào entity pointer.
    - Trả về kết quả cho Use Case.

7.  **Application Layer (Use Case) - Trả về**:
    - Nhận kết quả từ Repository.
    - Nếu thành công → Chuyển đổi Entity sang Response DTO.
    - Trả DTO về cho Handler.

8.  **Adapter Layer (Handler) - Response**:
    - Handler nhận DTO từ Use Case.
    - `response.Success(c, 201, res)` — serialize thành JSON, translate message qua i18n.
    - Gửi HTTP Response về Client.

### Sơ đồ luồng dữ liệu

```
Client (HTTP)
   │
   ▼
[Infrastructure] HTTP Server / Router
   │
   ▼
[Middleware] RequestID → Language → Auth → RateLimit → Logger → Recovery
   │
   ▼
[Adapter] Handler (ShouldBindJSON → ValidationBadRequest if fail)
   │         (DTO)
   ▼
[Application] Use Case (Business Logic, Error mapping)
   │         (Entity)
   ▼
[Adapter] Repository (Entity ↔ Model mapping, Circuit Breaker)
   │         (DB Model)
   ▼
[Infrastructure] Database (PostgreSQL) / Redis
```
