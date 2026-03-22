# Coding Style

## Naming — No abbreviations, must be immediately readable

- **Receiver name**: Use meaningful names, NEVER use single-character abbreviations (`p`, `u`, `v`, `m`, `r`, `s`...).
  - Bad: `func (p *UserProfile) CompleteOnboarding(...)`
  - Good: `func (profile *UserProfile) CompleteOnboarding(...)`
  - Good: `func (repo *VocabularyRepository) FindByID(...)`
  - Good: `func (handler *VocabularyHandler) CreateVocabulary(...)`
- **Variables and parameters**: Names must be self-explanatory, no abbreviations unless it's a widely accepted convention.
  - Bad: `v`, `f`, `gp`, `m`
  - Good: `vocab`, `folder`, `grammarPoint`, `model`
- **Allowed abbreviations**: `ctx`, `err`, `db`, `cfg`, `id`, `req`, `res`, `tx` (transaction).

## Logging — Module prefix required

- All log messages MUST include a module prefix in uppercase brackets: `[MODULE_NAME]`.
- Prefix is based on the module the code belongs to, not the package name.
- This applies to all log levels (Error, Warn, Info, Debug).

| Module | Prefix |
|--------|--------|
| Auth | `[AUTH]` |
| Vocabulary | `[VOCABULARY]` |
| OCR | `[OCR]` |
| Server / Middleware | `[SERVER]` |

**Examples:**
```go
// Good
logger.WithContext(ctx).Warn("[VOCABULARY] error fetching topics", zap.Error(err))
logger.WithContext(ctx).Debug("[AUTH] rejected", zap.String("reason", "invalid token"))

// Bad — missing prefix
logger.WithContext(ctx).Error("error saving folder", zap.Error(err))
```

## Error handling — Correct status code first, descriptive i18n key second

### Constructors are pure — logging happens at the handler layer

Error constructors (`InternalServerError`, `ServiceUnavailable`, etc.) are **pure functions** with no side effects. They do NOT log. Logging for server errors (500/503) happens once in the handler's `handleError` function, which:
- Reads `domErr.Message()` for the i18n key
- Reads `domErr.Unwrap()` for the cause
- Logs with `[MODULE]` prefix + i18n key + cause

This means:
- **Use case layer**: Only creates and returns errors. No `logger.Error(...)` calls for error returns.
- **Handler layer**: Single point of logging for all server errors. No duplicate logs possible.
- **Tests**: Error creation has no side effects. No log noise.

### Choosing the right constructor

- `BadRequest` (400) — Client sent bad data: invalid UUID, malformed JSON, FK violation (referencing non-existent ID)
- `Unauthorized` (401) — Authentication failed or missing
- `Forbidden` (403) — Authenticated but not allowed
- `NotFound` (404) — The requested resource does not exist
- `Conflict` (409) — Resource state conflict (duplicate, already exists)
- `UnprocessableEntity` (422) — Request well-formed but fails validation (missing required field, invalid value range, etc.)
- `InternalServerError` (500) — **Only** for truly unexpected system errors: DB connection lost, unhandled panic, unknown errors after all known types are filtered out
- `ServiceUnavailable` (503) — External service is down (OCR service, SSO, circuit breaker open, etc.)

**Never use `InternalServerError` as a default catch-all.** Before returning 500, ask: "Is this really a server-side system failure, or is it caused by bad client input?" FK violations, missing references, constraint errors → 400, not 500.

### Error message must be an i18n key

Never use hardcoded English strings.
- Bad: `sharederror.InternalServerError("failed to save vocabulary", err)`
- Good: `sharederror.NotFound("vocabulary.not_found")`
- Good: `sharederror.BadRequest("vocabulary.invalid_topic_id")`

### Always check `IsAppError` before falling back to `InternalServerError`

Repos may return typed AppErrors (not-found, bad-request from FK violations). The use case must propagate those, not swallow them into 500:

```go
if err := repo.Save(ctx, entity); err != nil {
    if _, ok := sharederror.IsAppError(err); ok {
        return err  // propagate 400/404/etc. from repo
    }
    return sharederror.InternalServerError("entity.save_failed", err)
}
```

### No sentinels — only constructors

No sentinel variables (`ErrNotFound`, etc.). Always use constructors with i18n keys:
- **Client errors (4xx)**: `BadRequest(key)`, `Unauthorized(key)`, `Forbidden(key)`, `NotFound(key)`, `Conflict(key)`, `UnprocessableEntity(key)` — no cause needed.
- **Server errors (5xx)**: `InternalServerError(key, cause)`, `ServiceUnavailable(key, cause)` — always carry cause for handler-layer logging.

### `Error()` vs `Message()`

- `Error()` returns `"CODE: i18n_key"` — for debug/logs, does NOT leak cause details.
- `Message()` returns just the i18n key — for API response translation.
- `Unwrap()` returns the underlying cause — for handler-layer logging.

### Manual logging in use cases

Only use `logger.WithContext(ctx).Warn(...)` or `.Info(...)` / `.Debug(...)` for **non-critical, non-error situations** (e.g., fetching optional related data fails but we continue). Always include `[MODULE]` prefix. Do NOT use `logger.Error(...)` in use cases — server error logging is handled by the handler layer.
