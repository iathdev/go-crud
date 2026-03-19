# Auth Module — API Contract (Flow B: Prep Token Direct)

> Tài liệu cho mobile team. Go backend chỉ có 1 endpoint `GET /me`. Tất cả auth (login, logout) gọi trực tiếp Prep platform.

---

## 1. Tổng quan

Mobile app quản lý **1 token duy nhất**: Prep access_token. Token này dùng cho cả Prep platform lẫn Go backend.

- **Login/Logout**: Mobile gọi trực tiếp Prep platform
- **Mọi request tới Go backend**: Gửi Prep token trong header `Authorization: Bearer <prep_token>`
- **Go backend**: Mỗi request gọi Prep `/me` để validate token + lấy user info

---

## 2. Luồng Login

```
┌─────────────┐     ┌──────────────────┐     ┌─────────────────┐
│  Mobile App  │     │  Prep Platform    │     │  Go Backend     │
└──────┬───────┘     └────────┬─────────┘     └────────┬────────┘
       │                      │                        │
       │  1. Login             │                        │
       │  (Google/Phone/Pass)  │                        │
       │─────────────────────>│                        │
       │                      │                        │
       │  2. Prep access_token │                        │
       │<─────────────────────│                        │
       │                      │                        │
       │  3. Lưu token local   │                        │
       │  (SecureStorage)      │                        │
       │                      │                        │
       │  4. GET /api/me       │                        │
       │  Bearer: prep_token   │                        │
       │──────────────────────────────────────────────>│
       │                      │                        │
       │                      │  5. GET /me (Prep)      │
       │                      │  Bearer: prep_token     │
       │                      │<───────────────────────│
       │                      │                        │
       │                      │  6. {id, name, email}   │
       │                      │───────────────────────>│
       │                      │                        │
       │                      │                        │ 7. Upsert user (DB)
       │                      │                        │
       │  8. Response:         │                        │
       │  {profile,            │                        │
       │   is_first_login}     │                        │
       │<──────────────────────────────────────────────│
       │                      │                        │
       │  [Nếu is_first_login] │                        │
       │  9. Hiện onboarding   │                        │
       │  PUT /api/me/         │                        │
       │      onboarding       │                        │
       │──────────────────────────────────────────────>│
```

### Giải thích:

1. Mobile login trực tiếp Prep platform (Google/Phone/Password)
2. Prep trả access_token (TTL 90 ngày)
3. Mobile lưu token vào SecureStorage
4. Mobile gọi Go backend `GET /api/me` để lấy profile local
5. Go backend validate token bằng cách gọi Prep `/me`
6. Prep xác nhận OK → trả user data
7. Go backend upsert user vào DB local
8. Trả profile local + `is_first_login` cho mobile
9. Nếu first login → mobile hiện onboarding

---

## 3. Luồng Logout

```
┌─────────────┐     ┌──────────────────┐     ┌─────────────────┐
│  Mobile App  │     │  Prep Platform    │     │  Go Backend     │
└──────┬───────┘     └────────┬─────────┘     └────────┬────────┘
       │                      │                        │
       │  1. POST /logout      │                        │
       │  Bearer: prep_token   │                        │
       │─────────────────────>│                        │
       │                      │                        │
       │  2. OK                │                        │
       │<─────────────────────│                        │
       │                      │                        │
       │  3. Xoá token local   │                        │
       │  (SecureStorage)      │                        │
```

### Giải thích:

1. Mobile gọi Prep logout trực tiếp → Prep revoke token
2. Prep trả OK
3. Mobile xoá token khỏi SecureStorage

Go backend **không liên quan**. Không cần gọi Go khi logout.

---

## 4. Luồng Request (protected)

```
┌─────────────┐     ┌──────────────────┐     ┌─────────────────┐
│  Mobile App  │     │  Prep Platform    │     │  Go Backend     │
└──────┬───────┘     └────────┬─────────┘     └────────┬────────┘
       │                      │                        │
       │  1. GET /api/vocabs   │                        │
       │  Bearer: prep_token   │                        │
       │──────────────────────────────────────────────>│
       │                      │                        │
       │                      │  2. GET /me (Prep)      │
       │                      │  Bearer: prep_token     │
       │                      │<───────────────────────│
       │                      │                        │
       │                      │  3. OK → user info      │
       │                      │───────────────────────>│
       │                      │                        │
       │                      │                        │ 4. Set user_id
       │                      │                        │    in context
       │                      │                        │
       │                      │                        │ 5. Execute handler
       │                      │                        │    (query vocabs)
       │                      │                        │
       │  6. Response: vocabs  │                        │
       │<──────────────────────────────────────────────│
```

### Giải thích:

1. Mobile gửi bất kỳ request nào tới Go backend với Prep token
2. AuthMiddleware gọi Prep `/me` để validate token
3. Prep OK → Go có user info
4. Gắn `user_id` vào context
5. Handler thực thi business logic
6. Trả response cho mobile

### Khi token hết hạn hoặc bị revoke:

- Bước 2-3: Prep trả 401 → Go trả 401 cho mobile → mobile redirect về màn login

---

## 5. API Endpoints

### 5.1 Prep Platform (mobile gọi trực tiếp)

| Method | Endpoint | Auth | Mô tả |
|---|---|---|---|
| `POST` | `/auth/api/v1/auth/account-login` | Không | Login bằng username/password |
| `POST` | `/auth/api/v1/auth/phone-login` | Không | Login bằng phone (password hoặc OTP) |
| `POST` | `/auth/api/v1/auth/google-login` | Không | Login bằng Google OAuth |
| `POST` | `/auth/api/v1/auth/logout` | Bearer token | Logout, revoke token |

> Base URL: `https://api-gw.testsprep.online` (test) / `https://api-gw.prep.vn` (prod)

### 5.2 Go Backend

#### `GET /api/me`

Lấy profile local của user. Đây là endpoint duy nhất của auth module trên Go backend.

**Request:**

```
GET /api/me
Authorization: Bearer <prep_token>
X-Lang: vi  (optional, default: en)
```

**Response 200:**

```json
{
  "success": true,
  "message": "OK",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "prep_user_id": 446416,
    "name": "thaidong",
    "email": "",
    "is_first_login": false,
    "created_at": "2026-03-19T10:00:00Z",
    "updated_at": "2026-03-19T10:00:00Z"
  }
}
```

- `id`: UUID local của Go app
- `prep_user_id`: ID trên Prep platform
- `is_first_login`: `true` nếu user chưa từng login Go app (chưa có profile local) → mobile hiện onboarding

**Response 401:**

```json
{
  "success": false,
  "message": "Unauthorized"
}
```

Prep token hết hạn hoặc bị revoke → mobile redirect về màn login.

**Response 503:**

```json
{
  "success": false,
  "message": "Service unavailable"
}
```

Circuit breaker trip (Prep API đang down) → mobile hiện thông báo thử lại sau.

---

## 6. Authentication cho tất cả Go API

Mọi request tới Go backend (vocabulary, folder, learning...) đều cần header:

```
Authorization: Bearer <prep_token>
```

Go backend validate token qua Prep `/me` ở middleware. Nếu fail → 401.

### Error responses chung:

| Status | Khi nào | Mobile xử lý |
|---|---|---|
| `401` | Token hết hạn / bị revoke / sai | Redirect về màn login Prep |
| `503` | Prep API down (circuit breaker) | Hiện thông báo "Thử lại sau" |

---

## 7. Lưu ý cho Mobile

1. **Chỉ quản lý 1 token**: Prep access_token. Lưu trong SecureStorage (Keychain/Keystore).
2. **Không cần refresh logic**: Prep token TTL 90 ngày. Hết hạn → login lại Prep.
3. **401 = login lại**: Bất kỳ API nào trả 401 → xoá token local → redirect màn login.
4. **503 = retry**: Go backend đang không kết nối được Prep → hiện UI retry, không xoá token.
5. **Gọi `GET /api/me` sau khi login Prep**: Để lấy profile local + kiểm tra `is_first_login` cho onboarding.
