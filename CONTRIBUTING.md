# Contributing to Radar Jentik API

Thank you for your interest in contributing to Radar Jentik API! This document provides guidelines and instructions for contributing to this project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Workflow](#development-workflow)
- [Coding Standards](#coding-standards)
- [Commit Message Guidelines](#commit-message-guidelines)
- [Pull Request Process](#pull-request-process)
- [Project Structure](#project-structure)
- [Testing Guidelines](#testing-guidelines)

## Code of Conduct

### Our Pledge

We are committed to providing a welcoming and inspiring community for everyone. Please be respectful and constructive in your interactions.

### Expected Behavior

- Use welcoming and inclusive language
- Be respectful of differing viewpoints and experiences
- Gracefully accept constructive criticism
- Focus on what is best for the community
- Show empathy towards other community members

## Getting Started

### Prerequisites

- Go 1.25.5 or higher
- Docker and Docker Compose
- Git
- A GitHub account

### Setting Up Your Development Environment

1. **Fork the repository** on GitHub

2. **Clone your fork:**
   ```bash
   git clone https://github.com/YOUR_USERNAME/radar-jentik-api.git
   cd radar-jentik-api
   ```

3. **Add upstream remote:**
   ```bash
   git remote add upstream https://github.com/ipincamp/radar-jentik-api.git
   ```

4. **Install dependencies:**
   ```bash
   go mod download
   ```

5. **Set up environment variables:**
   ```bash
   cp .env.example .env
   # Edit .env with your local configuration
   ```

6. **Start the database:**
   ```bash
   docker-compose up -d
   ```

7. **Run migrations:**
   ```bash
   go run cmd/migrate/main.go
   ```

8. **Start the server:**
   ```bash
   go run cmd/api/main.go
   ```

## Development Workflow

### Creating a New Feature

1. **Sync with upstream:**
   ```bash
   git checkout main
   git pull upstream main
   ```

2. **Create a feature branch:**
   ```bash
   git checkout -b feature/your-feature-name
   ```

3. **Make your changes** following the coding standards

4. **Test your changes:**
   ```bash
   go test ./...
   
   # Test with PostGIS if spatial features
   docker-compose up -d
   go run cmd/migrate/main.go
   go test ./internal/adapters/driven/postgres/repositories/...
   ```

5. **Commit your changes** (see commit guidelines below)

6. **Push to your fork:**
   ```bash
   git push origin feature/your-feature-name
   ```

7. **Create a Pull Request** on GitHub

### Branch Naming Convention

- **Features:** `feature/feature-name`
- **Bug fixes:** `fix/bug-description`
- **Documentation:** `docs/what-you-document`
- **Refactoring:** `refactor/what-you-refactor`
- **Tests:** `test/what-you-test`

Examples:
- `feature/user-profile`
- `fix/login-validation`
- `docs/api-endpoints`
- `refactor/auth-service`
- `test/auth-handler`

## Coding Standards

### Go Code Style

We follow the official [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments).

#### Key Guidelines

1. **Use `gofmt`** to format your code:
   ```bash
   gofmt -w .
   ```

2. **Use `golint`** to check your code:
   ```bash
   golint ./...
   ```

3. **Run `go vet`** to detect issues:
   ```bash
   go vet ./...
   ```

4. **Variable naming:**
   - Use camelCase for variables: `userRepo`, `tokenManager`
   - Use PascalCase for exported types: `UserRepository`, `AuthService`
   - Use short names for short-lived variables: `i`, `j`, `err`

5. **Error handling:**
   ```go
   // Good
   if err != nil {
       return fmt.Errorf("failed to save user: %w", err)
   }
   
   // Avoid
   if err != nil {
       return err
   }
   ```

6. **Comments:**
   - Add comments for exported functions, types, and constants
   - Comments should be complete sentences
   - Start with the name of the element being described

   ```go
   // UserRepository defines methods for user data access.
   type UserRepository interface {
       Save(ctx context.Context, user *domain.User) error
   }
   ```

### Hexagonal Architecture Guidelines

This project follows Hexagonal Architecture (Ports & Adapters):

```
internal/
├── adapters/      # Interface with external world
│   ├── driven/    # Infrastructure (database, external APIs)
│   └── driving/   # UI/API layer (HTTP handlers)
└── core/          # Business logic (framework-independent)
    ├── domain/    # Entities
    ├── ports/     # Interfaces
    └── services/  # Business logic
```

#### Rules

1. **Core should NOT depend on adapters**
   - ❌ Never import adapter packages in core
   - ✅ Core defines interfaces (ports), adapters implement them

2. **Use dependency injection**
   - Pass dependencies through constructors
   - Use interfaces, not concrete types

   ```go
   // Good
   type AuthService struct {
       userRepo ports.UserRepository
   }
   
   func NewAuthService(repo ports.UserRepository) *AuthService {
       return &AuthService{userRepo: repo}
   }
   ```

3. **Keep domain models pure**
   - No GORM tags in domain entities
   - No HTTP-specific logic in domain
   - Domain models should be framework-agnostic

4. **Adapters translate between layers**
   ```go
   // HTTP Request -> DTO -> Domain Model
   // Domain Model -> DTO -> HTTP Response
   ```

### File Organization

- One package per directory
- Related files together
- Test files alongside implementation: `user_repo.go`, `user_repo_test.go`

### Import Ordering

1. Standard library
2. External packages
3. Internal packages

```go
import (
    "context"
    "fmt"
    
    "github.com/gofiber/fiber/v2"
    "gorm.io/gorm"
    
    "github.com/ipincamp/radar-jentik-api/internal/core/domain"
    "github.com/ipincamp/radar-jentik-api/internal/core/ports"
)
```

## Commit Message Guidelines

We follow the [Conventional Commits](https://www.conventionalcommits.org/) specification.

### Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

- **feat**: A new feature
- **fix**: A bug fix
- **docs**: Documentation only changes
- **style**: Code style changes (formatting, missing semi colons, etc)
- **refactor**: Code refactoring
- **test**: Adding or updating tests
- **chore**: Maintenance tasks, dependency updates

### Examples

```
feat(auth): add password reset functionality

Implement password reset with email verification.
- Add reset token generation
- Create reset email template
- Add reset endpoint

Closes #123
```

```
fix(login): resolve validation error on empty username

The login handler was not properly validating empty usernames,
causing a 500 error instead of 400 Bad Request.

Fixes #456
```

```
docs(readme): update installation instructions

Add Docker setup instructions and troubleshooting section.
```

### Commit Message Rules

- Use imperative mood ("add" not "added" or "adds")
- Don't capitalize first letter of subject
- No period at the end of subject
- Limit subject line to 72 characters
- Wrap body at 72 characters
- Reference issues and pull requests in footer

## Pull Request Process

### Before Submitting

1. ✅ Code follows the style guidelines
2. ✅ Self-review completed
3. ✅ Comments added to complex code
4. ✅ Documentation updated (if needed)
5. ✅ Tests added/updated
6. ✅ All tests pass
7. ✅ No new warnings from linters

### PR Template

When creating a PR, include:

```markdown
## Description
Brief description of what this PR does.

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
Describe the tests you ran and how to reproduce them.

## Checklist
- [ ] Code follows style guidelines
- [ ] Self-review performed
- [ ] Comments added where needed
- [ ] Documentation updated
- [ ] Tests added/updated
- [ ] All tests pass
- [ ] No new warnings

## Related Issues
Closes #(issue number)
```

### Review Process

1. At least one maintainer approval required
2. All CI checks must pass
3. Conflicts must be resolved
4. Discussion points must be addressed

### After Approval

- Squash commits if needed
- Maintainer will merge the PR

## Project Structure

When adding new features, follow this structure:

### Adding a New Endpoint

1. **Define domain model** in `internal/core/domain/`
   - Keep models framework-agnostic
   - For spatial data, use `float64` for lat/lon in domain
   - No GORM tags in domain entities

2. **Define port interface** in `internal/core/ports/`
   - Define repository interface
   - Define service interface
   - Define DTOs for requests/responses

3. **Implement service** in `internal/core/services/`
   - Implement business logic
   - Handle validations

**Spatial Data Migrations:**

For tables with geometric columns:

```go
func (m *Migration20251214000002) Up(tx *gorm.DB) error {
    // Use raw SQL for PostGIS types
    return tx.Exec(`
        CREATE TABLE reports (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            location GEOMETRY(Point, 4326) NOT NULL,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );
        
        -- Create spatial index
        CREATE INDEX idx_reports_location ON reports USING GIST (location);
    `).Error
}

func (m *Migration20251214000002) Down(tx *gorm.DB) error {
    return tx.Exec(`DROP TABLE IF EXISTS reports CASCADE`).Error
}
```

**Important PostGIS Notes:**
- Always specify SRID (4326 for WGS84/GPS coordinates)
- Use `GIST` indexes for spatial columns
- Use appropriate geometry types: `Point`, `Polygon`, `MultiPolygon`
- Consider using `ST_Force2D` for 3D coordinates if only 2D needed
   - Transform between DTOs and domain models

4. **Implement repository** in `internal/adapters/driven/postgres/repositories/`
   - Create databas with PostGIS
- Clean up after tests
- Test spatial queries

**Example: Spatial Repository Test**

```go
func TestReportRepo_Create_WithLocation(t *testing.T) {
    // Setup test database with PostGIS
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)
    
    repo := NewReportRepo(db)
    
    // Test data
    report := &domain.Report{
        ID:        "test-id",
        Latitude:  -7.250445,
        Longitude: 112.768845,
    }
    
    // Execute
    err := repo.Create(context.Background(), report)
    
    // Assert
    assert.NoError(t, err)
    
    // Verify spatial data
    var lat, lon float64
    err = db.Raw(
        "SELECT ST_Y(location), ST_X(location) FROM reports WHERE id = ?",
        report.ID,
    ).Row().Scan(&lat, &lon)
    
    assert.NoError(t, err)
    assert.InDelta(t, -7.250445, lat, 0.000001)
    assert.InDelta(t, 112.768845, lon, 0.000001)
}
```

### Running Tests

```bash
# All tests
go test ./...

# Specific package
go test ./internal/core/services

# Repository tests (requires PostGIS)
docker-compose up -d
go test ./internal/adapters/driven/postgres/repositories/...

# With coverage
go test -cover ./...

# Verbose output
go test -v ./...

# Race detection
go test -race ./...
```

### Black Box Testing

Use Postman or similar tools to test endpoints:

1. **Authentication Flow**
   - Register new user
   - Login to get token
   - Use token for protected endpoints

2. **Report Management**
   - Create report with valid coordinates
   - Verify report appears in list
   - Test pagination
   - Validate report as petugas

3. **Spatial Features**
   - Create reports at different locations
   - Test heatmap generation with different parameters
   - Verify GeoJSON export for areas

4. **Role-Based Access**
- Document PostGIS functions used

**Example:**
```go
// CreateReport stores a new report with geolocation data.
// The location is stored as a PostGIS POINT geometry using ST_MakePoint.
// Coordinates must be in WGS84 (SRID 4326) format.
func (r *ReportRepo) CreateReport(ctx context.Context, report *domain.Report) error {
    // Implementation
}
```

### API Documentation

- Update `API.md` for new endpoints
- Include request/response examples
- Document error cases
- Document query parameters
- Include cURL examples
- For spatial endpoints, document coordinate format

### README Updates

- Update README for significant changes
- Keep installation instructions current
- Update architecture diagrams if needed
- Document new dependencies (e.g., PostGIS)

### Changelog

Update `.vscode/changelog.txt` with:
- Date and time
- Feature description
- Technical implementation details
- Testing verification notes
}

// Repository (internal/adapters/driven/postgres/repositories/report_repo.go)
type ReportModel struct {
    ID       string `gorm:"primaryKey"`
    Location string `gorm:"type:geometry(Point,4326)"` // PostGIS geometry
}

func (r *ReportRepo) Create(ctx context.Context, report *domain.Report) error {
    // Convert domain to database model with PostGIS
    model := &ReportModel{
        ID: report.ID,
    }
    
    // Use raw SQL with ST_MakePoint
    return r.db.Exec(
        "INSERT INTO reports (id, location) VALUES (?, ST_SetSRID(ST_MakePoint(?, ?), 4326))",
        model.ID, report.Longitude, report.Latitude,
    ).Error
}
```

### Adding a New Migration

1. Copy `_TEMPLATE.txt` in `internal/adapters/driven/postgres/migrations/`
2. Name it: `YYYYMMDDHHMMSS_description.go`
3. Implement `Up()` and `Down()` methods
4. Register in `registry.go`

## Testing Guidelines

### Unit Tests

- Test business logic in services
- Mock dependencies using interfaces
- Aim for >80% coverage

```go
func TestAuthService_Register(t *testing.T) {
    // Setup
    mockRepo := &MockUserRepository{}
    mockToken := &MockTokenManager{}
    service := NewAuthService(mockRepo, mockToken)
    
    // Test
    err := service.Register(context.Background(), ports.RegisterRequest{
        Name: "Test User",
        Username: "testuser",
        Password: "password123",
    })
    
    // Assert
    if err != nil {
        t.Errorf("expected no error, got %v", err)
    }
}
```

### Integration Tests

- Test adapter implementations
- Use test database
- Clean up after tests

### Running Tests

```bash
# All tests
go test ./...

# Specific package
go test ./internal/core/services

# With coverage
go test -cover ./...

# Verbose output
go test -v ./...
```

## Documentation

### Code Documentation

- Document all exported functions, types, and constants
- Use godoc-style comments
- Provide examples where helpful

### API Documentation

- Update `API.md` for new endpoints
- Include request/response examples
- Document error cases

### README Updates

- Update README for significant changes
- Keep installation instructions current
- Update architecture diagrams if needed

## Questions?

- Open an issue for questions
- Join discussions in existing issues
- Contact maintainers

## License

By contributing, you agree that your contributions will be licensed under the same license as the project.

---

Thank you for contributing to Radar Jentik API! 🎉
