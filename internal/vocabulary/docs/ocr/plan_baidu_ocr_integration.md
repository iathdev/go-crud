# Plan: Tích hợp Baidu OCR API

## Context

Google Vision đã tích hợp (Phase 1). Theo `plan_ocr_engine.md`, Baidu OCR là engine chính cho **handwritten Chinese** — accuracy vượt trội (PP-OCRv5 vượt GPT-4o). Khi có Baidu, routing `handwritten + zh` sẽ là: **Baidu → PaddleOCR → Google Vision**.

Ngoài ra, với mọi request `language=zh` (kể cả `auto`, `printed`), khi cần cascading fallback (confidence thấp) → Baidu là secondary engine thay vì PaddleOCR.

### Quyết định đã chốt

| Quyết định | Chọn | Lý do |
|---|---|---|
| **API approach** | REST API (HTTP POST) | Baidu không có official Go SDK. REST đơn giản, chỉ cần HTTP client |
| **Authentication** | OAuth2 (API Key + Secret Key → Access Token) | Cách duy nhất Baidu hỗ trợ |
| **Token caching** | Redis, TTL 29 ngày | Token valid 30 ngày, cache tránh round trip ~200-500ms mỗi request |
| **Per-character confidence** | `recognize_granularity=small` | Baidu mặc định trả per-line. Param này bật per-character — cần cho classify confirmed/low_confidence |
| **Confidence normalization** | ÷ 100 (Baidu: 0-100 → 0.0-1.0) | Đồng nhất với Google Vision (0.0-1.0) |
| **Routing priority (zh)** | Baidu > PaddleOCR > Google Vision | Baidu accuracy handwritten zh tốt nhất. PaddleOCR cùng model, free fallback. Google Vision cuối cùng |

### Khác biệt so với Google Vision adapter

| | Google Vision | Baidu OCR |
|---|---|---|
| **Protocol** | gRPC (Go SDK) | REST API (HTTP POST) |
| **Auth** | Service Account JSON file | OAuth2 token (API Key + Secret → token 30 ngày) |
| **Token management** | SDK tự handle | Adapter tự quản lý: cache Redis, refresh khi hết hạn |
| **Confidence range** | 0.0-1.0 (giữ nguyên) | 0-100 (cần ÷ 100) |
| **Per-character** | Native (Symbol level) | Cần param `recognize_granularity=small` |
| **Dependency** | `cloud.google.com/go/vision/v2` (~30 packages) | Không thêm dependency — chỉ `net/http` + `encoding/json` |
| **Cleanup** | Close gRPC connection | Không cần (stateless HTTP) |

---

## Đăng ký tài khoản Baidu & lấy API credentials

### Bước 1: Tạo tài khoản Baidu

1. Vào [Baidu Account Registration](https://passport.baidu.com/v2/?reg)
2. Chọn đăng ký bằng **email** (ổn định hơn phone từ VN)
3. Điền thông tin → verify email
4. Nếu yêu cầu số điện thoại TQ: dùng Google Voice / TextNow hoặc nhờ người ở TQ

> **Lưu ý từ VN**: UI tiếng Trung. Dùng Chrome auto-translate. OTP quốc tế hay fail → ưu tiên đăng ký bằng email. Nếu vẫn không được → dùng VPN node TQ hoặc nhờ người ở TQ đăng ký giúp.

### Bước 2: Kích hoạt Baidu AI Cloud

1. Vào [Baidu AI Cloud Console](https://console.bce.baidu.com/)
2. Lần đầu sẽ yêu cầu **实名认证 (xác thực danh tính)**:
   - Cá nhân: CCCD/passport + selfie
   - Doanh nghiệp: giấy phép kinh doanh
   - **Nếu không có giấy tờ TQ**: chọn "个人认证" (cá nhân) → dùng passport nước ngoài (một số trường hợp chấp nhận)
3. Sau khi xác thực → vào **产品服务 (Product Services)** → **文字识别 (Text Recognition)**

### Bước 3: Tạo ứng dụng OCR

1. Vào [Baidu AI Console - OCR](https://console.bce.baidu.com/ai/#/ai/ocr/overview/index)
2. Click **创建应用 (Create Application)**
3. Điền:
   - 应用名称 (App name): `ocr-backend` (tuỳ ý)
   - 接口选择 (API selection): chọn **手写文字识别 (Handwriting Recognition)** + **通用文字识别 (General Recognition)**
   - 应用描述 (Description): `Chinese character OCR for vocabulary app`
4. Submit → nhận **API Key** và **Secret Key**
5. Copy cả hai vào `.env`:
   ```
   BAIDU_OCR_API_KEY=<your_api_key>
   BAIDU_OCR_SECRET_KEY=<your_secret_key>
   ```

### Bước 4: Verify credentials

```bash
# Lấy access token
curl -s -X POST "https://aip.baidubce.com/oauth/2.0/token?grant_type=client_credentials&client_id=<API_KEY>&client_secret=<SECRET_KEY>" | jq .

# Expected: {"access_token": "24.xxx...", "expires_in": 2592000, ...}
```

### Bước 5: Test OCR trực tiếp

```bash
# Lấy token
TOKEN=$(curl -s -X POST "https://aip.baidubce.com/oauth/2.0/token?grant_type=client_credentials&client_id=<API_KEY>&client_secret=<SECRET_KEY>" | jq -r .access_token)

# Test handwriting OCR
curl -X POST "https://aip.baidubce.com/rest/2.0/ocr/v1/handwriting?access_token=$TOKEN" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "image=$(base64 < test_image.jpg)&recognize_granularity=small"
```

### Free tier

| Endpoint | Free quota | Sau khi hết |
|---|---|---|
| 手写文字识别 (Handwriting) | 500 calls/ngày | ¥0.01/call (~$0.0014) |
| 通用文字识别 (General) | 50,000 calls/ngày | ¥0.004/call |

MVP (180 calls/ngày handwritten zh): **Free tier đủ dùng.**

### Fallback nếu không đăng ký được

Nếu không đăng ký được Baidu (xác thực fail, OTP fail):
- `handwritten + zh` sẽ fallback sang **PaddleOCR** (cùng model PP-OCRv5, accuracy tương đương, nhưng per-line confidence)
- App vẫn hoạt động bình thường — chỉ thiếu per-character confidence cho handwritten zh

---

## Implementation (đã hoàn thành)

### Step 1: Thêm Baidu config

File: `internal/infrastructure/config/ocr.go`

```go
type OCRConfig struct {
    OCRServiceURL                string `mapstructure:"OCR_SERVICE_URL"`
    GoogleApplicationCredentials string `mapstructure:"GOOGLE_APPLICATION_CREDENTIALS"`
    BaiduOCRAPIKey               string `mapstructure:"BAIDU_OCR_API_KEY"`
    BaiduOCRSecretKey            string `mapstructure:"BAIDU_OCR_SECRET_KEY"`
}
```

File: `.env.example`

```
# Baidu OCR API (handwritten Chinese)
# Get at: https://console.bce.baidu.com/ → 文字识别 → 创建应用
# Leave empty to skip Baidu (PaddleOCR/Google Vision remain as fallback)
BAIDU_OCR_API_KEY=
BAIDU_OCR_SECRET_KEY=
```

### Step 2: Tạo Baidu OCR adapter

File: `internal/vocabulary/adapter/service/baidu_ocr_service.go`

```go
type BaiduOCRService struct {
    apiKey    string
    secretKey string
    client    *http.Client
    breaker   *circuitbreaker.Breaker
    redis     *redis.Client  // cache access token
}

func NewBaiduOCRService(apiKey, secretKey string, breaker *circuitbreaker.Breaker, redis *redis.Client) port.OCRServicePort

func (svc *BaiduOCRService) Recognize(ctx context.Context, req port.OCRRequest) (*port.OCRResult, error)
```

**Recognize flow:**

1. Circuit breaker wrap
2. Lấy access token:
   - Check Redis cache `baidu_ocr:access_token`
   - Nếu có → dùng
   - Nếu không → gọi `POST https://aip.baidubce.com/oauth/2.0/token` → cache Redis TTL 29 ngày
3. Encode image → base64
4. Chọn endpoint theo loại text:
   - Handwriting: `POST https://aip.baidubce.com/rest/2.0/ocr/v1/handwriting`
   - General (printed): `POST https://aip.baidubce.com/rest/2.0/ocr/v1/general_basic`
5. Request body: `image=<base64>&recognize_granularity=small`
6. Parse response → normalize confidence (÷ 100) → map sang `port.OCRResult`
7. Filter CJK nếu `language == "zh"`
8. Deduplicate

**Token management:**

```go
const (
    baiduTokenKey = "baidu_ocr:access_token"
    baiduTokenTTL = 29 * 24 * time.Hour // 29 days (token valid 30 days)
)

func (svc *BaiduOCRService) getAccessToken(ctx context.Context) (string, error) {
    // 1. Check Redis
    token, err := svc.redis.Get(ctx, baiduTokenKey).Result()
    if err == nil && token != "" {
        return token, nil
    }

    // 2. Request new token
    // POST https://aip.baidubce.com/oauth/2.0/token
    // ?grant_type=client_credentials&client_id=<apiKey>&client_secret=<secretKey>

    // 3. Cache in Redis
    svc.redis.Set(ctx, baiduTokenKey, token, baiduTokenTTL)
    return token, nil
}
```

**Response parsing:**

Baidu response format:
```json
{
  "words_result": [
    {
      "words": "学习中文",
      "chars": [
        {"char": "学", "probability": 98},
        {"char": "习", "probability": 95},
        {"char": "中", "probability": 92},
        {"char": "文", "probability": 87}
      ]
    }
  ]
}
```

Map to `port.OCRResult`:
```go
for _, word := range response.WordsResult {
    for _, ch := range word.Chars {
        // Normalize: 98 → 0.98
        characters = append(characters, port.OCRCharacter{
            Text:       ch.Char,
            Confidence: float64(ch.Probability) / 100.0,
        })
    }
}
// Engine = "baidu_ocr"
// Pinyin rỗng → use case layer enrich (giống Google Vision)
```

**Error handling:**
- Token request fail → log `[OCR] Baidu token refresh failed` + `apperr.ServiceUnavailable("ocr.service_connection_failed", err)`
- API call fail → log `[OCR] Baidu OCR API error` + `apperr.ServiceUnavailable("ocr.service_error", err)`
- Invalid response → log `[OCR] Baidu OCR invalid response` + `apperr.ServiceUnavailable("ocr.service_invalid_response", err)`

### Step 3: Circuit breaker riêng cho Baidu

```go
baiduBreaker := circuitbreaker.NewBreaker(circuitbreaker.BreakerConfig{
    Name: "baidu-ocr",
}, successPredicate)
```

### Step 4: Wire vào DI container

File: `internal/infrastructure/di/container.go`

```go
// Baidu OCR adapter (chỉ tạo nếu có credentials)
if cfg.BaiduOCRAPIKey != "" && cfg.BaiduOCRSecretKey != "" {
    baiduBreaker := circuitbreaker.NewBreaker(circuitbreaker.BreakerConfig{
        Name: "baidu-ocr",
    }, successPredicate)

    baiduAdapter := vocabservice.NewBaiduOCRService(
        cfg.BaiduOCRAPIKey, cfg.BaiduOCRSecretKey,
        baiduBreaker, pst.redis,
    )
    ocrEngines[vocabport.OCREngineBaiduOCR] = baiduAdapter
}
```

Không cần cleanup function (stateless HTTP, không có gRPC connection).

### Step 5: Cập nhật routing logic

File: `internal/vocabulary/application/usecase/ocr_command.go`

Routing hiện tại **đã đúng** cho `handwritten + zh`:
```go
case "handwritten":
    if language == "zh" {
        return useCase.getFirstAvailable(
            port.OCREngineBaiduOCR,    // primary
            port.OCREnginePaddleOCR,   // fallback 1
            port.OCREngineGoogleVision, // fallback 2
        )
    }
```

**Cần thêm**: cascading `auto + zh` khi confidence thấp — hiện tại đã cascade sang Baidu:
```go
if req.Type == "auto" && req.Language == "zh" && len(ocrResult.Characters) > 0 {
    if avgConfidence(ocrResult.Characters) < ocrCascadingThreshold {
        if fallback, _ := useCase.getEngine(port.OCREngineBaiduOCR); fallback != nil {
            // ...
        }
    }
}
```

**Không cần sửa routing logic.**

### Step 6: Thêm i18n keys

Dùng chung các key đã có: `ocr.service_connection_failed`, `ocr.service_error`, `ocr.service_invalid_response`. Log nội bộ phân biệt qua message prefix `[OCR] Baidu ...`.

**Không cần thêm i18n key mới.**

---

## Files đã tạo

| File | Mô tả |
|---|---|
| `internal/vocabulary/adapter/service/baidu_ocr_service.go` | Baidu OCR adapter: OAuth2 token caching Redis, REST API, confidence ÷100, per-character via `recognize_granularity=small` |

## Files đã sửa

| File | Thay đổi |
|---|---|
| `internal/infrastructure/config/ocr.go` | Thêm `BaiduOCRAPIKey`, `BaiduOCRSecretKey` |
| `internal/infrastructure/di/container.go` | Conditional wire Baidu adapter với circuit breaker |
| `.env.example` | Thêm `BAIDU_OCR_API_KEY`, `BAIDU_OCR_SECRET_KEY` |

## Files KHÔNG cần sửa

| File | Lý do |
|---|---|
| `application/port/outbound.go` | `OCREngineBaiduOCR` constant đã có |
| `application/usecase/ocr_command.go` | Routing + cascading logic đã handle Baidu |
| `adapter/handler/handler.go` | Handler engine-agnostic |
| `module.go` | Nhận `OCREngineRegistry`, engine-agnostic |
| `application/dto/dto.go` | DTOs không thay đổi |
| `resources/i18n/*/common.json` | Error keys đã đủ |
| `go.mod` | Không thêm dependency mới — chỉ dùng `net/http` + `encoding/json` |

---

## Dependency

Không thêm dependency mới. Baidu adapter chỉ dùng stdlib (`net/http`, `encoding/json`, `encoding/base64`) + Redis client đã có.

---

## Testing

### Test token

```bash
curl -X POST "https://aip.baidubce.com/oauth/2.0/token?grant_type=client_credentials&client_id=<API_KEY>&client_secret=<SECRET_KEY>"
```

### Test OCR trực tiếp

```bash
curl -X POST "https://aip.baidubce.com/rest/2.0/ocr/v1/handwriting?access_token=<TOKEN>" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "image=$(base64 < test_image.jpg)&recognize_granularity=small"
```

### Test qua Go backend

```bash
curl -X POST http://localhost:8001/vocabularies/ocr-scan \
  -H "Content-Type: application/json" \
  -d '{"image_url":"https://example.com/handwritten_chinese.jpg","type":"handwritten","language":"zh"}'
```

Response `metadata.engine_used` sẽ là `"baidu_ocr"`.

### So sánh accuracy

Cùng ảnh handwritten Chinese, test 3 engine:
- `type: "handwritten"` + có Baidu credentials → Baidu
- `type: "handwritten"` + không có Baidu → PaddleOCR (fallback)
- `type: "printed"` → Google Vision

So sánh: total_detected, per-character confidence distribution, missed characters.

---

## Routing tổng kết (sau khi có Baidu)

| Request | Engine chain |
|---|---|
| `printed` (any lang) | Google Vision → 503 |
| `handwritten + zh` | **Baidu** → PaddleOCR → Google Vision → 503 |
| `handwritten + other` | Google Vision → 503 |
| `auto + zh` | Google Vision → cascade **Baidu** (nếu avg conf < 75%) |
| `auto + other` | Google Vision → 503 |

---

## Cost estimate

| Tier | Giá (CNY) | ~USD |
|---|---|---|
| ≤ 500 calls/ngày | Free | Free |
| 500-20K calls/tháng | ¥0.01/call | ~$0.0014/call |
| 20K-50K calls/tháng | ¥0.008/call | ~$0.0011/call |

MVP (30% of 600 req/ngày = ~180 Baidu calls/ngày): **Free tier đủ dùng**.

---

## Risks & Mitigation

| Risk | Mitigation |
|---|---|
| Đăng ký Baidu từ VN khó | VPN hoặc nhờ người ở TQ. Nếu không đăng ký được → PaddleOCR vẫn là fallback cho handwritten zh |
| Data gửi sang server TQ | Ảnh user là transient (không lưu server). Baidu terms cho phép commercial use. Nếu concern → dùng PaddleOCR thay thế |
| Token refresh race condition | Redis SET NX + TTL — chỉ 1 goroutine refresh, các goroutine khác chờ |
| Baidu API rate limit (10 QPS default) | MVP: 0.01 QPS → OK. Scale: mua thêm QPS package |
| Baidu down | Circuit breaker → fallback PaddleOCR → Google Vision |
