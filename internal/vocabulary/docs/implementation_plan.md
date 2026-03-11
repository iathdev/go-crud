# Vocabulary Module — Kế hoạch triển khai

> **Cập nhật**: 2026-03-18
> **Sprint 1**: Apr 1-14 — OCR Scan → Flashcard → Recall → Review
> **Deadline API contract**: Mar 28 (unblock mobile team)

---

## Bối cảnh

Module Vocabulary là module đầu tiên được triển khai cho Prep Chinese Vocab. Ban đầu module chỉ có CRUD cơ bản (Vocabulary + Folder). Cần mở rộng data model để hỗ trợ: examples, radicals, stroke data, grammar points, topics, OCR scan, HSK import.

## Cách tiếp cận: Hybrid (JSONB + Relational)

- **JSONB** cho dữ liệu gắn chặt vào từng word: `examples`, `radicals` (TEXT[])
- **Bảng quan hệ** cho shared entities: `topics` (10 chủ đề hệ thống), `grammar_points` (80 cho HSK 1-3)
- Giữ `meaning_vi`/`meaning_en` columns cho MVP (2 ngôn ngữ). Chuyển sang bảng `vocabulary_meanings` khi cần multi-language

### So sánh các phương án

| | Full JSONB | **Hybrid (đã chọn)** | Full Relational |
|---|---|---|---|
| Thời gian | Nhanh nhất | Vừa phải | Chậm nhất |
| Query cross-cutting | Khó, cần GIN index | Dễ (topics là table) | Dễ nhất |
| Cập nhật grammar toàn cục | Phải update mọi row | Dễ (grammar là table) | Dễ |
| Độ phức tạp | Thấp | Trung bình | Cao |
| Sẵn sàng multi-language | Kém | Tốt (topics có i18n columns) | Tốt nhất |

---

## Các phase triển khai

### Phase 0: API Contract + Docs (unblock mobile) — HOÀN THÀNH

- [x] Tạo `internal/vocabulary/docs/api_contract.md` — tài liệu API cho mobile team
- [x] Mở rộng DTOs: `ExampleDTO`, `TopicResponse`, `GrammarPointResponse`, `VocabularyDetailResponse`, `VocabularyListResponse`, `OCRScanRequest/Response`, `BulkImportRequest/Response`
- [x] Định nghĩa API endpoints mới:
  - `GET /api/topics` — Liệt kê 10 chủ đề hệ thống
  - `GET /api/vocabularies/topic/:slug` — Từ vựng theo chủ đề (phân trang)
  - `GET /api/vocabularies/:id/detail` — Chi tiết đầy đủ với examples, grammar
  - `POST /api/vocabularies/ocr-scan` — Xử lý kết quả OCR
  - `POST /api/admin/vocabularies/import` — Nhập hàng loạt từ vựng HSK (admin)

### Phase 1: Schema Migrations — HOÀN THÀNH

- [x] **Migration 00006** — Mở rộng bảng `vocabularies`:
  - Thêm: `examples JSONB`, `radicals TEXT[]`, `stroke_count INT`, `stroke_data_url VARCHAR`, `recognition_only BOOLEAN`, `frequency_rank INT`
  - Index: GIN cho examples, unique trên hanzi (WHERE deleted_at IS NULL)

- [x] **Migration 00007** — Tạo `topics` + `vocabulary_topics`:
  - Bảng `topics` (id, name_cn, name_vi, name_en, slug, sort_order)
  - Bảng join `vocabulary_topics`
  - Seed 10 chủ đề hệ thống

- [x] **Migration 00008** — Tạo `grammar_points` + `vocabulary_grammar_points`:
  - Bảng `grammar_points` (id, code, pattern, example_cn, example_vi, rule, common_mistake, hsk_level)
  - Bảng join `vocabulary_grammar_points`

- [x] **Migration 00009** — Xoá cột `topic VARCHAR` cũ khỏi `vocabularies`

### Phase 2: Domain Layer — HOÀN THÀNH

- [x] Mở rộng entity `Vocabulary`:
  - Thêm `Example` value object, các field mới (Radicals, StrokeCount, StrokeDataURL, RecognitionOnly, FrequencyRank)
  - Refactor constructor sang `VocabularyParams` struct (tránh 10+ tham số)
  - Thêm `NewVocabularyFromParams()`, `UpdateFromParams()`
  - Giữ `NewVocabulary()`/`Update()` cũ làm wrapper để backward compatible

- [x] Entity mới:
  - `domain/topic.go` — Topic entity (read-only, hệ thống định nghĩa)
  - `domain/grammar_point.go` — GrammarPoint entity

- [x] Domain errors: `ErrDuplicateHanzi`, `ErrTopicNotFound`, `ErrGrammarPointNotFound`

- [x] Cập nhật tests: thêm `TestNewVocabularyFromParams`, `TestVocabulary_UpdateFromParams`

### Phase 3: Ports & Adapters — HOÀN THÀNH

- [x] **Outbound ports mới:**
  - `TopicRepositoryPort`: FindAll, FindBySlug, FindByIDs, FindByVocabularyID
  - `GrammarPointRepositoryPort`: FindByVocabularyID, FindByHSKLevel, FindByCode, FindByIDs
  - Mở rộng `VocabularyRepositoryPort`: FindByHanzi, FindByHanziList, SaveBatch, FindByTopicID, CountByTopicID, SetTopics, SetGrammarPoints

- [x] **Inbound ports mới:**
  - `TopicQueryPort`: ListTopics
  - `OCRCommandPort`: ProcessOCRScan
  - `ImportCommandPort`: ImportVocabularies
  - Mở rộng `VocabularyQueryPort`: GetVocabularyDetail, ListByTopic

- [x] **Repository models:**
  - `VocabularyModel` cập nhật (JSONB/array columns, `ExamplesJSON` custom type)
  - Thêm `TopicModel`, `GrammarPointModel`, `VocabularyTopicModel`, `VocabularyGrammarPointModel`

- [x] **Repository implementations:**
  - `topic_postgres.go` — TopicRepository
  - `grammar_point_postgres.go` — GrammarPointRepository
  - `postgres.go` — Cập nhật VocabularyRepository với các method mới

- [x] **Use case implementations:**
  - `topic_query.go` — ListTopics
  - `ocr_command.go` — ProcessOCRScan (phân loại từ mới vs đã có)
  - `import_command.go` — ImportVocabularies (batch save, bỏ qua trùng)
  - `vocabulary_query.go` — Cập nhật GetVocabularyDetail, ListByTopic

### Phase 4: Handlers & Wiring — HOÀN THÀNH

- [x] Handler thêm 3 ports mới (topicQry, ocrCmd, importCmd) và 5 endpoint handlers
- [x] Module wiring tạo tất cả repos/use cases mới
- [x] Routes đăng ký đầy đủ

### Phase 5: HSK Data Seeding — CHƯA TRIỂN KHAI

- [ ] Seed script (`cmd/seed/main.go` hoặc Makefile target `make seed-hsk`)
- [ ] File JSON cho HSK 1-3 (~1000 từ)
- [ ] 80 grammar points cho HSK 1-3
- [ ] Gọi ImportCommandPort

---

## Cấu trúc file sau khi hoàn thành

```
internal/vocabulary/
├── docs/
│   ├── api_contract.md          ← Tài liệu API cho mobile
│   └── implementation_plan.md   ← File này
├── domain/
│   ├── vocabulary.go            ← Entity + Example value object (enriched)
│   ├── vocabulary_test.go       ← Unit tests
│   ├── folder.go                ← Folder entity
│   ├── topic.go                 ← Topic entity (NEW)
│   └── grammar_point.go         ← GrammarPoint entity (NEW)
├── application/
│   ├── port/
│   │   ├── inbound.go           ← Driving ports (enriched)
│   │   └── outbound.go          ← Driven ports (enriched)
│   ├── dto/
│   │   └── dto.go               ← DTOs (enriched)
│   └── usecase/
│       ├── vocabulary_command.go ← CRUD vocabulary (enriched)
│       ├── vocabulary_query.go   ← Query vocabulary (enriched)
│       ├── folder_command.go     ← CRUD folder
│       ├── folder_query.go       ← Query folder
│       ├── topic_query.go        ← Query topics (NEW)
│       ├── ocr_command.go        ← OCR scan (NEW)
│       └── import_command.go     ← Bulk import (NEW)
├── adapter/
│   ├── handler/
│   │   └── handler.go           ← HTTP handlers (enriched)
│   └── repository/
│       ├── model.go             ← GORM models (enriched)
│       ├── postgres.go          ← Vocab + Folder repos (enriched)
│       ├── topic_postgres.go    ← Topic repo (NEW)
│       └── grammar_point_postgres.go ← GrammarPoint repo (NEW)
└── module.go                    ← Wiring + routes (enriched)
```

## Migrations

```
internal/infrastructure/database/migration/
├── 00003_create_vocabularies_table.{up,down}.sql     (có sẵn)
├── 00004_create_folders_table.{up,down}.sql           (có sẵn)
├── 00005_create_learning_cards_table.{up,down}.sql    (có sẵn)
├── 00006_enrich_vocabularies_table.{up,down}.sql      (NEW)
├── 00007_create_topics_table.{up,down}.sql            (NEW)
├── 00008_create_grammar_points_table.{up,down}.sql    (NEW)
└── 00009_drop_topic_column_from_vocabularies.{up,down}.sql (NEW)
```

---

## Kiểm tra

1. `make migrate-up` — tất cả migrations chạy không lỗi
2. `go build ./...` — project compile thành công
3. `go test ./internal/vocabulary/...` — tất cả tests pass
4. Test thủ công: `GET /api/topics` trả về 10 chủ đề
5. Test thủ công: `POST /api/vocabularies` với enriched fields
6. Test thủ công: `GET /api/vocabularies/:id/detail` trả về chi tiết đầy đủ
7. Test thủ công: `POST /api/vocabularies/ocr-scan` phát hiện trùng lặp
8. Test thủ công: `POST /api/admin/vocabularies/import` nhập hàng loạt

---

## Việc còn lại (Phase 5 — Sprint 1)

1. **Seed data HSK**: Tạo file JSON + script seed cho ~1000 từ HSK 1-3
2. **Seed grammar points**: 80 grammar points cho HSK 1-3
3. **Admin middleware**: Phân quyền admin cho endpoint import (hiện tại chỉ cần JWT)
4. **OCR integration**: Kết nối với service OCR thực tế trên mobile
