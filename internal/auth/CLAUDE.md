# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in the auth module.

## Module Overview

SSO authentication module. Go backend **does not manage login directly** — mobile app logs in via Prep platform, receives a Prep access_token, then sends it to Go backend to create a local session with JWT.

## Key Domain Concepts

- **User**: Local user entity mapped from Prep platform user. Identified by `prep_user_id` (int64, from Prep) and local `id` (UUID).
- **PrepUser**: Value object representing user data returned from Prep's `/me` endpoint. Used only during login flow.
- **Two token systems**: Prep access_token (90-day, used once to validate) → Go JWT access_token (15min) + refresh_token (7 days).

## Module Structure

```
internal/auth/
├── domain/
│   ├── user.go              # User entity
│   └── prep_user.go         # PrepUser value object
├── application/
│   ├── port/
│   │   ├── inbound.go       # AuthUseCasePort
│   │   └── outbound.go      # PrepUserServicePort, UserRepositoryPort, TokenServicePort, RefreshTokenStorePort
│   ├── dto/
│   │   └── dto.go           # LoginRequest, LoginResponse, TokenResponse, etc.
│   └── usecase/
│       └── auth_usecase.go  # SSO login, refresh, logout, get profile
├── adapter/
│   ├── handler/
│   │   └── handler.go       # HTTP handlers (Gin)
│   ├── repository/
│   │   ├── model/
│   │   │   └── user_model.go
│   │   ├── user_repository.go           # Postgres user repository
│   │   └── refresh_token_repository.go  # Redis refresh token repository
│   └── service/
│       ├── jwt_service.go          # JWT token generation (implements TokenServicePort)
│       └── prep_user_service.go    # Prep API HTTP client with circuit breaker (implements PrepUserServicePort)
└── module.go                # Module wiring + RegisterRoutes(public, protected)
```

## Routes

```
POST /login    — Public, rate-limited. Send Prep token → get JWT pair
POST /refresh  — Public, rate-limited. Refresh expired access_token
POST /logout   — Protected. Invalidate refresh token
GET  /profile  — Protected. Get current user profile
```

## Login Flow

1. Mobile app logs in with Prep platform (Google/Phone/Password) → gets Prep access_token
2. Mobile sends Prep token to `POST /login`
3. Go backend calls Prep `GET /me` (via `PrepUserService` with circuit breaker) to validate token
4. Upsert user in Postgres (create or update name/email)
5. Generate local JWT access_token + refresh_token (stored in Redis as SHA-256 hash)
6. Return JWT pair + profile + `is_first_login` to mobile

## Module-Specific Patterns

- **Circuit breaker**: `PrepUserService` wraps Prep API calls with gobreaker. When Prep is down, returns `ErrServiceUnavailable` (HTTP 503) immediately instead of waiting for timeout.
- **Refresh token storage**: Tokens stored in Redis as SHA-256 hashes. Each user has a set of active token hashes (`user_tokens:{userID}`) for bulk revocation on logout.
- **Upsert on login**: Every login upserts the user record — creates if new, updates name/email if existing. Uses Postgres `ON CONFLICT` on `prep_user_id`.

## Database Tables

- `users` — unique index on `prep_user_id`, soft delete via `deleted_at`
