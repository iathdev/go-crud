# Go Clean Architecture CRUD Example

Dự án mẫu triển khai **REST API** với **Go (Golang)** theo kiến trúc **Hexagonal Architecture (Ports and Adapters)**.

## Tài liệu chi tiết

Vui lòng xem chi tiết tại thư mục `doc/`:

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
    - `POST /api/products`: Tạo sản phẩm (Cần Token)
    - `GET /api/products`: Lấy danh sách sản phẩm
