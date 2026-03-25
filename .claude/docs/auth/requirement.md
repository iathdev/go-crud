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

## 2. Dependencies — PRD Section 17

### 5.1 External Dependencies (PRD Section 17.2)

| Dependency | Owner | Description |
|---|---|---|
| Prep User Service | Platform team | Sử dụng chung User Service (không tạo riêng). SSO, `learner_id` consistent. |
| Prep HSK subscription system | Growth Squad | Đồng bộ trạng thái Pro nếu Premium = Prep HSK subscriber. |

---
