# OCR Engine Research — Vocabulary Module

> Research phục vụ quyết định OCR engine cho feature OCR Scan Hán tự → Auto Flashcards.

---

## 1. Yêu cầu từ PRD

- Target accuracy: ≥ 90% printed (chữ in), ≥ 80% handwritten (chữ viết tay)
  - VD: ảnh chụp có 100 chữ Hán → printed phải nhận đúng ≥ 90 chữ, handwritten phải nhận đúng ≥ 80 chữ
- Latency budget: 1-3s cho toàn bộ OCR processing
- Confidence < 80% → "Did you mean X?" + top-3 candidates
- Confidence < 70% → show top-3 cho user chọn
- Lọc chỉ lấy Hán tự từ mixed content (CN + VN + EN)
- Use case: học sinh chụp vở ghi chép — thường có **cả chữ in (bài tập, đề bài) lẫn chữ viết tay (ghi chép)** trên cùng 1 trang

**Các lựa chọn engine từ PRD:**
- Google Cloud Vision API (primary)
- Baidu OCR API (fallback cho handwritten)
- Tesseract (open-source) fine-tuned trên dataset viết tay tiếng Trung

### 1.2 Yêu cầu phi chức năng

| Tiêu chí | Target | Lý do |
|---|---|---|
| **Availability** | 99.5% (~3.6h downtime/tháng) | OCR không phải critical path — nếu OCR down, user vẫn import thủ công. Nhưng downtime quá nhiều ảnh hưởng trải nghiệm Trụ 1 |
| **Latency** | p50 < 1.5s, p99 < 3s (end-to-end: upload ảnh → trả kết quả) | PRD quy định 1-3s. Budget breakdown: network upload ~200-500ms, OCR engine ~300-1500ms, post-processing + DB lookup ~100-300ms |
| **Scalability** | MVP: 1K req/ngày (~50 users). Target 12 tháng: 500K req/ngày (~50K MAU) | Từ bảng chi phí section 3. Architecture phải không cần redesign khi scale 500x |
| **Consistency** | Eventual consistency chấp nhận được | OCR là stateless operation — không có shared mutable state. Duplicate detection dùng DB read → có thể stale vài ms, chấp nhận được (worst case: user thấy "từ mới" nhưng thực tế đã có → handled ở preview screen) |
| **Durability** | Ảnh gốc: KHÔNG lưu server-side (mobile-only). OCR results → flashcards: durable trong DB (backed by Postgres WAL + daily backup) | Ảnh gốc thuộc về user device. Server chỉ nhận ảnh → xử lý → trả kết quả → discard. Flashcards tạo từ OCR phải persistent |
| **Correctness** | Printed ≥ 90%, handwritten ≥ 80% (character-level accuracy). Confidence threshold: < 80% → suggest, < 70% → manual pick | Sai số OCR được buffer bởi preview screen — user luôn confirm trước khi tạo flashcard. Nên correctness ở mức 80-90% + human review là chấp nhận được |

**Trade-off chính:**
- **Latency vs Correctness:** Cascading fallback tăng accuracy nhưng tăng latency cho 10-20% requests (p99 có thể lên 5-6s). Chọn: chấp nhận p99 cao hơn vì preview screen buffer thời gian chờ.
- **Cost vs Availability:** Multi-engine fallback tăng availability nhưng tăng cost. Chọn: cascading (chỉ gọi engine 2 khi engine 1 fail/low confidence) thay vì parallel.
- **Simplicity vs Scalability:** Google Cloud Vision zero-ops nhưng cost tuyến tính. PaddleOCR rẻ hơn ở scale lớn nhưng phức tạp ops. Chọn: Google Vision cho MVP, migration path sang PaddleOCR khi vượt 50K req/ngày.

---

## 2. So sánh 4 Engine

| Factor | Google Cloud Vision | Baidu OCR API | PaddleOCR (self-hosted) | Tesseract |
|---|---|---|---|---|
| **Printed Chinese (recognition)** | Cao | Rất cao | Rất cao (cùng model PP-OCRv5) | Trung bình |
| **Handwritten Chinese (recognition)** | Trung bình | **Tốt nhất** (PP-OCRv5 vượt GPT-4o) | **Tốt nhất** (cùng model) | Rất kém — gần như không dùng được |
| **Confidence granularity** | **Per-symbol** (per-character native) | **Per-character** (cần `recognize_granularity=small`) | **Per-line only** — không hỗ trợ per-character | Per-word |
| **Pricing** | $1.50/1K req (free 1K/tháng) | ~5K free/tháng, packages tính CNY | Free (open-source) | Free (open-source) |
| **Go SDK** | Official (first-party) | Không có — community only | Không — Python ecosystem, cần sidecar service | cgo wrapper (gosseract) |
| **Docs** | English | Tiếng Trung | English + Tiếng Trung | English |
| **Đăng ký từ VN** | Dễ (Google account) | Khó — UI tiếng Trung, OTP quốc tế hay fail | N/A (self-hosted) | N/A |
| **Data residency** | US/EU (global endpoint) | Server Trung Quốc — data privacy concern | Server của mình — không lo privacy | Server của mình |
| **Latency từ VN** | ~500ms-1.5s | ~300ms-1s | Local (100ms-2s) | Local (100ms-2s) |
| **Ops overhead** | Không — managed service | Không — managed service | Cao — cần deploy, scale, monitor Python service | Trung bình — cần install C++ libs |
| **Fine-tune** | Không | Không | Có thể fine-tune trên PaddlePaddle framework | Hàng tuần, effort cực lớn, kết quả không chắc |

> **⚠️ Accuracy ≠ Phù hợp cho bài toán project.** Bảng trên đánh giá **text recognition** (đọc đúng chữ). Bài toán project cần thêm **per-character classification** (phân loại từng chữ: confirmed / low_confidence) để tạo flashcard. PaddleOCR recognition accuracy "Rất cao" nhưng thiếu per-character confidence → classify sai. Ví dụ: line "你好鑫" score 0.75 → cả 3 character đều nhận 0.75 → nhưng thực tế "你好" ~0.95, "鑫" ~0.50. Google Vision / Baidu trả per-character confidence → classify chính xác hơn. **Kết luận: PaddleOCR chỉ phù hợp dev/evaluate, production phải dùng cloud APIs.**
>
> **PaddleOCR confidence limitation**: PaddleOCR chỉ trả confidence per text line (recognition score của cả dòng). Khi segment thành nhiều words/characters, mỗi word nhận cùng confidence của line — không phản ánh độ chính xác per-character. Không có workaround khả thi qua API hiện tại (cắt ảnh từng character rồi recognize riêng chậm gấp N lần). Google Vision hỗ trợ per-symbol confidence native. Baidu OCR API cùng model PP-OCRv5 với PaddleOCR nhưng thêm server-side post-processing layer — khi set `recognize_granularity=small`, cloud service tách line result thành per-character confidence (value-add của cloud, không phải feature model gốc). Đây là lý do chính để dùng cloud APIs cho production.

> **PaddleOCR vs Baidu OCR API:** Cùng gốc Baidu, cùng model PP-OCRv5, accuracy tương đương. Khác biệt: PaddleOCR là open-source self-hosted (free, không lo data privacy) còn Baidu OCR API là cloud service (data gửi sang server TQ, trả phí). Trade-off của PaddleOCR: Python ecosystem → cần chạy như sidecar service (gRPC/HTTP) bên cạnh Go backend + tự chịu ops.

### Đánh giá từng engine

**Tesseract** — Loại cho handwritten. Accuracy chữ viết tay Hán tự gần như không dùng được. Fine-tune cần hàng nghìn mẫu labeled, tốn hàng tuần, kết quả vẫn kém xa cloud API. Chỉ viable cho printed text clean, nhưng vẫn thua xa Google/Baidu/PaddleOCR.

**Baidu OCR API** — Accuracy handwritten tốt nhất, nhưng friction cao: docs tiếng Trung, không official Go SDK, đăng ký overseas không ổn định, data gửi sang server TQ. Không có lợi thế gì so với PaddleOCR (cùng model) ngoài việc không cần tự host.

**PaddleOCR** — Accuracy ngang Baidu API (cùng PP-OCRv5), free, không lo data privacy. Có thể fine-tune thêm. Nhưng ops overhead cao: deploy Python service, tự scale, tự monitor. Phù hợp nếu team sẵn sàng chịu ops cost.

**Google Cloud Vision** — Lựa chọn thực tế nhất cho MVP: official Go SDK, docs English, dễ integrate, zero ops, accuracy printed cao, handwritten "đủ dùng", latency trong budget 1-3s.

---

## 3. Chi phí dài hạn — Ảnh hưởng của tần suất sử dụng

### 3.0 Ước lượng quy mô

**Traffic:**

| Giai đoạn | MAU | Scan/user/ngày | Req/ngày | QPS trung bình | QPS đỉnh (3x) |
|---|---|---|---|---|---|
| MVP (tháng 1-3) | 50-200 | 2-3 | 100-600 | ~0.01 | ~0.03 |
| Growth (tháng 4-12) | 1K-10K | 3-5 | 3K-50K | ~0.6-3.5 | ~2-10 |
| Scale (năm 2) | 10K-50K | 5-10 | 50K-500K | ~3.5-35 | ~10-100 |
| Target (năm 3+) | 50K-100K | 5-10 | 250K-1M | ~17-70 | ~50-200 |

> Ghi chú: QPS đỉnh ước tính 3x average (homework time 18h-22h chiếm ~60% traffic). Mỗi scan request = 1 ảnh.

**Storage:**

| Loại dữ liệu | Kích thước/request | Lưu? | Storage/tháng (@ 500K req/ngày) |
|---|---|---|---|
| Ảnh gốc (upload) | 1-3 MB | **Không** — xử lý xong discard | 0 (transient only, xử lý trong memory/temp) |
| OCR raw response (JSON) | 2-10 KB | **Không** — extract Hanzi rồi discard | 0 |
| Flashcards tạo từ OCR | ~0.5-1 KB/từ, ~10-30 từ/scan | **Có** — Postgres | ~75-450 GB/năm (worst case, shared với toàn bộ vocabulary data) |

> Ảnh không lưu server-side → storage cost gần như 0 cho OCR pipeline. Storage chính là flashcard data trong Postgres.

**Bandwidth:**

| Hướng | Kích thước | @ 500K req/ngày |
|---|---|---|
| Upload (mobile → server) | 1-3 MB/ảnh | ~500 GB - 1.5 TB/ngày |
| Server → OCR engine (nếu cloud) | 1-3 MB/ảnh (forward) | ~500 GB - 1.5 TB/ngày |
| OCR engine → Server (response) | 2-10 KB | ~1-5 GB/ngày |
| Server → mobile (response) | 1-5 KB | ~0.5-2.5 GB/ngày |

> **Bottleneck:** Upload bandwidth. Cần image compression ở mobile trước khi upload (target < 500KB sau compress, giảm 3-6x). Google Vision chấp nhận ảnh ≤ 20MB, recommend ≤ 1500px longest edge.

**Read/Write ratio:**

OCR pipeline là **write-heavy**: mỗi scan = 1 read (duplicate check: `FindByHanziList`) + N writes (tạo flashcards). Ước tính **1:3 đến 1:10** (read:write) tùy số từ mới/scan.

> Khác biệt với vocabulary module tổng thể (read-heavy: browse, search, review). OCR chỉ là 1 input channel — sau khi flashcard tạo xong, traffic chuyển sang read-heavy learning modes.

### 3.1 Pricing chi tiết từng engine

**Google Cloud Vision API:**

| Volume/tháng | Giá / 1K requests |
|---|---|
| ≤ 1,000 | Free |
| 1,001 - 5,000,000 | $1.50 |
| 5,000,001+ | $0.60 |

- Default QPS: 1,800 req/phút (30 RPS)
- Chi phí ẩn: network egress fees, Cloud Storage nếu dùng batch mode

**Baidu OCR API — Handwriting Recognition (pay-per-use):**

| Volume/tháng | CNY/call |
|---|---|
| ≤ 20,000 | 0.0100 |
| 20,001 - 50,000 | 0.0080 |
| 50,001 - 100,000 | 0.0065 |
| 100,001 - 200,000 | 0.0055 |
| 200,001 - 300,000 | 0.0050 |
| 300,000+ | 0.0045 |

Free tier: 500-1,000 call/tháng. Prepaid packages rẻ hơn (VD: 5M calls = 19,000 CNY = ~$2,639).

**PaddleOCR (self-hosted):**

Software free. Chi phí = infrastructure. Performance benchmark (PP-OCRv5, NVIDIA T4 GPU):

| Model | GPU latency/ảnh | CPU latency/ảnh | Throughput (GPU) |
|---|---|---|---|
| Server (accuracy cao) | ~100-120ms | ~400ms | ~8-14 ảnh/giây/GPU |
| Mobile (lightweight) | ~16ms | ~80ms | ~60-125 ảnh/giây/GPU |

**Tesseract (self-hosted):**

Software free. Chi phí = infrastructure.
- ~1-7.5 giây/ảnh (tùy complexity), trung bình ~0.5-1 ảnh/giây/core
- RAM: ~100-300 MB/process
- CJK accuracy: ~82% — kém hơn PaddleOCR (91-92%)

### 3.2 So sánh chi phí theo quy mô (USD/tháng)

| Quy mô | Users ước tính | Google Vision | Baidu Handwriting | PaddleOCR (self-hosted) | Tesseract (self-hosted) |
|---|---|---|---|---|---|
| **1K req/ngày** | ~50 | **$43** | ~$39 | ~$124 | ~$124 |
| **10K req/ngày** | ~500 | $449 | **~$253** | ~$384 | ~$750 |
| **100K req/ngày** | ~5,000 | $4,499 | ~$1,826 | **~$768-$1,152** | ~$6,250 |
| **500K req/ngày** | ~50,000 (target MAU) | $13,499 | ~$7,917 | **~$3,072-$3,840** | N/A |
| **1M req/ngày** | ~100,000 | **$21,000** | ~$13,500 | **~$5,000-$6,000** | N/A |

> Ghi chú: PaddleOCR/Tesseract tính theo AWS EC2 (g4dn.xlarge cho GPU ~$384/tháng, c5.xlarge cho CPU ~$124/tháng). Baidu tính prepaid packages cho 500K/ngày.

### 3.3 Điểm giao cắt chi phí (crossover point)

```
Chi phí/tháng ($)

14,000 |                                          G ──────
       |                                     ·····
12,000 |                                ····
       |                           ····
10,000 |                      ····                  B ─────
       |                 ····                  ·····
 8,000 |            ····                  ····
       |       ····                  ····
 6,000 |  ····                  ····          T ─── (không viable)
       |                   ····
 4,000 |              ····                          P ─────
       |         ····                          ·····
 2,000 |    ····                          ····
       |····                         ····
     0 └──────────────────────────────────────────────────
       1K      10K      50K     100K    300K    500K     1M  req/ngày

       G = Google Cloud Vision    B = Baidu OCR API
       P = PaddleOCR (self-hosted) T = Tesseract
```

**Nhận xét:**
- **< 10K req/ngày:** Cloud API (Google/Baidu) rẻ hơn self-hosted vì không phải trả server cost
- **~10K-50K req/ngày:** Baidu rẻ nhất trong cloud APIs. PaddleOCR bắt đầu cạnh tranh
- **> 50K req/ngày:** PaddleOCR self-hosted rẻ nhất — gap càng lớn khi scale lên
- **500K req/ngày:** Google Vision đắt gấp ~3.5x PaddleOCR, gấp ~1.7x Baidu
- **1M req/ngày:** Google Vision cao nhất (~$21K/tháng) — gấp ~4x PaddleOCR (~$5-6K), gấp ~1.6x Baidu (~$13.5K). Pricing per-request tuyến tính, không có economy of scale mạnh (chỉ giảm từ $1.50 → $0.60/1K khi vượt 5M/tháng, sau đó giữ nguyên)
- **Tesseract:** Đắt hơn PaddleOCR ở mọi quy mô (throughput thấp hơn 10-60x → cần nhiều server hơn) và accuracy kém hơn → không có lý do chọn

### 3.4 Chi phí ẩn cần lưu ý

| Engine | Chi phí ẩn |
|---|---|
| **Google Vision** | Network egress fees, Cloud Storage (batch mode), chi phí tăng tuyến tính — không có economy of scale mạnh |
| **Baidu OCR** | Cả successful lẫn failed calls đều tính tiền. QPS mặc định chỉ 10 (mua thêm QPS tốn phí). Payment bằng CNY |
| **PaddleOCR** | DevOps time (deploy, monitor, scale, upgrade model). GPU spot instances rẻ hơn ~60-70% nhưng có thể bị interrupt. Cần team biết Python infra |
| **Tesseract** | CPU-bound → scale kém. Cần rất nhiều instances cho volume lớn. Maintenance effort cho CJK models |

---

## 4. Các app tương tự đang dùng gì?

### Chinese dictionary / learning apps

| App | OCR Engine | Handwritten? | On-device / Cloud | Ghi chú |
|---|---|---|---|---|
| **Pleco** (Chinese dictionary, phổ biến nhất) | Licensed engine (template-matching). iOS đang thêm **Apple Vision** | Printed only. Experimental neural HWR | On-device | Camera live feed. Struggle với calligraphy, chữ nhỏ/mờ |
| **有道词典 / Youdao** (NetEase) | **Self-developed** (proprietary). 97% printed, ~90% stylized | **Có** — handwritten, rotated, curved text | Both (offline + cloud) | Offline neural network chạy nhanh hơn 20% vs phiên bản trước. API tại ai.youdao.com |
| **HanYou / Yomiwa** | **Custom in-house** (6+ năm phát triển) | Printed + handwriting input | On-device, offline | 13,000+ characters. Detect trong fraction of a second |
| **Hanping Chinese Camera** | Custom/proprietary. 99.5% simplified, 98.7% traditional | Printed only | On-device, offline | 6,703 simplified + 5,401 traditional characters |
| **金山词霸 / Kingsoft** | Không công bố | Có screen/photo OCR từ 2015 | Không rõ | Cải thiện 50% accuracy, 30% speed trong bản 2015 |
| **Trainchinese** | Không công bố | Có photo capture cho dictionary lookup + flashcard import | Không rõ | |
| **Quizlet** (flashcard) | Không công bố | Printed + handwritten notes | Cloud-based (Pro only) | Không tối ưu cho Chinese |
| **Anki** (flashcard plugin) | **Tesseract** (via AnkiOCR) | Printed only — "not handwritten" | On-device | ~80% accuracy Chinese. Compound characters thường sai |

> **HelloChinese, SuperChinese, ChineseSkill, LingoDeer, Du Chinese, The Chairman's Bao** — không có camera OCR/scan feature.
> **百词斩, 墨墨背单词, 不背单词** — không có OCR. Tập trung vào spaced repetition / flashcard.
> **Skritter** — không có OCR. Tích hợp Pleco/Hanping cho word lookup.

### Các công ty TQ lớn — self-developed OCR

| Công ty | Engine | Handwritten? | Đặc điểm |
|---|---|---|---|
| **Baidu** (百度) | **PaddleOCR** (open-source, PP-OCRv5) | **Có** — best-in-class. CNN+Transformer, 95%+ | ~3.5MB mobile model. 100+ ngôn ngữ. Apache 2.0 license |
| **Tencent** (腾讯) | **Tencent Youtu Lab** (self-developed) | **Có** — handwriting >80%, single char <15ms | WeChat built-in OCR. Cloud API tại cloud.tencent.com |
| **NetEase Youdao** (网易有道) | **Self-developed** | **Có** — 97% printed, ~90% special text | Offline + cloud. Dictionary Pen hardware 98.3% accuracy |
| **iFlytek** (科大讯飞) | **Self-developed** (deep neural network) | **Có** — education focus (answer sheet grading) | Handwriting OCR cloud-only. Offline SDK chỉ cho speech |
| **Hanwang** (汉王) | **Self-developed** (pattern recognition pioneer) | **Có** — core competency, free-writing mode | Mobile offline SDK + server private deployment |

### OCR engines phổ biến trên mobile

| Engine | Chinese accuracy | Handwritten | Mobile size | Ghi chú |
|---|---|---|---|---|
| **Google ML Kit v2** | Có (cần Chinese script dependency ~4MB) | Không cho camera OCR. Digital Ink Recognition riêng (stroke input) | ~4MB | Cross-platform. Reported issues mixed CJK+Latin. Chinese chỉ có "Accurate" mode |
| **Apple Vision** | Simplified Chinese only | Không tối ưu cho handwritten | Built-in iOS | iOS only. "Accurate" path 6x chậm hơn ML Kit. Không có Fast path cho Chinese |
| **PaddleOCR Mobile** | **Excellent** — primary focus, 91-96% | **Có** (PP-OCRv5) | **~3.5MB** | iOS/Android via Paddle Lite. Best open-source cho Chinese |
| **EasyOCR** | Good (~82.8%) | Limited | ~50+ MB | Python-based. Nặng hơn PaddleOCR nhiều |
| **Tesseract** | Poor (~38-85%, varies greatly) | Rất kém | ~15+ MB | Không app serious nào dùng cho Chinese |

### Nhận xét

- **Không app Chinese learning/dictionary nào dùng Google Cloud Vision API** — đa số dùng on-device engine hoặc self-developed
- **Các công ty TQ lớn đều tự build OCR engine** (Baidu, Tencent, NetEase, iFlytek, Hanwang) — chỉ Baidu open-source (PaddleOCR)
- **Youdao** (NetEase) là reference tốt nhất cho use case tương tự: self-developed, offline + cloud, 97% printed, hỗ trợ handwritten, có dictionary pen hardware
- **PaddleOCR** là lựa chọn open-source mạnh nhất: 3.5MB mobile, offline, accuracy cao, cả printed + handwritten, 100+ ngôn ngữ
- **Google ML Kit v2** hỗ trợ Chinese nhưng không phải thế mạnh, **không có camera-based handwriting OCR**
- **Tesseract** chỉ xuất hiện trong Anki plugin — accuracy kém nhất, không app nào serious dùng cho Chinese

---

## 5. Cơ chế Classification khi dùng nhiều model

Khi sử dụng > 1 OCR model, cần cơ chế quyết định model nào xử lý request nào. Có 4 cơ chế phổ biến:

### 5.1 Pre-classification → Route

```
Image → [Classifier: printed? handwritten? mixed?] → Route to Model A or B
```

- Dùng lightweight classifier (CNN nhỏ, ~1-2MB) phân loại ảnh trước
- Classifier dựa trên: stroke regularity, intensity distribution, edge patterns
- **Vấn đề:** 1 ảnh có cả printed + handwritten (vở học sinh) → classifier phải segment theo vùng, không phải toàn ảnh → phức tạp hơn nhiều
- **Thêm latency:** ~50-100ms cho classification step

### 5.2 Parallel execution → Pick best

```
Image → [Model A] ─┐
      → [Model B] ─┤→ Compare confidence → Pick best result
```

- Gửi ảnh cho tất cả models đồng thời
- So sánh confidence scores, chọn kết quả cao nhất (per-character hoặc per-region)
- **Ưu điểm:** không cần classifier, không lo route sai
- **Vấn đề:** gấp đôi/gấp ba cost + infrastructure. Ở 500K req/ngày = 1M-1.5M actual API calls
- **Latency:** bằng model chậm nhất (không cộng dồn vì chạy song song)

### 5.3 Cascading / Fallback

```
Image → [Model A (primary)]
    → If confidence ≥ threshold → Return
    → If confidence < threshold → [Model B (fallback)] → Return
```

- Primary model xử lý trước. Chỉ gọi fallback khi confidence thấp
- Threshold VD: avg confidence < 75% → trigger fallback
- **Ưu điểm:** tiết kiệm cost — đa số request chỉ cần 1 model. Chỉ ~10-20% request cần fallback
- **Vấn đề:** latency tăng gấp đôi cho những request cần fallback (sequential). Threshold bao nhiêu là đủ? Cần tuning

### 5.4 User-specified routing

```
API request { image, type: "handwritten" | "printed" | "auto" }
    → "handwritten" → Model B
    → "printed" → Model A
    → "auto" → Cascading hoặc Parallel
```

- Caller (mobile app) cho biết loại content
- Mobile app có thể hỏi user: "Bạn đang scan chữ in hay chữ viết tay?"
- **Ưu điểm:** đơn giản, không cần classifier, user biết rõ nhất content của mình
- **Vấn đề:** user chọn sai hoặc content mixed → vẫn cần fallback logic. Thêm 1 bước UX

### 5.5 So sánh

| Cơ chế | Latency | Cost | Accuracy | Complexity |
|---|---|---|---|---|
| **Pre-classification** | +50-100ms mọi request | 1x + classifier cost | Phụ thuộc classifier accuracy. Sai = cascade error | Cao — cần train + maintain classifier |
| **Parallel** | = model chậm nhất | **2-3x** (gọi tất cả models) | **Cao nhất** — always pick best | Trung bình — chỉ cần compare logic |
| **Cascading** | 1x (80-90% requests), 2x (10-20% requests) | **~1.1-1.2x** (tiết kiệm nhất) | Tốt — fallback bắt được cases primary miss | Thấp — chỉ cần threshold tuning |
| **User-specified** | 1x | 1x | Phụ thuộc user chọn đúng | **Thấp nhất** — nhưng thêm UX step |

### 5.6 Kết hợp

Có thể kết hợp nhiều cơ chế. VD: **User-specified (5.4) + Cascading (5.3)**:

```
API request { image, type: "handwritten" | "printed" | "auto" }
    → "handwritten" → Model B              ← User-specified routing
    → "printed" → Model A                  ← User-specified routing
    → "auto" → Model A first               ← Cascading/Fallback
        → confidence ≥ 75% → Return
        → confidence < 75% → Model B → Return
```

- User biết rõ content → route thẳng, nhanh, 1x cost
- User không chắc hoặc mixed content → hệ thống tự xử lý qua cascading

---

## 6. Xử lý các bài toán khó

### 6.1 Fault Tolerance — OCR engine down

| Scenario | Xử lý | Fallback |
|---|---|---|
| **Google Vision API timeout** (> 3s) | Retry 1 lần với exponential backoff (base 500ms). Nếu vẫn fail → trả lỗi | Circuit breaker (gobreaker): sau 5 failures liên tiếp → open 30s → half-open probe. Khi open → trả lỗi ngay, không chờ timeout |
| **Google Vision API 5xx** | Không retry (server-side error thường kéo dài). Trả lỗi cho user | Circuit breaker auto-trigger. Nếu có PaddleOCR sidecar → route sang đó |
| **Google Vision API quota exceeded** | Trả lỗi "OCR temporarily unavailable, please try again later" | Alert ops team. Nếu có PaddleOCR → route sang |
| **PaddleOCR sidecar crash** (future) | Health check fail → Kubernetes restart pod | Traffic tự động route về Google Vision (cascading ngược) |
| **Cả 2 engine down** | Trả lỗi 503 + message "OCR service unavailable" | User vẫn import thủ công. Push notification khi service recovered (Phase 2) |

**Circuit breaker config (gobreaker v2 — đã có trong `infrastructure/circuitbreaker/`):**
- `MaxRequests`: 3 (half-open state)
- `Interval`: 60s (closed state counter reset)
- `Timeout`: 30s (open → half-open)
- `ReadyToTrip`: 5 consecutive failures hoặc failure rate > 50% trong 10 requests
- Success condition: chỉ `nil` error và HTTP 2xx từ OCR engine

### 6.2 Rate Limiting & Spike Traffic

**Spike pattern:** Homework time 18h-22h → traffic tăng 3-5x average. Đặc biệt trước kỳ thi → burst gấp 10x.

| Layer | Mechanism | Config |
|---|---|---|
| **Per-user** | Token bucket (đã có middleware) | Free: 3 scan/ngày (PRD). Pro: 50 scan/ngày. Enforce qua Redis counter `ocr:{user_id}:{date}` |
| **Global API** | Rate limiter trên route `/api/vocabularies/ocr-scan` | 100 RPS (MVP), tăng theo capacity |
| **OCR engine** | Google Vision mặc định 1,800 req/phút | Monitor usage vs quota. Request quota increase trước khi đạt 70% |
| **Backpressure** | Nếu Google Vision latency > 2s liên tục → giảm accept rate | Semaphore pattern: max 50 concurrent OCR requests. Vượt → queue hoặc reject 429 |

**Semaphore cho concurrent OCR calls:**
```
maxConcurrent = 50  // configurable via env OCR_MAX_CONCURRENT
```
Khi 50 OCR requests đang xử lý đồng thời → request thứ 51 nhận HTTP 429 "Too many OCR requests, please retry". Tránh overload OCR engine + tránh goroutine leak.

### 6.3 Deduplication — Cùng ảnh submit 2 lần

| Scenario | Giải pháp |
|---|---|
| **User tap "Scan" 2 lần liên tiếp** (double tap) | Client-side debounce (300ms). Server-side: idempotency key trong request header `X-Idempotency-Key: {uuid}`. Cache key + response trong Redis TTL 5 phút. Request trùng key → trả cached response |
| **Cùng ảnh, khác thời điểm** | Không cần dedup — OCR là stateless, kết quả không side-effect (chỉ trả về JSON, chưa tạo flashcard). Flashcard tạo ở bước confirm sau → duplicate check đã có ở vocabulary layer (`FindByHanziList`) |
| **Nhiều user scan cùng ảnh** (cùng bài tập) | Không cache theo image hash (privacy concern + cache hit rate thấp). Mỗi user xử lý độc lập — cost chấp nhận được ở quy mô MVP-Growth |

### 6.4 Image Preprocessing Edge Cases

| Edge case | Ảnh hưởng | Giải pháp |
|---|---|---|
| **Ảnh mờ / out of focus** | Accuracy giảm mạnh (< 50%) | Client-side: detect blur score trước khi upload (Laplacian variance < threshold → warn "Ảnh quá mờ, chụp lại"). Server-side: nếu avg confidence < 50% → trả warning "Image quality too low" |
| **Ảnh nghiêng / rotated** | Google Vision tự handle (auto-detect orientation). PaddleOCR cũng có auto-rotation | Không cần xử lý thêm — cả 2 engine đều support |
| **Ánh sáng kém / shadow** | Accuracy giảm 10-20% | Client-side: auto-enhance (brightness/contrast) trước upload. Server-side: preprocessing pipeline (adaptive thresholding) — Phase 2, MVP chấp nhận accuracy loss |
| **Ảnh chứa quá nhiều text** | Latency tăng, có thể vượt 3s budget | Limit: max 1 ảnh/request, recommend crop vùng cần scan. Server: timeout 5s hard limit → trả partial results nếu có |
| **Ảnh không chứa text** | Waste API call + confusing response | Server-side: nếu OCR trả 0 characters → response `{ new_items: [], existing_items: [], message: "No Chinese characters detected" }` |
| **Mixed orientation (ngang + dọc trên 1 trang)** | Một số dòng bị miss | Google Vision handle multi-orientation. Nếu miss → user thêm thủ công ở preview screen |

### 6.5 Network & Upload

| Bài toán | Giải pháp |
|---|---|
| **Upload ảnh 3MB trên 3G** | Client compress trước upload: JPEG quality 70-80%, resize max 1500px longest edge → target < 500KB. Google Vision recommend ≤ 1500px |
| **Upload bị ngắt giữa chừng** | Client-side retry (exponential backoff, max 3 lần). Không cần resumable upload (ảnh < 1MB sau compress) |
| **Timeout** | Client timeout: 10s (bao gồm upload + processing). Server timeout: 5s cho OCR processing. Mismatch buffer: 5s cho upload |


