# Coding Style

## Naming

- **No single-character receivers.** Use meaningful names: `profile`, `repo`, `handler` — not `p`, `r`, `h`.
- **No abbreviations** except: `ctx`, `err`, `db`, `cfg`, `id`, `req`, `res`, `tx`.

## Logging

All log messages MUST include module prefix: `[AUTH]`, `[VOCABULARY]`, `[OCR]`, `[SERVER]`.

```go
logger.WithContext(ctx).Warn("[VOCABULARY] error fetching topics", zap.Error(err))
```

- **Use cases**: `Warn`/`Info`/`Debug` only for non-error situations. Never `Error` — just create and return errors.
- **Handlers**: No logging. Just `response.HandleError(c, err)`. Request logger middleware auto-logs 5xx.

## Error handling

### Flow

```
Repo (raw error / nil) → Use case (wraps into AppError) → Handler (response.HandleError) → HTTP status + i18n
```

### Constructors

**4xx — no cause:**

| Constructor | When |
|---|---|
| `BadRequest(key)` 400 | Invalid input, malformed JSON, FK violation |
| `Unauthorized(key)` 401 | Auth failed or missing |
| `Forbidden(key)` 403 | Authenticated but not allowed |
| `NotFound(key)` 404 | Resource doesn't exist |
| `Conflict(key)` 409 | Duplicate / already exists |
| `UnprocessableEntity(key)` 422 | Validation failure, domain entity errors |

**5xx — always carry cause:**

| Constructor | When |
|---|---|
| `InternalServerError(key, err)` 500 | Unexpected system errors only |
| `ServiceUnavailable(key, err)` 503 | External service down, circuit breaker open |

Never use 500 as catch-all. FK violations → 400. Domain validation → 422.

### Key: i18n key, not English

```go
// Bad
apperr.InternalServerError("failed to save", err)
// Good
apperr.NotFound("vocabulary.not_found")
```

### Repos: raw errors only, no AppError

- **Not found** → `(nil, nil)`. Use case checks `if result == nil` → `apperr.NotFound(key)`.
- **Other errors** → raw error. Use case wraps → `apperr.InternalServerError(key, err)`.

### Domain validation errors

Domain uses `errors.New()` sentinels. Use cases map via mapper → `UnprocessableEntity(key)`.

### AppError methods

- `Error()` → `"NOT_FOUND: vocabulary.not_found"` (debug/logs)
- `Message()` → `"vocabulary.not_found"` (i18n key for response)
- `Unwrap()` → cause (5xx only)
