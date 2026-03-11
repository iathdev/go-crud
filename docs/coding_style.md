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
