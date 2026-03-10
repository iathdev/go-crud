# Danh sách Công nghệ và Thư viện (Tech Stack)

Tài liệu này liệt kê các công nghệ, công cụ và thư viện quan trọng được sử dụng trong dự án.

## 1. Công nghệ Cốt lõi (Core Technologies)

| Công nghệ | Phiên bản (Khuyến nghị) | Mô tả |
| :--- | :--- | :--- |
| **Go (Golang)** | 1.25+ | Ngôn ngữ lập trình chính, hiệu năng cao, concurrency tốt. |
| **PostgreSQL** | 15+ | Hệ quản trị cơ sở dữ liệu quan hệ (RDBMS) chính. |
| **Docker** | Latest | Containerization platform để đóng gói và chạy ứng dụng. |
| **Docker Compose** | Latest | Công cụ định nghĩa và chạy multi-container Docker applications (App + DB). |

## 2. Thư viện Go Quan trọng (Key Go Packages)

Các thư viện chính được liệt kê trong `go.mod`:

### Web Framework & Networking
- **[Gin Gonic](https://github.com/gin-gonic/gin)** (`github.com/gin-gonic/gin`):
  - Web framework siêu nhanh (high-performance) cho Go.
  - Xử lý HTTP requests, routing, middleware (Auth, Logger).
  - Binding và validation dữ liệu JSON đầu vào.

### Database & ORM
- **[GORM](https://gorm.io/)** (`gorm.io/gorm`):
  - ORM (Object Relational Mapping) thư viện phổ biến nhất cho Go.
  - Giúp tương tác với Database bằng Go struct thay vì viết SQL thuần.
  - Hỗ trợ Hooks, Associations, Transactions.
- **[GORM Postgres Driver](https://github.com/go-gorm/postgres)** (`gorm.io/driver/postgres`):
  - Driver để GORM kết nối với PostgreSQL.

### Configuration
- **[Viper](https://github.com/spf13/viper)** (`github.com/spf13/viper`):
  - Thư viện quản lý cấu hình (Configuration Management) mạnh mẽ.
  - Hỗ trợ đọc config từ nhiều nguồn: file (`.env`, `.yaml`, `.json`), biến môi trường (Environment Variables), command line flags.
  - Trong dự án này dùng để load config từ `.env`.

### Authentication & Security
- **[JWT Go](https://github.com/golang-jwt/jwt)** (`github.com/golang-jwt/jwt/v5`):
  - Thư viện tạo và xác thực JSON Web Tokens (JWT).
  - Dùng cho tính năng Đăng nhập và bảo vệ API (Auth Middleware).
- **[Go Crypto](https://pkg.go.dev/golang.org/x/crypto)** (`golang.org/x/crypto`):
  - Cung cấp các thuật toán mã hóa bổ sung.
  - Cụ thể dùng `bcrypt` để hash password người dùng trước khi lưu vào DB.

### Utilities
- **[Google UUID](https://github.com/google/uuid)** (`github.com/google/uuid`):
  - Tạo và xử lý UUID (Universally Unique Identifier).
  - Dùng làm Primary Key cho các bảng trong Database thay vì ID tự tăng (Integer).

## 3. Kiến trúc (Architecture)
- **Hexagonal Architecture (Ports and Adapters)**: Giúp tách biệt Business Logic khỏi Framework và Infrastructure.
- **Dependency Injection (DI)**: Manual DI (constructor injection) để quản lý sự phụ thuộc giữa các components (Handler -> UseCase -> Repository).

## 4. Công cụ Phát triển (Development Tools)
- **Makefile**: Tự động hóa các tác vụ thường gặp (run, build, docker-up).
- **Postman / cURL**: Dùng để test API.
