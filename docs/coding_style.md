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
logger.WithContext(ctx).Error("[VOCABULARY] error saving folder", zap.Error(err))
logger.WithContext(ctx).Error("[OCR] service request failed", zap.Error(err))
logger.WithContext(ctx).Debug("[AUTH] rejected", zap.String("reason", "invalid token"))

// Bad — missing prefix
logger.WithContext(ctx).Error("error saving folder", zap.Error(err))
```

## Error handling

- Use **existing sentinel errors** (`ErrInternal`, `ErrServiceUnavailable`, `ErrNotFound`, etc.) — do NOT create new error instances with constructors.
- Sentinel error messages are **i18n keys** (e.g., `"common.internal_server_error"`). The `handleError` function passes `AppError.Message()` to the response layer for translation.
- **Status code must match the error**: DB unavailable → `ErrServiceUnavailable` (503), not `ErrInternal` (500). Invalid input → `ErrInvalidInput` (400), not `ErrInternal`.
- **Log the detail, return the appropriate error**: `logger.Error(...)` captures root cause for debugging. The returned error tells the client what category of problem it is.

**Examples:**
```go
// Good — log detail, return correct sentinel error
logger.WithContext(ctx).Error("[VOCABULARY] error saving folder", zap.Error(err))
return nil, sharederror.ErrInternal

// Good — service down → correct status 503
logger.WithContext(ctx).Error("[OCR] service request failed", zap.Error(err))
return nil, sharederror.ErrServiceUnavailable

// Bad — wrong status code (DB error is not "invalid input")
return nil, sharederror.ErrInvalidInput

// Bad — creating new error instance when sentinel exists
return nil, sharederror.NewInternal("some message", err)
```
