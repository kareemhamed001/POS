# POS (Point of Sale) API

A production-ready RESTful API for a Point of Sale system built with Go, featuring clean architecture, comprehensive testing, and Docker deployment.

## 🎯 Features

- **JWT Authentication**: Secure token-based authentication with registration and login endpoints
- **Role-Based Authorization**: Admin and customer roles with role-specific middleware for protected endpoints
- **User Management**: Create, read, update, and delete users with role-based access (customer/admin)
- **Product Catalog**: Manage products with flexible pricing and discount strategies
- **Order Processing**: Complete order lifecycle with item management and calculations
- **Address Management**: Multi-address support for users with primary address designation
- **Advanced Search**: Case-insensitive search across all resources with pagination
- **Discount System**: Support for both fixed and percentage-based discounts at item and order levels
- **Password Security**: Bcrypt hashing for secure password storage
- **Global Logging**: Production-grade logging with Zap logger accessible throughout the application
- **Hot Reload**: Air integration for rapid development without rebuilding
- **Comprehensive Testing**: Unit tests with SQLite in-memory database isolation
- **Docker Ready**: Complete Docker and docker-compose setup for easy deployment

## 🛠️ Tech Stack

- **Language**: Go 1.25.3
- **Framework**: Gin Web Framework
- **Database**: PostgreSQL (production) / SQLite (testing)
- **ORM**: GORM with generics
- **Authentication**: JWT (golang-jwt/jwt/v5)
- **Password Hashing**: bcrypt
- **Validation**: go-playground/validator with custom validators
- **Logging**: Uber Zap Logger
- **Hot Reload**: Air
- **Containerization**: Docker & Docker Compose
- **Testing**: Testify (assert/require)

## 📋 Prerequisites

- Go 1.25.3 or higher
- Docker & Docker Compose
- PostgreSQL 16 (if running without Docker)
- Git

## 🚀 Quick Start

### Using Docker (Recommended)

1. **Clone the repository**

```bash
git clone <repository-url>
cd POS
```

2. **Start the application**

```bash
docker-compose up app -d --build
```

The API will be available at `http://localhost:8081`

3. **Stop the application**

```bash
docker-compose down
```

### Local Development

1. **Install dependencies**

```bash
go mod download
go mod vendor
```

2. **Set up environment variables**

```bash
cp .env.example .env
```

3. **Start PostgreSQL** (if not using Docker)

```bash
# Make sure PostgreSQL is running and accessible
```

4. **Run the application**

```bash
go run ./cmd/api
```

Or with Air for hot reload:

```bash
air -c .air.toml
```

## 📚 API Documentation

### Base URL

```
http://localhost:8081/api
```

### Health Check

```
GET /health
```

### Authentication Endpoints

**Note**: Authentication is required for all endpoints except `/health`, `/auth/register`, and `/auth/login`.

| Method | Endpoint         | Description                 |
| ------ | ---------------- | --------------------------- |
| POST   | `/auth/register` | Register a new user account |
| POST   | `/auth/login`    | Login and get JWT token     |

**Register Example:**

```bash
curl -X POST http://localhost:8081/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "phone": "+201234567890",
    "password": "SecurePass123!"
  }'
```

**Response:**

```json
{
  "success": true,
  "message": "User registered successfully",
  "data": {
    "user": {
      "id": 1,
      "name": "John Doe",
      "email": "john@example.com",
      "phone": "+201234567890",
      "role": "customer"
    },
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

**Login Example:**

```bash
curl -X POST http://localhost:8081/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "SecurePass123!"
  }'
```

**Response:**

```json
{
  "success": true,
  "message": "User logged in successfully",
  "data": {
    "user": {
      "id": 1,
      "name": "John Doe",
      "email": "john@example.com",
      "phone": "+201234567890",
      "role": "customer"
    },
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

**Using the JWT Token:**

```bash
# Add the token to the Authorization header for protected endpoints
curl -X GET http://localhost:8081/api/users \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Users Endpoints

**Note**: All user endpoints require authentication. Creating and deleting users requires `admin` role.

| Method | Endpoint     | Description                               | Required Role          |
| ------ | ------------ | ----------------------------------------- | ---------------------- |
| GET    | `/users`     | List all users with pagination and search | Any authenticated user |
| POST   | `/users`     | Create a new user                         | Admin only             |
| GET    | `/users/:id` | Get user by ID                            | Any authenticated user |
| PUT    | `/users/:id` | Update user information                   | Any authenticated user |
| DELETE | `/users/:id` | Delete a user                             | Admin only             |

**Create User Example:**

```bash
curl -X POST http://localhost:8081/api/users \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "phone": "+201234567890",
    "password": "SecurePass123!",
    "role": "customer"
  }'
```

### Products Endpoints

**Note**: All product endpoints require authentication. Creating products requires `admin` role.

| Method | Endpoint        | Description          | Required Role          |
| ------ | --------------- | -------------------- | ---------------------- |
| GET    | `/products`     | List all products    | Any authenticated user |
| POST   | `/products`     | Create a new product | Admin only             |
| GET    | `/products/:id` | Get product by ID    | Any authenticated user |

**Create Product Example:**

```bash
curl -X POST http://localhost:8081/api/products \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "name": "Wireless Headphones",
    "short_description": "High-quality wireless headphones",
    "description": "Premium wireless headphones with noise cancellation",
    "price": 199.99,
    "discount_type": "percent",
    "discount_value": 15,
    "image_url": "https://example.com/headphones.jpg",
    "quantity": 50
  }'
```

### Addresses Endpoints

**Note**: All address endpoints require authentication.

| Method | Endpoint                   | Description                  | Required Role          |
| ------ | -------------------------- | ---------------------------- | ---------------------- |
| POST   | `/addresses`               | Add a new address for a user | Any authenticated user |
| GET    | `/addresses/user/:user_id` | Get all addresses for a user | Any authenticated user |

### Orders Endpoints

**Note**: All order endpoints require authentication. Updating order status requires `admin` role.

| Method | Endpoint             | Description                   | Required Role          |
| ------ | -------------------- | ----------------------------- | ---------------------- |
| POST   | `/orders`            | Create a new order with items | Any authenticated user |
| GET    | `/orders/:id`        | Get order with details        | Any authenticated user |
| PATCH  | `/orders/:id/status` | Update order status           | Admin only             |

**Create Order Example:**

```bash
curl -X POST http://localhost:8081/api/orders \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "user_id": 1,
    "address_id": 1,
    "shipping_cost": 10.00,
    "discount_type": "percent",
    "discount_value": 5,
    "items": [
      {
        "product_id": 1,
        "quantity": 2,
        "price": 199.99,
        "discount_type": "percent",
        "discount_value": 15
      }
    ]
  }'
```

### Order Status Update Example:\*\*

```bash
curl -X PATCH http://localhost:8081/api/orders/1/status \
  -H "Content-Type: application/json" \  -H "Authorization: Bearer YOUR_JWT_TOKEN" \  -d '{"status": "completed"}'
```

Valid statuses: `pending`, `completed`, `cancelled`

## 📁 Project Structure

```
POS/
├── cmd/
│   └── api/
│       └── main.go                 # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go              # Configuration management
│   ├── db/
│   │   └── db.go                  # Database initialization
│   ├── delivery/
│   │   └── http/
│   │       ├── handler/           # HTTP request handlers
│   │       ├── middleware/        # HTTP middlewares
│   │       ├── request/           # Request DTOs
│   │       ├── response/          # Response helpers
│   │       ├── routes/            # Route definitions
│   │       └── validator/         # Custom validators
│   ├── entity/                    # Domain entities
│   ├── repository/
│   │   └── postgres/              # Data access layer
│   └── usecase/                   # Business logic layer
├── pkg/
│   └── logger/                    # Global logger package
├── docker-compose.yml             # Docker Compose configuration
├── Dockerfile                     # Production Dockerfile
├── Dockerfile.dev                 # Development Dockerfile with Air
├── .air.toml                      # Air configuration
└── README.md
```

## 🧪 Testing

### Run All Tests

```bash
go test ./...
```

### Run Specific Tests

```bash
go test -v ./internal/repository/postgres -run TestUserRepository
```

### Run Tests with Coverage

```bash
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Run Tests with Coverage Report

```bash
go test -v -coverprofile=coverage.out ./internal/repository/postgres
```

## 📊 Architecture

### Clean Architecture Pattern

The project follows Clean Architecture principles:

```
┌─────────────────────────────────────────────────────┐
│              API Layer (Handlers)                   │
├─────────────────────────────────────────────────────┤
│              Use Cases (Business Logic)             │
├─────────────────────────────────────────────────────┤
│              Repository (Data Access)              │
├─────────────────────────────────────────────────────┤
│              Database (PostgreSQL)                  │
└─────────────────────────────────────────────────────┘
```

### Data Flow

1. **HTTP Request** → Handler receives request
2. **Validation** → Request is validated with custom rules
3. **Use Case** → Business logic processes the request
4. **Repository** → Data is persisted/retrieved from database
5. **Response** → Result is formatted and sent back to client

## 🌍 Environment Variables

Create a `.env` file in the project root:

```env
APP_ENV=development
APP_PORT=8081

DB_DRIVER=postgres
DB_HOST=postgres
DB_PORT=5432
DB_USER=admin
DB_PASSWORD=admin
DB_NAME=pos_db

JWT_PRIVATE_KEY=your-super-secret-jwt-key-change-this-in-production
JWT_TOKEN_DURATION=24
```

**Environment Variables Explained:**

- `APP_ENV`: Environment (development/production)
- `APP_PORT`: Server port
- `DB_*`: Database connection settings
- `JWT_PRIVATE_KEY`: Secret key for signing JWT tokens (change in production!)
- `JWT_TOKEN_DURATION`: Token expiration time in hours (default: 24)

## 🔧 Configuration

### Database Configuration

Configure database connection in `internal/config/config.go`:

```go
type Config struct {
    AppEnv       string
    AppPort      int
    DBDriver     string
    DBHost       string
    DBPort       int
    DBUser       string
    DBPassword   string
    DBName       string
    JWTPrivateKey string
}
```

## 📝 Available Postman Collection

A complete Postman collection is included: `POS_API.postman_collection.json`

**To import:**

1. Open Postman
2. Click "Import" → "Upload Files"
3. Select `POS_API.postman_collection.json`
4. Update `base_url` variable to your API endpoint

## 🔐 Validation Rules

### User Creation

- **Name**: Required, 2-50 characters
- **Email**: Required, valid email format, unique
- **Phone**: Required, valid Egyptian phone format, unique
- **Password**: Required, minimum length enforced

### Product Creation

- **Name**: Required
- **Price**: Required, positive number
- **Description**: Required
- **Quantity**: Required, positive integer

### Address Creation

- **User ID**: Required, valid user must exist
- **Name**: Required
- **Country**: Required
- **City**: Required
- **Address**: Required

## 🚦 Running Tests in Docker

```bash
# Run tests inside Docker
docker-compose exec app go test ./...

# Run tests with coverage
docker-compose exec app go test -cover ./...
```

## 📦 Dependencies

Key dependencies:

- `gin-gonic/gin` - Web framework
- `gorm.io/gorm` - ORM
- `jackc/pgx` - PostgreSQL driver
- `go-playground/validator` - Validation
- `uber-go/zap` - Logging
- `air-verse/air` - Hot reload

Run `go mod tidy` to manage dependencies.

## 🤝 Contributing

Contributions are welcome! Please follow these steps:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 📈 Future Enhancements

- [x] Authentication & Authorization (JWT) ✅ **Completed**
- [x] Role-based access control (Admin/Customer) ✅ **Completed**
- [ ] Redis caching layer
- [ ] Advanced reporting and analytics
- [ ] Inventory management
- [ ] Payment processing integration
- [ ] Real-time notifications
- [ ] GraphQL API support
- [ ] API rate limiting
- [ ] Refresh token mechanism
- [ ] Email verification for new users

## 👤 Author

**Kareem Hamed**

- LinkedIn: [https://www.linkedin.com/in/kareemhamed001/]
- GitHub: [https://github.com/kareemhamed001]
- Email: [kareemhamedibrahim@gmail.com]

## 🙏 Acknowledgments

- Gin Web Framework community
- GORM developers
- Go community for excellent libraries and tools

---

**Last Updated**: January 16, 2026
**Go Version**: 1.25.3
