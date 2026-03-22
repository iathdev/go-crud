# Plan: Cài đặt PaddleOCR và tích hợp thử với Go project

## Context

Project đã có OCR plan chi tiết (`plan_ocr_engine.md`) nhắm tới Google Vision + Baidu OCR (cloud APIs). Tuy nhiên, muốn **thử PaddleOCR trước** — engine OCR open-source, miễn phí, chạy local. Mục tiêu: setup nhanh để evaluate accuracy trước khi commit vào cloud APIs.

Current state: endpoint `POST /api/vocabularies/ocr-scan` chỉ nhận danh sách hanzi đã extract sẵn, chưa nhận image.

### Alignment với `plan_ocr_engine.md`

Thiết kế Go side (ports, DTOs, use case) **đồng nhất** với plan_ocr_engine.md để sau này swap PaddleOCR → Google Vision / Baidu chỉ cần thay adapter, không sửa interface:

| Yếu tố | plan_ocr_engine.md | Implementation này |
|---|---|---|
| `OCRServicePort` interface | Đã định nghĩa (section 3.3) | **Dùng y nguyên** |
| Language codes | `"zh"`, `"vi"`, `"en"` | **Dùng y nguyên** (Python service map `"zh"` → PaddleOCR `"ch"`) |
| Response format | `new_items`, `existing_items`, `low_confidence_items`, `metadata` | **Dùng y nguyên** |
| Confidence thresholds | ≥80% confirmed, <70% low_confidence | **Dùng y nguyên** |
| Endpoint | `POST /api/vocabularies/ocr-scan` (replace, Alternative A) | Tạo `POST /api/vocabularies/ocr-image` (giữ cũ backward compat — **sau này rename khi chốt engine**) |
| Circuit breaker | gobreaker v2, cùng pattern `PrepUserService` | **Dùng y nguyên** |

### Multi-engine design (Tesseract future)

Python OCR service hỗ trợ **nhiều engine** qua `engine` parameter:
- `paddleocr` (default) — PaddleOCR PP-OCRv5
- `tesseract` — Tesseract OCR (cài thêm khi cần evaluate)

Go side **không cần thay đổi** khi thêm engine mới — chỉ truyền `engine` param xuống Python service.

Sau này khi chốt dùng Google Vision / Baidu (cloud):
- Tạo `GoogleVisionAdapter` / `BaiduOCRAdapter` implement cùng `OCRServicePort`
- Swap adapter trong DI container
- Python service retire

## Approach

Tạo một **Python HTTP service** wrap PaddleOCR (+ Tesseract), Go backend gọi qua HTTP. Phù hợp hexagonal architecture — PaddleOCR service là một adapter implement `OCRServicePort`.

---

## Steps

### Step 1: Tạo Python OCR service (`scripts/ocr-service/`)

Tạo thư mục `scripts/ocr-service/` với:

- **`requirements.txt`** — `paddlepaddle`, `paddleocr`, `fastapi`, `uvicorn`, `python-multipart`
- **`main.py`** — FastAPI app với endpoint:
  ```
  POST /recognize
  Content-Type: application/json
  Body: {
    "image": "<base64-encoded image>",
    "language": "zh",
    "engine": "paddleocr"
  }

  Response: {
    "characters": [
      {"text": "你好", "confidence": 0.95, "candidates": []},
      ...
    ],
    "engine": "paddleocr"
  }
  ```
  - Nhận base64 image → decode → chạy engine → trả JSON với text + confidence per line
  - Filter chỉ giữ CJK characters (U+4E00–U+9FFF) nếu language = "zh"
  - Map language codes: API dùng `"zh"` → PaddleOCR dùng `"ch"` (internal mapping)
  - Tesseract engine: optional, chỉ hoạt động nếu `pytesseract` đã cài

### Step 2: Thêm `OCRServicePort` interface vào Go project

File: `internal/vocabulary/application/port/outbound.go`

```go
// Matches plan_ocr_engine.md section 3.3 exactly
type OCRServicePort interface {
    ExtractCharacters(ctx context.Context, req OCRExtractRequest) (*OCRExtractResult, error)
}

type OCRExtractRequest struct {
    Image    []byte
    Type     string  // "printed" | "handwritten" | "auto"
    Language string  // "zh" | "vi" | "en"
}

type OCRExtractResult struct {
    Characters []OCRCharacter
    EngineUsed string  // "paddleocr" | "tesseract" | "google_vision" | "baidu_ocr"
}

type OCRCharacter struct {
    Text       string
    Confidence float64
    Candidates []string
}
```

### Step 3: Tạo OCR Service adapter

File: `internal/vocabulary/adapter/service/ocr_service.go`

- Implement `OCRServicePort`
- Gọi HTTP POST tới Python service (`OCR_SERVICE_URL/recognize`) với **JSON body** chứa base64 image
- Request format: `{"image": "<base64>", "language": "zh"}` — đồng nhất với Google Vision / Baidu OCR APIs
- Parse JSON response → map sang `OCRExtractResult`
- Dùng circuit breaker từ `infrastructure/circuitbreaker/` (cùng pattern auth module `PrepUserService`)

### Step 4: Update `OCRCommand` use case

File: `internal/vocabulary/application/usecase/ocr_command.go`

- Thêm `OCRServicePort` dependency
- Thêm method mới `ProcessOCRImage()`:
  - Nhận image bytes → gọi `OCRServicePort.ExtractCharacters()`
  - Classify by confidence (matching plan_ocr_engine.md section 5.1):
    - ≥ 80% → confirmed → check new/existing qua `VocabularyRepositoryPort.FindByHanziList()`
    - < 70% → low_confidence (trả kèm candidates)
  - Trả response với metadata (engine_used, total_detected, processing_time_ms)
- Giữ nguyên `ProcessOCRScan()` cũ (backward compatible)

### Step 5: Update DTOs

File: `internal/vocabulary/application/dto/dto.go`

- Thêm `OCRImageHTTPRequest` (JSON binding: `image_url`, type, language) — handler input
- Thêm `OCRImageRequest` (image bytes, type, language) — internal use case input
- Thêm `OCRImageResponse` matching plan_ocr_engine.md response format:
  - `new_items`, `existing_items`, `low_confidence_items`
  - `metadata` (engine_used, total_detected, processing_time_ms)

### Step 6: Update handler

File: `internal/vocabulary/adapter/handler/handler.go`

- Thêm handler method `ProcessOCRImage()` — nhận **JSON body** với `image_url` (URL ảnh)
- BE download ảnh từ URL → validate size ≤ 5MB, format JPEG/PNG → truyền bytes cho use case
- Use case convert sang base64 rồi gửi cho OCR service (đồng nhất Google Vision / Baidu)
- Giữ nguyên handler `ProcessOCRScan()` cũ

### Step 7: Thêm inbound port

File: `internal/vocabulary/application/port/inbound.go`

- Thêm method `ProcessOCRImage` vào `OCRCommandPort` interface

### Step 8: Wire everything

- `internal/vocabulary/module.go` — inject `OCRServicePort` vào `OCRCommand`
- `internal/infrastructure/di/container.go` — tạo `OCRService` với circuit breaker + config URL
- `internal/infrastructure/config/ocr.go` — thêm `OCRConfig` struct
- `.env.example` — thêm `OCR_SERVICE_URL=http://localhost:8000`
- `internal/vocabulary/module.go` — thêm route `POST /api/vocabularies/ocr-image`

---

## Files to create

- `scripts/ocr-service/requirements.txt`
- `scripts/ocr-service/main.py`
- `scripts/ocr-service/README.md`
- `internal/vocabulary/adapter/service/ocr_service.go`
- `internal/infrastructure/config/ocr.go`

## Files to modify

- `internal/vocabulary/application/port/outbound.go` — thêm `OCRServicePort` interface
- `internal/vocabulary/application/port/inbound.go` — thêm `ProcessOCRImage` method
- `internal/vocabulary/application/dto/dto.go` — thêm DTOs mới
- `internal/vocabulary/application/usecase/ocr_command.go` — thêm `OCRServicePort` dep + `ProcessOCRImage()`
- `internal/vocabulary/adapter/handler/handler.go` — thêm multipart handler
- `internal/vocabulary/module.go` — wire adapter + route
- `internal/infrastructure/di/container.go` — tạo adapter
- `internal/infrastructure/config/config.go` — embed `OCRConfig`
- `.env.example` — thêm config

---

## Chạy Python OCR Service

### Cài đặt

```bash
cd scripts/ocr-service

# Tạo virtual environment
python3 -m venv .venv
source .venv/bin/activate

# Cài PaddleOCR (default engine)
pip install -r requirements.txt

# (Optional) Cài Tesseract engine
# macOS:
brew install tesseract tesseract-lang
pip install pytesseract

# Ubuntu:
# sudo apt install tesseract-ocr tesseract-ocr-chi-sim
# pip install pytesseract
```

### Chạy service

```bash
cd scripts/ocr-service
source .venv/bin/activate
uvicorn main:app --host 0.0.0.0 --port 8000

# Hoặc với auto-reload (dev):
uvicorn main:app --host 0.0.0.0 --port 8000 --reload
```

Service chạy tại `http://localhost:8000`. Swagger docs tại `http://localhost:8000/docs`.

### Test trực tiếp với Python service

```bash
# Test PaddleOCR (default)
curl -X POST http://localhost:8000/recognize \
  -H "Content-Type: application/json" \
  -d '{"image": "'$(base64 < test_image.jpg)'", "language": "zh"}'

# Test PaddleOCR với tiếng Anh
curl -X POST http://localhost:8000/recognize \
  -H "Content-Type: application/json" \
  -d '{"image": "'$(base64 < english_text.png)'", "language": "en"}'

# Test Tesseract engine (nếu đã cài)
curl -X POST http://localhost:8000/recognize \
  -H "Content-Type: application/json" \
  -d '{"image": "'$(base64 < test_image.jpg)'", "language": "zh", "engine": "tesseract"}'

# Health check
curl http://localhost:8000/health
```

**Response mẫu:**
```json
{
  "characters": [
    {"text": "你好", "confidence": 0.9521, "candidates": []},
    {"text": "世界", "confidence": 0.9234, "candidates": []}
  ],
  "engine": "paddleocr"
}
```

### Test qua Go backend (full integration)

```bash
# 1. Start Python OCR service (terminal 1)
cd scripts/ocr-service && source .venv/bin/activate && uvicorn main:app --port 8000

# 2. Start Go server (terminal 2) — cần .env có OCR_SERVICE_URL=http://localhost:8000
make run

# 3. Test endpoint (terminal 3) — cần JWT token
curl -X POST http://localhost:8001/api/vocabularies/ocr-image \
  -H "Authorization: Bearer <JWT_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"image_url": "https://example.com/test_image.jpg", "type": "auto", "language": "zh"}'
```

**Response mẫu:**
```json
{
  "success": true,
  "data": {
    "new_items": [
      {"hanzi": "新词", "confidence": 0.92, "candidates": []}
    ],
    "existing_items": [
      {"id": "uuid", "hanzi": "你好", "pinyin": "nǐ hǎo", "meaning_vi": "Xin chào", "confidence": 0.98, "candidates": []}
    ],
    "low_confidence_items": [
      {"hanzi": "鑫", "confidence": 0.55, "candidates": ["鑫", "森", "淼"]}
    ],
    "metadata": {
      "engine_used": "paddleocr",
      "total_detected": 15,
      "processing_time_ms": 1234
    }
  }
}
```

---

## Verification

1. **Python service**: `cd scripts/ocr-service && source .venv/bin/activate && uvicorn main:app` → test với curl commands ở trên
2. **Go build**: `make build` — verify compile thành công
3. **Existing tests**: `go test ./...` — verify không break gì
4. **Integration test**: Start Python service → start Go server → test với curl qua Go endpoint
5. **So sánh engines**: Test cùng ảnh với `engine=paddleocr` vs `engine=tesseract` → so accuracy

---

## Migration path → Cloud APIs

Khi chốt dùng Google Vision / Baidu (theo `plan_ocr_engine.md`):

1. Tạo `GoogleVisionAdapter` implement cùng `OCRServicePort` → swap trong DI
2. Tạo `BaiduOCRAdapter` implement cùng `OCRServicePort` → swap trong DI
3. Thêm routing logic trong use case (type + language → chọn engine)
4. Retire Python service
5. Rename endpoint `ocr-image` → `ocr-scan` (replace current JSON endpoint)

**Không cần sửa**: ports, DTOs, use case logic, handler, response format.
