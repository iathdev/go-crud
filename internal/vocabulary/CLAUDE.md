# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in the vocabulary module.

## Module Overview

CRUD for Chinese vocabulary entries (Hanzi) and user-created folders for organizing them. Vocabularies are **global shared data** (no user ownership); folders are **user-scoped**. Uses CQRS with separate Command/Query use cases.

## Key Domain Concepts

- **Vocabulary**: Hanzi + Pinyin + meanings (VI and/or EN) + HSK level (1-9) + optional topic/audio. Validated at construction — at least one meaning required.
- **Folder**: User-owned collection of vocabularies. Many-to-many via `folder_vocabularies` junction table.
- **Ownership model**: Vocabularies are global; folders belong to a user. All folder operations verify `folder.UserID == request user_id` via `getOwnedFolder()` helper.

## CQRS Structure

```
VocabularyCommand  →  CreateVocabulary, UpdateVocabulary, DeleteVocabulary
VocabularyQuery    →  GetVocabulary, ListByHSKLevel, SearchVocabulary
FolderCommand      →  CreateFolder, UpdateFolder, DeleteFolder, AddVocabulary, RemoveVocabulary
FolderQuery        →  ListFolders, ListVocabularies
```

`FolderCommand` depends on both `FolderRepositoryPort` and `VocabularyRepositoryPort` (verifies vocab exists before linking).

## Routes (all protected)

```
POST/GET/PUT/DELETE  /vocabularies[/:id]
GET                  /vocabularies/hsk/:level    (paginated, ordered by pinyin ASC)
GET                  /vocabularies/search?q=...  (LIKE on hanzi/pinyin/meaning_vi/meaning_en)
POST/GET/PUT/DELETE  /folders[/:id]
POST/GET/DELETE      /folders/:id/vocabularies[/:vocab_id]
```

## Module-Specific Patterns

- **Search**: Case-insensitive LIKE with `%query%` wildcards across 4 columns. Ordered by HSK level ASC, pinyin ASC.
- **Pagination**: Normalized in `normalizePagination()` — defaults page=1, pageSize=10, max 100. Dual queries (COUNT + SELECT) for paginated endpoints.
- **Shared helpers**: `getOwnedFolder()`, `classifyRepoError()`, `normalizePagination()`, `toFolderResponse()` are package-level functions shared across Command/Query use cases.
- **Folder-vocabulary junction**: `folder_vocabularies` table with composite PK `(folder_id, vocabulary_id)` + `added_at` timestamp. No domain entity for this — handled purely at repository level.
- **Domain error mapping**: `mapVocabEntityError()` converts domain errors (ErrHanziRequired, etc.) to `AppError` with `CodeInvalidInput`.

## Database Tables

- `vocabularies` — indexes on `hsk_level`, `topic`, `deleted_at`
- `folders` — index on `user_id`, `deleted_at`
- `folder_vocabularies` — composite PK `(folder_id, vocabulary_id)`
