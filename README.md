# Go Modular Monolith — Chinese Vocabulary Learning API

Dự án **REST API** với **Go (Golang)** theo kiến trúc **Modular Monolith** + **Hexagonal Architecture** cho từng module.

## Tài liệu chi tiết

Vui lòng xem chi tiết tại thư mục `docs/`:

- [Kiến trúc Dự án và Vòng đời Request](docs/architecture.md)
- [Danh sách Công nghệ và Thư viện](docs/tech_stack.md)

## Cài đặt và Chạy

### Yêu cầu
- Docker & Docker Compose
- Go 1.23+

### Chạy ứng dụng

1.  **Khởi động Database & App**:
    ```bash
    make docker-up
    ```

2.  **Chạy ứng dụng cục bộ (Dev mode)**:
    ```bash
    make run
    ```

3.  **API Endpoints**:
    - `POST /register`: Đăng ký tài khoản
    - `POST /login`: Đăng nhập lấy Token
    - `POST /refresh`: Refresh Token
    - `POST /api/logout`: Đăng xuất (Cần Token)
    - `POST /api/vocabularies`: Tạo từ vựng (Cần Token)
    - `GET /api/vocabularies/:id`: Lấy từ vựng theo ID
    - `GET /api/vocabularies/hsk/:level`: Lấy từ vựng theo HSK level
    - `GET /api/vocabularies/search?q=...`: Tìm kiếm từ vựng
    - `PUT /api/vocabularies/:id`: Cập nhật từ vựng
    - `DELETE /api/vocabularies/:id`: Xóa từ vựng
    - `POST /api/folders`: Tạo folder
    - `GET /api/folders`: Danh sách folder
    - `POST /api/learning/start`: Bắt đầu học từ vựng
    - `GET /api/learning/review`: Lấy danh sách cần ôn tập
    - `GET /health`: Health check
