# Plan: Tích hợp Google Cloud Vision OCR

## Context

Hiện tại OCR pipeline dùng PaddleOCR (Python service local) qua HTTP adapter. Theo `plan_ocr_engine.md`, Google Cloud Vision là engine chính cho:
- **Printed text** (mọi ngôn ngữ)
- **Handwritten** (mọi ngôn ngữ trừ Chinese)
- **Auto mode** (primary engine, fallback sang Baidu nếu confidence thấp + lang=zh)

Architecture hiện tại đã sẵn sàng — chỉ cần tạo adapter mới implement `OCRServicePort` rồi đăng ký vào `OCREngineRegistry`. Không cần sửa port, use case, handler, hay routing logic.

### Scope

- Tạo `GoogleVisionAdapter` implement `OCRServicePort`
- Wire vào DI container
- Thêm config cho Google Vision credentials
- Pinyin conversion đặt ở use case layer (post-processing chung cho mọi engine)
- PaddleOCR vẫn giữ làm fallback / dev engine

### Quyết định đã chốt

| Quyết định | Chọn | Lý do |
|---|---|---|
| **SDK vs REST API** | Official Go SDK (`cloud.google.com/go/vision/v2/apiv1`) | Type-safe, auto-retry, gRPC efficient, connection pooling, SDK tự handle auth. Project đã dùng managed infrastructure pattern |
| **Authentication** | Service Account (`GOOGLE_APPLICATION_CREDENTIALS`) mọi môi trường | Đồng nhất dev/staging/prod. Scoped permissions, audit trail, không cần logic detect auth method |
| **Pinyin** | Post-processing ở use case layer | Pinyin không phải responsibility của OCR engine. Đặt ở use case → mọi engine (PaddleOCR, Google Vision, Baidu) đều có pinyin mà adapter không cần biết. Dùng Go pinyin library |

---

## Research cần làm trước khi code

### R1: Google Vision response format

Cần verify:
- `TEXT_DETECTION` vs `DOCUMENT_TEXT_DETECTION` — cái nào phù hợp cho Chinese characters?
- Confidence score format: per-character hay per-block? Cần normalize thế nào?
- Response có `pages > blocks > paragraphs > words > symbols` — extract ở level nào?

**Dự kiến**: `DOCUMENT_TEXT_DETECTION` cho structured text, extract ở **symbol level** cho Chinese (mỗi symbol = 1 character), **word level** cho các ngôn ngữ khác.

### R2: Go pinyin library

Cần evaluate:
- `github.com/mozillazg/go-pinyin` — popular, maintained, tone marks support?
- Accuracy so với Python `pypinyin` (PaddleOCR hiện dùng)
- Performance: convert 100 characters < 1ms?

---

## Steps

### Step 1: Thêm Google Vision config

File: `internal/infrastructure/config/ocr.go`

```go
type OCRConfig struct {
    OCRServiceURL                string `mapstructure:"OCR_SERVICE_URL"`
    GoogleApplicationCredentials string `mapstructure:"GOOGLE_APPLICATION_CREDENTIALS"`
}
```

- `GOOGLE_APPLICATION_CREDENTIALS` — path tới service account JSON file
- Nếu rỗng → không đăng ký Google Vision adapter (graceful skip)

**Tạo Service Account & lấy credentials:**

1. Vào [Google Cloud Console](https://console.cloud.google.com/)
2. Tạo project (hoặc chọn project có sẵn)
3. **Enable Vision API:**
   - Menu → APIs & Services → Library
   - Tìm "Cloud Vision API" → **Enable**
4. **Tạo Service Account:**
   - Menu → IAM & Admin → Service Accounts
   - **Create Service Account**
   - Name: `ocr-backend` (hoặc tên tuỳ ý)
   - Role: chọn **Cloud Vision AI Service Agent** (chỉ cho phép Vision API, không gì khác)
   - Done
5. **Tạo JSON key:**
   - Click vào service account vừa tạo
   - Tab **Keys** → Add Key → Create new key → **JSON**
   - File JSON tự download → lưu vào nơi an toàn (VD: `~/.gcp/ocr-service-account.json`)
   - **KHÔNG commit file này vào git**
6. **Set env var:**
   ```
   GOOGLE_APPLICATION_CREDENTIALS=/Users/thaidong/.gcp/ocr-service-account.json
   ```

> **Lưu ý billing**: Google Cloud yêu cầu gắn billing account vào project. Vision API miễn phí 1,000 units/tháng — vượt mới tính tiền. Nếu chưa có billing account, tạo tại Billing → sẽ được $300 free credit cho 90 ngày đầu.

File: `.env.example`

```
# Google Cloud Vision OCR — path to Service Account JSON file
# Setup guide: xem plan_google_vision_integration.md Step 1
GOOGLE_APPLICATION_CREDENTIALS=
```

### Step 2: Tạo Google Vision adapter

File: `internal/vocabulary/adapter/service/google_vision_service.go`

```go
type GoogleVisionService struct {
    client  *vision.ImageAnnotatorClient
    breaker *circuitbreaker.Breaker
}

func NewGoogleVisionService(credFile string, breaker *circuitbreaker.Breaker) (port.OCRServicePort, func(), error)
// Returns: adapter, cleanup function (closes gRPC connection), error

func (svc *GoogleVisionService) Recognize(ctx context.Context, req port.OCRRequest) (*port.OCRResult, error)
```

**Recognize flow:**
1. Circuit breaker wrap
2. Tạo `vision.Image` từ `req.Image` bytes
3. Gọi `client.DetectDocumentText(ctx, image)` (DOCUMENT_TEXT_DETECTION)
4. Parse response:
   - Chinese (`req.Language == "zh"`): extract ở **symbol level** — mỗi symbol = 1 Hán tự
   - Khác: extract ở **word level**
5. Với mỗi detected text:
   - `Text`: raw character/word
   - `Confidence`: symbol/word confidence (Google trả 0-1, giữ nguyên)
   - `Pinyin`: **rỗng** (use case layer sẽ enrich sau)
   - `Candidates`: rỗng (Google không trả alternatives)
6. Filter CJK characters nếu `language == "zh"` (U+4E00–U+9FFF)
7. Deduplicate (seen set)
8. Return `OCRResult{Characters: [...], Engine: "google_vision"}`

**Error handling:**
- Network/gRPC error → log `[OCR] Google Vision connection failed` + `apperr.ServiceUnavailable("ocr.service_connection_failed", err)`
- API error (quota, invalid image) → log `[OCR] Google Vision API error` + `apperr.ServiceUnavailable("ocr.service_error", err)`
- 0 results → return empty `OCRResult` (không phải error)

**Constructor:**
```go
func NewGoogleVisionService(credFile string, breaker *circuitbreaker.Breaker) (port.OCRServicePort, func(), error) {
    ctx := context.Background()
    client, err := vision.NewImageAnnotatorClient(ctx,
        option.WithCredentialsFile(credFile),
    )
    if err != nil {
        return nil, nil, err
    }
    cleanup := func() { client.Close() }
    return &GoogleVisionService{client: client, breaker: breaker}, cleanup, nil
}
```

### Step 3: Pinyin post-processing ở use case layer

File: `internal/vocabulary/application/usecase/ocr_command.go`

Thêm pinyin enrichment **sau** khi nhận `OCRResult` từ engine, **trước** khi classify:

```go
func (useCase *OCRCommand) ProcessOCRScan(...) {
    // ... resolve engine, call Recognize ...

    // Enrich pinyin cho characters chưa có
    for i, ch := range ocrResult.Characters {
        if ch.Pinyin == "" && req.Language == "zh" {
            ocrResult.Characters[i].Pinyin = convertToPinyin(ch.Text)
        }
    }

    // ... classify by confidence, check DB ...
}
```

Hàm `convertToPinyin` dùng Go pinyin library, đặt package-level trong `ocr_command.go` hoặc tách file `pinyin.go` nếu phức tạp.

**Impact**: PaddleOCR adapter có thể bỏ pinyin conversion ở Python side (hoặc giữ, use case sẽ skip nếu đã có pinyin). Google Vision adapter không cần biết pinyin.

### Step 4: Tạo circuit breaker riêng cho Google Vision

Theo `plan_ocr_engine.md` section 6.1, mỗi engine có CB riêng:

```go
gvBreaker := circuitbreaker.NewBreaker(circuitbreaker.BreakerConfig{
    Name:         "google-vision",
    MaxRequests:  3,       // half-open: chỉ cho 3 request thử
    Interval:     60,      // reset counter mỗi 60s
    Timeout:      30,      // open → half-open sau 30s
    FailureRatio: 0.5,     // trip khi >50% fail
    MinRequests:  10,      // cần ít nhất 10 request mới evaluate
}, successPredicate)
```

### Step 5: Wire vào DI container

File: `internal/infrastructure/di/container.go`

```go
// Google Vision adapter (chỉ tạo nếu có credentials)
if cfg.GoogleApplicationCredentials != "" {
    gvBreaker := circuitbreaker.NewBreaker(circuitbreaker.BreakerConfig{
        Name: "google-vision",
    }, successPredicate)

    gvAdapter, gvCleanup, err := vocabservice.NewGoogleVisionService(
        cfg.GoogleApplicationCredentials, gvBreaker,
    )
    if err != nil {
        logger.Warn("[DI] Google Vision adapter init failed, skipping", zap.Error(err))
    } else {
        ocrEngines[vocabport.OCREngineGoogleVision] = gvAdapter
        // Thêm gvCleanup vào cleanup chain
    }
}
```

**Quan trọng**: Nếu init fail → log warning + skip, không crash app. PaddleOCR vẫn là fallback.

### Step 6: Verify routing logic

Use case `resolveEngine()` hiện tại đã route đúng:
- `printed` / `auto` / `handwritten + non-zh` → `OCREngineGoogleVision` → **sẽ hit Google Vision adapter**
- `handwritten + zh` → `OCREngineBaiduOCR` → fallback sang engine có sẵn (PaddleOCR hoặc Google Vision)

**Không cần sửa routing logic.** Chỉ cần đăng ký adapter vào registry.

---

## Files cần tạo

| File | Mô tả |
|---|---|
| `internal/vocabulary/adapter/service/google_vision_service.go` | Google Vision adapter implement `OCRServicePort` |

## Files cần sửa

| File | Thay đổi |
|---|---|
| `internal/infrastructure/config/ocr.go` | Thêm `GoogleApplicationCredentials` |
| `internal/infrastructure/di/container.go` | Tạo + đăng ký Google Vision adapter + cleanup |
| `.env.example` | Thêm `GOOGLE_APPLICATION_CREDENTIALS` |
| `go.mod` | Thêm `cloud.google.com/go/vision/v2` + Go pinyin library |
| `internal/vocabulary/application/usecase/ocr_command.go` | Thêm pinyin enrichment sau khi nhận `OCRResult` |

## Files KHÔNG cần sửa

| File | Lý do |
|---|---|
| `application/port/outbound.go` | Interface đã đủ |
| `application/port/inbound.go` | Không thay đổi |
| `adapter/handler/handler.go` | Handler layer không biết engine nào |
| `module.go` | Nhận `OCREngineRegistry`, không biết engine cụ thể |
| `application/dto/dto.go` | DTOs không thay đổi |
| `resources/i18n/*/common.json` | Error keys đã đủ |

---

## Dependency mới

```bash
go get cloud.google.com/go/vision/v2
go get github.com/mozillazg/go-pinyin  # hoặc library khác sau khi evaluate
```

SDK dependency tree: `cloud.google.com/go/vision/v2` → `google.golang.org/api` → `google.golang.org/grpc` (~20-30 transitive dependencies).

---

## Testing

### Unit test

- Mock `vision.ImageAnnotatorClient` → test adapter normalize response đúng
- Test CJK filter, dedup, confidence mapping
- Test pinyin enrichment ở use case layer

### Integration test (manual)

```bash
# 1. Set credentials
export GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account.json

# 2. Start Go server
make run

# 3. Test
curl -X POST http://localhost:8001/vocabularies/ocr-scan \
  -H "Content-Type: application/json" \
  -d '{"image_url": "https://example.com/chinese_text.jpg", "type": "printed", "language": "zh"}'
```

### So sánh accuracy

Test cùng ảnh với PaddleOCR vs Google Vision:
- `type: "printed"` → Google Vision vs PaddleOCR
- So sánh: total_detected, confidence distribution, missed characters, pinyin accuracy

---

## Rollout

| Phase | Mô tả |
|---|---|
| **Dev** | Service Account với dev project. Test với sample images. So sánh accuracy với PaddleOCR |
| **Staging** | Service Account staging project. Load test: verify latency < 3s p99 |
| **Production** | Enable Google Vision. PaddleOCR giữ làm fallback khi Google down |

---

## Cost estimate

| Tier | Giá | Budget |
|---|---|---|
| First 1,000 units/month | Free | MVP đủ dùng |
| 1,001 - 5M units/month | $1.50/1000 units | Growth: ~$75/month @ 50K req/month |

1 unit = 1 image. `DOCUMENT_TEXT_DETECTION` = 1 unit/image.

---

## Risks & Mitigation

| Risk | Mitigation |
|---|---|
| SDK dependency nặng (~30 packages) | Chấp nhận — official SDK stable, well-maintained |
| Credentials leak | Service Account JSON file không commit. `.gitignore` đã có. CI/CD inject secrets |
| Google Vision down | Circuit breaker + fallback sang PaddleOCR (đã có trong registry) |
| Latency cao (cold start) | SDK tạo gRPC connection pool, reuse sau lần đầu |
| Go pinyin library accuracy | Evaluate trước, so sánh với Python `pypinyin`. Nếu không đủ tốt → giữ PaddleOCR Python làm pinyin source |
