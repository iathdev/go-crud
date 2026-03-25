# Plan: Cài đặt và tích hợp Tesseract OCR

## Context

Python OCR service (`scripts/ocr-service/main.py`) đã hỗ trợ Tesseract engine — gọi với `engine: "tesseract"` là chạy. Tuy nhiên Go side chưa có cách route request tới Tesseract:

- `OCRService` adapter (`ocr_service.go`) luôn gửi request **không có** field `engine` → Python service mặc định dùng `paddleocr`
- `OCREngineRegistry` chưa có `OCREngineTesseract` key
- `resolveEngine()` chưa biết Tesseract tồn tại

**Mục tiêu:** Đăng ký Tesseract như 1 engine riêng trong Go registry, route qua **cùng Python service** với `engine: "tesseract"`, không cần cgo/gosseract.

**Giới hạn:** Tesseract chỉ phù hợp **printed text**. Handwritten Chinese accuracy gần 0 (từ `research_ocr_engine.md`).

### Quyết định đã chốt

| Quyết định | Chọn | Thay vì | Tại sao |
|---|---|---|---|
| **Native Go (gosseract/cgo)** vs **Python service** | Python service | gosseract cgo wrapper | Python service đã hỗ trợ sẵn Tesseract (`_extract_tesseract`). Không cần cgo → không phức tạp build, không cần C compiler, không tăng Docker image. Reuse infrastructure có sẵn |
| **1 adapter chung** vs **2 adapter riêng** | 1 adapter `OCRService` dùng chung, thêm `engineName` field | Tạo `TesseractService` adapter riêng | Cùng protocol (HTTP → Python), cùng request/response format. Chỉ khác field `engine`. DRY — không duplicate code |
| **Routing** | Tesseract là fallback cuối cho `printed` + `auto` | Thêm vào `handwritten` chain | Research confirm Tesseract handwritten Chinese accuracy gần 0. Chỉ có giá trị cho printed text |

---

## Step 1: Cài Tesseract system deps + Python deps

**macOS:**
```bash
brew install tesseract tesseract-lang
cd scripts/ocr-service && source .venv/bin/activate
pip install pytesseract Pillow
```

**Ubuntu/Debian:**
```bash
sudo apt install tesseract-ocr tesseract-ocr-chi-sim tesseract-ocr-chi-tra tesseract-ocr-vie
cd scripts/ocr-service && source .venv/bin/activate
pip install pytesseract Pillow
```

Language data cần: `chi_sim` (Simplified Chinese), `chi_tra` (Traditional Chinese), `vie` (Vietnamese), `eng` (default).

Thêm vào `scripts/ocr-service/requirements.txt`:
```
pytesseract
Pillow
```

---

## Step 2: Thêm `OCREngineTesseract` constant

**File:** `internal/vocabulary/application/port/outbound.go`

```go
const (
    OCREnginePaddleOCR    OCREngineKey = "paddleocr"
    OCREngineGoogleVision OCREngineKey = "google_vision"
    OCREngineBaiduOCR     OCREngineKey = "baidu_ocr"
    OCREngineTesseract    OCREngineKey = "tesseract"      // NEW
)
```

---

## Step 3: Refactor `OCRService` adapter để hỗ trợ `engine` parameter

**File:** `internal/vocabulary/adapter/service/ocr_service.go`

Hiện tại `ocrExtractRequest` không gửi `engine` field → Python mặc định `paddleocr`. Cần thêm:

```go
type ocrExtractRequest struct {
    Image    string `json:"image"`
    Language string `json:"language"`
    Engine   string `json:"engine,omitempty"`   // NEW
}
```

Thêm `engineName` field vào `OCRService` struct:
```go
type OCRService struct {
    baseURL    string
    client     *http.Client
    breaker    *circuitbreaker.Breaker
    engineName string    // "paddleocr" or "tesseract"
}
```

Update constructor:
```go
func NewOCRService(baseURL string, engineName string, breaker *circuitbreaker.Breaker) port.OCRServicePort
```

Trong `Recognize()`, set `payload.Engine = svc.engineName`.

Approach này cho phép **tạo 2 instance** của cùng `OCRService` adapter — 1 cho PaddleOCR, 1 cho Tesseract — mỗi instance gửi `engine` name khác nhau tới cùng Python service.

---

## Step 4: Thêm config

**File:** `internal/infrastructure/config/ocr.go`

```go
TesseractEnabled bool `mapstructure:"TESSERACT_ENABLED"`
```

**File:** `.env.example`
```
# Tesseract OCR — requires Python OCR service + pytesseract installed
# Dev/fallback engine for printed text only. Accuracy thấp hơn Google/Baidu/PaddleOCR.
TESSERACT_ENABLED=false
```

---

## Step 5: Wire vào DI container

**File:** `internal/infrastructure/di/ocr.go`

```go
// PaddleOCR (existing — thêm engine name parameter)
if cfg.OCRServiceURL != "" {
    paddleAdapter := vocabservice.NewOCRService(cfg.OCRServiceURL, "paddleocr", newOCRBreaker("paddle-ocr"))
    result.register(vocabport.OCREnginePaddleOCR, withRetry(paddleAdapter))
}

// Tesseract — cùng Python service, khác engine name
if cfg.OCRServiceURL != "" && cfg.TesseractEnabled {
    tessAdapter := vocabservice.NewOCRService(cfg.OCRServiceURL, "tesseract", newOCRBreaker("tesseract"))
    result.register(vocabport.OCREngineTesseract, withRetry(tessAdapter))
}
```

Cả 2 gọi cùng `OCR_SERVICE_URL`, chỉ khác `engine` parameter gửi tới Python service.

---

## Step 6: Update routing logic

**File:** `internal/vocabulary/application/usecase/ocr_command.go` — `resolveEngine()`

```go
case "printed":
    return useCase.getFirstAvailable(
        port.OCREngineGoogleVision,
        port.OCREnginePaddleOCR,
        port.OCREngineTesseract,    // last fallback cho printed
    )

case "handwritten":
    if language == "zh" {
        // Tesseract KHÔNG có ở đây — accuracy handwritten gần 0
        return useCase.getFirstAvailable(
            port.OCREngineBaiduOCR,
            port.OCREnginePaddleOCR,
            port.OCREngineGoogleVision,
        )
    }
    return useCase.getEngine(port.OCREngineGoogleVision)

default: // "auto"
    return useCase.getFirstAvailable(
        port.OCREngineGoogleVision,
        port.OCREnginePaddleOCR,
        port.OCREngineTesseract,    // last fallback
    )
```

### Routing summary

| ocrType | Language | Engine chain |
|---|---|---|
| `printed` | any | GoogleVision → PaddleOCR → **Tesseract** |
| `handwritten` | `zh` | BaiduOCR → PaddleOCR → GoogleVision (KHÔNG có Tesseract) |
| `handwritten` | other | GoogleVision (KHÔNG có Tesseract) |
| `auto` | any | GoogleVision → PaddleOCR → **Tesseract** |

Tesseract **KHÔNG BAO GIỜ** là primary engine. Chỉ activate khi tất cả engine khác unavailable.

---

## Files tổng hợp

### Files cần sửa

| File | Thay đổi |
|---|---|
| `scripts/ocr-service/requirements.txt` | Thêm `pytesseract`, `Pillow` |
| `internal/vocabulary/application/port/outbound.go` | Thêm `OCREngineTesseract` const |
| `internal/vocabulary/adapter/service/ocr_service.go` | Thêm `engineName` field vào struct + constructor, gửi `engine` trong JSON request |
| `internal/infrastructure/config/ocr.go` | Thêm `TesseractEnabled` bool field |
| `internal/infrastructure/di/ocr.go` | Thêm block tạo Tesseract adapter instance |
| `internal/vocabulary/application/usecase/ocr_command.go` | Update `resolveEngine()` fallback chains |
| `.env.example` | Thêm `TESSERACT_ENABLED=false` |

### Files KHÔNG cần sửa

| File | Lý do |
|---|---|
| `scripts/ocr-service/main.py` | Đã hỗ trợ `engine: "tesseract"` sẵn |
| `application/port/outbound.go` (interface) | `OCRServicePort` interface không đổi |
| `application/dto/dto.go` | DTOs không thay đổi |
| `adapter/handler/handler.go` | Handler không biết engine nào |
| `vocabulary/module.go` | Nhận `OCREngineRegistry`, không biết engine cụ thể |
| `resources/i18n/*/common.json` | Error keys đã đủ |

---

## Verification

1. Cài Tesseract system + Python deps:
   ```bash
   brew install tesseract tesseract-lang
   cd scripts/ocr-service && source .venv/bin/activate && pip install pytesseract Pillow
   ```

2. `go build ./...` — verify compile

3. Set `.env`:
   ```
   OCR_SERVICE_URL=http://localhost:8000
   TESSERACT_ENABLED=true
   ```

4. Start Python service:
   ```bash
   cd scripts/ocr-service && source .venv/bin/activate && uvicorn main:app --port 8000
   ```

5. Test trực tiếp Python service:
   ```bash
   curl -X POST http://localhost:8000/recognize \
     -H "Content-Type: application/json" \
     -d '{"image": "'$(base64 < scripts/ocr-service/test_image.png)'", "language": "zh", "engine": "tesseract"}'
   ```

6. Start Go server: `make run`

7. Test qua Go endpoint (không có Google/Baidu keys, không có PaddleOCR → fallback tới Tesseract cho printed):
   ```bash
   curl -X POST http://localhost:8001/api/v1/vocabularies/ocr-scan \
     -H "Authorization: Bearer <JWT>" \
     -H "Content-Type: application/json" \
     -d '{"image_url": "https://example.com/printed_chinese.jpg", "type": "printed", "language": "zh"}'
   ```

8. Verify response có `"engine_used": "tesseract"`
