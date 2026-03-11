# Prep Chinese Vocab — PRD v3

**Smart Chinese Vocabulary Learning App with AI-Powered Pronunciation & Video Flashcards**

| Field | Detail |
|---|---|
| Team | Learning Utilities Squad |
| PO + BA | Nguyễn Thị Tuyến (50% — song song Learning Squad, đánh giá lại sau Phase 1) |
| Commercial Lead | Lâm Phụng Nhi (100%) |
| Product Designer | Phạm Đình Khánh (≥50% Phase 1) |
| Backend Engineer | Đồng Phú Thái (100%) |
| Mobile Engineer | Nghiêm Bá Cường (100%) |
| QC | Nguyễn Thanh Quỳnh Chi (100%) + Phan Thị Kim Tuyết (support) |
| Sponsors | CTO (Nam) — Architecture review, conflict escalation · CEO — Strategic alignment · Academic Director — Học thuật review, SME support |
| Shared Resources | DevOps: Tuấn Bùi (on-demand) |
| Resources | PRD v2 (Draft 2), HSK 3.0 syllabus (Nov 2025), MyVocab English PRD, PrepAI Scoring Engine docs, Learning Utilities Squad Charter v2.0 |
| Status | **Draft 3 — For Review** |
| Sprint 0 Start | March 15, 2026 |
| MVP Clock Start | April 1, 2026 |
| MVP Target | April 29, 2026 (stretch, 4w) — May 13, 2026 (realistic, 6w) |
| Last Updated | Monday, March 9, 2026 |

---

## 1. Problem Alignment

### 1.1 Customer Pain

Học viên tiếng Trung gặp 3 rào cản đặc thù mà tiếng Anh không có:

1. **Hán tự là hệ thống chữ viết hoàn toàn khác biệt** — không có alphabet, không đoán được cách đọc từ mặt chữ. Ghi nhớ đòi hỏi luyện viết lặp lại, hiểu bộ thủ (radical), nắm stroke order. Đa số app hiện tại bỏ qua hoặc xử lý sơ sài phần này.

2. **Thanh điệu (tones) quyết định nghĩa** — cùng âm "ma" nhưng 4 thanh = 4 nghĩa hoàn toàn khác. Học viên Việt Nam có lợi thế (tiếng Việt cũng có thanh điệu) nhưng hệ thanh không map 1:1, vẫn cần drill riêng.

3. **Ngữ pháp dựa trên trật tự từ và hư từ** — không biến đổi hình thái, dễ hiểu sai nếu chỉ học từ vựng rời rạc mà không nắm cấu trúc câu.

Ngoài ra, học viên thường phải tự nhập tay từ vựng đã ghi chép → cảm thấy double work. Chưa có công cụ nào giải quyết triệt để điểm này.

### 1.2 Business Pain

- **Prep HSK thiếu top-of-funnel tool.** Hiện chỉ có sản phẩm trả phí, không có phễu acquisition free users.
- **HSK là vertical chiến lược** trong kế hoạch giảm phụ thuộc IELTS xuống dưới 40% revenue. App miễn phí tăng brand awareness và tạo habit loop trước khi convert.
- **Dữ liệu học tập** (vocab level, learning progress, weak areas) feed ngược vào Prep HSK để personalize lộ trình, tạo competitive advantage so với đối thủ chỉ bán khóa học tĩnh.

### 1.3 Evidence & Insights

- **Thị trường HSK tăng trưởng:** Số thí sinh HSK tại Việt Nam tăng ~30% YoY (2023-2025), đặc biệt Gen Z có nhu cầu làm việc với doanh nghiệp Trung Quốc.
- **Gap trên thị trường:** Không app nào kết hợp đầy đủ: Hán tự stroke + SRS + AI pronunciation + grammar context trong 1 flow liền mạch.
- **Insight từ MyVocab English:** Scan notes → auto flashcard có viral potential cao (đã validate qua user research).
- **Dữ liệu nội bộ Prep HSK:** ~60% học viên báo khó khăn lớn nhất là ghi nhớ từ vựng và Hán tự.
- **Insight viral:** Công cụ chuyển note sang flashcard nhanh hơn có sức hút rất lớn trên MXH.

---

## 2. Product Vision & Positioning

### 2.1 Vision Statement

Prep Chinese Vocab là ứng dụng học từ vựng tiếng Trung **mobile-first**, thiết kế cho tất cả người học tiếng Trung — đặc biệt phù hợp cho ôn thi HSK. App đóng vai trò **phễu acquisition (free tool)** đồng thời là **công cụ bổ trợ (learning utility)** cho học viên Prep HSK.

### 2.2 Core Differentiators (so với market)

| # | Differentiator | Mô tả | Competitor Gap |
|---|---|---|---|
| 1 | **Video Flashcard (Seedance 2.0)** | AI-generated short video 3-5s minh họa nghĩa từ trong context thực tế. "Flashcard sống" — không app nào có. | Zero competitors |
| 2 | **PrepAI Pronunciation Scoring** | Engine scoring chi tiết đến từng âm/thanh, xác định chính xác user phát âm sai ở đâu, tạo learning plan cải thiện cá nhân hóa. | HelloChinese/SuperChinese chỉ có basic speech recognition, không scoring engine |
| 3 | **Handwriting Multi-mode** | Guided Writing → Recall Writing → Speed Writing. AI chấm stroke order + quality + confusable detection. | Competitors chỉ có 1 mode viết cơ bản |
| 4 | **HSK 3.0 Native** | Hệ thống 9 levels, 4-dimension tracking, recognition vs. writing flag. | Đa số stuck ở HSK 2.0 |
| 5 | **OCR Scan → Auto Flashcard** | Chụp vở ghi chép → auto-generate learning cards. | Không ai giải quyết triệt để |

> **⚠️ Cần validate:** Competitor gaps ở trên dựa trên desk research ban đầu (HSK 3.0 Deep Research Report). Nhi (Commercial Lead) sẽ nghiên cứu sâu hơn trong Sprint 0 Market Brief — bao gồm: verify các claims, cập nhật competitive landscape mới nhất, đánh giá mức độ thực sự của từng gap, và xác định có differentiator nào đã bị competitors thu hẹp kể từ lần research trước không. Bảng trên sẽ được cập nhật sau khi Market Brief v1 hoàn thành (Mar 31).

### 2.3 Product Type

**Standalone App, Deep Ecosystem Integration:**

App phát hành độc lập trên App Store / Play Store để tối ưu acquisition funnel. Tuy nhiên, app được thiết kế như một phần mở rộng của hệ sinh thái Prep — không phải standalone thuần túy:

- **User Service chung:** Sử dụng chung User Service của Prep platform (không tạo user system riêng) → tránh data migration về sau, đảm bảo `learner_id` consistent xuyên suốt ecosystem.
- **Deep data sync:** Learning progress, vocab level, weak areas sync realtime với Prep HSK → feed vào adaptive learning system (khi sẵn sàng).
- **SSO:** Single Sign-On giữa Prep Chinese Vocab ↔ Prep HSK ↔ các app Prep khác.
- **Có thể tích hợp vào app Prep hiện tại** dưới dạng embedded module (WebView hoặc native module) nếu strategy thay đổi — architecture phải support cả hai hướng.
- **Standalone cho acquisition:** User có thể dùng app không cần Prep HSK subscription, nhưng account vẫn là Prep account.

> **Lưu ý:** Team Charter (v2.0, mục 1.5) ghi "utility products kết nối trực tiếp với nền tảng Prep — không phải standalone apps mà là phần mở rộng của ecosystem." PRD approach: standalone packaging cho acquisition, nhưng bản chất là ecosystem extension về mặt data/account/integration.

### 2.4 Target Market

- **V1 Focus:** Việt Nam — người học tiếng Trung (đặc biệt ôn thi HSK)
- **Định hướng Global từ ngày đầu:** Architecture, UI/UX, content structure sẵn sàng cho multi-language (VN → TH → ID → KR → JP → Global). Ngôn ngữ giao diện MVP: Tiếng Việt + English.

### 2.5 Target Platforms

**Mobile only (iOS + Android)** — React Native hoặc Flutter. Web version không nằm trong scope V1.

---

## 3. Goals & Success Metrics

### 3.1 Goals

| # | Goal | Metric | Target (3 tháng post-launch) |
|---|---|---|---|
| 1 | Tạo phễu acquisition cho Prep HSK | MAU Free users | 50,000 |
| 2 | Conversion sang paid | Free → Pro conversion rate | ≥ 3% |
| 3 | Giúp ghi nhớ Hán tự hiệu quả | % users đạt "Finish Learning" cho ≥ 50% vocab sau 30 ngày | ≥ 70% |
| 4 | Xây dựng habit loop | D7 retention / D30 retention | ≥ 40% / ≥ 20% |
| 5 | Cross-sell Prep HSK | Prep HSK cross-sell từ app | ≥ 1.5% MAU |
| 6 | Downloads | Total downloads | 100,000 |
| 7 | Engagement | Avg. words learned/user/week | ≥ 30 |

### 3.2 Non-goals (MVP)

1. Không xây module ngữ pháp độc lập (Phase 2).
2. Không hỗ trợ ngôn ngữ khác ngoài tiếng Trung.
3. Không build social features (chat nhóm, ghép cặp, vocab battle) — Phase 2.
4. Không build Video Flashcard — Phase 2 (cần evaluate cost/latency Seedance 2.0).
5. Không hỗ trợ offline mode.
6. Không hỗ trợ Traditional Chinese (Phase 2).

---

## 4. Phased Roadmap

### Phase 1 — MVP

**Timeline:** Sprint 0 (Mar 15-31) → Sprint 1 (Apr 1-14) → Sprint 2 (Apr 15-28) → Sprint 3 stretch ship (Apr 29) hoặc buffer + polish (May 13). Chi tiết xem mục 18.

Tập trung vào 3 trụ cột cốt lõi:

**Trụ 1: Nhập từ vựng thông minh**
- OCR Scan Hán tự → auto flashcards (printed ≥ 90%, handwritten ≥ 80% accuracy)
- Import thủ công + HSK Built-in Wordlists (HSK 1-9, theo chuẩn HSK 3.0 Nov 2025 syllabus)

**Trụ 2: Chu trình học đa dạng (Smart Learning Path) — 7 Modes**
- Discover (flashcards) → Recall → Stroke & Recall → Pinyin Drill → AI Chat → Review (SM-2) → Mastery Check
- Kết hợp SRS (SM-2) + Memory Score tracking
- Grammar gắn context: mỗi từ vựng kèm cấu trúc ngữ pháp liên quan

**Trụ 3: Hệ thống ghi nhớ & phân loại (Vocabulary Retention Logic)**
- Memory Score per word, 6 trạng thái: Start Learning → Still Learning → Almost Learnt → Finish Learning → Memory Mode → Mastered
- Dashboard theo dõi tiến độ 4-dimension (Syllables, Characters, Vocabulary, Grammar)

### Phase 2 — Growth & Differentiation (Q3-Q4 2026)

| Feature | Priority | Mô tả |
|---|---|---|
| **Video Flashcard (Seedance 2.0)** | P0 | AI-generated short video cho mỗi từ vựng. "Flashcard sống." |
| **Social Learning — Vocab Sharing** | P0 | Share flashcard sets, folder vocab, leaderboard, streaks sharing |
| **Social Learning — AI Group Chat** | P1 | Chat nhóm (max 4 người) với AI personas, ghép cặp ngẫu nhiên |
| **30-second Vocab Battle** | P1 | Thi đấu chọn nghĩa đúng nhanh nhất + leaderboard |
| **Translation Practice Module** | P0 | Kỹ năng mới HSK 3.0 Level 4+. CN↔VN → CN↔TH, CN↔ID |
| **Grammar Module riêng** | P1 | Bài học ngữ pháp theo HSK level, bài tập, so sánh patterns |
| **PrepAI Custom Model** | P1 | Thay thế SpeechSuper bằng custom pronunciation model fine-tuned cho Mandarin |
| **Handwriting Quality Scoring** | P1 | AI chấm chất lượng nét viết (stroke quality), không chỉ order |
| **CBT Simulation** | P1 | HSK 3.0 thi máy: pinyin input, headset audio, strict timing |
| **Radical Explorer** | P2 | Interactive map bộ thủ, nhóm Hán tự theo radical |
| **Rapid Recall 2x** | P2 | Mode lướt reels tốc độ x2 — true/false, 10 từ trong 20s |
| **Traditional Chinese** | P2 | Hỗ trợ phồn thể cho TW/HK market |
| **HSK 2.0 → 3.0 Migration Wizard** | P2 | Placement test + gap analysis |

### Phase 3 — Future Ideas (Backlog)

- **Ghép từ Hán Việt (Sino-Vietnamese Compound Mapping):** Tận dụng lợi thế người Việt — ~60% từ vựng tiếng Việt gốc Hán Việt. Mapping Hán tự → âm Hán Việt → từ ghép. User học 1 chữ → unlock cả cluster. Differentiation cực mạnh cho VN market, mở rộng sang JP (on'yomi) và KR (hanja).
- **Tích hợp từ điển Prep AI.**
- **Integration Teacher Bee AI.**

---

## 5. Key Features — Chi tiết MVP

### 5.1 OCR Scan Hán tự → Auto Flashcards

**Mô tả:** Chụp ảnh vở ghi chép / sách giáo khoa → OCR nhận diện Hán tự → tự động tạo flashcard gồm: Hán tự, pinyin, nghĩa tiếng Việt, ví dụ câu, hình ảnh minh họa, audio phát âm.

**Flow chi tiết:**

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

**Xử lý edge cases:**

| Case | Solution |
|---|---|
| OCR nhận sai chữ Hán (viết tay, font lạ) | Preview + "Confirm or Edit" từng mục. Cảnh báo từ sai chính tả → "Did you mean X?" khi confidence < 80%. Hiển thị top-3 candidates. |
| Vở có mix tiếng Trung + Việt + Anh | Lọc chỉ lấy Hán tự. Hiển thị "Detected X Chinese characters" trước khi tạo cards. |
| Chữ quá xấu / ký tự đặc biệt | UI review với thông tin tối thiểu: word, pinyin (auto-suggest), definition, IPA, example (opt). Hệ thống so sánh từ điển → đưa IPA + audio + ảnh nếu có. |
| Scan trùng từ đã có | Check duplicate → thông báo → View / Ignore / Merge. Merge luôn vào card cũ, hiển thị số vocab trùng + CTA review. |
| Note hs ghi sai | Tôn trọng nội dung user viết. Chỉ gợi ý nếu phát hiện sai sót (ưu tiên thấp hơn so với user input). |

**Technical Requirements:**
- OCR Engine: Google Cloud Vision API (primary) hoặc Baidu OCR API (fallback cho handwritten). Tesseract (open-source) fine-tuned trên dataset viết tay tiếng Trung.
- Target accuracy: ≥ 90% printed, ≥ 80% handwritten.
- Fallback: confidence < 70% → show top-3 candidates cho user chọn.

### 5.2 HSK Built-in Wordlists

**Cấu trúc theo HSK 3.0 (syllabus Nov 2025):**

| HSK Level | Stage | Từ vựng (tích lũy) | Hán tự nhận diện | Hán tự viết | Syllables | Access |
|---|---|---|---|---|---|---|
| HSK 1 | Elementary (A1) | 300 | 246 | — | 269 | Free |
| HSK 2 | Elementary (A1) | 500 | 424 | — | 468 | Free |
| HSK 3 | Elementary (A2) | 1,000 | 636 | — | 608 | Free |
| HSK 4 | Intermediate (B1) | 2,000 | 1,096 | — | 724 | Pro |
| HSK 5 | Intermediate (B1) | 3,600 | 1,527 | 150 | 822 | Pro |
| HSK 6 | Intermediate (B2) | 5,400 | 1,940 | 300 | 908 | Pro |
| HSK 7 | Advanced (C1) | 7,000 | 2,421 | 400 | 1,020 | Pro (Phase 2) |
| HSK 8 | Advanced (C1) | 9,000 | 2,753 | 450 | 1,070 | Pro (Phase 2) |
| HSK 9 | Advanced (C2) | 11,000 | 3,088 | 500 | 1,110 | Pro (Phase 2) |

**Data model per word:**

```
{
  "hanzi": "学习",          // Simplified Chinese
  "pinyin": "xuéxí",        // với tone marks
  "meaning_vi": "học tập",  // Primary
  "meaning_en": "to study", // Secondary
  "examples": [
    { "cn": "我每天学习中文。", "vi": "Tôi mỗi ngày học tiếng Trung.", "audio_url": "..." }
  ],
  "audio_url": "...",
  "hsk_level": 1,
  "topic": "学习教育",
  "radicals": ["子", "冖", "习"],
  "stroke_count": 11,
  "stroke_data_url": "...",    // SVG path from Make Me a Hanzi
  "grammar_points": ["gp_001"],
  "recognition_only": true,    // true for L1-4, false for L5+
  "frequency_rank": 42
}
```

### 5.3 Smart Learning Path — 7 Modes

#### 5.3.1 Discover Mode (Free & Pro)

Flashcard dọc full-screen. Mặt trước: Hán tự (lớn) + pinyin + tone marks + hình ảnh minh họa. Tap flip → mặt sau: nghĩa Việt, ví dụ câu, grammar tip.

| Element | Spec |
|---|---|
| Flashcard Layout | Full-screen dọc, swipe up = next card |
| Audio Button | Góc trên phải — phát âm chuẩn (male/female voice toggle) |
| Radical Badge | Badge nhỏ hiển thị bộ thủ + nghĩa. Tap → character decomposition popup |
| Progress | "Card 2 of 20" |
| Phase 2 Upgrade | **Video Flashcard (Seedance 2.0)** — thay hình tĩnh bằng video 3-5s AI-generated |

#### 5.3.2 Recall Mode (Free & Pro)

| Element | Spec |
|---|---|
| Quiz Types | (1) Nghe audio → chọn Hán tự, (2) Xem Hán tự → chọn nghĩa (MCQ), (3) Xem pinyin → viết Hán tự, (4) Matching: nối Hán tự-nghĩa, (5) Fill pinyin tones, (6) Odd-one-out, (7) Group words by category |
| Feedback | Chấm tức thì + mini-tip giải thích (VD: "休 = 人 + 木 → nghỉ dưới gốc cây") |
| Accuracy Tracker | Tính % chính xác → cập nhật Memory Score |
| Polysemy handling | Đáp án đúng = nghĩa user đã save. Nếu từ trùng ở nhiều folder → dạng "pick from list", chọn đủ các nghĩa đã học mới tính pass |

#### 5.3.3 Stroke & Recall Mode

Mode kết hợp 3 sub-modes trong flow liền mạch, giải quyết pain point: "học nhưng không viết → quên mặt chữ."

**Sub-mode A: Guided Writing (mặc định cho từ mới, Memory Score 0-39)**

| Element | Spec |
|---|---|
| Stroke Animation | Animation viết từng nét theo đúng thứ tự. User xem rồi viết theo. |
| Writing Canvas | Vùng viết tay + grid ô vuông (田字格) hướng dẫn tỷ lệ. |
| Stroke Order Check | Đúng = xanh, sai = đỏ + gợi ý nét tiếp theo. |
| Stroke Count | "Nét 3/8" |
| Radical Highlight | Sau viết xong → highlight bộ thủ + giải thích |
| Retry Logic | Sai > 2 nét → "Viết lại". Đạt → "Next Character" |

**Sub-mode B: Recall Writing (auto-unlock khi Memory Score ≥ 40)**

| Element | Spec |
|---|---|
| Purpose | Viết Hán tự từ trí nhớ — KHÔNG gợi ý stroke, KHÔNG animation. Chuyển từ "nhận diện" sang "tái tạo." |
| Prompt | Hiển thị: pinyin + nghĩa Việt + audio. KHÔNG hiển thị Hán tự. |
| Canvas | Grid ô vuông, không hướng dẫn stroke order. |
| AI Evaluation | So sánh output vs. chuẩn: (1) Đúng chữ? (2) Đúng stroke order? (3) Tỷ lệ nét chính xác. 3 mức: Perfect / Acceptable (đúng chữ, sai 1-2 stroke) / Incorrect. |
| Hint System | Stuck >10s → hint L1: hiện radical. +10s → hint L2: hiện số nét. "Show answer" → hiện Hán tự + animation (trừ điểm recall). |
| Interleaving | Mỗi từ viết đúng 3 lần xen kẽ = Pass. |
| Confusable Detection | Viết sai ra chữ Hán hợp lệ khác → so sánh 2 chữ cạnh nhau, highlight nét khác biệt (VD: 拔 vs 拨). |

**Sub-mode C: Speed Writing (Phase 2, auto-unlock khi Memory Score ≥ 70)**

| Element | Spec |
|---|---|
| Purpose | Viết nhanh liên tục 10 chữ trong time pressure. Train automaticity — viết không cần nghĩ. |
| Flow | Hiện nghĩa/pinyin → user viết → 5s timeout → next. Chấm binary: pass/fail. |
| Gamification | Timer bar + streak counter. Personal best tracking. |

**Flow chuyển đổi giữa sub-modes:**

| Điều kiện | Sub-mode |
|---|---|
| Từ mới (Memory Score = 0) | Guided Writing |
| Memory Score < 40 (Still Learning) | Guided Writing (option tự chuyển Recall) |
| Memory Score ≥ 40 (Almost Learnt+) | Auto-switch → Recall Writing |
| Memory Score ≥ 70 (Finish Learning+) | Auto-switch → Speed Writing (Phase 2) |
| Recall fail 2 lần liên tiếp | Quay Guided, 1 session sau thử lại Recall |

**Session Summary:** Số từ viết đúng/tổng, Guided vs Recall ratio, danh sách từ cần ôn, confusable pairs.

**Access:**

| Free | Pro |
|---|---|
| Guided Writing: xem animation only | Guided Writing: viết tay + AI check |
| Recall Writing: 5 từ/ngày | Recall Writing: unlimited |
| Speed Writing: N/A | Speed Writing: unlimited (Phase 2) |

#### 5.3.4 Pinyin & Tone Drill Mode — PrepAI Pronunciation

Đây là **core differentiator #2**: sử dụng PrepAI scoring engine để đánh giá phát âm chi tiết đến từng âm/thanh.

| Element | Spec |
|---|---|
| Prompt | Hiển thị từ vựng 1-by-1, phân tách các âm đọc kèm IPA + trọng âm. Speaker button nghe mẫu. |
| Recording | User tap mic icon → ghi âm → AI processing (< 2s) |
| **PrepAI Scoring** | Chấm 3 chiều riêng biệt: **Initial** (âm đầu: zh/z, sh/s, ch/c...), **Final** (vần), **Tone** (thanh điệu 1-4 + thanh nhẹ). Mỗi chiều cho điểm 0-100. |
| Tone Visualization | Contour diagram 4 thanh (+ thanh nhẹ). Overlay: thanh user đọc (đỏ) vs. thanh đúng (xanh). Real-time pitch contour nếu possible. |
| Result Display | Xanh (đúng) / Đỏ (sai) trên từng âm phân tách. VD: "**x**uéx**í**" → x=đúng, ué=sai tone 2, xí=đúng |
| Common Mistakes | Gợi ý lỗi phổ biến người Việt: nhầm zh/z, sh/s, thanh 2 vs 3 |
| **Personalized Weakness Report** | Sau mỗi session → tổng hợp: "Bạn hay sai thanh 2 (rising tone) và âm đầu zh. Gợi ý drill thêm 5 từ có zh + thanh 2." |
| **Learning Plan** | Tích lũy data pronunciation → tạo personalized improvement plan: focus vào âm/thanh yếu nhất. Hiển thị progress over time. |
| CTA | "Retry" nếu sai hoàn toàn. "Next Word." "Skip Mode" (dòng nhỏ trên Record button) → chuyển thẳng AI Chat. |
| Access | Free: trial 3 từ/ngày. Pro: unlimited. |

**Technical — PrepAI Integration:**

| Component | MVP | Phase 2 |
|---|---|---|
| Scoring Engine | SpeechSuper API (SDK có sẵn cho Mandarin tone scoring) | Custom PrepAI model fine-tuned cho Mandarin tones (tái sử dụng architecture DeBERTa/PANN từ IELTS Speaking) |
| Granularity | Initial + Final + Tone scoring per syllable | + Fluency + Rhythm + Sentence-level prosody |
| Data Collection | Ghi âm user pronunciation (opt-in consent) → training data cho custom model | 10,000+ audio samples × 4 tones × common syllables |
| Weakness Detection | Per-session summary | Cross-session learning plan + adaptive drill |

#### 5.3.5 AI Chat Mode (Pro only)

| Element | Spec |
|---|---|
| Interface | Messenger-style. AI bên trái, user bên phải. |
| Topic Context | Gắn chủ đề learning card (VD: 点餐 gọi món, 旅行 du lịch) |
| AI Personas | 3 persona: (1) **朋友 Bạn bè** — tiếng Trung đơn giản, gần gũi, không quan trọng ngữ pháp; (2) **老板 Sếp** — từ vựng công việc, yêu cầu trang trọng; (3) **老师 Gia sư** — sửa lỗi ngữ pháp, yêu cầu dùng synonyms cao cấp hơn |
| Interaction | Text input (ưu tiên) + voice input. AI phát hiện đúng từ đã học → highlight xanh, thả tim + cộng Recall. |
| Grammar Feedback | Phản hồi tự nhiên, gợi ý sửa lỗi / từ đồng nghĩa. VD: "把 dùng khi tác động lên đối tượng: 把书放在桌子上" |
| Session Summary | Words used correctly, New words introduced, XP gained |
| Scoring Link | Update Memory Score (mode weight 2) |

#### 5.3.6 Review Mode (Free & Pro)

| Element | Spec |
|---|---|
| Purpose | Spaced Repetition theo SM-2. |
| Daily Review CTA | Tự động hiển thị khi spacing due. Nút CTA review bên ngoài learning path. |
| Mini Games | Word scramble, True/False, Crosswords, Matching word-picture, Odd-one-out |
| SM-2 Integration | Chấm q-score 0-5 → cập nhật Spacing_Score (+0–2) |
| Feedback | Kết thúc: "Next Review In: X days" |
| Access | Free: chỉ review Still Learning & Almost Learnt. Pro: tất cả trạng thái. |

#### 5.3.7 Mastery Check (Pro only)

| Element | Spec |
|---|---|
| Goal | Kiểm tra toàn bộ từ trong learning card, xác nhận "Mastered." |
| Structure | (1) Fill-in-the-blank, (2) 5 short-sentence chunking — tập phát âm câu ví dụ, (3) Viết pinyin đúng tone, (4) Sắp xếp từ thành câu đúng ngữ pháp |
| Result | "Mastery Achieved: 87%" + spacing tiếp theo (30 days) |
| Gamification | Double XP + badge "Mastered Island" theo HSK level |
| Trùng lặp Recall | Max 15% câu hỏi trùng dạng bài với Recall mode |

---

## 6. Vocabulary Classification System (Memory Score)

### 6.1 Trạng thái ghi nhớ

| Trạng thái | Điều kiện | Review Interval |
|---|---|---|
| **Start Learning** | Memory Score = 0. Chưa hoàn thành mode nào. | I(1) = 1 day |
| **Still Learning** | Memory Score < 40. Đã học nhưng sai nhiều. | I(2) = 2 days |
| **Almost Learnt** | 40 ≤ Memory Score < 60. Nhớ cơ bản, chưa ổn định. | I(3) = 3-5 days |
| **Finish Learning** | 60 ≤ Memory Score < 80. Nhớ trong đa phần context. | I(4) = 7 days |
| **Memory Mode** | Memory Score ≥ 80 + spacing đúng ≥ 1 lần. Củng cố dài hạn. | I(5) = 14 days |
| **Mastered** | Memory Score ≥ 90 + spacing đúng ≥ 2 lần + không sai 10-14 ngày. | I(6) = 30 days. Reset nếu không ôn 60 ngày. |

### 6.2 Memory Score Formula

```
Memory Score (per word) =
  Σ(Mode_Score × Mode_Weight) + Spacing_Score
  ─────────────────────────────────────────────  × 100
                   Max_Points
```

**Mode Weights:**

| Mode | Weight | Max Weighted | Ghi chú |
|---|---|---|---|
| Discover | 1 | 1 | Recognition |
| Recall | 2 | 4 | Active recall — mode chính |
| Stroke (Guided) | 1 | 1 | Muscle memory |
| Stroke (Recall) | 2 | 4 | Deep character retention |
| Pinyin Drill | 1 | 1 | Secondary reinforcement |
| Chat (AI) | 2 | 4 | Deep retention, productive |
| Review (SM-2) | 2 | 4 | Long-term retention |
| Mastery Check | 2 | 4 | Validation |

**Max_Points:** Free = 11 (Discover 1 + Recall 4 + Review 4 + Spacing 2). Pro = 25 (all modes 23 + Spacing 2).

**Spacing_Score:**

| Condition | Score |
|---|---|
| Review đúng ngày (±1 day) và q ≥ 4 | +2 |
| Review trễ ≤ 3 ngày, q ≥ 3 | +1 |
| Review sớm hơn 2 ngày | +0.5 |
| Review trễ > 3 ngày hoặc q < 3 | 0 |

### 6.3 Upgrade Free → Pro Migration

Khi user upgrade: giữ toàn bộ data cũ → chuẩn hóa điểm từ thang max 11 sang 25 → nếu score sau chuẩn hóa < 40 → giảm 1 bậc state → giữ spacing history → mode mới sẽ update score dần.

---

## 7. Character Decomposition System

Giúp học viên hiểu cấu trúc Hán tự thay vì ghi nhớ thuần hình ảnh. Nghiên cứu cho thấy hiểu radical tăng tốc ghi nhớ 40-60%.

Mỗi flashcard hiển thị:
1. **Radical (bộ thủ):** Thành phần chính → nhóm nghĩa. VD: 语 → bộ 讠(ngôn ngữ) + 五 + 口
2. **Breakdown animation:** Tách chữ thành từng thành phần
3. **Memory hook:** Câu chuyện ghi nhớ. VD: 休 = 人 + 木 → "Người dựa vào cây = nghỉ ngơi"
4. **Related characters:** Các chữ cùng bộ thủ. VD: 讠→ 说, 话, 语, 读, 认

**Data:** Unihan database + CC-CEDICT + CJK Decomposition Data Project. Memory hooks: AI-generated, Phase 2 thêm human review cho top 500 từ.

---

## 8. Grammar Context System (MVP)

Không xây module riêng. Grammar gắn vào từng từ vựng dưới dạng Grammar Tips.

Mỗi learning card có section "Grammar":
1. **Pattern:** VD: 把 → `S + 把 + O + V + Complement`
2. **Example:** 1-2 câu highlight pattern. VD: 我**把**书**放在**桌子上。
3. **Rule ngắn:** 1-2 câu giải thích.
4. **Common mistake:** Lỗi người Việt hay mắc. VD: "Không dùng 把 với 是, 有, 知道."

**Data:** Chinese Grammar Wiki (AllSet Learning, CC) + HSK Standard Course index. MVP cover 80 grammar points cho HSK 1-3.

---

## 9. Video Flashcard — Seedance 2.0 (Phase 2)

### 9.1 Concept

Thay hình ảnh tĩnh trong Discover mode bằng short video 3-5 giây AI-generated, thể hiện nghĩa từ trong context thực tế.

VD: 跑步 (chạy bộ) → video người chạy trong công viên. 吃饭 (ăn cơm) → video người ngồi ăn.

**USP: "Flashcard sống"** — không app nào trên thị trường có. Viral potential cao khi user share video flashcard lên MXH.

### 9.2 Technical Evaluation (cần trước Phase 2)

| Factor | Question | Target |
|---|---|---|
| Cost per video | Seedance 2.0 pricing per generation? | < $0.05/video để viable ở scale |
| Latency | Time to generate 1 video? | < 10s (acceptable), < 5s (ideal) |
| Quality consistency | Video output stable chất lượng? | ≥ 85% usable without manual review |
| Caching strategy | Pre-generate cho HSK wordlists vs. on-demand? | Pre-generate HSK 1-6 (~5,400 words), on-demand cho user-created |
| Storage & CDN | Video storage cost + delivery? | Estimate cho 5,400 × 3-5s videos |
| Fallback | Nếu video generation fail? | Fallback to static image |

### 9.3 Viral Potential

- User share video flashcard lên TikTok/Douyin/Instagram Reels → organic reach
- "Xem video flashcard của tôi" → CTA download app
- GV share bộ video flashcard cho cả lớp

---

## 10. PrepAI Pronunciation — Detailed Architecture

### 10.1 MVP: SpeechSuper API

SpeechSuper đã có SDK cho Mandarin Chinese, hỗ trợ:
- Tone detection per syllable
- Initial/Final accuracy scoring
- Real-time feedback

**Integration flow:**
```
User tap mic → Record audio (WAV, 16kHz)
    → Send to SpeechSuper API
    → Receive: per-syllable scores (initial, final, tone)
    → Display results on UI
    → Update Memory Score (Pinyin Drill weight = 1)
```

**Pricing concern:** Cần confirm pricing tier cho expected volume (50K MAU × avg 10 checks/day = 500K API calls/day). Negotiate enterprise rate.

### 10.2 Phase 2: Custom PrepAI Model

Tái sử dụng architecture từ Prep IELTS Speaking (DeBERTa/PANN) nhưng fine-tune cho Mandarin:

| Component | Spec |
|---|---|
| Tone Detection | Pitch contour analysis model, fine-tuned trên Mandarin 4 tones + neutral |
| Initial/Final Scoring | Phoneme-level ASR → accuracy scoring per phoneme |
| Training Data | Cần ~10,000 labeled audio samples (4 tones × common syllables × multiple speakers) |
| Personalized Learning Plan | Aggregate user pronunciation data → identify weak phonemes/tones → generate adaptive drill queue |
| Advantage vs. SpeechSuper | Lower per-call cost at scale, deeper integration với Memory Score, richer weakness analysis, Prep ecosystem data flywheel |

### 10.3 Pronunciation Weakness Detection & Improvement Plan

Feature unique của Prep Chinese Vocab — không competitor nào có:

```
After N sessions (N ≥ 5):
    → Aggregate pronunciation scores across all words
    → Identify: Top 3 weakest initials, Top 2 weakest tones
    → Generate: "Your Pronunciation Focus"
        - "Thanh 2 (rising): 62% accuracy → Drill 10 words with tone 2"
        - "Âm đầu zh: 55% accuracy → Practice: 中, 知, 张, 住, 找"
    → Track improvement over time (weekly chart)
    → Adjust drill queue to prioritize weak areas
```

---

## 11. Social Features Roadmap (Phase 2)

### 11.1 Objective

Scale users & virality thông qua social mechanics. Không phải social network — mà là **social learning.**

### 11.2 Feature Stack (ordered by priority)

**P0 — Sharing & Discovery:**
- Folder vocab shareable → link mở app. Xem list từ nhưng phải login mới học được. CTA: "Save to track your own score."
- HSK Level Certificate: đạt milestone → certificate shareable MXH. (A1 <1,500 words, A2 ~2,500, B1 ~3,250, B2 ~4,000, C1 ~8,000, C2 ~8,000+)
- Streak Challenge: "Bạn đã học liên tục X ngày" → shareable card
- Character of the Day: 1 Hán tự/ngày + etymology + fun fact → shareable widget

**P1 — Competitive & Interactive:**
- 30-second Vocab Battle: thi đấu chọn nghĩa nhanh nhất + leaderboard
- Rapid Recall 2x: lướt reels tốc độ, true/false, 10 từ trong 20s. Hold 3s trái/phải → swipe down = 2x speed
- GV tạo class → assign wordlist → track progress học sinh

**P2 — AI Social Chat:**
- AI Chat room nâng cao: chat với nhân vật AI trong chủ đề
- Chat 1-1 bạn bè (AI hoặc real users)
- Nhóm chat (max 4 người)
- Ghép cặp ngẫu nhiên: matching users, chat đoạn ngắn, phải sử dụng 5 từ vựng đề bài

### 11.3 Viral Metrics Target (6 tháng post-Phase 2)

| Metric | Target |
|---|---|
| K-factor (viral coefficient) | ≥ 0.3 |
| Shareable content generated/MAU | ≥ 2/month |
| Organic installs from shares | ≥ 20% total installs |

---

## 12. Topic & Category System

10 topic chuẩn HSK:

1. 日常生活 (Daily Life) — Chào hỏi, gia đình, thời gian, số đếm
2. 饮食 (Food & Drink) — Món ăn, gọi món, nấu ăn
3. 交通旅行 (Travel & Transport) — Phương tiện, hỏi đường
4. 学习教育 (Education) — Trường học, thi cử
5. 工作商务 (Work & Business) — Công việc, thương mại
6. 健康医疗 (Health) — Bệnh viện, triệu chứng
7. 科技 (Technology) — Internet, thiết bị
8. 自然环境 (Nature) — Thời tiết, động vật
9. 文化娱乐 (Culture & Entertainment) — Phim, nhạc, lễ hội
10. 社会 (Society) — Luật pháp, kinh tế

**Lưu ý:** Tách khái niệm Topic (system-defined) và Folder/Deck (user-created). Học sinh tự tạo folder khi scan/import, có thể chọn topic tag (optional). 1 từ có thể thuộc nhiều topic nếu polysemy. Không giới hạn số từ per learning card.

---

## 13. Funnel & Monetization

### 13.1 Free vs. Premium

| Feature | Free | Premium (Prep HSK subscribers) |
|---|---|---|
| Cards/day | Max 20 | Unlimited |
| Scan/day | Max 3 ảnh | Unlimited |
| HSK Wordlists | HSK 1-3 | HSK 1-9 |
| Flashcard type | Text only | Text + Images (Phase 2: + Video) |
| Stroke & Recall | Guided xem only + Recall 5 từ/ngày | Full Guided + unlimited Recall + Speed Writing (Phase 2) |
| Pronunciation | Trial 3 từ/ngày | Unlimited + Weakness Report |
| Learning Modes | Discover + Recall + Review | All 7 modes (incl. Chat, Mastery) |
| Grammar | Tips giới hạn | Full context + Phase 2 module |
| AI Chat | Không | Unlimited |
| Ads | Non-intrusive | Ad-free |

> **⚠️ PROPOSED — Chưa finalize.** Monetization model dự kiến Freemium. Bảng trên là proposal ban đầu. Nhi Lâm (Commercial Lead) sẽ chủ trì thảo luận với team và xin ý kiến Sponsors (CTO + CEO) về monetization model **trước khi bắt đầu Sprint 1** (trước 01/04/2026). Tuyến (PO) tham gia thảo luận, cả hai cần đồng thuận trước khi đưa ra team (theo Team Charter mục 3.2). Free/Premium tiers, pricing, paywall triggers có thể thay đổi dựa trên quyết định này.

### 13.2 Conversion Triggers

1. **Soft paywall tại Stroke Practice:** Free xem animation nhưng không viết → CTA "Unlock Stroke Practice"
2. **HSK Level Gate:** Hoàn thành HSK 3 → "Ready for HSK 4? Upgrade."
3. **AI Chat preview:** Free 1 session/week → "Want more? Go Pro."
4. **Memory Score ceiling:** Free pathway rất chậm → "Reach Mastered faster with Pro."
5. **Weekly progress email:** "You learned X words. Unlock AI Chat to learn 2x faster." (data-driven nudge)

---

## 14. Learning Stats Dashboard

| Component | Spec |
|---|---|
| Header Stats | Words to Review Today + "Review Now" button. Conversations practiced count. |
| Memory State Breakdown | Tổng từ đã học + breakdown 6 trạng thái (bar chart hoặc donut) |
| HSK Level Progress | Progress bar per level (VD: HSK 1: 245/300) |
| **4-Dimension Tracking** | Progress theo 4 chiều HSK 3.0: Syllables, Characters, Vocabulary, Grammar |
| XP Tracker | Tổng XP tích lũy |
| Review Progress | Topics đã/đang/chưa review + "Finish Review Now" CTA |
| Pronunciation Insight | (Pro) Weak phonemes/tones summary + improvement trend |

---

## 15. Edge Cases & Solutions

### 15.1 Polysemy & Duplicates

| Case | Solution |
|---|---|
| 1 Hán tự nhiều nghĩa → topic nào? | Check duplicate word ID. Cùng nghĩa → 1 bản, "also appears in X topic." Khác nghĩa → flashcard riêng, user chọn topic. |
| Scan trùng từ ở nhiều folder | Same word ≠ meaning → keep both. Same word = meaning → confirm override? |
| Cùng pinyin khác Hán tự (homophone) | Recall quiz pinyin → hiển thị tất cả Hán tự cùng pinyin để phân biệt |
| Traditional vs Simplified | MVP: Simplified only. Settings toggle để hiển thị Traditional as reference (Phase 2). |

### 15.2 Learning Mode Issues

| Case | Solution |
|---|---|
| Discover: nghĩa câu ví dụ không khớp user input | Ưu tiên nghĩa user đã save. Câu ví dụ phải match meaning context. |
| AI Chat: phản hồi lạc topic | Kỹ thuật check chặt topic. Rate limiting + context buffer cho spam. |
| AI Chat: user dùng synonym không tính XP | NLP matching linh hoạt, nhận diện synonyms |
| Mastery trùng Recall | Max 15% câu trùng dạng bài |
| Stroke order nhiều chuẩn | Mặc định GB (Trung Quốc đại lục). Ghi chú nếu khác biệt phổ biến. |
| Pronunciation AI nghe sai (noise/accent) | Retry 2 lần. Vẫn fail → "Mark as correct" (không tính điểm) + suggest headphones. |

---

## 16. Technical Considerations

### 16.1 OCR Engine

- Primary: Google Cloud Vision API (printed text)
- Secondary: Baidu OCR API (handwritten Chinese)
- Open-source fallback: Tesseract fine-tuned
- Target: ≥ 90% printed, ≥ 80% handwritten
- Confidence < 70% → top-3 candidates

### 16.2 Stroke Order Data

- Make Me a Hanzi (open-source, ARPHIC License, 9,000+ characters, SVG paths)
- Stroke validation: compare user input stroke vs. reference (order + direction)
- Standard: GB (Trung Quốc đại lục)
- **Legal:** Cần confirm commercial use license với legal team

### 16.3 Character Decomposition

- Unihan database (Unicode Consortium) → radical data
- CC-CEDICT → nghĩa + pinyin
- CJK Decomposition Data Project → thành phần cấu tạo

### 16.4 Grammar Data

- Chinese Grammar Wiki (AllSet Learning) — Creative Commons
- HSK Standard Course textbook grammar index
- Custom mapping: HSK vocab → grammar points

### 16.5 Mobile Tech Stack

Cần quyết định: React Native vs. Flutter (ADR trong Sprint 0, Thái + CTO review). Factors:
- Handwriting canvas performance (critical cho Stroke mode)
- Audio recording/playback quality (critical cho Pronunciation)
- Existing Prep team expertise
- Animation performance cho character decomposition

### 16.6 Development Approach — API-First + Vibe Coding

**API-First (from Team Charter):**
- Thái (BE) define và deliver API contracts/mocks **trước Sprint 1** → Cường (Mobile) và Chi (QC) develop/test song song, không stuck chờ backend.
- Contract-first development: FE code against mocks → integration khi real API ready.

**Vibe Coding Boundaries (CTO define trong Sprint 0):**

| Cho phép Vibe Code (Claude Code) | Cần Manual coding |
|---|---|
| UI components, screens, layouts | Data layer, database schema |
| CRUD endpoints | Authentication, authorization |
| Test case generation | Core business logic (Memory Score, SM-2) |
| API client code | Encryption, security |
| Styling, animations | Integration points với Prep platform |

Mọi AI-generated code phải qua PR review (Thái ↔ Cường cross-review) + CI pass.

### 16.7 Data Strategy — Future-proof cho Adaptive Learning

Không tích hợp adaptive learning system từ ngày đầu, nhưng data phải sẵn sàng:

- **Log mọi learning event:** word learned, quiz result, time spent, difficulty level, pronunciation scores, stroke accuracy, revision history.
- **`learner_id` consistent** với Prep platform User Service (dùng chung, không tạo riêng).
- **Event log pattern:** khi integrate adaptive system, có thể **replay toàn bộ history (backfill)** vào ATLAS/BKT.
- **CTO review data schema** tuần đầu Sprint 1 để đảm bảo không phải redesign.
- Schema phải support multi-dimension tracking: Syllables, Characters, Vocabulary, Grammar (HSK 3.0 requirements).

---

## 17. Dependencies & Integration

### 17.1 Internal Squad Dependencies

| Dependency | PIC trong Squad | Description |
|---|---|---|
| API contracts & mocks | Thái (BE) | Sẵn sàng trước Sprint 1. CTO review. |
| Mobile app build | Cường (Mobile) | Develop trên API mocks → integration |
| QA strategy & testing | Chi + Tuyết | Shift-left: test cases trước sprint start |
| UI/UX design | Khánh | AI x Figma lean approach, focus UX flow + brand |
| Product spec & acceptance criteria | Tuyến (PO) | Feature spec, backlog, acceptance criteria |
| Market brief & GTM | Nhi (Commercial Lead) | Market research, monetization, GTM plan |
| Monetization model | Nhi (PIC) + Tuyến (co-own) | Thảo luận + Sponsors approval trước Sprint 1 |

### 17.2 External Dependencies

| Dependency | Owner | Description |
|---|---|---|
| Prep User Service | Platform team | Sử dụng chung User Service (không tạo riêng). SSO, `learner_id` consistent. |
| Prep HSK subscription system | Growth Squad | Đồng bộ trạng thái Pro nếu Premium = Prep HSK subscriber. |
| PrepAI Scoring Engine | AI Squad | SpeechSuper API (MVP), Custom Mandarin model (Phase 2) |
| Content: HSK wordlists + grammar | Content team | 1,000 từ HSK 1-3: wordlists, grammar tips, audio. Academic Director review accuracy. |
| Design System | Design team | Khánh dùng Design System có sẵn cho lean approach |
| DevOps | Tuấn Bùi (shared) | CI/CD pipeline, dev environment setup |
| CTO | Nam | Architecture review (Sprint 0 Gate 1), data schema review (Sprint 1 W1), conflict escalation |
| Academic Director | — | Review HSK wordlist accuracy, grammar points, cung cấp SMEs nếu cần |
| Seedance 2.0 API | AI Squad / External | Phase 2: Video Flashcard evaluation |

### 17.3 RACI cho Key Decisions (from Team Charter)

| Decision | Tuyến (PO) | Nhi (Commercial) | Engineer | CTO |
|---|---|---|---|---|
| Feature spec & acceptance criteria | **R/A** | Informed | Consulted | — |
| Roadmap prioritization | **A (final call)** | Consulted (propose & debate) | Consulted | Consulted |
| Architecture & tech stack | Consulted | — | **R/A** | **Review** |
| Pricing & monetization | Consulted | **R/A** | — | Consulted |
| GTM & growth | Consulted | **R/A** | — | Informed |
| Kill / pivot / scale | **A** | **A** | Consulted | **A** |

---

## 18. Timeline & Milestones (aligned with Team Charter)

### 18.1 Sprint 0 — Explore & Kick-off (Mar 15-31, 2026)

Không có delivery pressure. CTO support team làm product brief + backlog, giảm tải cho Tuyến (50% allocation).

| Deliverable | PIC | Deadline |
|---|---|---|
| PRD v3 Approved (tài liệu này) | Tuyến + CTO support | Mar 15 |
| Product Brief v1 (target user, core features MVP, success metrics, out-of-scope) | Tuyến + CTO support | Mar 31 |
| MVP Backlog v1 (user stories, priority, acceptance criteria) | Tuyến + CTO support | Mar 31 |
| Market Brief v1 (competitive landscape, target segment, initial GTM hypothesis) | Nhi | Mar 31 |
| **Monetization model decision** | **Nhi (PIC) + Tuyến + Sponsors** | **Mar 31** |
| Architecture Decision Record (data model, API contracts, tech stack, integration points, vibe coding boundaries) | Thái (CTO review) | Mar 28 |
| API contracts & mocks ready (API-first: FE + QC không bị stuck chờ BE) | Thái | Mar 28 |
| Dev environment setup, CI/CD pipeline, vibe coding workflow | Cường + Tuấn Bùi (DevOps) | Mar 28 |
| Core UI patterns mapped to Design System, key screens wireframe | Khánh | Mar 31 |
| QA strategy document (AI testing spike plan, quality gates, test coverage) | Chi + Tuyết | Mar 31 |
| Standard templates (user story, product brief, market brief, ADR, test case) | Tuyến + CTO | Mar 28 |
| **CTO review data schema** (đảm bảo compatible ATLAS/BKT cho backfill) | CTO | Sprint 1 tuần 1 |
| HSK 1-3 wordlist content prep started | Content team | Mar 31 |
| Academic Director review: HSK wordlist accuracy, grammar points | Academic Director | Ongoing |

### 18.2 Phase 1 — MVP Build (Apr 1 → mid-May 2026)

MVP clock bắt đầu **01/04/2026**. Engineering capacity: **70% product / 30% growth experiments** (Charter mục 5.1). PO plans trên 70% capacity.

| Sprint | Dates | Focus | QC Track |
|---|---|---|---|
| **Sprint 1** | Apr 1-14 | Core flow: OCR Scan → Flashcard → Recall → Review. API integration (Thái API → Cường mobile). SpeechSuper API integration. | AI testing spike bắt đầu. Test cases từ API contracts. |
| **Sprint 2** | Apr 15-28 | Stroke & Recall mode. Pinyin Drill. Memory Score logic. Dashboard basic. | AI testing spike tiếp tục. Testing trên mocks + real API. Sprint 2 Review: Collaboration Check (Charter milestone). |
| **Sprint 3 (stretch)** | Apr 29-May 12 | AI Chat mode. Mastery Check. HSK wordlist full integration. Polish + performance. | Spike evaluation. Integration testing. |
| **Buffer + Polish** | May 13-26 | Bug fix, performance optimization, App Store/Play Store submission. | Chi sign-off. QA final. |

**Milestones:**

| Milestone | Date | Description |
|---|---|---|
| API contracts + mocks ready | Mar 28 | Gate 1 pass → FE + QC can start parallel |
| CTO architecture review | Mar 28 | Gate 1: data model, vibe coding boundaries |
| HSK 1-3 Content Ready | Apr 14 (Sprint 1 end) | 1,000 words: wordlists + grammar tips + decomposition + audio |
| SpeechSuper Integration working | Apr 14 | Pronunciation scoring end-to-end |
| Alpha Build (core flow) | Apr 14 | Scan → Flashcard → Recall → Review working |
| Beta Build (full features) | Apr 28 | All 7 modes + Memory Score + Dashboard |
| **MVP Stretch Target** | **Apr 29** | If team velocity high |
| QA Sign-off | May 12 | Chi + Tuyết final testing |
| **MVP Realistic Target** | **May 13** | Public release target |
| App Store submission buffer | May 13-26 | Review process + hotfix buffer |

### 18.3 Phase 2 Timeline

| Milestone | Date | Description |
|---|---|---|
| Phase 2 Kickoff | Post-MVP | Video Flashcard eval + Social features spec |
| Phase 2 Launch | Q4 2026 | Video Flashcard + Social + Translation module |

### 18.4 Quality Gates (from Team Charter)

**Gate 1 — Architecture Review (Sprint 0, CTO review):**
- Data model, API contract, tech stack decisions
- API-first: contracts/mocks sẵn sàng trước Sprint 1
- Integration points với Prep platform (auth, user system, event logging)
- Vibe coding boundaries: data layer + auth → manual; UI components + CRUD → vibe code okay
- Data schema compatible ATLAS/BKT để backfill adaptive learning sau

**Gate 2 — Code Review (mỗi PR, Sprint 1+):**
- Vibe coding (Claude Code) → cross PR review (Thái ↔ Cường) → CI pass
- Automated: linting, type checking, unit test coverage minimum
- Không merge nếu CI fail

**Gate 3 — QA Sign-off (per sprint):**
- Chi + Tuyết define test cases trước sprint start (shift-left)
- AI-generated test cases (spike) + manual exploratory testing
- Release checklist trước mỗi deployment

---

## 19. HSK 3.0 Alignment

### 19.1 Key Changes Affecting App

| Change | Impact | Action |
|---|---|---|
| 9 levels (thay vì 6) | Progression system, UI, wordlists, badges | P0: Implement 9-level system |
| Vocabulary +120% (11,000 từ) | Database, content scope | P0: Dùng syllabus Nov 2025 |
| 4-dimension assessment | Dashboard, progress tracking | P0: Track Syllables + Characters + Vocab + Grammar |
| Recognition vs. Writing | Stroke Practice mode | P1: "Recognition Drill" (L1-4), "Writing Drill" (L5+) |
| Speaking mandatory (L3+) | Pinyin Drill + AI Chat | P1: Stronger speaking component từ HSK 3 |
| Translation (L4+, mới) | Phase 2 opportunity | P0 Phase 2: CN↔VN → CN↔TH, CN↔ID |

### 19.2 Competitive Gap

Từ phân tích sơ bộ 7 đối thủ (SuperTest, HelloChinese, Duolingo, ChineseSkill, SuperChinese, HSK Lord, Pleco):

1. Không ai fully cover HSK 3.0 với 9 levels
2. Không có AI Writing + Speaking scoring engine
3. Zero translation practice (kỹ năng mới HSK 3.0)
4. Weak SEA localization
5. Không có transition support HSK 2.0 → 3.0

> **⚠️ Cần nghiên cứu thêm:** Phân tích trên dựa trên desk research ban đầu và có thể chưa phản ánh đầy đủ competitive landscape hiện tại. Nhi (Commercial Lead, R/A cho Market Research theo Team Charter RACI) sẽ thực hiện competitive analysis chuyên sâu trong Sprint 0 Market Brief, bao gồm: (1) verify từng gap claim với data cập nhật, (2) đánh giá các competitors mới hoặc feature updates gần đây, (3) phân tích pricing & positioning của từng đối thủ, (4) xác định sustainable moats vs. temporary gaps. Kết quả sẽ feed vào Sprint 1 prioritization.

### 19.3 Timeline Alignment

| Event | Date | Implication |
|---|---|---|
| HSK 3.0 Trial Exams (168 countries) | Jan 2026 | User awareness tăng |
| Squad Sprint 0 Kick-off | Mar 15, 2026 | Explore & align |
| MVP Clock Start | Apr 1, 2026 | Sprint 1 bắt đầu |
| **Prep Chinese Vocab MVP (stretch)** | **Apr 29, 2026** | **Trước HSK 3.0 deploy ~2.5 tháng** |
| **Prep Chinese Vocab MVP (realistic)** | **May 13, 2026** | **Trước HSK 3.0 deploy ~2 tháng** |
| HSK 3.0 Full Deployment | July 2026 | Monitor syllabus changes cuối cùng |
| Phase 2 Target | Q4 2026 | Video Flashcard + Social + Translation |

---

## 20. Risks & Mitigations

| # | Risk | Severity | Mitigation |
|---|---|---|---|
| 1 | **PO 50% allocation:** Tuyến song song Learning Squad → spec/acceptance criteria bị delay → block cả team | **High** | CTO support Sprint 0 (làm product brief + backlog). Nhi support product work nếu cần (Charter cho phép). Đánh giá lại workload sau Phase 1. |
| 2 | **70/30 capacity split:** Chỉ 70% engineering capacity cho product → timeline tighter | **High** | Sprint planning chỉ plan trên 70%. Growth experiments (30%) do Nhi drive, không ảnh hưởng architecture. Negotiate 100% cho Sprint 1 nếu cần. |
| 3 | **Content bottleneck:** 1,000 từ HSK 1-3 cần grammar tips + ví dụ + audio trước Sprint 1 end | **Medium** | Start content prep từ Sprint 0. Academic Director review. Phân chia: AI generate draft → human validate. |
| 4 | **SpeechSuper API cost:** 500K calls/day ở scale có thể expensive | **Medium** | Negotiate enterprise rate sớm. Fallback: rate limit free users. Phase 2: custom model giảm per-call cost. |
| 5 | **Handwriting canvas performance:** Mobile writing experience cần smooth — nếu lag sẽ kill UX | **Medium** | Tech spike trong Sprint 0 (Cường). Framework choice ảnh hưởng trực tiếp. |

## 21. Open Questions

| # | Question | Status | Owner | Deadline |
|---|---|---|---|---|
| 1 | SpeechSuper API pricing cho 500K calls/day? Enterprise rate? | **Open** | Thái | Sprint 0 |
| 2 | HSK 3.0 Nov 2025 wordlist final chưa? Monitor CTI. | **Open** | Content team | Sprint 0 |
| 3 | Make Me a Hanzi ARPHIC License — commercial use OK? | **Open** | Legal | Sprint 0 |
| 4 | Content team bandwidth cho 1,000 từ HSK 1-3 trước Sprint 1 end? | **Open** | Content Lead | Sprint 0 |
| 5 | Mobile stack: React Native vs Flutter? (ADR) | **Open** | Thái + Cường (CTO review) | Mar 28 |
| 6 | Seedance 2.0 pricing & latency evaluation? | **Open** | AI Squad | Phase 2 |
| 7 | Monetization model final (Freemium tiers, pricing)? | **Open** | Nhi + Tuyến + Sponsors | Mar 31 |
| 8 | App direction: Chinese-only vs. multi-language vocab? | **Open** | Tuyến + Nhi | Sprint 0 |

---

## 22. Appendix: Key Flows Summary

### Flow 1: First-time User

```
Download → Onboarding (chọn HSK level hoặc "I don't know")
    → Nếu "I don't know" → Quick placement test (20 words, 2 min)
    → Chọn level → Load wordlist → Discover mode (first 10 words)
    → Complete Discover → Prompt: "Ready to test yourself?" → Recall mode
    → Complete Recall → Dashboard showing progress
    → Push notification next day: "5 words to review!"
```

### Flow 2: Scan Notes

```
Home → Tap "Scan" → Camera / Gallery
    → OCR processing → Preview detected characters
    → Edit/Confirm → Duplicate check → Assign folder
    → Generate flashcards → Start learning
```

### Flow 3: Daily Learning Session

```
Open app → Dashboard shows: "12 words to review" + "Continue HSK 2"
    → Option A: "Review Now" → Review mode (SM-2 scheduled words)
    → Option B: "Continue" → Next mode in Smart Learning Path
    → Session complete → Summary: words learned, XP, streak
```

### Flow 4: Pronunciation Improvement

```
Pinyin Drill → Record pronunciation → PrepAI scores per syllable
    → Red/Green overlay on each phoneme
    → Session end → Weakness summary
    → After 5+ sessions → "Your Pronunciation Focus" generated
    → Adaptive drill queue prioritizes weak areas
```

---

*PRD v3 — Aligned với Learning Utilities Squad Charter v2.0. Cập nhật từ v2: timeline theo Sprint cadence (Sprint 0: Mar 15, MVP: Apr 29 stretch / May 13 realistic), standalone + ecosystem integration positioning, team roles RACI, API-first + vibe coding approach, data strategy cho adaptive learning backfill, Quality Gates, Risk section. Thêm: Video Flashcard (Seedance 2.0) roadmap, PrepAI pronunciation architecture, Handwriting multi-mode, Social features roadmap, Global-ready positioning.*

*Pending review: Tuyến (PO), Nhi (Commercial Lead), Thái (BE), Cường (Mobile), Chi (QC), Khánh (Design), CTO (Nam), Academic Director.*
