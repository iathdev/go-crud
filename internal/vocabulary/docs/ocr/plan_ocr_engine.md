# Plan — OCR Engine Integration

> **Quyết định công nghệ đã chốt:**
> - Printed (mọi ngôn ngữ) → **Google Cloud Vision**
> - Handwritten Chinese → **Baidu OCR API**
> - Handwritten các ngôn ngữ khác → **Google Cloud Vision**
> - Classification: **User-specified + Cascading**
>
> Tham khảo chi tiết: [`research_ocr_engine.md`](research_ocr_engine.md)

---

## 1. Xác định yêu cầu

### 1.1 Chức năng

**Core feature:** User chụp ảnh (vở ghi chép, sách giáo khoa) → hệ thống nhận diện Hán tự qua OCR → phân loại từ mới/từ đã có → user confirm → tạo flashcards.

**Ai dùng:**

| Loại user | Hành vi | Tần suất |
|---|---|---|
| **Học sinh (Free)** | Scan vở ghi chép sau giờ học. Chủ yếu handwritten Chinese | 2-3 scan/ngày (limit 3/ngày) |
| **Học sinh (Pro)** | Scan sách giáo khoa + vở. Mix printed + handwritten | 5-10 scan/ngày (limit 50/ngày) |
| **Content team (Admin)** | Không dùng OCR — import qua bulk endpoint | N/A |

**Input/Output:**

```
Input:  { image: base64/multipart, type: "printed"|"handwritten"|"auto", language: "zh"|"vi"|"en"|... }
Output: { items: [{ hanzi, confidence, candidates[] }], new_items[], existing_items[], low_confidence_items[] }
```

So với endpoint hiện tại (`POST /api/vocabularies/ocr-scan` nhận `{ items: [{ hanzi }] }`), plan này thêm:
- Nhận **image** thay vì danh sách hanzi đã extract sẵn
- Trả thêm **confidence** per character và **candidates** cho low-confidence items
- **Low-confidence items** tách riêng để mobile hiển thị "Did you mean X?"

### 1.2 Phi chức năng

| Tiêu chí | Target | Lý do |
|---|---|---|
| **Availability** | 99.5% (~3.6h downtime/tháng) | OCR không phải critical path — user vẫn import thủ công khi down. Dual-engine (Google + Baidu) tăng availability: 1 engine down → engine còn lại xử lý được (degraded mode, accuracy thấp hơn cho handwritten Chinese nếu Baidu down) |
| **Latency** | p50 < 1.5s, p99 < 3s | Budget: upload ~200-500ms + engine ~300-1500ms + post-processing ~100-300ms. Cascading fallback: p99 có thể lên 5-6s (chấp nhận — preview screen buffer thời gian chờ) |
| **Scalability** | MVP: 1K req/ngày. Target 12 tháng: 500K req/ngày | Cả Google Vision và Baidu OCR đều managed service → scale theo demand, không cần tự provision. Bottleneck: Go backend concurrent requests → semaphore pattern |
| **Consistency** | Eventual consistency | OCR stateless — không có shared mutable state. Duplicate check dùng DB read, stale vài ms chấp nhận được |
| **Durability** | Ảnh gốc: KHÔNG lưu server. Flashcards: durable (Postgres) | Server nhận ảnh → xử lý → discard. Chỉ flashcards persist |
| **Correctness** | Printed ≥ 90%, handwritten Chinese ≥ 80% | Sai số buffer bởi preview screen — user luôn confirm. **Tại sao Baidu cho handwritten Chinese:** Google Vision accuracy handwritten Chinese chỉ "trung bình", Baidu tốt nhất (PP-OCRv5 vượt GPT-4o). **Tại sao Google cho ngôn ngữ khác:** Baidu tối ưu cho Chinese, accuracy ngôn ngữ khác không đảm bảo |

**Trade-off:**

| Quyết định | Chọn | Thay vì | Tại sao |
|---|---|---|---|
| **Dual-engine (Google + Baidu)** vs **Single-engine (Google only)** | Dual-engine | Single → handwritten Chinese accuracy thấp | Baidu handwritten Chinese vượt trội. Cost tăng ~10-20% (chỉ handwritten Chinese mới gọi Baidu). Complexity tăng vừa phải (thêm 1 adapter + routing logic) |
| **Cascading + User-specified** vs **Parallel execution** | Cascading + User-specified | Parallel → 2x cost, 2x API calls | Cascading tiết kiệm cost (~1.1-1.2x). User-specified giảm cascading rate. Parallel chỉ tốt hơn accuracy ~5% nhưng gấp đôi chi phí |
| **Server-side OCR** vs **On-device OCR** | Server-side | On-device → khó control quality, fragmented device support | Server-side: consistent accuracy, dễ monitor, dễ switch engine. On-device: offline support nhưng accuracy thấp hơn, phụ thuộc device capability. **Phase 2:** xem xét on-device cho printed (Google ML Kit) để giảm latency |

---

## 2. Ước lượng quy mô

### Traffic

| Giai đoạn | MAU | Req/ngày | QPS avg | QPS peak (3x) | Google Vision calls | Baidu calls (~30% handwritten CN) |
|---|---|---|---|---|---|---|
| MVP (tháng 1-3) | 50-200 | 100-600 | ~0.01 | ~0.03 | ~70-420 | ~30-180 |
| Growth (tháng 4-12) | 1K-10K | 3K-50K | ~0.6-3.5 | ~2-10 | ~2.1K-35K | ~900-15K |
| Scale (năm 2) | 10K-50K | 50K-500K | ~3.5-35 | ~10-100 | ~35K-350K | ~15K-150K |

> **Tại sao ~30% handwritten Chinese:** Use case chính là học sinh scan vở ghi chép — đa số mixed (printed đề bài + handwritten ghi chú). Ước tính ~30% requests là handwritten Chinese cần Baidu. Con số này cần validate bằng production data.

### Storage

| Loại | Size/req | Lưu? | Ghi chú |
|---|---|---|---|
| Ảnh upload | 1-3 MB (raw), < 500KB (compressed) | Không | Xử lý trong memory → discard |
| OCR raw response | 2-10 KB | Không | Extract kết quả → discard |
| Flashcards | ~0.5-1 KB/từ, 10-30 từ/scan | Có (Postgres) | Shared với vocabulary table hiện tại |
| Idempotency cache | ~1-5 KB/entry, TTL 5 phút | Redis | Max ~50 concurrent × 5 min = ~250 entries |

→ OCR pipeline gần như không tạo thêm storage. Flashcards tái sử dụng vocabulary table.

### Bandwidth

Bottleneck: upload ảnh từ mobile. Giải pháp: client compress JPEG 70-80%, resize max 1500px → target < 500KB/ảnh.

| Hướng | Size/req | @ 50K req/ngày |
|---|---|---|
| Mobile → Server | < 500KB | ~25 GB/ngày |
| Server → Google/Baidu | < 500KB (forward) | ~25 GB/ngày |
| Response → Mobile | 1-5 KB | ~250 MB/ngày |

### Read/Write ratio

OCR pipeline: **1:3 → 1:10** (read:write). Mỗi scan = 1 read (`FindByHanziList`) + N writes (tạo flashcards ở bước confirm).

→ **Kết luận:** Quy mô MVP đơn giản — single Go instance + managed cloud APIs đủ xử lý. Không cần message queue, sharding, hay distributed processing cho đến khi vượt 50K req/ngày.

---

## 3. Định nghĩa interface

### 3.1 API Contract — OCR Scan (mới)

**Thay thế endpoint hiện tại** `POST /api/vocabularies/ocr-scan` (nhận hanzi list) bằng endpoint nhận image.

#### Alternative A: Endpoint nhận image trực tiếp (Recommend)

```
POST /api/vocabularies/ocr-scan
Content-Type: multipart/form-data
Headers: X-Idempotency-Key: {uuid}

Fields:
  image: file (JPEG/PNG, max 5MB)
  type: "printed" | "handwritten" | "auto"   (default: "auto")
  language: "zh" | "vi" | "en" | ...          (default: "zh")
```

**Response** `200`:
```json
{
  "success": true,
  "data": {
    "new_items": [
      {
        "hanzi": "新词",
        "confidence": 0.92,
        "candidates": []
      }
    ],
    "existing_items": [
      {
        "id": "uuid",
        "hanzi": "你好",
        "pinyin": "nǐ hǎo",
        "meaning_vi": "Xin chào",
        "meaning_en": "Hello",
        "hsk_level": 1,
        "confidence": 0.98,
        "candidates": []
      }
    ],
    "low_confidence_items": [
      {
        "hanzi": "鑫",
        "confidence": 0.65,
        "candidates": ["鑫", "森", "淼"]
      }
    ],
    "metadata": {
      "engine_used": "google_vision",
      "total_detected": 15,
      "processing_time_ms": 1234
    }
  }
}
```

**Ưu điểm:** Đơn giản — 1 request làm tất cả. Mobile không cần biết OCR engine.
**Nhược điểm:** Server phải handle file upload lớn. Latency cao hơn (upload + OCR + classify trong 1 request).

#### Alternative B: 2 endpoints tách biệt (OCR + Classify)

```
POST /api/vocabularies/ocr-extract     ← Mới: nhận image → trả raw characters
POST /api/vocabularies/ocr-scan        ← Giữ nguyên: nhận hanzi list → classify
```

**Ưu điểm:** Tách responsibility rõ ràng. Client có thể retry classify mà không cần re-upload. Giữ backward compatibility.
**Nhược điểm:** 2 round trips. Client phải orchestrate 2 calls. Latency tổng cao hơn.

**Recommend: Alternative A.** Tại sao: user experience quan trọng hơn — 1 tap = kết quả. Mobile không cần biết internal flow. Nếu cần retry classify, server tự xử lý trong cùng request. Backward compatibility không quan trọng vì endpoint hiện tại chưa có mobile integration.

### 3.2 Giao tiếp nội bộ

```
Handler (HTTP)
  │
  ▼
OCRCommand (Use Case) ── orchestrate ──┐
  │                                     │
  ▼                                     ▼
OCRServicePort ◄────────────    VocabularyRepositoryPort
(Output Port)                    (Output Port — đã có)
  │
  ├── GoogleVisionAdapter   (sync HTTP — official Go SDK)
  └── BaiduOCRAdapter       (sync HTTP — REST API call)
```

| Giao tiếp | Protocol | Lý do chọn |
|---|---|---|
| Mobile → Go backend | HTTP REST (multipart) | Đã có Gin router. Multipart cho file upload |
| Go backend → Google Vision | **Sync HTTP** (Go SDK `cloud.google.com/go/vision/v2`) | Official first-party SDK. Latency ~500ms-1.5s chấp nhận được |
| Go backend → Baidu OCR | **Sync HTTP** (REST API call) | Không có official Go SDK → HTTP client trực tiếp. Latency ~300ms-1s |
| OCR result → Classify | **In-process** (function call) | Cùng request context. Không cần async |

**Tại sao sync thay vì async (message queue)?**
- OCR là request-response pattern — user chờ kết quả real-time (1-3s)
- Message queue thêm complexity + latency mà không giải quyết vấn đề gì ở quy mô MVP
- Nếu cần async (VD: batch OCR 10 ảnh), Phase 2 thêm queue cho batch endpoint riêng

### 3.3 Port Interface mới

```go
// Output port — mới
type OCRServicePort interface {
    // ExtractCharacters gửi ảnh tới OCR engine, trả về danh sách ký tự + confidence
    ExtractCharacters(ctx context.Context, req OCRExtractRequest) (*OCRExtractResult, error)
}

type OCRExtractRequest struct {
    Image    []byte   // compressed image bytes
    Type     string   // "printed" | "handwritten" | "auto"
    Language string   // "zh" | "vi" | "en" | ...
}

type OCRExtractResult struct {
    Characters []OCRCharacter
    EngineUsed string  // "google_vision" | "baidu_ocr"
}

type OCRCharacter struct {
    Text       string    // detected character
    Confidence float64   // 0.0 - 1.0 (normalized)
    Candidates []string  // top-3 alternatives (nếu có)
}
```

---

## 4. Mô hình dữ liệu

### 4.1 Không cần schema mới

OCR pipeline là **stateless** — không persist dữ liệu OCR riêng. Flow:

```
Image (transient) → OCR engine → Characters (transient) → Classify → Response (transient)
                                                              │
                                                              ▼
                                                    vocabularies table (đã có)
```

Flashcards tạo từ OCR sử dụng **vocabulary table hiện tại** — không tạo table mới. Lý do: vocabulary là global shared data, không phân biệt nguồn gốc (OCR, manual, import).

### 4.2 Cache

| Loại cache | Ở đâu | TTL | Tại sao cần | Nếu không cache thì sao |
|---|---|---|---|---|
| **Idempotency cache** | Redis | 5 phút | User double-tap "Scan" hoặc mobile retry khi network flaky → cùng 1 ảnh gửi 2 lần → 2 API calls tới Google/Baidu (tốn tiền), 2 response có thể khác nhau (OCR non-deterministic). Cache đảm bảo: cùng request → cùng response, không tốn thêm cost. Key: `idem:{idempotency_key}`, Value: serialized response | Mỗi double-tap tốn thêm $0.0015 (Google) hoặc ¥0.01 (Baidu). Ở 500K req/ngày nếu 5% double-tap = 25K duplicate calls = ~$37.5/ngày lãng phí. Nhỏ nhưng hoàn toàn có thể tránh |
| **Baidu access token** | Redis hoặc in-memory | 29 ngày (token expires 30 ngày) | Baidu OAuth2 yêu cầu gọi `/oauth/2.0/token` để lấy access token trước mỗi API call. Token valid 30 ngày — gọi lại mỗi request là lãng phí 1 round trip (~200-500ms tới server Baidu TQ). Cache token = bỏ hoàn toàn round trip này cho 29/30 ngày | Mỗi OCR request thêm 200-500ms latency (token call) → tổng latency vượt 3s budget. Ở 15K Baidu calls/ngày (30% of 50K) = 15K round trips thừa/ngày. Ngoài ra, token endpoint có rate limit riêng — spam gọi có thể bị block |
| **OCR result cache** | **Không cache** | — | — | — |

**Tại sao không cache OCR results?**

| Lý do | Chi tiết |
|---|---|
| **Privacy** | Ảnh user là private data (vở ghi chép, bài tập). Lưu image hash trong Redis = fingerprinting — biết user X đã scan ảnh nào. Ngay cả hash cũng có thể dùng để correlate requests |
| **Hit rate cực thấp** | Mỗi ảnh chụp khác nhau (góc, ánh sáng, crop). Ước tính cache hit < 1%. Storage cost cho cache > tiền tiết kiệm được từ duplicate API calls |
| **OCR non-deterministic** | Cùng ảnh, cùng engine, kết quả có thể khác nhau giữa các lần gọi (Google Vision update model liên tục). Cache result cũ có thể kém hơn result mới |
| **Không side-effect** | OCR scan chỉ trả JSON — không tạo/sửa/xóa data. Duplicate call chỉ tốn tiền, không gây data inconsistency. Trade-off: chấp nhận duplicate cost thay vì privacy risk |

### 4.3 Indexing

Tái sử dụng index hiện tại. Cần verify:

- **`vocabularies.hanzi`** — cần index cho `FindByHanziList` (batch IN query). Nếu chưa có → thêm index.
- Không cần full-text index — OCR trả exact characters, matching bằng `IN (...)` clause.

#### Alternative A: Gin index trên `hanzi` column (Recommend)

```sql
CREATE INDEX idx_vocabularies_hanzi ON vocabularies (hanzi);
```

**Ưu điểm:** Đơn giản, đủ cho `WHERE hanzi IN (...)` queries.
**Nhược điểm:** Không cover fuzzy matching (chữ gần giống).

#### Alternative B: GIN trigram index cho fuzzy matching

```sql
CREATE EXTENSION pg_trgm;
CREATE INDEX idx_vocabularies_hanzi_trgm ON vocabularies USING gin (hanzi gin_trgm_ops);
```

**Ưu điểm:** Support fuzzy matching cho low-confidence characters (tìm chữ "gần giống").
**Nhược điểm:** Index lớn hơn, write overhead. Overkill cho MVP — OCR engine đã trả candidates.

**Recommend: Alternative A.** Tại sao: MVP chỉ cần exact matching. Low-confidence candidates do OCR engine trả về, không cần DB fuzzy search. Phase 2 xem xét trigram nếu cần server-side candidate generation.

---

## 5. Kiến trúc tổng quan

### 5.1 Data flow end-to-end

```
Mobile App                    Go Backend                              External Services
──────────                    ──────────                              ─────────────────

1. User chụp ảnh
   + chọn type
   (printed/handwritten/auto)
        │
        │  POST /api/vocabularies/ocr-scan
        │  multipart { image, type, language }
        │  Header: X-Idempotency-Key
        ▼
                              2. Middleware chain:
                                 Auth JWT → Rate limit
                                 (Free: 3/ngày, Pro: 50/ngày)
                                      │
                                      ▼
                              3. OCR Handler
                                 - Validate image (size ≤ 5MB, format)
                                 - Check idempotency cache (Redis)
                                 - If cached → return cached response
                                      │
                                      ▼
                              4. OCRCommand (Use Case)
                                 - Routing logic:
                                      │
                          ┌───────────┼───────────────┐
                          │           │               │
                    type="printed"  type="handwritten" type="auto"
                          │           │               │
                          ▼           ▼               ▼
                    5a. Google    5b. Route by     5c. Google Vision
                        Vision       language:         (primary)
                          │        ┌────┴────┐         │
                          │     lang="zh"  other    confidence
                          │        │        │       ≥ 75%? ──► Return
                          │        ▼        ▼          │
                          │     Baidu    Google      < 75%
                          │     OCR      Vision        │
                          │        │        │          ▼
                          │        │        │     5d. Baidu OCR
                          │        │        │     (fallback, if lang=zh)
                          │        │        │          │
                          └────────┴────────┴──────────┘
                                      │
                                      ▼
                              6. Post-process:
                                 - Filter: chỉ giữ CJK characters
                                   (Unicode range U+4E00–U+9FFF)
                                 - Normalize confidence (Google: 0-1,
                                   Baidu: 0-100 → 0-1)
                                 - Classify by confidence:
                                   ≥ 80% → confirmed
                                   70-80% → suggest ("Did you mean?")
                                   < 70% → low_confidence (top-3)
                                      │
                                      ▼
                              7. VocabularyRepo
                                 FindByHanziList(confirmed hanzi)     ──► PostgreSQL
                                      │
                                      ▼
                              8. Classify:
                                 new_items / existing_items / low_confidence_items
                                      │
                                      ▼
                              9. Cache response (Redis, TTL 5min)
                                 Return JSON
                                      │
        ◄─────────────────────────────┘
        │
10. Preview screen
    - User confirm/edit/delete
    - Low confidence: pick from candidates
        │
        │  POST /api/vocabularies (bulk create confirmed)
        │  POST /api/folders/:id/vocabularies (add to folder)
        ▼
                              11. Existing CRUD flow
```

### 5.2 Các thành phần chính

```
┌─────────────────────────────────────────────────────────────┐
│                        Go Backend                           │
│                                                             │
│  ┌──────────┐   ┌──────────────┐   ┌─────────────────────┐ │
│  │   Gin    │──►│  OCR Handler │──►│    OCRCommand        │ │
│  │  Router  │   │              │   │    (Use Case)        │ │
│  │          │   │  - validate  │   │                      │ │
│  │ Middleware│   │  - idem     │   │  - routing logic     │ │
│  │ - Auth   │   │    check    │   │  - confidence merge  │ │
│  │ - Rate   │   │              │   │  - post-process      │ │
│  │   Limit  │   └──────────────┘   │  - classify          │ │
│  └──────────┘                      └──────┬──────┬────────┘ │
│                                           │      │          │
│                              ┌────────────┘      └────┐     │
│                              ▼                        ▼     │
│                   ┌────────────────┐      ┌──────────────┐  │
│                   │ OCRServicePort │      │ VocabRepo    │  │
│                   │ (Output Port)  │      │ (Output Port)│  │
│                   └───┬────────┬───┘      └──────┬───────┘  │
│                       │        │                 │          │
│              ┌────────┘        └────────┐        │          │
│              ▼                          ▼        │          │
│  ┌─────────────────────┐  ┌──────────────────┐   │          │
│  │ GoogleVisionAdapter │  │ BaiduOCRAdapter  │   │          │
│  │                     │  │                  │   │          │
│  │ - official Go SDK   │  │ - HTTP REST      │   │          │
│  │ - circuit breaker   │  │ - circuit breaker│   │          │
│  │ - retry (1x)        │  │ - token mgmt     │   │          │
│  └──────────┬──────────┘  └────────┬─────────┘   │          │
│             │                      │             │          │
└─────────────┼──────────────────────┼─────────────┼──────────┘
              │                      │             │
              ▼                      ▼             ▼
     Google Cloud Vision      Baidu OCR API    PostgreSQL
              │                      │             │
              │                      │             │
         ┌────┘                      │             │
         ▼                           │             │
       Redis ◄───────────────────────┘             │
       - idempotency cache                         │
       - Baidu access token                        │
       - rate limit counters                       │
```

**Scale strategy:**

| Component | Scale độc lập? | Cách scale |
|---|---|---|
| Go Backend | Có | Horizontal — multiple instances behind LB. Stateless |
| Google Vision | Managed | Google tự scale. Client-side: connection pool + semaphore |
| Baidu OCR | Managed | Baidu tự scale. Default QPS 10 → mua thêm khi cần |
| PostgreSQL | Có | Read replicas khi read-heavy. Hiện tại single instance đủ |
| Redis | Có | Single instance đủ cho MVP. Cluster khi cần |

---

## 6. Xử lý các bài toán khó

### 6.1 Fault Tolerance

**Circuit Breaker** (gobreaker v2 — `infrastructure/circuitbreaker/`):

Mỗi OCR engine có circuit breaker riêng:

| Config | Google Vision CB | Baidu OCR CB |
|---|---|---|
| `MaxRequests` (half-open) | 3 | 3 |
| `Interval` (counter reset) | 60s | 60s |
| `Timeout` (open → half-open) | 30s | 30s |
| `ReadyToTrip` | 5 consecutive failures HOẶC > 50% failure rate trong 10 req | 5 consecutive failures HOẶC > 50% failure rate trong 10 req |

**Failure scenarios & recovery:**

| Scenario | Xử lý |
|---|---|
| **Google Vision down + type="printed"** | CB open → trả lỗi 503. User retry sau 30s. Không fallback sang Baidu (Baidu không tốt hơn cho printed) |
| **Google Vision down + type="auto"** | CB open → route thẳng sang Baidu (nếu lang=zh). Nếu lang khác → trả lỗi 503 |
| **Baidu OCR down + type="handwritten" + lang=zh** | CB open → fallback sang Google Vision (accuracy thấp hơn cho handwritten Chinese, ~trung bình thay vì tốt nhất, nhưng vẫn trả kết quả). Response kèm warning: `"engine_degraded": true` |
| **Cả 2 engine down** | Trả 503 "OCR service unavailable". User vẫn import thủ công |

**Retry policy:**

| Engine | Retry | Backoff | Timeout |
|---|---|---|---|
| Google Vision | 1 lần | 500ms | 3s per attempt |
| Baidu OCR | 1 lần | 500ms | 3s per attempt |

Tại sao chỉ retry 1 lần: latency budget 1-3s. Retry 2+ lần → vượt budget. Circuit breaker bảo vệ khỏi cascade failure.

### 6.2 Rate Limiting & Spike

**Spike pattern:** 18h-22h homework time → 3-5x average. Trước kỳ thi → burst 10x.

| Layer | Config |
|---|---|
| **Per-user per-day** | Redis: `ocr:{user_id}:{date}` counter. Free: 3, Pro: 50. TTL: 24h |
| **Global concurrent** | Semaphore: `OCR_MAX_CONCURRENT=50`. Vượt → HTTP 429 |
| **Google Vision quota** | 1,800 req/phút (default). Monitor → request increase khi đạt 70% |
| **Baidu OCR QPS** | Default 10 QPS (rất thấp!). Mua package QPS khi cần. MVP: 10 QPS đủ cho ~30% of 0.03 QPS peak |

**Baidu QPS 10 — có đủ không?**
- MVP peak: ~0.03 QPS × 30% handwritten Chinese = ~0.01 QPS → đủ thừa
- Growth peak: ~10 QPS × 30% = ~3 QPS → vẫn đủ
- Scale peak: ~100 QPS × 30% = ~30 QPS → cần mua thêm QPS package

### 6.3 Deduplication

| Scenario | Giải pháp |
|---|---|
| **Double-tap** | Client debounce 300ms + server idempotency key. Redis cache `idem:{key}` TTL 5 phút |
| **Cùng ảnh, khác thời điểm** | Không dedup — OCR stateless, chỉ trả JSON. Flashcard dedup ở vocabulary layer (`FindByHanziList`) |

### 6.4 Edge Cases

| Case | Xử lý |
|---|---|
| **Image > 5MB** | Handler reject 413. Client nên compress trước upload |
| **Image không phải ảnh** | Handler validate Content-Type. Reject 400 |
| **0 characters detected** | Response: `{ new_items: [], existing_items: [], low_confidence_items: [], metadata: { total_detected: 0 } }` + message "No Chinese characters detected" |
| **Image mờ / quality kém** | Nếu avg confidence < 50% → response kèm warning "Image quality too low, please retake" |
| **Mixed printed + handwritten trên cùng 1 ảnh** | type="auto" → Google Vision scan toàn bộ → nếu avg confidence < 75% → Baidu rescan (cascading). Baidu có thể detect handwritten tốt hơn cho phần viết tay |
| **Baidu response format khác Google** | Normalize trong adapter: Baidu confidence 0-100 → chia 100 thành 0-1. Cả 2 adapter trả cùng `OCRExtractResult` struct |
| **Baidu access token expired** | Adapter tự refresh trước khi call. Cache token trong Redis TTL 29 ngày. Nếu refresh fail → circuit breaker |

### 6.5 Confidence Normalization

2 engine có thang confidence khác nhau:

| Engine | Thang gốc | Normalize | Ghi chú |
|---|---|---|---|
| Google Vision | 0.0 - 1.0 (float) | Giữ nguyên | Confidence ở text block level, không per-character |
| Baidu OCR | 0 - 100 (int) | ÷ 100 → 0.0 - 1.0 | Confidence per line |

**Vấn đề:** Google trả confidence per text block, Baidu per line → granularity khác nhau. Normalize: lấy min confidence trong block/line → assign cho mỗi character trong block/line đó. Conservative approach — thà flag false positive (low confidence khi thực tế đúng) hơn là miss false negative.

---

## 7. Giám sát & Vận hành

### 7.1 Metrics

**Business metrics:**

| Metric | Alert khi |
|---|---|
| `ocr.requests.total` (by engine, status) | Error rate > 5% trong 5 phút |
| `ocr.accuracy.implicit` (tính từ user edits ở preview) | Drop > 10% so với 7-day baseline |
| `ocr.daily_usage.per_user` | > 100/ngày (possible abuse) |

**Infrastructure metrics (OpenTelemetry):**

| Metric | Alert khi |
|---|---|
| `ocr.latency.p50` / `p99` | p50 > 2s hoặc p99 > 5s |
| `ocr.engine.{google,baidu}.latency` | p99 > 2s |
| `ocr.circuit_breaker.{google,baidu}.state` | State = open |
| `ocr.concurrent.active` | > 80% of semaphore (40/50) |
| `ocr.cascading.fallback_rate` | > 30% (primary engine có vấn đề) |
| `ocr.cost.google_estimated` | > 80% monthly budget |

### 7.2 Alerting

| Severity | Condition | Action |
|---|---|---|
| **P1** | CB open > 5 phút HOẶC error rate > 20% | PagerDuty. Auto-fallback nếu có engine còn sống |
| **P2** | p99 > 5s HOẶC fallback rate > 30% HOẶC cost > 80% budget | Slack #ocr-alerts. Investigate 4h |
| **P3** | Accuracy drop > 10% HOẶC daily requests > 80% quota | Slack #ocr-monitoring. Review sprint planning |

### 7.3 Accuracy Reconciliation

**Implicit feedback loop:** User edit kết quả OCR ở preview screen → log edits → tính:

```
actual_accuracy = 1 - (edits / total_detected)
```

Breakdown by: engine, confidence bucket, image source. Weekly report so sánh implicit accuracy vs engine confidence → detect calibration drift.

---

## 8. Implementation Phases

### Phase 1: Google Vision only (MVP)

| Task | Mô tả |
|---|---|
| `OCRServicePort` interface | Định nghĩa output port trong `application/port/outbound.go` |
| `GoogleVisionAdapter` | Implement adapter trong `adapter/external/`. Dùng official Go SDK. Circuit breaker wrap |
| Update `OCRCommand` | Thêm `OCRServicePort` dependency. Nhận image → call engine → post-process → classify |
| Update `OCR Handler` | Đổi từ JSON body sang multipart. Thêm idempotency check |
| Update DTOs | Thêm confidence, candidates, low_confidence_items, metadata |
| Update `module.go` | Wire `GoogleVisionAdapter` → `OCRCommand` |
| Config | Thêm `GOOGLE_VISION_*` env vars |
| Hanzi index | Verify/add `idx_vocabularies_hanzi` |

### Phase 2: Thêm Baidu OCR

| Task | Mô tả |
|---|---|
| `BaiduOCRAdapter` | Implement adapter. HTTP REST client + OAuth2 token management |
| `OCRRouter` | Routing logic trong use case: type + language → chọn engine. Cascading fallback |
| Confidence normalization | Baidu 0-100 → 0-1. Merge results khi cascading |
| Config | Thêm `BAIDU_OCR_*` env vars |

### Phase 3: Observability & Hardening

| Task | Mô tả |
|---|---|
| Metrics | OpenTelemetry spans + custom metrics cho OCR pipeline |
| Accuracy tracking | Log user edits → implicit accuracy dashboard |
| Rate limiting per-user per-day | Redis counter cho OCR-specific limits |
