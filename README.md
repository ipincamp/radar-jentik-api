# Radar Jentik API

A RESTful API service for the Radar Jentik (Mosquito Larvae Radar) application, built with Go using Hexagonal Architecture (Ports & Adapters) pattern.

## 📋 Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Tech Stack](#tech-stack)
- [Architecture](#architecture)
- [Project Structure](#project-structure)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Configuration](#configuration)
- [Running the Application](#running-the-application)
- [API Endpoints](#api-endpoints)
- [Database Migrations](#database-migrations)
- [Security](#security)
- [Testing](#testing)
- [Changelog](#changelog)
- [License](#license)

## 🎯 Overview

Radar Jentik API is a backend service designed to support the Radar Jentik application for monitoring and managing mosquito larvae detection. The API implements authentication and authorization features with secure token-based authentication using PASETO (Platform-Agnostic Security Tokens).

## ✨ Features

- **User Authentication & Authorization**
  - User Registration with role assignment (kader/petugas)
  - User Login with secure password verification
  - Stateless Logout
  - Role-based access control (RBAC)
  - User management for administrators
  
- **Report Management**
  - Create mosquito larvae reports with geolocation
  - View paginated reports with role-based filtering
  - Report validation workflow (pending → valid/rejected)
  - Status tracking with verifier information
  - Geographic data storage using PostGIS
  
- **Spatial Analysis**
  - Area boundary management with MultiPolygon support
  - GeoJSON import/export functionality
  - Heatmap generation using Inverse Distance Weighting (IDW) algorithm
  - Customizable risk estimation grid
  
- **Security**
  - Password hashing with Argon2id
  - Token generation using PASETO v2
  - Role-based data scoping (kader vs petugas)
  - Protected endpoints with middleware authentication
  
- **Infrastructure**
  - PostgreSQL 16 with PostGIS extension
  - GORM ORM with spatial data support
  - Docker containerization
  - Database migrations system
  - CLI seeder tool for area data
  - Environment-based configuration

## 🛠 Tech Stack

- **Language:** Go 1.25.5
- **Web Framework:** Fiber v2.52.10
- **Database:** PostgreSQL 16 with PostGIS 3.5
- **ORM:** GORM v1.31.1 with spatial support
- **Token:** PASETO (o1egl/paseto v1.0.0)
- **Password Hashing:** Argon2id (alexedwards/argon2id v1.0.0)
- **Validation:** go-playground/validator v10.29.0
- **Migration:** gormigrate v2.1.5
- **Configuration:** godotenv v1.5.1
- **Spatial:** PostGIS for geographic data (ST_MakePoint, ST_GeomFromGeoJSON, etc.)
- **Containerization:** Docker & Docker Compose

## 🏗 Architecture

This project follows **Hexagonal Architecture** (also known as Ports & Adapters pattern), which provides:

- **Clear separation of concerns**
- **Independent of frameworks**
- **Testable business logic**
- **Flexible adapters**

### Architecture Layers

```
┌─────────────────────────────────────────────────────┐
│                  Driving Adapters                   │
│              (HTTP Handlers / REST API)             │
└─────────────────────┬───────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────┐
│                   Application Core                  │
│                                                      │
│  ┌──────────────────────────────────────────────┐  │
│  │            Ports (Interfaces)                │  │
│  └──────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────┐  │
│  │         Services (Business Logic)            │  │
│  └──────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────┐  │
│  │            Domain (Entities)                 │  │
│  └──────────────────────────────────────────────┘  │
└─────────────────────┬───────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────┐
│                  Driven Adapters                    │
│           (Database / External Services)            │
└─────────────────────────────────────────────────────┘
```

## 📁 Project Structure

```
radar_jentik_api/
├── cmd/                          # Application entry points
│   ├── api/                      # Main API application
│   │   └── main.go               # Application bootstrap
│   ├── migrate/                  # Migration CLI tool
│   │   └── main.go               # Migration runner
│   └── seeder/                   # Data seeder CLI tool
│       └── main.go               # GeoJSON seeder for areas
│
├── internal/                     # Private application code
│   ├── adapters/                 # Adapter layer
│   │   ├── driven/               # Infrastructure adapters
│   │   │   └── postgres/         # PostgreSQL adapter
│   │   │       ├── db.go         # Database connection
│   │   │       ├── migrations/   # Migration files
│   │   │       │   ├── 20251214000001_create_users_table.go
│   │   │       │   ├── 20251214000002_create_reports_table.go
│   │   │       │   ├── 20251214000003_create_areas_table.go
│   │   │       │   └── registry.go
│   │   │       └── repositories/ # Data access implementations
│   │   │           ├── user_repo.go
│   │   │           ├── report_repo.go
│   │   │           └── area_repo.go
│   │   └── driving/              # UI/API adapters
│   │       └── http/             # HTTP/REST adapter
│   │           ├── fiber.go      # Fiber server setup
│   │           ├── handlers/     # HTTP request handlers
│   │           │   ├── auth_handler.go
│   │           │   ├── report_handler.go
│   │           │   └── area_handler.go
│   │           └── middleware/   # HTTP middleware
│   │               └── auth.go   # PASETO token validation
│   │
│   └── core/                     # Business logic core
│       ├── domain/               # Domain entities
│       │   ├── user.go           # User entity
│       │   ├── report.go         # Report entity
│       │   └── area.go           # Area entity
│       ├── ports/                # Interface definitions
│       │   ├── auth.go           # Auth service interface
│       │   ├── report.go         # Report service interface
│       │   └── area.go           # Area service interface
│       └── services/             # Business logic implementations
│           ├── auth_service.go   # Auth service implementation
│           ├── report_service.go # Report & heatmap logic
│           └── area_service.go   # Area service implementation
│
├── pkg/                          # Public reusable packages
│   ├── auth/                     # Authentication utilities
│   │   └── token.go              # PASETO token manager
│   └── config/                   # Configuration management
│       └── godotenv.go           # Environment config loader
│
├── docker-compose.yml            # Docker services configuration
├── go.mod                        # Go module definition
├── go.sum                        # Go module checksums
├── .env                          # Environment variables (not in repo)
├── API.md                        # API documentation
├── DEPLOYMENT.md                 # Deployment guide
├── CONTRIBUTING.md               # Contribution guidelines
├── LICENSE                       # Project license
└── README.md                     # This file
```

### Directory Explanation

- **`cmd/`**: Contains the application entry points. Each subdirectory is a separate executable.
- **`internal/`**: Private application code that cannot be imported by other projects.
  - **`adapters/driven/`**: Implements infrastructure concerns (database, external APIs).
  - **`adapters/driving/`**: Implements user-facing concerns (HTTP handlers, CLI).
  - **`core/`**: Contains the business logic, independent of external frameworks.
- **`pkg/`**: Contains reusable packages that can be imported by external projects.

## ⚙️ Prerequisites

- **Go** 1.25.5 or higher
- **Docker** and **Docker Compose**
- **PostgreSQL** 16 (if running without Docker)
- **Git**

## 📦 Installation

### 1. Clone the Repository

```bash
git clone https://github.com/ipincamp/radar-jentik-api.git
cd radar-jentik-api
```

### 2. Install Go Dependencies

```bash (PostgreSQL with PostGIS)
go mod download
```

### 3. Setup Environment Variables

Create a `.env` file in the root directory:

```bash
cp .env.example .env
```

Edit the `.env` file with your configuration (see [Configuration](#configuration) section).

## 🔧 Configuration

Create a `.env` file with the following variables:

```env
# Application Settings
APP_PORT=:3000
APP_ENV=development

# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_NAME=radar_jentik
DB_USERNAME=your_db_user
DB_PASSWORD=your_db_password
DB_TIMEZONE=Asia/Jakarta
DB_SSL_MODE=disable

# PASETO Token Configuration
PASETO_SECRET_KEY=your-32-character-secret-key!!
PASETO_EXP_DURATION=24h
PASETO_AUDIENCE=radar-jentik-app
PASETO_ISSUER=radar-jentik-api
```

### Configuration Details

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `APP_PORT` | Application port | `:3000` | No |
| `APP_ENV` | Environment (development/production) | `development` | No |
| `DB_HOST` | Database host | - | **Yes** |
| `DB_PORT` | Database port | `5432` | No |
| `DB_NAME` | Database name | - | **Yes** |
| `DB_USERNAME` | Database username | - | **Yes** |
| `DB_PASSWORD` | Database password | - | **Yes** |
| `DB_TIMEZONE` | Database timezone | `Asia/Jakarta` | No |
| `DB_SSL_MODE` | SSL mode (disable/require) | `disable` | No |
| `PASETO_SECRET_KEY` | Secret key for token (32 chars) | - | **Yes** |
| `PASETO_EXP_DURATION` | Token expiration duration | `24h` | No |
| `PASETO_AUDIENCE` | Token audience claim | `radar-jentik-app` | No |
| `PASETO_ISSUER` | Token issuer claim | `radar-jentik-api` | No |

**Important:** 
- `PASETO_SECRET_KEY` **must be exactly 32 characters** for PASETO v2.
- Generate a secure random key: `openssl rand -base64 32`

## 🚀 Running the Application

### Option 1: Using Docker Compose (Recommended)

1. **Start the database:**

```bash
docker-compose up -d
```

This will start:
- PostgreSQL 16 with PostGIS 3.5 extension on `localhost:5432`

2. **Run database migrations:**

```bash
go run cmd/migrate/main.go
```

3. **Seed area data (optional):**

```bash
go run cmd/seeder/main.go /path/to/geojson/file.txt
```

4. **Start the API server:**

```bash
go run cmd/api/main.go
```

### Option 2: Running Locally

1. **Ensure PostgreSQL is running** on your system.

2. **Run migrations:**

```bash
go run cmd/migrate/main.go
```

3. **Start the server:**

```bash
go run cmd/api/main.go
```

### Using Air for Hot Reload (Development)

Install Air:
```bash
go install github.com/cosmtrek/air@latest
```

Run with hot reload:
```bash
air
For comprehensive API documentation, see [API.md](API.md).

### Quick Reference

#### Authentication (Public)
- `POST /api/v1/auth/register` - Register new user
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/logout` - User logout

#### User Management (Protected - Petugas Only)
- `GET /api/v1/auth/users` - List all users

#### Reports (Protected)
- `POST /api/v1/reports` - Create new report
- `GET /api/v1/reports` - Get paginated reports (role-based filtering)
- `PATCH /api/v1/reports/:id/validate` - Validate report
- `GET /api/v1/reports/heatmap` - Get IDW heatmap data

#### Areas (Protected)
- `GET /api/v1/areas` - Get all areas (GeoJSON format)

### Example: Create Report

**Endpoint:** `POST /api/v1/reports`

**Headers:**
```
Authorization: Bearer <token>
Content-Type: application/json
```

**Request Body:**
```json
{
  "location_name": "Jl. Merdeka No. 10",
  "latitude": -7.250445,
  "longitude": 112.768845,
  "description": "Found mosquito larvae in water container"
}
```

**Success Response (201):**
```json
{
  "id": "uuid",
  "user_id": "user-uuid",
  "location_name": "Jl. Merdeka No. 10",
  "latitude": -7.250445,
  "longitude": 112.768845,
  "description": "Found mosquito larvae in water container",
  "status": "pending",
  "created_at": "2025-12-15T09:00:00Z"
}
```

### Example: Get Heatmap

**Endpoint:** `GET /api/v1/reports/heatmap?res=50&p=2`

**Query Parameters:**
- `res` (optional): Grid resolution, default 50
- `p` (optional): IDW power parameter, default 2

**Success Response (200):**
```json
[
  {
    "latitude": -7.250445,
    "longitude": 112.768845,
    "risk_value": 0.85
  }
]*Description:** Since the API uses stateless authentication, logout is handled client-side by discarding the token.

**Success Response (200):**
```json
{
  "message": "Logout berhasil"
}
```

## 🗄 Database Migrations

### Running Migrations

Migrations are located in `internal/adapters/driven/postgres/migrations/`

**Run migrations:**
```bash
go run cmd/migrate/main.go
```

### Creating New Migrations

1. Copy the template:
```bash
cp internal/adapters/driven/postgres/migrations/_TEMPLATE.txt \
   internal/adapters/driven/postgres/migrations/20251214000002_your_migration_name.go
```

2. Edit the new migration file with your changes.

3. Register the migration in `registry.go`

4. Run migrations:
```bash
go run cmd/migrate/main.go
```

### Current Migrations
   - Roles: kader (default), petugas

2. **20251214000002_create_reports_table.go**
   - Creates `reports` table with PostGIS geometry
   - Fields: id, user_id, location_name, location (geometry), description, status, verified_by, verified_at, timestamps
   - Uses ST_MakePoint for location storage
   - Status values: pending, verified, rejected

3. **20251214000003_create_areas_table.go**
   - Creates `areas` table with MultiPolygon support
   - Fields: id, name, geometry (MultiPolygon), timestamps
   - Uses ST_GeomFromGeoJSON for data import
   - Supports 2D/3D GeoJSON with ST_Force2D

1. **20251214000001_create_users_table.go**
   - Creates `users` table
   - Fields: id, name, username, password, role, created_at, updated_at
   - Indexes: unique index on username

## 🔐 Security

### Password Hashing

- **Algorithm:** Argon2id
- **Implementation:** `github.com/alexedwards/argon2id`
- **Parameters:** Default (memory=64MB, iterations=1, parallelism=2)

### Token Management

- **Algorithm:** PASETO v2 (Platform-Agnostic Security Tokens)
- **Encryption:** Symmetric encryption with 32-byte secret key
- **Claims:**
  - `aud`: Audience (application identifier)
  - `iss`: Issuer (API identifier)
  - `jti`: JWT ID (user ID)
  - `sub`: Subject (user ID)
  - `iat`: Issued At
  - `exp`: Expiration
  - `nbf`: Not Before
- **Footer:** Contains user role for authorization

**Authentication:**
- ✅ User Registration
- ✅ User Login
- ✅ User Logout
- ✅ List Users (Petugas only)

**Reports:**
- ✅ Create Report
- ✅ Get Reports (with pagination)
- ✅ Validate Report (status transitions)
- ✅ Get Heatmap Data (IDW algorithm)

**Areas:**
- ✅ Get Areas (GeoJSON export)
- ✅ Seeder Tool (GeoJSON import)gorithm confusion attacks
- No need to specify algorithms (misuse-resistant)
- Encrypted payload (confidentiality)
- Built-in expiration handling

## 🧪 Testing

### Black Box Testing Status

All endpoints have been tested and passed:
- ✅ User Registration
- ✅ User Login
- ✅ User Logout

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...
```

### Manual Testing with cURL

**Register:**
```bash
curl -X 2.0] - 2025-12-15

#### Added - Spatial Analysis & RBAC
- Inverse Distance Weighting (IDW) algorithm for risk estimation heatmap
- Customizable power parameter and grid resolution for heatmap
- Role-based data scoping (Kader: own reports, Petugas: all reports)
- User management endpoint (Petugas-only access)
- Enhanced middleware with role extraction from PASETO token footer

#### Improvements
- Explicit data filtering bypass for global heatmap visualization
- Extended repository interfaces for flexible user filtering
- Comprehensive RBAC implementation across all endpoints

### [v1.1.0] - 2025-12-14 (Evening)

#### Added - Geographic Features
- Area boundary management with PostGIS MultiPolygon
- GeoJSON import/export functionality
- CLI seeder tool for area data initialization
- ST_GeomFromGeoJSON and ST_AsGeoJSON integration
- Support for 2D/3D GeoJSON with ST_Force2D
- Protected endpoint for area data retrieval

### [v1.0.1] - 2025-12-14 (Afternoon/Evening)

#### Added - Report Management
- Report creation with PostGIS geometry (ST_MakePoint)
- Paginated report retrieval (GET /reports)
- Report validation workflow (pending → valid/rejected)
- Status tracking with verifier information and timestamps
- Protected middleware for token validation
- User ID injection into request context
- FindByID and Update methods in repository
- Comprehensive report management business logic

### [v1.0.0] - 2025-12-14 (Initial Release)

#### Added
- Initial project setup with hexagonal architecture
- Docker with PostGIS support (postgis/postgis:16-3.5-alpine)
- Environment-based configuration with godotenv
- Fiber web framework integration
- GORM ORM with PostgreSQL and PostGIS support
- Database migration system
- User authentication system
  - User registration with role assignment
  - User login with token generation
  - Stateless logout
- PASETO v2 token generation and management
- Argon2id password hashing
- Input validation with go-playground/validator
- Black box testing for all authentication endpoints

#### Security
- Implemented Argon2id for secure password hashing
- Implemented PASETO v2 for secure token management
- Role-based access control (kader/petugas)

#### Infrastructure
- PostgreSQL 16 with PostGIS 3.5 extension
- HealAdvanced spatial queries (radius search, polygon containment)
- [ ] Real-time notifications for new reports
- [ ] User profile management and updates
- [ ] Password reset functionality with email
- [ ] Email verification for registration
- [ ] Rate limiting per endpoint
- [ ] API documentation with Swagger/OpenAPI
- [ ] Comprehensive unit and integration tests
- [ ] CI/CD pipeline with GitHub Actions
- [ ] Monitoring and logging (Prometheus, Grafana)
- [ ] Report photo upload and storage
- [ ] Report analytics and statistics dashboard
- [ ] Mobile app support (Push notifications)
- [ ] Multi-language support (i18n)
- [ ] Export reports to CSV/Excel
  - User registration endpoint
  - User login endpoint
  - User logout endpoint
- PASETO v2 token generation and management
- Argon2id password hashing
- Input validation with go-playground/validator
- Black box testing for all authentication endpoints

#### Security
- Implemented Argon2id for secure password hashing
- Implemented PASETO v2 for secure token management
- Role-based access control structure

#### Infrastructure
- PostgreSQL 16 with Docker containerization
- Health checks for database container
- Resource limits for containers
- Logging configuration

## 🏛 Architecture Decisions

### Why Hexagonal Architecture?

1. **Testability**: Business logic is isolated and easy to test
2. **Flexibility**: Easy to swap adapters (e.g., switch from Fiber to Gin)
3. **Maintainability**: Clear separation of concerns
4. **Domain-Driven**: Business logic is the core, not the framework

### Why Fiber?

- High performance (built on fasthttp)
- Express-like API (easy to learn)
- Rich middleware ecosystem
- Great documentation

### Why PASETO over JWT?

- More secure by design (no algorithm confusion)
- Encrypted payloads
- Simpler to use correctly
- Better for new projects

### Why Argon2id?

- Winner of Password Hashing Competition
- Memory-hard (resistant to GPU attacks)
- Configurable parameters
- Recommended by OWASP

## 🔮 Future Enhancements

- [ ] JWT middleware for protected routes
- [ ] Role-based authorization middleware
- [ ] User profile management
- [ ] Password reset functionality
- [ ] Email verification
- [ ] Rate limiting
- [ ] API documentation with Swagger
- [ ] Unit and integration tests
- [ ] CI/CD pipeline
- [ ] Monitoring and logging
- [ ] API versioning strategy

## 🤝 Contributing

Contributions are welcome! Please follow these steps:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the terms specified in the [LICENSE](LICENSE) file.

## 👨‍💻 Author

**ipincamp**

## 📧 Contact

For questions or support, please open an issue on the GitHub repository.

---

**Built with ❤️ using Go and Hexagonal Architecture**
