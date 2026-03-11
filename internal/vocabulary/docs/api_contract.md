# Vocabulary Module — API Contract

> **Trạng thái**: Sẵn sàng cho mobile tích hợp
> **Base URL**: `/api` (yêu cầu JWT auth)
> **Cập nhật**: 2026-03-18

---

## Tổng quan

Module Vocabulary quản lý từ vựng tiếng Trung, bao gồm:
- CRUD từ vựng (enriched với examples, radicals, stroke data)
- Quản lý chủ đề (topics) — 10 chủ đề hệ thống
- Điểm ngữ pháp (grammar points) liên kết với từ vựng
- OCR Scan — nhận diện từ mới vs từ đã có
- Bulk Import — nhập hàng loạt từ vựng HSK
- Thư mục (folders) — người dùng tự tổ chức từ vựng

---

## Chủ đề (Topics)

### `GET /api/topics`

Lấy danh sách 10 chủ đề hệ thống.

**Response** `200`:
```json
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "name_cn": "日常生活",
      "name_vi": "Cuộc sống hằng ngày",
      "name_en": "Daily Life",
      "slug": "daily-life"
    }
  ]
}
```

**10 chủ đề mặc định:**

| Slug | Tiếng Trung | Tiếng Việt | Tiếng Anh |
|------|-------------|------------|-----------|
| daily-life | 日常生活 | Cuộc sống hằng ngày | Daily Life |
| food-drink | 饮食 | Ẩm thực | Food & Drink |
| transportation | 交通出行 | Giao thông | Transportation |
| shopping | 购物 | Mua sắm | Shopping |
| work-career | 工作 | Công việc | Work & Career |
| education | 教育 | Giáo dục | Education |
| health | 健康 | Sức khỏe | Health |
| travel | 旅游 | Du lịch | Travel |
| culture | 文化 | Văn hóa | Culture |
| technology | 科技 | Khoa học công nghệ | Technology |

---

## Từ vựng (Vocabularies)

### `POST /api/vocabularies`

Tạo mới một từ vựng.

**Request**:
```json
{
  "hanzi": "你好",
  "pinyin": "nǐ hǎo",
  "meaning_vi": "Xin chào",
  "meaning_en": "Hello",
  "hsk_level": 1,
  "audio_url": "https://...",
  "examples": [
    {
      "sentence_cn": "你好，我是小明",
      "sentence_vi": "Xin chào, tôi là Tiểu Minh",
      "audio_url": "https://..."
    }
  ],
  "radicals": ["亻", "尔"],
  "stroke_count": 6,
  "stroke_data_url": "https://...",
  "recognition_only": false,
  "frequency_rank": 1
}
```

| Field | Kiểu | Bắt buộc | Mô tả |
|-------|------|----------|-------|
| hanzi | string | **Có** | Chữ Hán |
| pinyin | string | **Có** | Phiên âm pinyin |
| meaning_vi | string | Ít nhất 1 | Nghĩa tiếng Việt |
| meaning_en | string | Ít nhất 1 | Nghĩa tiếng Anh |
| hsk_level | int | **Có** | Cấp HSK (1-9) |
| audio_url | string | Không | URL file âm thanh |
| examples | array | Không | Danh sách câu ví dụ (JSONB) |
| radicals | string[] | Không | Bộ thủ của chữ |
| stroke_count | int | Không | Số nét |
| stroke_data_url | string | Không | URL dữ liệu nét viết |
| recognition_only | bool | Không | Chỉ cần nhận diện (không cần viết) |
| frequency_rank | int | Không | Thứ hạng tần suất sử dụng |

**Response** `201`: `VocabularyResponse`

### `GET /api/vocabularies/:id`

Lấy từ vựng theo ID (response cơ bản).

**Response** `200`: `VocabularyResponse`

### `GET /api/vocabularies/:id/detail`

Lấy từ vựng kèm đầy đủ chi tiết: topics và grammar points.

**Response** `200`:
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "hanzi": "你好",
    "pinyin": "nǐ hǎo",
    "meaning_vi": "Xin chào",
    "meaning_en": "Hello",
    "hsk_level": 1,
    "audio_url": "https://...",
    "examples": [
      {
        "sentence_cn": "你好，我是小明",
        "sentence_vi": "Xin chào, tôi là Tiểu Minh",
        "audio_url": "https://..."
      }
    ],
    "radicals": ["亻", "尔"],
    "stroke_count": 6,
    "stroke_data_url": "https://...",
    "recognition_only": false,
    "frequency_rank": 1,
    "created_at": "2026-03-18T...",
    "topics": [
      {
        "id": "uuid",
        "name_cn": "日常生活",
        "name_vi": "Cuộc sống hằng ngày",
        "name_en": "Daily Life",
        "slug": "daily-life"
      }
    ],
    "grammar_points": [
      {
        "id": "uuid",
        "code": "HSK1-GP-001",
        "pattern": "S + V + O",
        "example_cn": "我学中文",
        "example_vi": "Tôi học tiếng Trung",
        "rule": "Trật tự từ cơ bản SVO",
        "common_mistake": "...",
        "hsk_level": 1
      }
    ]
  }
}
```

### `GET /api/vocabularies/hsk/:level`

Liệt kê từ vựng theo cấp HSK (có phân trang).

**Query params**: `page` (mặc định 1), `page_size` (mặc định 10, tối đa 100)

**Response** `200`: Phân trang `VocabularyResponse[]`

### `GET /api/vocabularies/topic/:slug`

Liệt kê từ vựng theo chủ đề (có phân trang).

**Query params**: `page`, `page_size`

**Response** `200`: Phân trang `VocabularyResponse[]`

### `GET /api/vocabularies/search?q=...`

Tìm kiếm từ vựng theo hanzi, pinyin hoặc nghĩa.

**Query params**: `q` (bắt buộc), `page`, `page_size`

**Response** `200`: Phân trang `VocabularyResponse[]`

### `PUT /api/vocabularies/:id`

Cập nhật từ vựng. Hỗ trợ gán topics và grammar points.

**Request**: Giống create, thêm `topic_ids` và `grammar_point_ids` (mảng UUID string).

**Response** `200`: `VocabularyResponse`

### `DELETE /api/vocabularies/:id`

Xoá mềm (soft-delete) từ vựng.

**Response** `200`

---

## OCR Scan

### `POST /api/vocabularies/ocr-scan`

Xử lý kết quả quét OCR. Trả về danh sách từ mới (chưa có trong DB) và từ đã tồn tại.

**Request**:
```json
{
  "items": [
    { "hanzi": "你好" },
    { "hanzi": "学习" },
    { "hanzi": "新词" }
  ]
}
```

**Response** `200`:
```json
{
  "success": true,
  "data": {
    "new_items": [
      { "hanzi": "新词" }
    ],
    "existing_items": [
      {
        "id": "uuid",
        "hanzi": "你好",
        "pinyin": "nǐ hǎo",
        "meaning_vi": "Xin chào",
        "meaning_en": "Hello",
        "hsk_level": 1
      }
    ]
  }
}
```

**Luồng sử dụng trên mobile:**
1. Người dùng chụp ảnh sách/bài tập
2. Mobile OCR trích xuất danh sách hanzi
3. Gọi API này để phân loại từ mới vs từ đã học
4. Hiển thị cho người dùng chọn thêm từ mới vào thư mục

---

## Admin — Nhập hàng loạt

### `POST /api/admin/vocabularies/import`

Nhập hàng loạt từ vựng. Tự động bỏ qua từ trùng (theo hanzi).

**Request**:
```json
{
  "vocabularies": [
    {
      "hanzi": "你好",
      "pinyin": "nǐ hǎo",
      "meaning_vi": "Xin chào",
      "meaning_en": "Hello",
      "hsk_level": 1,
      "examples": [],
      "radicals": ["亻", "尔"],
      "stroke_count": 6,
      "recognition_only": false,
      "frequency_rank": 1
    }
  ]
}
```

**Response** `200`:
```json
{
  "success": true,
  "data": {
    "imported": 95,
    "skipped": 5,
    "total": 100
  }
}
```

| Field | Mô tả |
|-------|-------|
| imported | Số từ đã nhập thành công |
| skipped | Số từ bị bỏ qua (trùng hoặc invalid) |
| total | Tổng số từ trong request |

---

## Thư mục (Folders)

### `POST /api/folders`

Tạo thư mục mới cho người dùng.

**Request**: `{ "name": "HSK1 - Tuần 1", "description": "Từ vựng tuần đầu" }`

**Response** `201`: `FolderResponse`

### `GET /api/folders`

Liệt kê thư mục của người dùng (sắp xếp theo ngày tạo mới nhất).

**Response** `200`: `FolderResponse[]`

### `PUT /api/folders/:id`

Cập nhật thư mục (chỉ owner mới được).

**Response** `200`: `FolderResponse`

### `DELETE /api/folders/:id`

Xoá thư mục (chỉ owner).

**Response** `200`

### `POST /api/folders/:id/vocabularies`

Thêm từ vựng vào thư mục.

**Request**: `{ "vocabulary_id": "uuid" }`

**Response** `200`

### `DELETE /api/folders/:id/vocabularies/:vocab_id`

Xoá từ vựng khỏi thư mục.

**Response** `200`

### `GET /api/folders/:id/vocabularies`

Liệt kê từ vựng trong thư mục (phân trang, sắp xếp theo ngày thêm mới nhất).

**Query params**: `page`, `page_size`

**Response** `200`: Phân trang `VocabularyResponse[]`

---

## Định dạng Response chung

Tất cả response đều theo cấu trúc:
```json
{
  "success": true,
  "message": "i18n key hoặc message đã dịch",
  "data": { ... },
  "metadata": { "total": 100, "page": 1, "page_size": 10, "total_pages": 10 }
}
```

## Mã lỗi

| HTTP | Code | Ý nghĩa |
|------|------|---------|
| 400 | INVALID_INPUT | Dữ liệu đầu vào không hợp lệ |
| 401 | UNAUTHORIZED | Thiếu/sai JWT token |
| 404 | NOT_FOUND | Không tìm thấy tài nguyên |
| 500 | INTERNAL | Lỗi server |
| 503 | SERVICE_UNAVAILABLE | DB/service không khả dụng |
