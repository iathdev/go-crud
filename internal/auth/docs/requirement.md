# Auth Module — Requirement (trích từ PRD v3)

> Trích nguyên văn từ `docs/requirement.md` — Prep Chinese Vocab PRD v3. Chỉ lấy các phần liên quan auth module.

---

## 1. Product Type — PRD Section 2.3

**Standalone App, Deep Ecosystem Integration:**

App phát hành độc lập trên App Store / Play Store để tối ưu acquisition funnel. Tuy nhiên, app được thiết kế như một phần mở rộng của hệ sinh thái Prep — không phải standalone thuần túy:

- **User Service chung:** Sử dụng chung User Service của Prep platform (không tạo user system riêng) → tránh data migration về sau, đảm bảo `learner_id` consistent xuyên suốt ecosystem.
- **Deep data sync:** Learning progress, vocab level, weak areas sync realtime với Prep HSK → feed vào adaptive learning system (khi sẵn sàng).
- **SSO:** Single Sign-On giữa Prep Chinese Vocab ↔ Prep HSK ↔ các app Prep khác.
- **Có thể tích hợp vào app Prep hiện tại** dưới dạng embedded module (WebView hoặc native module) nếu strategy thay đổi — architecture phải support cả hai hướng.
- **Standalone cho acquisition:** User có thể dùng app không cần Prep HSK subscription, nhưng account vẫn là Prep account.

> **Lưu ý:** Team Charter (v2.0, mục 1.5) ghi "utility products kết nối trực tiếp với nền tảng Prep — không phải standalone apps mà là phần mở rộng của ecosystem." PRD approach: standalone packaging cho acquisition, nhưng bản chất là ecosystem extension về mặt data/account/integration.

---

## 2. First-time User Flow — PRD Section 22, Flow 1

```
Download → Onboarding (chọn HSK level hoặc "I don't know")
    → Nếu "I don't know" → Quick placement test (20 words, 2 min)
    → Chọn level → Load wordlist → Discover mode (first 10 words)
    → Complete Discover → Prompt: "Ready to test yourself?" → Recall mode
    → Complete Recall → Dashboard showing progress
    → Push notification next day: "5 words to review!"
```

---

## 3. Free vs Premium — PRD Section 13.1

> **⚠️ PROPOSED — Chưa finalize.** Monetization model dự kiến Freemium. Nhi Lâm (Commercial Lead) sẽ chủ trì thảo luận với team và xin ý kiến Sponsors (CTO + CEO) về monetization model **trước khi bắt đầu Sprint 1** (trước 01/04/2026). Free/Premium tiers, pricing, paywall triggers có thể thay đổi dựa trên quyết định này.

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

---

## 4. Conversion Triggers — PRD Section 13.2

1. **Soft paywall tại Stroke Practice:** Free xem animation nhưng không viết → CTA "Unlock Stroke Practice"
2. **HSK Level Gate:** Hoàn thành HSK 3 → "Ready for HSK 4? Upgrade."
3. **AI Chat preview:** Free 1 session/week → "Want more? Go Pro."
4. **Memory Score ceiling:** Free pathway rất chậm → "Reach Mastered faster with Pro."
5. **Weekly progress email:** "You learned X words. Unlock AI Chat to learn 2x faster." (data-driven nudge)

---

## 5. Dependencies — PRD Section 17

### 5.1 External Dependencies (PRD Section 17.2)

| Dependency | Owner | Description |
|---|---|---|
| Prep User Service | Platform team | Sử dụng chung User Service (không tạo riêng). SSO, `learner_id` consistent. |
| Prep HSK subscription system | Growth Squad | Đồng bộ trạng thái Pro nếu Premium = Prep HSK subscriber. |

---

## 6. Technical Considerations — PRD Section 16

### 6.1 Development Approach (PRD Section 16.6)

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

### 6.2 Data Strategy (PRD Section 16.7)

- **Log mọi learning event:** word learned, quiz result, time spent, difficulty level, pronunciation scores, stroke accuracy, revision history.
- **`learner_id` consistent** với Prep platform User Service (dùng chung, không tạo riêng).
- **Event log pattern:** khi integrate adaptive system, có thể **replay toàn bộ history (backfill)** vào ATLAS/BKT.
- **CTO review data schema** tuần đầu Sprint 1 để đảm bảo không phải redesign.

---

## 7. Timeline — PRD Section 18

| Milestone | Date | Description |
|---|---|---|
| API contracts + mocks ready | Mar 28 | Gate 1 pass → FE + QC can start parallel |
| CTO architecture review | Mar 28 | Gate 1: data model, vibe coding boundaries |
| **Monetization model decision** | **Mar 31** | Nhi (PIC) + Tuyến + Sponsors |
| Alpha Build (core flow) | Apr 14 | Scan → Flashcard → Recall → Review working |
| MVP Stretch Target | Apr 29 | If team velocity high |
| MVP Realistic Target | May 13 | Public release target |

---

## 8. Quality Gates — PRD Section 18.4

**Gate 1 — Architecture Review (Sprint 0, CTO review):**
- Data model, API contract, tech stack decisions
- API-first: contracts/mocks sẵn sàng trước Sprint 1
- Integration points với Prep platform (auth, user system, event logging)
- Vibe coding boundaries: data layer + auth → manual; UI components + CRUD → vibe code okay
- Data schema compatible ATLAS/BKT để backfill adaptive learning sau
