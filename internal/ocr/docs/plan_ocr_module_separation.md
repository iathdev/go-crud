# Plan: Tách OCR thành module riêng + Multilang response

## Context

OCR code hiện nằm trong vocabulary module. Cần tách ra `internal/ocr/` để:
1. OCR chỉ lo **recognize text từ ảnh** — không biết vocabulary tồn tại
2. Vocabulary module orchestrate: gọi OCR → classify new/existing
3. Response format chung cho mọi ngôn ngữ (không chỉ Chinese)

### Response hiện tại (Chinese-only)
```json
{"hanzi": "谢谢", "pinyin": "xiè xiè", "confidence": 0.97}
```

### Response mới (multilang)
```json
{"text": "谢谢", "pronunciation": "xiè xiè", "confidence": 0.97}
```

`text` + `pronunciation` — FE render giống nhau cho mọi ngôn ngữ.

---

## Alternative A: Rename field chỉ ở DTO layer

Giữ internal struct dùng `Text`/`Pinyin`, chỉ đổi JSON tag ở response DTO.

**Pros:** Thay đổi nhỏ nhất, chỉ sửa JSON tags
**Cons:** Code nội bộ vẫn dùng `Hanzi`/`Pinyin` — misleading cho non-Chinese

## Alternative B: Rename toàn bộ từ internal struct đến response (Recommend)

Đổi tất cả: domain types, DTOs, response — thống nhất `Text`/`Pronunciation` xuyên suốt.

**Pros:** Consistent, code tự document đúng ý nghĩa multilang
**Cons:** Refactor rộng hơn, nhưng đang tách module nên sửa luôn 1 lần

### Recommend: Alternative B — refactor toàn bộ trong lần tách này.

---

## Cấu trúc sau tách

```
internal/ocr/
├── application/
│   ├── port/
│   │   ├── inbound.go      ← OCRCommandPort
│   │   └── outbound.go     ← OCRServicePort, OCREngineRegistry
│   ├── dto/
│   │   └── dto.go          ← OCRScanRequest, OCRScanResponse (multilang)
│   └── usecase/
│       └── ocr_command.go  ← recognize + enrich pronunciation + classify confidence
├── adapter/
│   ├── handler/
│   │   └── handler.go      ← ProcessOCRScan + downloadImage
│   └── service/
│       ├── google_vision_service.go
│       ├── baidu_ocr_service.go
│       ├── ocr_service.go       ← PaddleOCR/Tesseract (Python service)
│       └── ocr_retry.go
├── docs/                        ← move từ vocabulary/docs/ocr/
└── module.go
```

```
internal/vocabulary/
├── application/
│   ├── port/inbound.go     ← bỏ OCRCommandPort
│   ├── port/outbound.go    ← bỏ OCRServicePort, OCREngineRegistry
│   ├── dto/dto.go          ← bỏ OCR DTOs
│   └── usecase/            ← bỏ ocr_command.go
├── adapter/
│   ├── handler/handler.go  ← bỏ ProcessOCRScan, downloadImage
│   └── service/            ← bỏ OCR service files
└── module.go               ← bỏ ocrEngines dependency
```

---

## OCR Module response (multilang)

### OCR module trả (raw characters):
```json
{
  "items": [
    {"text": "谢谢", "pronunciation": "xiè xiè", "confidence": 0.97, "candidates": []},
    {"text": "鑫", "pronunciation": "xīn", "confidence": 0.55, "candidates": ["鑫", "森"]}
  ],
  "metadata": {
    "engine_used": "google_vision",
    "total_detected": 2,
    "processing_time_ms": 540
  }
}
```

### Vocabulary module wraps (classify new/existing):
```json
{
  "new_items": [
    {"text": "谢谢", "pronunciation": "xiè xiè", "confidence": 0.97}
  ],
  "existing_items": [
    {"id": "uuid", "text": "你好", "pronunciation": "nǐ hǎo", ...vocabulary fields..., "confidence": 0.98}
  ],
  "low_confidence_items": [
    {"text": "鑫", "pronunciation": "xīn", "confidence": 0.55, "candidates": ["鑫", "森"]}
  ],
  "metadata": {"engine_used": "google_vision", "total_detected": 2, "processing_time_ms": 540}
}
```

---

## Steps

### Step 1: Tạo `internal/ocr/` module structure

Tạo directories + files mới.

### Step 2: Move OCR ports

Move từ `vocabulary/application/port/outbound.go`:
- `OCRServicePort`, `OCRRequest`, `OCRResult`, `OCRCharacter` → `ocr/application/port/outbound.go`
- `OCREngineKey`, `OCREngineRegistry`, constants → `ocr/application/port/outbound.go`

Tạo `ocr/application/port/inbound.go`:
- `OCRCommandPort` với method `ProcessScan` (trả raw characters, không classify)

### Step 3: Tạo OCR DTOs (multilang)

`ocr/application/dto/dto.go`:
```go
type OCRScanHTTPRequest struct {
    ImageURL string `json:"image_url" binding:"required,url"`
    Type     string `json:"type" binding:"omitempty,oneof=printed handwritten auto"`
    Language string `json:"language" binding:"omitempty,oneof=zh vi en"`
    Engine   string `json:"engine" binding:"omitempty,oneof=paddleocr tesseract google_vision baidu_ocr"`
}

type OCRScanRequest struct {
    Image    []byte
    Type     string
    Language string
    Engine   string
}

type OCRCharacterItem struct {
    Text          string   `json:"text"`
    Pronunciation string   `json:"pronunciation"`
    Confidence    float64  `json:"confidence"`
    Candidates    []string `json:"candidates,omitempty"`
}

type OCRScanMetadata struct {
    EngineUsed       string `json:"engine_used"`
    TotalDetected    int    `json:"total_detected"`
    ProcessingTimeMs int64  `json:"processing_time_ms"`
}

type OCRScanResponse struct {
    Items    []OCRCharacterItem `json:"items"`
    Metadata OCRScanMetadata    `json:"metadata"`
}
```

### Step 4: Move OCR use case

Move `vocabulary/application/usecase/ocr_command.go` → `ocr/application/usecase/ocr_command.go`

Simplify — bỏ DB lookup, chỉ:
1. resolveEngine
2. engine.Recognize
3. enrichPronunciation (pinyin for zh)
4. classify by confidence (items vs low_confidence)
5. return OCRScanResponse

### Step 5: Move OCR adapters

Move từ `vocabulary/adapter/service/`:
- `google_vision_service.go` → `ocr/adapter/service/`
- `baidu_ocr_service.go` → `ocr/adapter/service/`
- `ocr_service.go` → `ocr/adapter/service/`
- `ocr_retry.go` → `ocr/adapter/service/`

### Step 6: Tạo OCR handler

`ocr/adapter/handler/handler.go`:
- `ProcessOCRScan` — parse request, download image, call use case
- `downloadImage` — move từ vocabulary handler

### Step 7: Tạo OCR module.go

```go
func NewModule(engines port.OCREngineRegistry) *Module
func (module *Module) RegisterRoutes(public, protected *gin.RouterGroup)
```

Route: `public.POST("/ocr/scan", ...)`

### Step 8: Move pronunciation helper

`ConvertToPinyin` → `ocr/application/mapper/pronunciation.go`
(OCR module sở hữu pronunciation enrichment, không phải vocabulary)

### Step 9: Update vocabulary module

- Bỏ OCR dependencies khỏi `module.go`, `handler.go`
- Bỏ OCR DTOs, ports, use case
- Vocabulary handler **KHÔNG** gọi OCR nữa

### Step 10: Tạo vocabulary OCR orchestrator endpoint (nếu cần)

Nếu mobile cần 1 endpoint trả new/existing:
- Vocabulary module import `ocr` module exported port
- Tạo use case mới: gọi OCR → classify → return combined response
- Hoặc: FE gọi OCR endpoint → nhận raw → gọi vocabulary classify endpoint riêng

### Step 11: Update DI container

- `di/ocr.go` → tạo OCR module thay vì engine registry
- `di/container.go` → register OCR module routes
- `vocabulary.NewModule(db)` — bỏ `ocrEngines` param

### Step 12: Move docs

`vocabulary/docs/ocr/` → `ocr/docs/`

---

## Files tạo mới

| File | Mô tả |
|---|---|
| `internal/ocr/module.go` | Module wiring + RegisterRoutes |
| `internal/ocr/application/port/inbound.go` | OCRCommandPort |
| `internal/ocr/application/port/outbound.go` | OCRServicePort, OCREngineRegistry |
| `internal/ocr/application/dto/dto.go` | Multilang DTOs |
| `internal/ocr/application/usecase/ocr_command.go` | Recognize + enrich + classify confidence |
| `internal/ocr/application/mapper/pronunciation.go` | ConvertToPinyin |
| `internal/ocr/adapter/handler/handler.go` | HTTP handler + downloadImage |
| `internal/ocr/adapter/service/*.go` | Move 4 adapter files |

## Files sửa

| File | Thay đổi |
|---|---|
| `internal/vocabulary/application/port/outbound.go` | Xoá OCR types |
| `internal/vocabulary/application/port/inbound.go` | Xoá OCRCommandPort |
| `internal/vocabulary/application/dto/dto.go` | Xoá OCR DTOs |
| `internal/vocabulary/adapter/handler/handler.go` | Xoá ProcessOCRScan, downloadImage, ocrCmd |
| `internal/vocabulary/module.go` | Bỏ ocrEngines, bỏ OCR route |
| `internal/infrastructure/di/container.go` | Tạo OCR module, register routes |
| `internal/infrastructure/di/ocr.go` | Return OCR module thay vì engine registry |
| `internal/server/router.go` | Thêm OCR module param |

## Files xoá

| File | Lý do |
|---|---|
| `internal/vocabulary/application/usecase/ocr_command.go` | Moved to ocr module |
| `internal/vocabulary/adapter/service/google_vision_service.go` | Moved |
| `internal/vocabulary/adapter/service/baidu_ocr_service.go` | Moved |
| `internal/vocabulary/adapter/service/ocr_service.go` | Moved |
| `internal/vocabulary/adapter/service/ocr_retry.go` | Moved |
| `internal/vocabulary/application/mapper/vocabulary.go` | ConvertToPinyin moved to ocr |
