# Vocabulary Module — Requirement (trích từ PRD v3)

> Nguồn gốc: `docs/requirement.md` — Prep Chinese Vocab PRD v3

---

## 1. OCR Scan Hán tự → Auto Flashcards

**Mô tả:** Chụp ảnh vở ghi chép / sách giáo khoa → OCR nhận diện Hán tự → tự động tạo flashcard.

### 1.1 Flow

```
User tap "Scan" → Camera opens → Chụp ảnh / import từ gallery
    → OCR processing (1-3s)
    → Preview screen: hiển thị detected characters
        → Mỗi character có: confidence indicator, suggested pinyin, suggested meaning
        → User có thể: Edit / Delete / Add missing / Confirm all
    → Duplicate check: so sánh với existing collection
        → Trùng: hiện "X từ đã có" + option View / Ignore / Merge
    → Assign to folder/topic (optional, có thể chọn "Unsorted")
    → Generate flashcards → Done
```

### 1.2 Edge Cases

| Case | Solution |
|---|---|
| OCR nhận sai chữ Hán | Preview + "Confirm or Edit". Confidence < 80% → "Did you mean X?" + top-3 candidates |
| Mix tiếng Trung + Việt + Anh | Lọc chỉ lấy Hán tự. Hiển thị "Detected X Chinese characters" |
| Chữ quá xấu / ký tự đặc biệt | UI review với thông tin tối thiểu: word, pinyin (auto-suggest), definition |
| Scan trùng từ đã có | Check duplicate → thông báo → View / Ignore / Merge |
| Note học sinh ghi sai | Tôn trọng nội dung user viết. Chỉ gợi ý nếu phát hiện sai sót |

### 1.3 Technical

- OCR Engine: Google Cloud Vision API (primary) / Baidu OCR API (fallback handwritten)
- Target accuracy: ≥ 90% printed, ≥ 80% handwritten
- Fallback: confidence < 70% → show top-3 candidates

---

## 2. HSK Built-in Wordlists

### 2.1 Cấu trúc HSK 3.0 (syllabus Nov 2025)

| HSK Level | Stage | Từ vựng (tích lũy) | Hán tự nhận diện | Hán tự viết | Access |
|---|---|---|---|---|---|
| HSK 1 | Elementary (A1) | 300 | 246 | — | Free |
| HSK 2 | Elementary (A1) | 500 | 424 | — | Free |
| HSK 3 | Elementary (A2) | 1,000 | 636 | — | Free |
| HSK 4 | Intermediate (B1) | 2,000 | 1,096 | — | Pro |
| HSK 5 | Intermediate (B1) | 3,600 | 1,527 | 150 | Pro |
| HSK 6 | Intermediate (B2) | 5,400 | 1,940 | 300 | Pro |
| HSK 7-9 | Advanced (C1-C2) | 7,000-11,000 | 2,421-3,088 | 400-500 | Pro (Phase 2) |

### 2.2 Data Model per Word

```json
{
  "hanzi": "学习",
  "pinyin": "xuéxí",
  "meaning_vi": "học tập",
  "meaning_en": "to study",
  "examples": [
    { "cn": "我每天学习中文。", "vi": "Tôi mỗi ngày học tiếng Trung.", "audio_url": "..." }
  ],
  "audio_url": "...",
  "hsk_level": 1,
  "topic": "学习教育",
  "radicals": ["子", "冖", "习"],
  "stroke_count": 11,
  "stroke_data_url": "...",
  "grammar_points": ["gp_001"],
  "recognition_only": true,
  "frequency_rank": 42
}
```

---

## 3. Topic & Category System

10 topic chuẩn HSK:

| # | Tiếng Trung | Slug | Nội dung |
|---|---|---|---|
| 1 | 日常生活 | daily-life | Chào hỏi, gia đình, thời gian, số đếm |
| 2 | 饮食 | food-drink | Món ăn, gọi món, nấu ăn |
| 3 | 交通旅行 | transportation | Phương tiện, hỏi đường |
| 4 | 学习教育 | education | Trường học, thi cử |
| 5 | 工作商务 | work-career | Công việc, thương mại |
| 6 | 健康医疗 | health | Bệnh viện, triệu chứng |
| 7 | 科技 | technology | Internet, thiết bị |
| 8 | 自然环境 | nature | Thời tiết, động vật |
| 9 | 文化娱乐 | culture | Phim, nhạc, lễ hội |
| 10 | 社会 | society | Luật pháp, kinh tế |

**Quy tắc:**
- Tách khái niệm **Topic** (system-defined) và **Folder/Deck** (user-created)
- User tự tạo folder khi scan/import, có thể chọn topic tag (optional)
- 1 từ có thể thuộc nhiều topic nếu polysemy
- Không giới hạn số từ per learning card

---

## 4. Character Decomposition System

Mỗi flashcard hiển thị:

1. **Radical (bộ thủ):** Thành phần chính → nhóm nghĩa. VD: 语 → bộ 讠(ngôn ngữ) + 五 + 口
2. **Breakdown animation:** Tách chữ thành từng thành phần
3. **Memory hook:** Câu chuyện ghi nhớ. VD: 休 = 人 + 木 → "Người dựa vào cây = nghỉ ngơi"
4. **Related characters:** Các chữ cùng bộ thủ. VD: 讠→ 说, 话, 语, 读, 认

**Data sources:**
- Unihan database (Unicode Consortium) → radical data
- CC-CEDICT → nghĩa + pinyin
- CJK Decomposition Data Project → thành phần cấu tạo
- Memory hooks: AI-generated, Phase 2 thêm human review cho top 500 từ

---

## 5. Grammar Context System (MVP)

Không xây module riêng. Grammar gắn vào từng từ vựng dưới dạng Grammar Tips.

Mỗi learning card có section "Grammar":

1. **Pattern:** VD: 把 → `S + 把 + O + V + Complement`
2. **Example:** 1-2 câu highlight pattern. VD: 我**把**书**放在**桌子上。
3. **Rule ngắn:** 1-2 câu giải thích
4. **Common mistake:** Lỗi người Việt hay mắc. VD: "Không dùng 把 với 是, 有, 知道."

**Data:** Chinese Grammar Wiki (AllSet Learning, CC) + HSK Standard Course index. MVP cover **80 grammar points** cho HSK 1-3.

---

## 6. Stroke Order Data

- **Make Me a Hanzi** (open-source, ARPHIC License, 9,000+ characters, SVG paths)
- Stroke validation: compare user input vs. reference (order + direction)
- Standard: GB (Trung Quốc đại lục)
- ⚠️ Cần confirm commercial use license với legal team

---

## 7. Polysemy & Duplicate Handling

| Case | Solution |
|---|---|
| 1 Hán tự nhiều nghĩa → topic nào? | Cùng nghĩa → 1 bản, "also appears in X topic." Khác nghĩa → flashcard riêng |
| Scan trùng từ ở nhiều folder | Same word ≠ meaning → keep both. Same word = meaning → confirm override |
| Cùng pinyin khác Hán tự (homophone) | Quiz pinyin → hiển thị tất cả Hán tự cùng pinyin để phân biệt |
| Traditional vs Simplified | MVP: Simplified only. Phase 2: toggle Traditional as reference |

---

## 8. Folder System

- User tự tạo **Folder** (deck) để tổ chức từ vựng
- Folder thuộc sở hữu user (chỉ owner mới CRUD được)
- 1 từ có thể nằm trong nhiều folder
- Sắp xếp theo ngày thêm mới nhất trong folder
- Phase 2: Folder shareable → link mở app

---

## 9. Free vs Pro — Giới hạn liên quan Vocabulary

| Feature | Free | Pro |
|---|---|---|
| HSK Wordlists | HSK 1-3 | HSK 1-9 |
| Scan/day | Max 3 ảnh | Unlimited |
| Cards/day | Max 20 | Unlimited |
| Grammar | Tips giới hạn | Full context |

---

## 10. API Endpoints (hiện tại)

**Protected (JWT):**

| Method | Endpoint | Mô tả |
|---|---|---|
| GET | `/api/topics` | 10 chủ đề hệ thống |
| POST | `/api/vocabularies` | Tạo từ vựng |
| GET | `/api/vocabularies/:id` | Lấy từ vựng |
| GET | `/api/vocabularies/:id/detail` | Chi tiết + topics + grammar |
| GET | `/api/vocabularies/hsk/:level` | Theo HSK level (phân trang) |
| GET | `/api/vocabularies/topic/:slug` | Theo chủ đề (phân trang) |
| GET | `/api/vocabularies/search?q=` | Tìm kiếm |
| PUT | `/api/vocabularies/:id` | Cập nhật |
| DELETE | `/api/vocabularies/:id` | Xoá mềm |
| POST | `/api/vocabularies/ocr-scan` | Xử lý OCR scan |
| POST | `/api/admin/vocabularies/import` | Nhập hàng loạt |
| POST | `/api/folders` | Tạo folder |
| GET | `/api/folders` | Liệt kê folder |
| PUT | `/api/folders/:id` | Cập nhật folder |
| DELETE | `/api/folders/:id` | Xoá folder |
| POST | `/api/folders/:id/vocabularies` | Thêm từ vào folder |
| DELETE | `/api/folders/:id/vocabularies/:vocab_id` | Xoá từ khỏi folder |
| GET | `/api/folders/:id/vocabularies` | Từ trong folder (phân trang) |

---

## 11. Timeline liên quan

| Milestone | Date | Mô tả |
|---|---|---|
| API contracts + mocks ready | Mar 28 | Mobile team không bị block |
| HSK 1-3 Content Ready | Apr 14 | 1,000 từ: wordlists + grammar tips + audio |
| Alpha Build (core flow) | Apr 14 | Scan → Flashcard → Recall → Review working |

---

## 12. Content Dependencies

| Dependency | Mô tả |
|---|---|
| HSK wordlists + grammar | 1,000 từ HSK 1-3: wordlists, grammar tips, decomposition, audio |
| Academic Director | Review HSK wordlist accuracy, grammar points |
| Make Me a Hanzi | Stroke order SVG data (cần confirm license) |
| Chinese Grammar Wiki | AllSet Learning, CC license — grammar patterns |
