# Kiến trúc Dự án (Modular Monolith + Hexagonal Architecture)

Dự án này được xây dựng dựa trên kiến trúc **Modular Monolith** với **Hexagonal Architecture** (Ports and Adapters) cho từng module. Mục tiêu chính là tách biệt phần **Business Logic** (Core) khỏi các yếu tố bên ngoài (Frameworks, Database, UI, External APIs), đồng thời giữ mỗi subdomain độc lập trong cùng một binary.

## 1. Cấu trúc Thư mục

```
myapp/
├── cmd/
│   └── api/
│       └── main.go                    # Entry point, khởi tạo DI container
│
├── internal/
│   ├── auth/                          # Auth module (SSO via Prep platform)
│   │   ├── domain/                    # Entities (User, PrepUser)
│   │   ├── application/
│   │   │   ├── port/
│   │   │   │   ├── inbound.go         # AuthUseCasePort
│   │   │   │   └── outbound.go        # UserRepositoryPort, PrepUserServicePort
│   │   │   ├── dto/                   # Request/Response DTOs
│   │   │   └── usecase/               # AuthUseCase (GetMe)
│   │   ├── adapter/
│   │   │   ├── handler/               # HTTP handlers (Gin)
│   │   │   ├── repository/            # Postgres (User) + model/
│   │   │   └── service/               # PrepUserService (Prep API HTTP client, circuit breaker)
│   │   └── module.go                  # Module wiring + RegisterRoutes
│   │
│   ├── vocabulary/                    # Vocabulary module
│   │   ├── domain/                    # Entities (Vocabulary, Folder, Topic, GrammarPoint)
│   │   ├── application/
│   │   │   ├── port/
│   │   │   │   ├── inbound.go         # VocabularyCommand/QueryPort, FolderCommand/QueryPort, TopicQueryPort, OCRCommandPort, ImportCommandPort
│   │   │   │   └── outbound.go        # VocabularyRepositoryPort, FolderRepositoryPort, TopicRepositoryPort, GrammarPointRepositoryPort, OCRServicePort
│   │   │   ├── dto/
│   │   │   └── usecase/               # CQRS: vocabulary_command/query, folder_command/query, topic_query, ocr_command, import_command
│   │   ├── adapter/
│   │   │   ├── handler/               # HTTP handlers (vocabulary, folder, topic, OCR, import)
│   │   │   ├── repository/            # Postgres repos (vocabulary, folder, topic, grammar_point) + model/
│   │   │   └── service/               # OCR service (PaddleOCR HTTP adapter, circuit breaker)
│   │   └── module.go
│   │
│   ├── shared/                        # Shared kernel
│   │   ├── error/                     # AppError (codes: NOT_FOUND, BAD_REQUEST, UNAUTHORIZED, etc.)
│   │   ├── logger/                    # Logger interface + Field constructors
│   │   ├── ctxlog/                    # Context-aware log fields (request_id, trace_id)
│   │   ├── i18n/                      # Translation engine (5 languages: en, vi, zh, th, id)
│   │   ├── middleware/                # Auth, CORS, i18n, Logger, RateLimit, Recovery, RequestID, Security
│   │   ├── response/                  # APIResponse helpers (Success, HandleError, ValidationBadRequest)
│   │   ├── dto/                       # PaginationRequest/PaginatedResponse
│   │   └── common/                    # JSONB helper type
│   │
│   ├── server/                        # HTTP server + router
│   │   ├── router.go                  # Route registration + health check
│   │   └── server.go
│   │
│   └── infrastructure/                # Cross-cutting infrastructure
│       ├── di/                        # Container (NewApp), persistence init, observability init
│       ├── config/                    # Viper config (auth, db, redis, log, circuitbreaker, observability)
│       ├── database/                  # GORM postgres connection + custom GORM logger
│       ├── circuitbreaker/            # gobreaker v2 wrapper + registry
│       ├── logging/                   # Zap adapter (console, daily file, async OTLP)
│       ├── redis/                     # Redis client init
│       ├── sentry/                    # Sentry error tracking
│       └── tracing/                   # OpenTelemetry OTLP tracer
│
├── resources/
│   └── i18n/                          # Translation files (en, vi, th, zh, id)
│
├── migrations/                        # SQL migration files (golang-migrate)
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
[Middleware] SecurityHeaders → CORS → RequestID → OTEL → RequestLogger → Language → Recovery
    ↓
[Adapter] handler/
    ↓  calls interface
[Application] port/inbound.go (input port)
    ↓  implemented by
[Application] usecase/
    ↓  calls interface
[Application] port/outbound.go (output port)
    ↓  implemented by
[Adapter] repository/ | service/
    ↓
Database / Redis / External Services
```

## 2. Các Lớp (Layers)

### Domain Layer (`<module>/domain/`)
Đây là lớp trong cùng, chứa các quy tắc nghiệp vụ cốt lõi.
- **Entities**: Các đối tượng có định danh (Identity).
  - Auth: `User` (local user với PrepUserID), `PrepUser` (value object từ Prep platform).
  - Vocabulary: `Vocabulary` (Hanzi, pinyin, meanings, HSK level 1-9, examples, radicals, stroke data), `Folder` (user-owned collection), `Topic` (global topic với translations CN/VI/EN và slug), `GrammarPoint` (grammar pattern với code, example, rule, HSK level).
- **Entity Errors**: Lỗi đặc thù của entity (ví dụ: `ErrHanziRequired`, `ErrFolderNameRequired`).
- **UUID v7**: Tất cả entity IDs dùng `uuid.Must(uuid.NewV7())` — time-ordered, tốt cho DB indexing.
- **Đặc điểm**: Không phụ thuộc vào bất kỳ lớp nào khác bên ngoài. Không import framework, ORM, hay crypto libraries.

### Application Layer (`<module>/application/`)
Lớp này điều phối các hoạt động của ứng dụng.
- **Inbound Ports (`port/inbound.go`)**: Interfaces cho handlers gọi usecases. Vocabulary module dùng CQRS split: Command ports (write) và Query ports (read).
- **Outbound Ports (`port/outbound.go`)**: Interfaces cho usecases gọi repositories và external services (PrepUserServicePort, OCRServicePort, etc.).
- **Use Cases (`usecase/`)**: Triển khai các Inbound Ports.
  - Auth: `AuthUseCase` (GetMe — SSO profile via Prep).
  - Vocabulary: CQRS — `VocabularyCommand`, `VocabularyQuery`, `FolderCommand`, `FolderQuery`, `TopicQuery`, `OCRCommand`, `ImportCommand`.
- **DTOs (`dto/`)**: Data Transfer Objects với Gin binding tags cho validation (`required`, `email`, `min`, `max`).
- **Đặc điểm**: Chỉ phụ thuộc vào Domain Layer.

### Adapter Layer (`<module>/adapter/`)
Chứa các implementations cụ thể để kết nối Core với thế giới bên ngoài.
- **Handler (Driving)**: Nhận request từ bên ngoài. Bind JSON/Query → validate → gọi usecase qua inbound port. Validation errors trả field-level details qua `ValidationBadRequest()`.
- **Repository (Driven)**: GORM repositories implement outbound ports. Entity ↔ Model tách biệt với `toEntity()`/`fromEntity()`. Timestamps sync back sau Create/Save.
- **Service (Driven)**: External service adapters implement outbound ports.
  - Auth: `PrepUserService` — HTTP client gọi Prep API để validate token, có circuit breaker.
  - Vocabulary: `OCRService` — HTTP client gọi PaddleOCR, dùng `OCREngineRegistry` registry pattern (extensible cho Google Vision, Baidu).
- **Đặc điểm**: Phụ thuộc vào Application Layer (implement các Ports).

### Infrastructure Layer (`internal/infrastructure/`)
Cung cấp các công cụ và cấu hình để chạy ứng dụng.
- **Config**: Load biến môi trường qua Viper từ `.env`.
- **Database**: GORM postgres connection + custom GORM logger (slow query detection >200ms).
- **DI**: `container.go` → `initPersistence()` → module factories. Manual constructor injection.
- **Observability**: Zap structured logging + OTEL tracing + Sentry error tracking.
- **Resilience**: Circuit breaker (gobreaker v2) registry — used by PrepUserService và OCRService.

### Shared Kernel (`internal/shared/`)
Code dùng chung giữa các modules.
- **AppError**: Error codes layered: entity errors → `ErrInvalidInput`/`ErrNotFound` (usecase) → HTTP status + i18n key (handler via `response.HandleError()`).
- **Response**: `Success()`, `SuccessWithMetadata()` (pagination), `HandleError()` (map AppError.Code → HTTP status), `ValidationBadRequest()` (field-level validation errors).
- **Middleware**: SecurityHeaders, CORS, RequestID (UUID v7), OTEL (conditional), RequestLogger, Language (detect từ query/header), Recovery (panic → Sentry), RateLimit (public routes), Auth (JWT, protected routes).
- **Common**: JSONB helper type cho PostgreSQL JSON columns.
- **DTO**: `PaginationRequest`, `PaginationMeta`, `PaginatedResponse`.

---

## 3. Vòng đời của một API Request (Request Lifecycle)

**Ví dụ: Tạo từ vựng mới (Create Vocabulary)**

1.  **Client Request**:
    - Client gửi HTTP POST request tới `/api/v1/vocabularies` với JSON body (hanzi, pinyin, meaning).
    - Request đi kèm Header `Authorization: Bearer <token>`.

2.  **Infrastructure (Server)**:
    - `http.Server` nhận request.
    - Request đi qua **Router** (`gin.Engine`) và middleware chain.

3.  **Middleware Chain** (theo thứ tự thực thi):
    - `SecurityHeadersMiddleware` — set security headers (X-Content-Type-Options, etc.).
    - `CORSMiddleware` — xử lý CORS.
    - `RequestIDMiddleware` — gán UUID v7 request ID, propagate qua context.
    - `otelgin.Middleware` — OpenTelemetry tracing (conditional, khi OTLP_ENDPOINT được set).
    - `RequestLoggerMiddleware` — log request/response (skip sensitive paths).
    - `LanguageMiddleware` — detect ngôn ngữ từ query param > X-Lang header > Accept-Language header.
    - `RecoveryMiddleware` — catch panic, report Sentry.
    - `AuthMiddleware` — validate JWT token, set `user_id` vào context (chỉ áp dụng cho `/api/v1/*`).

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
    - **Database Execution**: GORM INSERT vào PostgreSQL.
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
[Middleware] SecurityHeaders → CORS → RequestID → OTEL → RequestLogger → Language → Recovery
   │                                                            (+ RateLimit on public, + Auth on /api/v1)
   ▼
[Adapter] Handler (ShouldBindJSON → ValidationBadRequest if fail)
   │         (DTO)
   ▼
[Application] Use Case (Business Logic, Error mapping)
   │         (Entity)
   ▼
[Adapter] Repository / Service (Entity ↔ Model mapping)
   │         (DB Model / HTTP Request)
   ▼
[Infrastructure] Database (PostgreSQL) / Redis / External APIs (Prep, PaddleOCR)
```
