# API Documentation - Radar Jentik API

## Base URL
```
http://localhost:3000
```

## API Version
```
v1
```

## API Endpoints

All API endpoints are prefixed with `/api/v1`

---

## Authentication

### 1. Register New User

Creates a new user account in the system.

**Endpoint:** `POST /api/v1/auth/register`

**Headers:**
```
Content-Type: application/json
```

**Request Body:**
```json
{
  "name": "string (required)",
  "username": "string (required)",
  "password": "string (required, min 6 characters)"
}
```

**Example Request:**
```json
{
  "name": "John Doe",
  "username": "johndoe",
  "password": "securepass123"
}
```

**Success Response:**

Status Code: `201 Created`

```json
{
  "message": "Registrasi berhasil"
}
```

**Error Responses:**

Status Code: `400 Bad Request`
```json
{
  "error": "Invalid input"
}
```
or
```json
{
  "error": "Key: 'RegisterInput.Password' Error:Field validation for 'Password' failed on the 'min' tag"
}
```

Status Code: `500 Internal Server Error`
```json
{
  "error": "username sudah terdaftar"
}
```

**cURL Example:**
```bash
curl -X POST http://localhost:3000/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "username": "johndoe",
    "password": "securepass123"
  }'
```

---

### 2. User Login

Authenticates a user and returns a PASETO token.

**Endpoint:** `POST /api/v1/auth/login`

**Headers:**
```
Content-Type: application/json
```

**Request Body:**
```json
{
  "username": "string (required)",
  "password": "string (required)"
}
```

**Example Request:**
```json
{
  "username": "johndoe",
  "password": "securepass123"
}
```

**Success Response:**

Status Code: `200 OK`

```json
{
  "token": "v2.local.Gdh5kiOTyyaQ3_bNykYDeYHO21Jg2aUJmJDRHdGFuZGFyZC1jbGFpbXMiLCJleHAiOiIyMDI1LTEyLTE1VDAwOjAwOjAwWiIsImlhdCI6IjIwMjUtMTItMTRUMDA6MDA6MDBaIiwiaXNzIjoicmFkYXItamVudGlrLWFwaSIsImp0aSI6IjEyMzQ1Njc4OTAiLCJuYmYiOiIyMDI1LTEyLTE0VDAwOjAwOjAwWiIsInN1YiI6IjEyMzQ1Njc4OTAifQ"
}
```

**Token Information:**
- Type: PASETO v2.local (encrypted)
- Algorithm: Symmetric encryption
- Expiration: Based on `PASETO_EXP_DURATION` (default: 24 hours)
- Claims:
  - `aud`: radar-jentik-app
  - `iss`: radar-jentik-api
  - `jti`: user ID
  - `sub`: user ID
  - `iat`: issued at timestamp
  - `exp`: expiration timestamp
  - `nbf`: not before timestamp
- Footer: Contains user role

**Error Responses:**

Status Code: `400 Bad Request`
```json
{
  "error": "Invalid input"
}
```

Status Code: `401 Unauthorized`
```json
{
  "error": "username atau password salah"
}
```

**cURL Example:**
```bash
curl -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "johndoe",
    "password": "securepass123"
  }'
```

**Usage of Token:**

Once you receive the token, include it in subsequent requests:

```bash
curl -X GET http://localhost:3000/api/v1/protected-endpoint \
  -H "Authorization: Bearer v2.local.Gdh5kiOTyyaQ3_bNykYDeYHO21Jg2..."
```

---

### 3. User Logout

Logs out the current user. Since the API uses stateless authentication, this endpoint instructs the client to discard the token.

**Endpoint:** `POST /api/v1/auth/logout`

**Headers:**
```
(none required)
```

**Request Body:**
```
(empty)
```

**Success Response:**

Status Code: `200 OK`

```json
{
  "message": "Logout berhasil"
}
```

**cURL Example:**
```bash
curl -X POST http://localhost:3000/api/v1/auth/logout
```

**Client-Side Implementation:**
After receiving the success response, the client application should:
1. Remove the token from storage (localStorage, sessionStorage, cookies, etc.)
2. Clear any user session data
3. Redirect to login page

---

## Error Handling

### Standard Error Response Format

All error responses follow this format:

```json
{
  "error": "Error message describing what went wrong"
}
```

### Common HTTP Status Codes

| Status Code | Meaning | Common Causes |
|-------------|---------|---------------|
| `200 OK` | Request successful | - |
| `201 Created` | Resource created successfully | Registration successful |
| `400 Bad Request` | Invalid request data | Missing fields, validation errors |
| `401 Unauthorized` | Authentication failed | Invalid credentials, expired token |
| `403 Forbidden` | Access denied | Insufficient permissions |
| `404 Not Found` | Resource not found | Invalid endpoint |
| `500 Internal Server Error` | Server error | Database error, unexpected error |

---

## Data Models

### User Model

```go
type User struct {
    ID        string    // UUID
    Name      string    // Full name
    Username  string    // Unique username
    Password  string    // Hashed with Argon2id
    Role      string    // User role (e.g., "kader")
    CreatedAt time.Time // Registration timestamp
    UpdatedAt time.Time // Last update timestamp
}
```

**Default Role:** `kader` (assigned during registration)

---

## Security

### Password Requirements

- Minimum length: 6 characters
- Stored as Argon2id hash
- Hash parameters:
  - Memory: 64MB
  - Iterations: 1
  - Parallelism: 2
  - Salt length: 16 bytes
  - Key length: 32 bytes

### Token Security

**PASETO v2.local** provides:
- Symmetric encryption (confidential payload)
- Authentication (integrity protection)
- No algorithm confusion attacks
- Built-in expiration handling

**Token Storage Recommendations:**
- Web: httpOnly cookies (recommended) or sessionStorage
- Mobile: Secure storage (Keychain/Keystore)
- Never store in localStorage if XSS is a concern

---

## Rate Limiting

*(Not yet implemented)*

Future implementation will include:
- Rate limiting per IP address
- Rate limiting per user account
- Configurable limits via environment variables

---

## Validation Rules

### Register Endpoint

| Field | Type | Required | Rules |
|-------|------|----------|-------|
| name | string | Yes | Non-empty |
| username | string | Yes | Non-empty, unique |
| password | string | Yes | Minimum 6 characters |

### Login Endpoint

| Field | Type | Required | Rules |
|-------|------|----------|-------|
| username | string | Yes | Non-empty |
| password | string | Yes | Non-empty |

---

## Testing

### Postman Collection

Import the following collection to test all endpoints:

```json
{
  "info": {
    "name": "Radar Jentik API",
    "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
  },
  "item": [
    {
      "name": "Auth",
      "item": [
        {
          "name": "Register",
          "request": {
            "method": "POST",
            "header": [
              {
                "key": "Content-Type",
                "value": "application/json"
              }
            ],
            "body": {
              "mode": "raw",
              "raw": "{\n  \"name\": \"Test User\",\n  \"username\": \"testuser\",\n  \"password\": \"password123\"\n}"
            },
            "url": {
              "raw": "http://localhost:3000/api/v1/auth/register",
              "protocol": "http",
              "host": ["localhost"],
              "port": "3000",
              "path": ["api", "v1", "auth", "register"]
            }
          }
        },
        {
          "name": "Login",
          "request": {
            "method": "POST",
            "header": [
              {
                "key": "Content-Type",
                "value": "application/json"
              }
            ],
            "body": {
              "mode": "raw",
              "raw": "{\n  \"username\": \"testuser\",\n  \"password\": \"password123\"\n}"
            },
            "url": {
              "raw": "http://localhost:3000/api/v1/auth/login",
              "protocol": "http",
              "host": ["localhost"],
              "port": "3000",
              "path": ["api", "v1", "auth", "login"]
            }
          }
        },
        {
          "name": "Logout",
          "request": {
            "method": "POST",
            "url": {
              "raw": "http://localhost:3000/api/v1/auth/logout",
              "protocol": "http",
              "host": ["localhost"],
              "port": "3000",
              "path": ["api", "v1", "auth", "logout"]
            }
          }
        }
      ]
    }
  ]
}
```

---

## Environment Variables

All configuration is managed through environment variables. See [README.md](README.md#configuration) for details.

---

## API Versioning

Current version: **v1**

All endpoints are prefixed with `/api/v1/`

Future versions will be accessible via:
- `/api/v2/`
- `/api/v3/`
- etc.

---

## Support

For issues or questions:
1. Check the [README.md](README.md)
2. Review this API documentation
3. Open an issue on GitHub

---

**Last Updated:** December 14, 2025
