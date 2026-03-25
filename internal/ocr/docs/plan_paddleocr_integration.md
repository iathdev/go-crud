# Plan: Cài đặt PaddleOCR và tích hợp thử với Go project

## Context

Project đã có OCR plan chi tiết (`plan_ocr_engine.md`) nhắm tới Google Vision + Baidu OCR (cloud APIs). Trước khi commit vào cloud APIs, **dùng PaddleOCR trước** — engine OCR open-source, miễn phí, chạy local — để evaluate accuracy.

### Alignment với `plan_ocr_engine.md`

Thiết kế Go side (ports, DTOs, use case) **đồng nhất** với plan_ocr_engine.md. Sau này swap PaddleOCR → Google Vision / Baidu chỉ cần thêm adapter + đăng ký vào engine registry, không sửa interface:

| Yếu tố | plan_ocr_engine.md | Implementation hiện tại |
|---|---|---|
| `OCRServicePort` interface | `ExtractCharacters()` (section 3.3) | `Recognize()` — cùng semantics, đổi tên cho gọn |
| Multi-engine | Google Vision + Baidu OCR | `OCREngineRegistry` map — PaddleOCR đã wire, Google/Baidu thêm sau |
| Engine routing | User-specified + cascading fallback | `resolveEngine()` trong use case — routing theo `type` + `language` |
| Language codes | `"zh"`, `"vi"`, `"en"` | **Dùng y nguyên** (Python service map `"zh"` → PaddleOCR `"ch"`) |
| Response format | `new_items`, `existing_items`, `low_confidence_items`, `metadata` | **Dùng y nguyên** |
| Confidence thresholds | ≥80% confirmed, <70% low_confidence | **Dùng y nguyên** (constants trong `ocr_command.go`) |
| Endpoint | `POST /api/vocabularies/ocr-scan` (Alternative A — single endpoint nhận image) | **Dùng y nguyên** — protected route, single endpoint |
| Circuit breaker | gobreaker v2 | **Dùng y nguyên** — wrap OCR adapter |

### Multi-engine design

**Go side:** `OCREngineRegistry` (`map[OCREngineKey]OCRServicePort`) cho phép đăng ký nhiều engine. Use case `resolveEngine()` chọn engine theo routing rules:

- `printed` (any lang) → `google_vision`
- `handwritten` + `zh` → `baidu_ocr`
- `handwritten` + other → `google_vision`
- `auto` → `google_vision` (primary)
- Fallback: dùng engine nào có sẵn trong registry

Hiện tại chỉ PaddleOCR được đăng ký → mọi request đều route tới PaddleOCR (fallback behavior).

**Python side:** Python OCR service hỗ trợ nhiều engine qua `engine` parameter:
- `paddleocr` (default) — PaddleOCR PP-OCRv5
- `tesseract` — Tesseract OCR (cài thêm khi cần evaluate)

Khi chốt dùng Google Vision / Baidu (cloud):
- Tạo `GoogleVisionAdapter` / `BaiduOCRAdapter` implement cùng `OCRServicePort`
- Đăng ký vào `OCREngineRegistry` trong DI container
- `resolveEngine()` tự route đúng engine
- Python service retire

---

## Implementation (đã hoàn thành)

### Step 1: Python OCR service (`scripts/ocr-service/`)

**`requirements.txt`:**
```
paddlepaddle
paddleocr
fastapi
uvicorn[standard]
pydantic
jieba
pypinyin
urllib3<2
```

**`main.py`** — FastAPI app:
- Endpoint: `POST /recognize`
- Request: `{"image": "<base64>", "language": "zh", "engine": "paddleocr"}`
- Response: `{"characters": [{"text": "你好", "pinyin": "nǐ hǎo", "confidence": 0.95, "candidates": []}], "engine": "paddleocr"}`
- Chinese processing: jieba word segmentation → CJK filter (U+4E00–U+9FFF) → pypinyin conversion
- Non-Chinese: trả raw text, pinyin rỗng
- Health check: `GET /health`

### Step 2: Output port — `OCRServicePort`

File: `internal/vocabulary/application/port/outbound.go`

```go
type OCRServicePort interface {
    Recognize(ctx context.Context, req OCRRequest) (*OCRResult, error)
}

type OCRRequest struct {
    Image    []byte
    Language string // "zh" | "vi" | "en"
}

type OCRResult struct {
    Characters []OCRCharacter
    Engine     string // "paddleocr" | "google_vision" | "baidu_ocr"
}

type OCRCharacter struct {
    Text       string
    Pinyin     string
    Confidence float64
    Candidates []string
}

type OCREngineKey string

const (
    OCREnginePaddleOCR    OCREngineKey = "paddleocr"
    OCREngineGoogleVision OCREngineKey = "google_vision"
    OCREngineBaiduOCR     OCREngineKey = "baidu_ocr"
)

type OCREngineRegistry map[OCREngineKey]OCRServicePort
```

### Step 3: OCR Service adapter

File: `internal/vocabulary/adapter/service/ocr_service.go`

- Implement `OCRServicePort` với method `Recognize()`
- Encode image → base64 → gọi `POST {OCR_SERVICE_URL}/recognize`
- Parse JSON response → map sang `port.OCRResult` (bao gồm `Pinyin`)
- Wrap trong circuit breaker (`infrastructure/circuitbreaker/`)

### Step 4: Input port — `OCRCommandPort`

File: `internal/vocabulary/application/port/inbound.go`

```go
type OCRCommandPort interface {
    ProcessOCRScan(ctx context.Context, req vdto.OCRScanRequest) (*vdto.OCRScanResponse, error)
}
```

Single method — nhận image bytes, trả kết quả đã classify.

### Step 5: DTOs

File: `internal/vocabulary/application/dto/dto.go`

```go
// Handler input (JSON binding)
type OCRScanHTTPRequest struct {
    ImageURL string `json:"image_url" binding:"required,url"`
    Type     string `json:"type" binding:"omitempty,oneof=printed handwritten auto"`
    Language string `json:"language" binding:"omitempty,oneof=zh vi en"`
}

// Use case input (image bytes)
type OCRScanRequest struct {
    Image    []byte
    Type     string // "printed" | "handwritten" | "auto"
    Language string // "zh" | "vi" | "en"
}

// Response types
type OCRScanCharacterItem struct {
    Hanzi      string   `json:"hanzi"`
    Pinyin     string   `json:"pinyin"`
    Confidence float64  `json:"confidence"`
    Candidates []string `json:"candidates,omitempty"`
}

type OCRScanExistingItem struct {
    VocabularyListResponse
    Confidence float64  `json:"confidence"`
    Candidates []string `json:"candidates,omitempty"`
}

type OCRScanMetadata struct {
    EngineUsed       string `json:"engine_used"`
    TotalDetected    int    `json:"total_detected"`
    ProcessingTimeMs int64  `json:"processing_time_ms"`
}

type OCRScanResponse struct {
    NewItems           []OCRScanCharacterItem `json:"new_items"`
    ExistingItems      []OCRScanExistingItem  `json:"existing_items"`
    LowConfidenceItems []OCRScanCharacterItem `json:"low_confidence_items"`
    Metadata           OCRScanMetadata        `json:"metadata"`
}
```

### Step 6: Use case — `OCRCommand`

File: `internal/vocabulary/application/usecase/ocr_command.go`

- Dependencies: `VocabularyRepositoryPort` + `OCREngineRegistry`
- `resolveEngine(type, language)` — routing logic theo plan_ocr_engine.md section 5.1
- `ProcessOCRScan()`:
  1. Resolve engine theo type + language
  2. Gọi `engine.Recognize()` với image bytes
  3. Classify theo confidence: ≥0.70 → confirmed, <0.70 → low_confidence
  4. Check confirmed items against DB (`FindByHanziList`)
  5. Split: new_items / existing_items / low_confidence_items
  6. Trả response kèm metadata (engine, total_detected, processing_time_ms)

### Step 7: Handler

File: `internal/vocabulary/adapter/handler/handler.go`

- `ProcessOCRScan(c *gin.Context)`:
  1. Parse `OCRScanHTTPRequest` từ JSON body
  2. Download image từ URL (max 5MB, chỉ JPEG/PNG)
  3. Default: type="auto", language="zh"
  4. Gọi `ocrCmd.ProcessOCRScan()`

### Step 8: Wiring

- **`internal/vocabulary/module.go`**: `NewModule(db, ocrEngines OCREngineRegistry)` → inject vào `OCRCommand` → register route `protected.POST("/vocabularies/ocr-scan", ...)`
- **`internal/infrastructure/di/container.go`**: Tạo circuit breaker → tạo `OCRService` adapter → đăng ký vào `OCREngineRegistry{OCREnginePaddleOCR: ocrAdapter}` → truyền vào module
- **`internal/infrastructure/config/ocr.go`**: `OCRConfig{OCRServiceURL}` load từ env
- **`.env.example`**: `OCR_SERVICE_URL=http://localhost:8000`

---

## Files đã tạo

- `scripts/ocr-service/requirements.txt`
- `scripts/ocr-service/main.py`
- `internal/vocabulary/adapter/service/ocr_service.go`
- `internal/infrastructure/config/ocr.go`

## Files đã sửa

- `internal/vocabulary/application/port/outbound.go` — `OCRServicePort`, `OCREngineRegistry`
- `internal/vocabulary/application/port/inbound.go` — `OCRCommandPort` với single method `ProcessOCRScan`
- `internal/vocabulary/application/dto/dto.go` — `OCRScan*` DTOs
- `internal/vocabulary/application/usecase/ocr_command.go` — `OCRCommand` với engine registry + routing
- `internal/vocabulary/adapter/handler/handler.go` — `ProcessOCRScan` handler + `downloadImage` helper
- `internal/vocabulary/module.go` — nhận `OCREngineRegistry`, wire route
- `internal/infrastructure/di/container.go` — tạo adapter + registry
- `internal/infrastructure/config/config.go` — embed `OCRConfig`
- `.env.example` — thêm `OCR_SERVICE_URL`

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
pip install pytesseract Pillow

# Ubuntu:
# sudo apt install tesseract-ocr tesseract-ocr-chi-sim
# pip install pytesseract Pillow
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
    {"text": "你好", "pinyin": "nǐ hǎo", "confidence": 0.9521, "candidates": []},
    {"text": "世界", "pinyin": "shì jiè", "confidence": 0.9234, "candidates": []}
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
curl -X POST http://localhost:8001/api/vocabularies/ocr-scan \
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
      {"hanzi": "新词", "pinyin": "xīn cí", "confidence": 0.92, "candidates": []}
    ],
    "existing_items": [
      {"id": "uuid", "hanzi": "你好", "pinyin": "nǐ hǎo", "meaning_vi": "Xin chào", "meaning_en": "Hello", "hsk_level": 1, "confidence": 0.98, "candidates": []}
    ],
    "low_confidence_items": [
      {"hanzi": "鑫", "pinyin": "xīn", "confidence": 0.55, "candidates": ["鑫", "森", "淼"]}
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

1. Tạo `GoogleVisionAdapter` implement `OCRServicePort` (method `Recognize`)
2. Tạo `BaiduOCRAdapter` implement `OCRServicePort` (method `Recognize`)
3. Đăng ký vào `OCREngineRegistry` trong DI container:
   ```go
   ocrEngines := vocabport.OCREngineRegistry{
       vocabport.OCREngineGoogleVision: googleAdapter,
       vocabport.OCREngineBaiduOCR:     baiduAdapter,
   }
   ```
4. `resolveEngine()` trong use case tự route đúng engine — **không cần sửa logic**
5. Retire Python service + PaddleOCR adapter

**Không cần sửa**: ports, DTOs, use case logic, handler, response format, routing logic.
