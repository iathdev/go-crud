# OCR Post-processing Research — Confidence Scoring & Candidate Ranking

> Research phục vụ thiết kế pipeline xử lý kết quả OCR sau khi nhận raw output từ engine (C2 trong `technical_challenges.md`).
>
> **Context:** OCR engine đã chốt (Google Cloud Vision + Baidu OCR). File này focus vào: nhận raw output → lọc → chuẩn hóa confidence → sinh candidates cho chữ không chắc chắn → enrich pinyin/meaning → trả kết quả cho mobile.

---

## 1. Yêu cầu từ PRD

- Confidence < 80% → "Did you mean X?" + top-3 candidates
- Confidence < 70% → show top-3 cho user chọn
- Lọc chỉ lấy Hán tự từ mixed content (CN + VN + EN)
- Auto-suggest pinyin + meaning cho detected characters
- Latency budget cho toàn bộ post-processing: < 300ms (trong tổng 1-3s OCR pipeline)

---

## 2. Confidence scoring — Mỗi engine trả gì?

### 2.1 Google Cloud Vision

Google trả cấu trúc phân cấp:

```
TextAnnotation → Page → Block → Paragraph → Word → Symbol
                  │        │          │         │       │
                  └ confidence (float 0-1) ở MỌI level
```

Với tiếng Trung, mỗi **Symbol** = 1 ký tự Hán → **có confidence per-character natively**. Đây là granularity tốt nhất trong các OCR engine thương mại.

**Response mẫu (simplified):**

```json
{
  "pages": [{
    "blocks": [{
      "paragraphs": [{
        "words": [{
          "symbols": [
            { "text": "学", "confidence": 0.98 },
            { "text": "习", "confidence": 0.95 }
          ],
          "confidence": 0.96
        }]
      }]
    }]
  }]
}
```

### 2.2 Baidu OCR API

Baidu có nhiều endpoint, confidence trả khác nhau:

| Endpoint | Confidence | Granularity |
|---|---|---|
| General (standard) | Không trả | Chỉ có `words` (text per line) |
| General (accurate) | `probability` per line | Line-level |
| General with location + `recognize_granularity=small` | `char_probability` per character | **Per-character** (cần request param) |
| Handwriting Recognition | `probability` per line | Line-level |

→ **Phải dùng endpoint có `recognize_granularity=small`** hoặc extract per-character từ line probability.

**Response mẫu (với char probability):**

```json
{
  "words_result": [
    {
      "words": "学习中文",
      "probability": { "average": 0.95, "min": 0.87 },
      "chars": [
        { "char": "学", "probability": 0.98 },
        { "char": "习", "probability": 0.95 },
        { "char": "中", "probability": 0.92 },
        { "char": "文", "probability": 0.87 }
      ]
    }
  ]
}
```

### 2.3 So sánh

| | Google Cloud Vision | Baidu OCR |
|---|---|---|
| **Range** | 0.0 - 1.0 (float) | 0.0 - 1.0 (float) |
| **Per-character** | Có (native, Symbol level) | Có (cần `recognize_granularity=small`) |
| **Cần normalize range?** | Không | Không |
| **Scores có comparable không?** | **Không.** Model khác nhau, training data khác nhau, calibration khác nhau. Google 0.85 ≠ Baidu 0.85 |

### 2.4 Calibration — Tại sao không thể so sánh trực tiếp

Confidence score từ mỗi engine phản ánh **internal model certainty**, không phải xác suất đúng tuyệt đối. Cùng 1 ảnh:
- Google có thể trả 0.92 (well-calibrated cho printed)
- Baidu có thể trả 0.78 (conservative hơn cho cùng image)

**Giải pháp:**

| Approach | Mô tả | Phù hợp khi |
|---|---|---|
| **Engine-specific thresholds** (Recommend MVP) | Google: high ≥ 0.90, medium 0.75-0.90, low < 0.75. Baidu: high ≥ 0.85, medium 0.70-0.85, low < 0.70. Tune dựa trên production data | Đơn giản, đủ cho 2 engines |
| **Platt scaling / Isotonic regression** | Train calibration model trên validation set (ảnh + ground truth) → map raw score → calibrated probability | Cần labeled dataset. Phase 2 khi có production data đủ lớn |
| **Treat as relative ranking** | Không so sánh cross-engine. Mỗi engine tự rank characters của nó. Cascading fallback dùng average confidence, không compare per-char | Đơn giản nhất. Đủ cho cascading logic |

**Recommend MVP:** Engine-specific thresholds + treat as relative ranking. Cascading trigger dùng average confidence (plan_ocr_engine.md §5.1 đã define: avg < 75% → fallback). Không cần calibration model cho MVP.

---

## 3. Mixed language filtering — Lọc chỉ lấy Hán tự

### 3.1 Vấn đề: Han Unification

Chinese, Japanese Kanji, Korean Hanja, Vietnamese Chữ Nôm **dùng chung codepoints** trong Unicode (Han Unification). Không có cách phân biệt ở Unicode level.

VD: U+4E2D (中) — cùng codepoint cho Chinese, Japanese, Korean, Vietnamese.

→ **Không thể filter "chỉ Chinese" vs "chỉ Japanese" bằng Unicode range.** Nhưng may mắn: use case của chúng ta không cần phân biệt — tất cả CJK characters đều là Hán tự hợp lệ cho flashcard.

### 3.2 Unicode ranges cần filter

| Block | Range | Ghi chú |
|---|---|---|
| **CJK Unified Ideographs** | U+4E00 – U+9FFF | Core block, 20,992 chars. **Đủ cho 99% use case** |
| CJK Extension A | U+3400 – U+4DBF | 6,592 chars hiếm |
| CJK Extension B-J | U+20000 – U+323AF | ~70K+ chars cực hiếm |
| CJK Compatibility Ideographs | U+F900 – U+FAFF | Variant forms |

**Cho MVP:** Chỉ cần filter core block **U+4E00 – U+9FFF** là đủ. HSK 1-9 toàn bộ 11,000 từ đều nằm trong block này.

### 3.3 Filtering trong Go

```go
func isCJK(r rune) bool {
    return unicode.Is(unicode.Han, r)
}

func filterCJKOnly(text string) []rune {
    var result []rune
    for _, r := range text {
        if isCJK(r) {
            result = append(result, r)
        }
    }
    return result
}
```

`unicode.Han` trong Go đã cover tất cả CJK blocks (core + extensions).

### 3.4 Phân biệt context xung quanh

OCR output mixed "我今天学习了10个新词very good" → sau filter CJK: `我今天学习了个新词`

| Ký tự | Unicode category | Filter |
|---|---|---|
| 我今天学习了 | CJK Unified Ideographs | **Giữ** |
| 10 | ASCII digits | Bỏ |
| 个新词 | CJK | **Giữ** |
| very good | ASCII Latin | Bỏ |

→ Kết quả: `[我, 今, 天, 学, 习, 了, 个, 新, 词]`. Từ đây group thành words bằng dictionary lookup.

---

## 4. Candidate generation — Chữ nào giống chữ nào?

### 4.1 Vấn đề

OCR engine trả "鑫" với confidence 0.65. User thực tế viết "森" hoặc "淼". Cần show top-3 candidates ["鑫", "森", "淼"] cho user chọn.

**OCR engine không phải lúc nào cũng trả candidates.** Google Vision chỉ trả 1 kết quả per symbol (character đúng nhất). Baidu tương tự. → Cần **server-side candidate generation** dựa trên visual similarity.

### 4.2 Datasets cho similar-looking characters

| Dataset | Mô tả | Format | Relevance |
|---|---|---|---|
| **[makemeahanzi](https://github.com/skishore/makemeahanzi)** | 9,000+ chars: SVG strokes, decomposition (IDS), radicals, components | NDJSON | Dùng decomposition để tính component similarity: chars chia sẻ ≥ 2 components = candidates |
| **[similar_chinese_characters](https://github.com/kris2808/similar_chinese_characters)** | 3 loại: 形近字 (visual similar), 同音字 (homophone), 近音字 (near-homophone) | CSV | Pre-built confusable pairs, dùng trực tiếp |
| **[ChineseCharacterSimilarity](https://github.com/liu-hanwen/ChineseCharacterSimilarity)** | Tính similarity score giữa 2 chars: pinyin distance (0.4) + stroke count (0.1) + four-corner encoding (0.3) + pixel matrix (0.2) | Python tool | Pre-compute similarity matrix offline, load khi app start |
| **[Wiktionary confusables](https://en.wiktionary.org/wiki/Appendix:Easily_confused_Chinese_characters)** | Curated list confusable pairs: 未/末, 人/入, 已/己/巳 | Wiki table | Bổ sung cho dataset trên |
| **Unicode TR39 confusables** | Cross-script confusables (Cyrillic/Latin). CJK-internal coverage hạn chế | Text file | Ít hữu ích cho Chinese-internal confusables |

### 4.3 Strategy sinh candidates

#### Alternative A: Pre-built lookup table (Recommend MVP)

```
Offline: Build confusable_map từ datasets
    similar_chinese_characters (形近字)
    + makemeahanzi (chars sharing ≥ 2 components)
    + Wiktionary confusables
    → map[rune][]rune  (mỗi char → list similar chars, sorted by similarity)

Runtime: char confidence < threshold → lookup confusable_map[char] → top-3
```

**Ưu điểm:** O(1) lookup. Zero latency. Precomputed offline.
**Nhược điểm:** Chỉ cover chars có trong dataset. Chars hiếm có thể không có candidates.
**Size:** ~5,000 chars có confusables × avg 5 candidates = ~25K entries. ~500KB in memory.

#### Alternative B: Real-time similarity computation

Tính similarity score on-the-fly: decompose char → find chars sharing components → rank by score.

**Ưu điểm:** Cover mọi char (kể cả hiếm). Flexible scoring.
**Nhược điểm:** Latency ~5-10ms per char. 100 low-confidence chars = 500ms-1s. Vượt budget.

**Loại cho hot path.** Dùng cho offline build lookup table (Alternative A).

#### Alternative C: OCR engine re-recognition

Crop vùng chứa low-confidence char → gửi lại OCR engine với hints. Hoặc dùng handwriting recognition mode.

**Ưu điểm:** Engine có thể trả kết quả tốt hơn lần 2 với context hẹp hơn.
**Nhược điểm:** Thêm 1 API call (cost + latency). Kết quả không đảm bảo khác.

**Phase 2.** Không cần cho MVP.

### 4.4 Ranking candidates

Khi có list candidates, rank bằng:

1. **Visual similarity score** (từ pre-built data) — cao nhất lên đầu
2. **Character frequency** — chữ phổ biến hơn likely đúng hơn. Dùng HSK level hoặc frequency rank từ vocabulary DB
3. **Context** (Phase 2) — bigram probability với chars xung quanh. VD: "明X" → P("明天") >> P("明夫") → "天" rank cao hơn "夫"

MVP chỉ cần (1) + (2). Context-based ranking là Phase 2.

---

## 5. Auto-suggest pinyin + meaning

### 5.1 CC-CEDICT dictionary

| | Chi tiết |
|---|---|
| **Format** | `Traditional Simplified [pin1yin1] /definition 1/definition 2/` |
| **Ví dụ** | `漢字 汉字 [han4 zi4] /Chinese character/CL:個\|个/` |
| **Size** | ~120,000 entries, ~4MB file |
| **License** | Creative Commons BY-SA 4.0 |

### 5.2 Lookup performance

Existing Go libraries (`jcramb/cedict`, `purohit/go-cc-cedict`) dùng linear scan O(n) per lookup → 300 lookups × 120K entries = 36M comparisons → chậm.

**Giải pháp: Build hashmap at startup**

```go
type DictService struct {
    bySimplified map[string][]*Entry   // "学习" → [Entry{pinyin: "xué xí", meanings: [...]}]
    byCharacter  map[rune][]*Entry     // '学' → [Entry for 学, Entry for 学习, Entry for 学生, ...]
}
```

- Load 120K entries → `map[string][]*Entry`: ~50ms startup, ~20-30MB RAM
- Single character lookup: **O(1)**, < 0.01ms
- 300 character lookups: **< 1ms total**

### 5.3 Polysemy — Chữ có nhiều nghĩa

VD: 打 có 20+ meanings trong CC-CEDICT (đánh, gọi, mở, chơi, từ...).

| Strategy | Mô tả | Phù hợp |
|---|---|---|
| **Show first 2-3 definitions** | CC-CEDICT sắp xếp roughly theo frequency. Lấy 2-3 đầu + "(+N more)" | MVP — đơn giản, đủ tốt |
| **Longest-match word segmentation** | Scan text trái → phải, tìm compound dài nhất trong CC-CEDICT. "打电话" → unambiguous (gọi điện), thay vì bare "打" | MVP — cải thiện accuracy đáng kể |
| **HSK level priority** | Cross-reference HSK data. Show nghĩa HSK-1/2 trước (phổ biến nhất) | MVP — dùng vocabulary DB đã có |
| **Context-based disambiguation** | Language model chọn nghĩa phù hợp context câu | Phase 2 — cần NLP |

**Recommend MVP: Longest-match + HSK priority.**

**Longest-match algorithm:**

```
Input:  "我打电话给朋友"
Output: [我, 打电话, 给, 朋友]  (không phải [我, 打, 电, 话, 给, 朋, 友])

Logic:
  i=0: thử "我打电话给朋友"(7) → miss, "我打电话给朋"(6) → miss, ... "我"(1) → hit → emit "我"
  i=1: thử "打电话给朋友"(6) → miss, ... "打电话"(3) → hit → emit "打电话"
  i=4: thử "给朋友"(3) → miss, "给朋"(2) → miss, "给"(1) → hit → emit "给"
  i=5: thử "朋友"(2) → hit → emit "朋友"
```

→ "打电话" match → meaning = "to make a phone call" (unambiguous). Không cần show 20+ meanings của "打".

---

## 6. Post-processing pipeline

### 6.1 Full pipeline

```
OCR Engine Raw Output (Google/Baidu)
  │
  ▼
[1. Normalize]
  │ - Baidu: extract char_probability nếu có, map struct về cùng format với Google
  │ - Unicode NFC normalization
  │ - Map CJK Compatibility Ideographs → Unified equivalents
  │ - Output: []OCRCharacter{ text, confidence, boundingBox }
  │
  ▼
[2. Filter CJK]
  │ - Giữ chỉ runes match unicode.Han
  │ - Bỏ English, Vietnamese, số, punctuation
  │ - Output: []OCRCharacter (CJK only)
  │
  ▼
[3. Deduplicate]
  │ - Cùng 1 character xuất hiện nhiều lần → giữ instance có confidence cao nhất
  │ - Output: []OCRCharacter (unique)
  │
  ▼
[4. Word segmentation + Enrich]
  │ - Longest-match against CC-CEDICT → group characters thành words
  │ - Lookup pinyin + meaning cho mỗi word/character
  │ - HSK level lookup từ vocabulary DB
  │ - Output: []OCRItem{ text, pinyin, meaning, hsk_level, confidence }
  │
  ▼
[5. Classify by confidence]
  │ - High (≥ 0.80): confirmed — show trực tiếp
  │ - Medium (0.70 - 0.80): suggest — "Did you mean X?" + top-3 candidates
  │ - Low (< 0.70): manual — show top-3 candidates, user phải chọn
  │ - Output: { confirmed[], suggested[], low_confidence[] }
  │
  ▼
[6. Generate candidates] (cho medium + low confidence items)
  │ - Lookup confusable_map[char] → similar characters
  │ - Rank by: visual similarity → character frequency → HSK level
  │ - Top-3 candidates per character
  │ - Output: items with candidates[]
  │
  ▼
[7. Match with DB]
  │ - FindByHanziList() — existing vocabulary check
  │ - Classify: new_items vs existing_items
  │ - Output: final OCRScanResponse
```

### 6.2 Latency budget

| Step | Estimated latency | Ghi chú |
|---|---|---|
| 1. Normalize | < 1ms | String processing |
| 2. Filter CJK | < 1ms | Rune iteration |
| 3. Deduplicate | < 1ms | Map operation |
| 4. Word segmentation + Enrich | < 5ms | CC-CEDICT hashmap O(1) lookups |
| 5. Classify by confidence | < 1ms | Compare thresholds |
| 6. Generate candidates | < 5ms | Pre-built confusable_map O(1) lookups |
| 7. Match with DB | < 10ms | Postgres IN query with index |
| **Total post-processing** | **< 25ms** | Trong budget 300ms |

### 6.3 OCR error types

| Error type | Ví dụ | Frequency | Xử lý |
|---|---|---|---|
| **Substitution** (wrong char) | 天 → 夫, 未 → 末 | Phổ biến nhất (~70% errors) | Confusable candidates + confidence threshold. Phase 2: bigram probability |
| **Insertion** (extra char) | OCR hallucinate char từ noise | ~20% errors | Low confidence + isolated char (không form word trong CC-CEDICT) → flag "possible noise" |
| **Deletion** (missed char) | OCR miss 1 char | ~10% errors | Khó detect server-side. User thêm thủ công ở preview screen |

---

## 7. Data structures cần build

### 7.1 Confusable map (pre-built, load at startup)

```go
// Load from: similar_chinese_characters CSV + makemeahanzi decomposition + Wiktionary
type ConfusableService struct {
    confusables map[rune][]SimilarChar  // '未' → [{char: '末', score: 0.95}, {char: '本', score: 0.82}, ...]
}

type SimilarChar struct {
    Char  rune
    Score float64  // 0-1, visual similarity
}
```

**Size:** ~5K chars × 5 candidates × 12 bytes = ~300KB RAM

**Build process (offline, 1 lần):**
1. Parse `similar_chinese_characters` CSV → extract 形近字 pairs
2. Parse `makemeahanzi` NDJSON → chars sharing ≥ 2 components
3. Merge + deduplicate + sort by similarity score
4. Export thành Go embed file hoặc JSON load at startup

### 7.2 Dictionary service (CC-CEDICT, load at startup)

```go
type DictService struct {
    bySimplified map[string][]*DictEntry   // word lookup: "学习" → entries
    byCharacter  map[rune][]*DictEntry     // single char: '学' → all entries containing 学
}

type DictEntry struct {
    Simplified string
    Pinyin     string
    Meanings   []string
}
```

**Size:** ~120K entries, ~20-30MB RAM. Load time ~50ms.

### 7.3 Tổng memory footprint

| Component | RAM |
|---|---|
| Confusable map | ~300KB |
| CC-CEDICT dictionary | ~20-30MB |
| HSK vocabulary | Đã có trong Postgres, query khi cần |
| **Total** | **~30MB** |

→ Chấp nhận được. Load 1 lần at startup, dùng cho mọi request.

---

## 8. Threshold tuning

### 8.1 Initial thresholds (cần validate với production data)

| Engine | High confidence | Medium | Low |
|---|---|---|---|
| **Google Vision** | ≥ 0.90 | 0.75 – 0.90 | < 0.75 |
| **Baidu OCR** | ≥ 0.85 | 0.70 – 0.85 | < 0.70 |

**Tại sao Baidu threshold thấp hơn:** Baidu handwriting recognition inherently có confidence thấp hơn Google printed. Handwritten text variability lớn → model less certain. Nếu dùng cùng threshold → quá nhiều false "low confidence" cho handwritten.

### 8.2 Tuning strategy

1. **MVP:** Hardcode thresholds trên. Log mọi character + confidence + user edit (ở preview screen)
2. **Sau 2 tuần production:** Phân tích logs:
   - Characters confidence ≥ 0.90 nhưng user sửa → threshold quá cao (false positive)
   - Characters confidence < 0.75 nhưng user không sửa → threshold quá thấp (false negative)
3. **Adjust:** Tìm threshold tối ưu minimize (false positive + false negative) per engine
4. **Phase 2:** Platt scaling nếu cần cross-engine comparison

---

## References

| Source | URL | Relevance |
|---|---|---|
| Google Cloud Vision — Text detection response | https://docs.cloud.google.com/vision/docs/fulltext-annotations | Confidence per Symbol (per-character) |
| Google Cloud Vision — REST reference | https://docs.cloud.google.com/vision/docs/reference/rest/v1/AnnotateImageResponse | Response structure hierarchy |
| Baidu OCR API — General recognition | https://intl.cloud.baidu.com/en/doc/BOS/s/akce62nbw-intl-en | Response format, probability field |
| PaddleOCR — Confidence discussion | https://github.com/PaddlePaddle/PaddleOCR/issues/5932 | Per-character confidence extraction |
| Make Me a Hanzi | https://github.com/skishore/makemeahanzi | 9K+ chars, SVG strokes, IDS decomposition → component similarity |
| similar_chinese_characters | https://github.com/kris2808/similar_chinese_characters | Pre-built 形近字/同音字/近音字 dataset |
| ChineseCharacterSimilarity | https://github.com/liu-hanwen/ChineseCharacterSimilarity | Multi-metric similarity scoring (pinyin + stroke + four-corner + pixel) |
| Wiktionary — Easily confused characters | https://en.wiktionary.org/wiki/Appendix:Easily_confused_Chinese_characters | Curated confusable pairs |
| Unicode TR39 — Confusables | https://unicode.org/reports/tr39/tr39-28.html | Cross-script confusables (limited CJK-internal) |
| CJK Unified Ideographs — Wikipedia | https://en.wikipedia.org/wiki/CJK_Unified_Ideographs | Unicode block ranges |
| Han Unification — Wikipedia | https://en.wikipedia.org/wiki/Han_unification | Tại sao CN/JP/KR/VN share codepoints |
| jcramb/cedict (Go) | https://github.com/jcramb/cedict | Go CC-CEDICT parser |
| CC-CEDICT Wiki | https://cc-cedict.org/wiki/ | Dictionary format + license (CC BY-SA 4.0) |
| Survey of Post-OCR Processing | https://dl.acm.org/doi/fullHtml/10.1145/3453476 | Academic survey: error types, correction approaches |
| Chinese Spelling Correction survey | https://arxiv.org/html/2502.11508v1 | Confusion sets, bigram correction, BERT-based approaches |
