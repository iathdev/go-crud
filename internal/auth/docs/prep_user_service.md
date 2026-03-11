# Prep User Service — API & Integration Analysis

> Phân tích từ codebase `prep-accounts-laravel` + test thực tế trên `api-gw.testsprep.online`

---

## 1. Tổng quan

Prep User Service là hệ thống auth trung tâm của Prep platform (Laravel + Passport). Go app **không tự quản lý login** — mobile app login trực tiếp với Prep, nhận Prep access_token, rồi gửi token đó cho Go backend.

Go backend hỗ trợ **2 flow authentication** tuỳ theo nhu cầu:

| Flow | Mô tả | Go backend làm gì |
|---|---|---|
| **Flow A: JWT Local** | Go tạo JWT riêng, mobile dùng JWT cho các request sau | Gọi Prep `/me` 1 lần khi login, sau đó verify JWT offline |
| **Flow B: Prep Token Direct** | Mobile dùng Prep token trực tiếp cho mọi request | Mỗi request Go đều gọi Prep `/me` để validate |

---

## 2. API Endpoints liên quan

### 2.1 Endpoints mobile gọi trực tiếp (Prep platform)

| Method | Endpoint | Auth | Mô tả |
|---|---|---|---|
| `POST` | `/auth/api/v1/auth/account-login` | Không | Login bằng username/password |
| `POST` | `/auth/api/v1/auth/phone-login` | Không | Login bằng phone (password hoặc OTP) |
| `POST` | `/auth/api/v1/auth/google-login` | Không | Login bằng Google OAuth |
| `POST` | `/auth/api/v1/auth/logout` | Bearer token | Logout Prep session |

### 2.2 Endpoint Go backend gọi (validate token)

| Method | Endpoint | Auth | Mô tả |
|---|---|---|---|
| `GET` | `/auth/api/v1.1/auth/me` | Bearer (Prep token) | Lấy user info từ Prep token |

### 2.3 Go backend API

| Method | Endpoint | Auth | Flow A | Flow B |
|---|---|---|---|---|
| `POST` | `/login` | Prep token | Validate + tạo JWT local | Không dùng |
| `POST` | `/refresh` | Refresh token | Refresh JWT local | Không dùng |
| `POST` | `/logout` | JWT | Xoá refresh token Redis | Không dùng |
| `GET` | `/me` | JWT hoặc Prep token | Trả profile từ DB local | Validate Prep token + trả profile |

### 2.4 Response Prep `/me` (đã confirm thực tế)

```json
{
  "data": {
    "id": 446416,
    "username": "FEDD000011",
    "phone": null,
    "email": "",
    "name": "thaidong",
    "is_first_login": false,
    "force_update_password": false,
    "avatar": null,
    "dob": null,
    "province_id": null,
    "has_agreed_policies_at": null,
    "b2b_id": 13876,
    "b2b_logo": "https://...",
    "user_country": null,
    "user_province": null,
    "career": null,
    "is_white_listed": 0
  },
  "message": "OK",
  "token": {
    "aud": "9",
    "jti": "f67496935fdb...",
    "sub": "446416",
    "exp": 1781597878.098545,
    "iat": 1773821878.102288,
    "nbf": 1773821878.102289
  }
}
```

### 2.5 Fields Go app sử dụng

| Field trong response | Map sang | Kiểu | Ghi chú |
|---|---|---|---|
| `data.id` | `prep_user_id` | `int64` | ID user trên Prep platform |
| `data.name` | `name` | `string` | Tên user |
| `data.email` | `email` | `string` | Có thể rỗng `""` |
| `data.is_first_login` | `is_first_login` | `bool` | User mới chưa login lần nào trên Prep |

### 2.6 Fields Go app KHÔNG sử dụng (thuộc Prep platform)

`username`, `phone`, `avatar`, `dob`, `province_id`, `b2b_id`, `b2b_logo`, `user_country`, `user_province`, `career`, `is_white_listed`, `has_agreed_policies_at`, `force_update_password`, `token.*`

---

## 3. Luồng Login

### 3.1 Flow A: JWT Local

Mobile login Prep → gửi Prep token cho Go 1 lần → Go tạo JWT local → mobile dùng JWT cho mọi request sau.

```
┌─────────────┐     ┌──────────────────┐     ┌─────────────────┐     ┌──────────────┐
│  Mobile App  │     │  Prep Platform    │     │  Go Backend     │     │   Redis      │
│              │     │ (accounts-laravel)│     │  (auth module)  │     │              │
└──────┬───────┘     └────────┬─────────┘     └────────┬────────┘     └──────┬───────┘
       │                      │                        │                     │
       │  1. Login             │                        │                     │
       │  (Google/Phone/Pass)  │                        │                     │
       │─────────────────────>│                        │                     │
       │                      │                        │                     │
       │  2. Prep access_token │                        │                     │
       │<─────────────────────│                        │                     │
       │                      │                        │                     │
       │  3. POST /login       │                        │                     │
       │  {token: "prep_..."}  │                        │                     │
       │──────────────────────────────────────────────>│                     │
       │                      │                        │                     │
       │                      │  4. GET /me             │                     │
       │                      │  Bearer: prep_token     │                     │
       │                      │<───────────────────────│                     │
       │                      │                        │                     │
       │                      │  5. {id, name, email}   │                     │
       │                      │───────────────────────>│                     │
       │                      │                        │                     │
       │                      │                        │  6. Upsert user     │
       │                      │                        │  (prep_user_id=id)  │
       │                      │                        │──── DB ────>        │
       │                      │                        │                     │
       │                      │                        │  7. Save refresh    │
       │                      │                        │     token hash      │
       │                      │                        │────────────────────>│
       │                      │                        │                     │
       │  8. Response:         │                        │                     │
       │  {access_token (JWT), │                        │                     │
       │   refresh_token,      │                        │                     │
       │   is_first_login,     │                        │                     │
       │   profile}            │                        │                     │
       │<──────────────────────────────────────────────│                     │
       │                      │                        │                     │
       │  [Nếu is_first_login] │                        │                     │
       │  9. Hiện onboarding   │                        │                     │
       │  → Chọn HSK level     │                        │                     │
       │  POST /api/me/        │                        │                     │
       │       onboarding      │                        │                     │
       │──────────────────────────────────────────────>│                     │
```

### Giải thích:

1. **Mobile app** login trực tiếp với Prep platform (Google/Phone/Password)
2. Prep trả **Prep access_token** (Laravel Passport Bearer token, TTL 90 ngày)
3. Mobile gửi Prep token tới **Go backend** qua `POST /login`
4. Go backend gọi **`GET /me`** tới Prep API Gateway, dùng Prep token
5. Prep xác nhận token hợp lệ → trả user data
6. Go backend **upsert** vào bảng `users` (tạo mới nếu chưa có, cập nhật name/email nếu đã có)
7. Go backend gen **JWT riêng** (24h access + 7 ngày refresh)
8. Trả về cho mobile: JWT local + profile + `is_first_login`
9. Nếu `is_first_login` → mobile hiện onboarding → user chọn HSK level

### Ưu điểm:

- **Nhanh**: Sau login, mọi request chỉ cần verify JWT signature (offline, ~μs). Không gọi HTTP ra ngoài.
- **Độc lập**: Prep down → user đã login vẫn dùng app bình thường cho đến khi JWT hết hạn.
- **Bảo mật**: JWT 24 giờ, token lộ thì hết hạn nhanh. Prep token 90 ngày thì rủi ro cao hơn.

### Nhược điểm:

- **Phức tạp hơn**: Go backend phải quản lý JWT, refresh token, Redis, logout.
- **Mobile phải gọi 2 hệ thống**: Login Prep + POST /login Go.

---

### 3.2 Flow B: Prep Token Direct

Mobile login Prep → dùng Prep token trực tiếp cho mọi request tới Go. Login, logout, refresh đều gọi Prep. Go chỉ phục vụ `GET /me` (profile local).

```
┌─────────────┐     ┌──────────────────┐     ┌─────────────────┐
│  Mobile App  │     │  Prep Platform    │     │  Go Backend     │
│              │     │ (accounts-laravel)│     │  (auth module)  │
└──────┬───────┘     └────────┬─────────┘     └────────┬────────┘
       │                      │                        │
       │  1. Login             │                        │
       │  (Google/Phone/Pass)  │                        │
       │─────────────────────>│                        │
       │                      │                        │
       │  2. Prep access_token │                        │
       │<─────────────────────│                        │
       │                      │                        │
       │  3. GET /me           │                        │
       │  Bearer: prep_token   │                        │
       │──────────────────────────────────────────────>│
       │                      │                        │
       │                      │  4. GET /me             │
       │                      │  Bearer: prep_token     │
       │                      │<───────────────────────│
       │                      │                        │
       │                      │  5. {id, name, email}   │
       │                      │───────────────────────>│
       │                      │                        │
       │                      │                        │  6. Upsert user
       │                      │                        │  (prep_user_id=id)
       │                      │                        │──── DB ────>
       │                      │                        │
       │  7. Response:         │                        │
       │  {profile,            │                        │
       │   is_first_login}     │                        │
       │<──────────────────────────────────────────────│
       │                      │                        │
       │  8. Logout            │                        │
       │  (khi cần)            │                        │
       │─────────────────────>│                        │
       │                      │                        │
```

### Giải thích:

1. **Mobile app** login trực tiếp với Prep platform (Google/Phone/Password)
2. Prep trả **Prep access_token** (TTL 90 ngày)
3. Mobile gửi mọi request tới Go backend với **Prep token** trong header `Authorization`
4. Go backend gọi Prep **`GET /me`** để validate token + lấy user info
5. Prep xác nhận token hợp lệ → trả user data
6. Go backend **upsert** user vào DB local
7. Trả profile local cho mobile
8. Logout → mobile gọi **trực tiếp Prep**, Go backend không liên quan

### Ưu điểm:

- **Đơn giản**: Go backend không cần quản lý JWT, refresh token, Redis cho auth. Không cần `/login`, `/refresh`, `/logout` trên Go.
- **Mobile chỉ quản lý 1 token**: Prep token dùng cho cả Prep platform lẫn Go backend.

### Nhược điểm:

- **Chậm**: Mỗi request protected đều gọi HTTP tới Prep `/me` (~100-500ms). Circuit breaker quan trọng hơn.
- **Phụ thuộc Prep**: Prep down → Go backend không thể xác thực bất kỳ request nào → app chết hoàn toàn.
- **Token TTL 90 ngày**: Token lộ thì attacker dùng được rất lâu, không có cơ chế revoke phía Go.

---

### 3.3 So sánh 2 flow

| Tiêu chí | Flow A: JWT Local | Flow B: Prep Token Direct |
|---|---|---|
| Độ phức tạp Go backend | Cao (JWT + Redis + refresh) | Thấp (chỉ gọi Prep `/me`) |
| Latency per request | Thấp (verify JWT offline) | Cao (HTTP call tới Prep mỗi request) |
| Phụ thuộc Prep runtime | Chỉ khi login | Mọi request |
| Token security | JWT 24 giờ | Prep token 90 ngày (do Prep platform quản lý) |
| API trên Go | `/login`, `/refresh`, `/logout`, `/me` | `/me` |
| Mobile complexity | Quản lý 2 token (Prep + JWT) | Quản lý 1 token (Prep) |
| Offline capability | Có (JWT còn hạn thì dùng được) | Không |
| Token revoke realtime | Không — JWT còn hạn vẫn dùng được (tối đa 24h) | Không cache: realtime. Có cache: trễ bằng cache TTL |

#### Vấn đề Token Revoke

Cả 2 flow đều có khoảng trễ khi Prep revoke token:

- **Flow A**: JWT đã phát hành thì không thể thu hồi. User bị revoke bên Prep nhưng JWT còn hạn → vẫn truy cập Go backend được tối đa 24h.
- **Flow B không cache**: Realtime 100% — mỗi request đều gọi Prep `/me`, token bị revoke thì reject ngay.
- **Flow B có cache**: Cache chưa hết hạn → Go vẫn cho qua dù Prep đã revoke.

**Giải pháp nếu cần realtime revoke:**

1. **Webhook từ Prep** — Prep gửi event khi revoke token → Go lưu vào blacklist Redis → middleware check blacklist trước khi cho qua. Cần Prep hỗ trợ.
2. **Short cache** — Flow B với cache TTL ngắn (1-2 phút). Cân bằng giữa performance và realtime.
3. **Flow A + revoke callback** — Prep gọi callback tới Go khi revoke → Go xoá refresh token + thêm JWT vào blacklist Redis (TTL = thời gian JWT còn lại). Cần Prep hỗ trợ.

> Nếu Prep không hỗ trợ webhook/callback thì chấp nhận khoảng trễ — đây là trade-off phổ biến trong các hệ thống distributed auth.

---

## 4. Key Design Decisions

### 4.1 `prep_user_id` là `int64`, không phải UUID

Prep platform dùng auto-increment integer ID. Go app lưu nguyên dạng `int64`.

### 4.2 `tier` là logic riêng của Go app

Prep platform có `user_markets` và `user_product_categories` nhưng không có field `tier` trực tiếp. `tier` (free/premium) là business logic riêng của Prep Chinese Vocab, sẽ handle sau.

### 4.3 `is_first_login` thay vì `is_new_user`

Đồng bộ tên field với Prep platform response (`data.is_first_login`). Tuy nhiên trong context Go app, `is_first_login` nghĩa là: **user chưa có local profile** (chưa từng login vào Prep Chinese Vocab), không phải chưa từng login Prep platform.

---

## 5. Prep Platform Config

| Config | Giá trị | Ghi chú |
|---|---|---|
| API Gateway | `https://api-gw.testsprep.online` | Test environment |
| SSO Cookie | `prep_sso_token` | Cookie domain: `.prep.vn` |
| Token TTL | 90 ngày (129,600 phút) | Prep access_token |
| Rate limit | 180 req/phút | Trên Prep endpoints |
| Auth framework | Laravel Passport | OAuth2 Personal Access Token |
