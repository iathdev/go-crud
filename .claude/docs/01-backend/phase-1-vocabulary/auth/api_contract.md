# Auth Module — API Contract

> Tài liệu API cho mobile team. Prep Account Service APIs (login, logout, token): mobile xem doc của Account Service.

---

## `GET /api/v1/auth/me`

Lấy profile local của user. Đây là endpoint duy nhất của auth module trên Go backend.

**Request:**

```
GET /api/v1/auth/me
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
    "force_update_password": false,
    "created_at": "2026-03-19T10:00:00Z",
    "updated_at": "2026-03-19T10:00:00Z"
  }
}
```

**Fields:**

- `id`: UUID local của Go app
- `prep_user_id`: ID trên Prep platform
- `is_first_login`: `true` nếu user chưa từng login Go app (chưa có profile local) → mobile hiện onboarding
- `force_update_password`: `true` nếu Prep yêu cầu user đổi mật khẩu → mobile hiện flow đổi password

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
