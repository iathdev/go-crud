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

## Error handling — Correct status code first, descriptive i18n key second

- **Choose the right error constructor based on the actual cause**, not as a catch-all:
  - `NewInvalidInput` (400) — Client sent bad data: invalid UUID, failed validation, FK violation (referencing non-existent ID)
  - `NewNotFound` (404) — The requested resource does not exist
  - `NewUnauthorized` (401) — Authentication failed or missing
  - `NewServiceUnavailable` (503) — External service is down (OCR service, SSO, etc.)
  - `NewInternal` (500) — **Only** for truly unexpected system errors: DB connection lost, unhandled panic, unknown errors after all known types are filtered out
- **Never use `NewInternal` as a default catch-all.** Before returning 500, ask: "Is this really a server-side system failure, or is it caused by bad client input?" FK violations, missing references, constraint errors → 400, not 500.
- **Error message must be an i18n key** that clearly describes the error. Never use hardcoded English strings.
  - Bad: `sharederror.NewInternal("failed to save vocabulary", err)` (hardcoded English)
  - Good: `sharederror.NewNotFound("vocabulary.not_found")`
  - Good: `sharederror.NewInvalidInput("vocabulary.invalid_topic_id")`
- **Always check `IsAppError` before falling back to `NewInternal`.** Repos may return typed AppErrors (not-found, invalid input from FK violations). The use case must propagate those, not swallow them into 500:
  ```go
  if err := repo.Save(ctx, entity); err != nil {
      if _, ok := sharederror.IsAppError(err); ok {
          return err  // propagate 400/404/etc. from repo
      }
      return sharederror.NewInternal(ctx, "entity.save_failed", err)
  }
  ```
- **`NewInternal` and `NewServiceUnavailable` auto-log.** They accept `ctx` and log the i18n key + cause error automatically. Do NOT add a separate `logger.WithContext(ctx).Error(...)` call before them — it would duplicate the log.
  - Bad (duplicated log):
    ```go
    logger.WithContext(ctx).Error("[VOCABULARY] error saving", zap.Error(err))
    return sharederror.NewInternal(ctx, "vocabulary.save_failed", err)
    ```
  - Good (single call, auto-logged):
    ```go
    return sharederror.NewInternal(ctx, "vocabulary.save_failed", err)
    ```
- **Logs stay in English** for developer debugging. i18n keys are only for the API response message (translated at the response layer). For `Warn`/`Info`/`Debug` logs (not tied to error constructors), use `[MODULE]` prefix manually.
