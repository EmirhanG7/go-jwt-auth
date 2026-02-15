
# Go Fiber Production-Ready Auth Boilerplate 🚀

A robust, secure, and scalable authentication starter kit built with **Go Fiber**, **GORM**, and **PostgreSQL**. This boilerplate implements modern security standards like JWT rotation, reuse detection, and HttpOnly cookies.

## ✨ Features

- **Authentication**: JWT based Access & Refresh Token logic.
- **Security First**:
    - **JWT Rotation**: Refresh tokens are rotated on every use.
    - **Reuse Detection**: Automatic session invalidation on token reuse.
    - **HttpOnly Cookies**: Protection against XSS attacks.
    - **Rate Limiting**: Brute-force protection for auth endpoints.
    - **Helmet**: Essential HTTP security headers.
    - **Validation**: Input validation using `go-playground/validator`.
- **Developer Experience**:
    - **Hot Reload**: Automatic recompilation with `Air`.
    - **Dockerized**: One-command setup with `docker-compose`.
    - **API Documentation**: Interactive documentation with `Swagger UI`.
    - **Graceful Shutdown**: Safe connection closing on server stop.

## 🛠️ Tech Stack

- **Framework**: [Fiber](https://gofiber.io/)
- **Database**: [PostgreSQL](https://www.postgresql.org/)
- **ORM**: [GORM](https://gorm.io/)
- **Authentication**: [JWT-Go](https://github.com/golang-jwt/jwt)
- **Validation**: [go-playground/validator](https://github.com/go-playground/validator)
- **Containerization**: [Docker](https://www.docker.com/)
- **Hot Reload**: [Air](https://github.com/cosmtrek/air)
- **API Docs**: [Swaggo](https://github.com/swaggo/swag)

## 🚀 Getting Started

### Prerequisites

- Docker and Docker Compose installed.
- **Go 1.25.6** installed locally (Optional if using Docker).

### Installation

1. **Clone the repository:**
```bash
git clone https://github.com/EmirhanG7/go-jwt-auth
cd jwt-auth
```

2. **Create environment file:**
```bash
cp .env.example .env
```


Or manually create `.env` file with:
```env
PORT=3000
ENV=development

DB_HOST=db
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=yourpassword
DB_NAME=jwt_auth

JWT_SECRET=your-super-secret-key-change-this-in-production
JWT_REFRESH_SECRET=your-refresh-secret-key-change-this-in-production
ALLOWED_ORIGINS=http://localhost:3000
```


3. **Start the application with Docker:**
```bash
docker-compose up -d --build
```


4. **Access the API:**
* API: http://localhost:3000
* Swagger Docs: http://localhost:3000/swagger/index.html



### Local Development (Without Docker)

1. **Install dependencies:**
```bash
go mod download
```


2. **Install Air for hot reload:**
```bash
go install https://github.com/air-verse/air@latest
```


3. **Install Swag for API documentation:**
```bash
go install https://github.com/swaggo/swag/cmd/swag@latest
```


4. **Generate Swagger docs:**
```bash
swag init
```

5. **Run the application:**
```bash
air
```


## 📚 API Endpoints

### Public Endpoints

| Method | Endpoint | Description | Rate Limit |
| --- | --- | --- | --- |
| POST | `/api/auth/register` | Register new user | 5 req/min |
| POST | `/api/auth/login` | Login user | 5 req/min |
| POST | `/api/auth/refresh` | Refresh access token | 5 req/min |

### Protected Endpoints (Requires Authentication)

| Method | Endpoint | Description | Rate Limit |
| --- | --- | --- | --- |
| GET | `/api/auth/profile` | Get user profile | 60 req/min |
| POST | `/api/auth/logout` | Logout from current device | - |
| POST | `/api/auth/logout-all` | Logout from all devices | - |

## 🔒 Security Features

### JWT Token Strategy

* **Access Token**: 15 min lifetime, stored in HttpOnly cookie (`/`).
* **Refresh Token**: 7 days lifetime, stored in HttpOnly cookie (`/api/auth/refresh`).
* **Token Rotation**: New refresh token issued on every refresh request; old one is immediately invalidated.
* **Reuse Detection**: If a used refresh token is presented, all sessions are terminated for security.

### Rate Limiting

* **Auth Endpoints**: 5 requests / minute (per IP).
* **Protected API**: 60 requests / minute (per IP).

### Cookie Security

* **HttpOnly**: Prevents JavaScript access (XSS protection).
* **SameSite=Lax**: Provides CSRF protection.
* **Secure**: Only sent over HTTPS (in `ENV=production` mode).
## 🏗️ Project Structure

```text
jwt-auth/
├── config/              # Database & App configurations
├── controllers/         # Request handlers (Business logic)
├── docs/               # Swagger generated files
├── middleware/         # Auth, Limiter, and Security middlewares
├── models/            # Database models (GORM)
├── routes/            # Route definitions
├── utils/             # JWT, Validator, and Helpers
├── .air.toml          # Air configuration
├── docker-compose.yaml # Docker orchestration
├── Dockerfile.dev     # Development image (with Air)
├── main.go            # Application entry point
└── README.md          # Project documentation

```

## 🐳 Docker Commands

| Task | Command |
| --- | --- |
| Start Services | `docker-compose up -d` |
| View Logs | `docker-compose logs -f app` |
| Stop Services | `docker-compose down` |
| Rebuild & Start | `docker-compose up -d --build` |
| Reset Database | `docker-compose down -v` |

## 📖 Swagger Documentation

Interactive API documentation: `http://localhost:3000/swagger/index.html`

To regenerate docs:

```bash
swag init
```

## 🤝 Contributing

1. Fork the repository.
2. Create your feature branch (`git checkout -b feature/AmazingFeature`).
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`).
4. Push to the branch (`git push origin feature/AmazingFeature`).
5. Open a Pull Request.


## 📧 Contact

Emirhan Gözükucuk - [@EmirhanG7](https://github.com/EmirhanG7)
Project Link: [https://github.com/EmirhanG7/go-jwt-auth](https://github.com/EmirhanG7/go-jwt-auth)

⭐ If you find this project useful, please consider giving it a star!