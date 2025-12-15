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

### Authentication Status
- 🌐 **Public**: Accessible without authentication
- 🔒 **Protected**: Requires valid PASETO token in Authorization header
- 👮 **Petugas Only**: Requires valid token with `petugas` role

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

### 4. List All Users (🔒 👮 Petugas Only)

Retrieves a list of all registered users. This endpoint is restricted to users with the `petugas` role for monitoring and management purposes.

**Endpoint:** `GET /api/v1/auth/users`

**Headers:**
```
Authorization: Bearer <token>
```

**Success Response:**

Status Code: `200 OK`

```json
[
  {
    "id": "uuid",
    "name": "John Doe",
    "username": "johndoe",
    "role": "kader",
    "created_at": "2025-12-14T10:00:00Z"
  },
  {
    "id": "uuid",
    "name": "Jane Smith",
    "username": "janesmith",
    "role": "petugas",
    "created_at": "2025-12-14T11:00:00Z"
  }
]
```

**Error Responses:**

Status Code: `401 Unauthorized`: "kader" or "petugas"
    CreatedAt time.Time // Registration timestamp
    UpdatedAt time.Time // Last update timestamp
}
```

**Roles:**
- `kader` (default): Regular user, can create and view own reports
- `petugas`: Officer/administrator, can view all reports, validate reports, and manage users

---

### Report Model

```go
type Report struct {
    ID           string     // UUID
    UserID       string     // Creator's user ID
    LocationName string     // Human-readable location
    Latitude     float64    // Latitude coordinate
    Longitude    float64    // Longitude coordinate
    Description  string     // Report description (optional)
    Status       string     // "pending", "verified", or "rejected"
    VerifiedBy   *string    // Verifier's user ID (null if pending)
    VerifiedAt   *time.Time // Verification timestamp (null if pending)
    CreatedAt    time.Time  // Creation timestamp
    UpdatedAt    time.Time  // Last update timestamp
}
```

**Status Values:**
- `pending`: Newly created, awaiting validation
- `verified`: Validated and confirmed by petugas
- `rejected`: Reviewed and rejected by petugas

**Database Storage:**
- Location stored as PostGIS geometry (POINT) in database
- Latitude/Longitude exposed in business logic and API responses

---

### Area Model

```go
type Area struct {
    ID        string    // UUID
    Name      string    // Area name (e.g., "Desa Panusupan")
    Geometry  string    // GeoJSON MultiPolygon string
    CreatedAt time.Time // Creation timestamp
    UpdatedAt time.Time // Last update timestamp
}
```

**Geometry Format:**
- Stored as PostGIS MultiPolygon in database
- Exposed as GeoJSON string in business logic
- Converted to standard GeoJSON in API responses

---

### Heatmap Point Model

```go
type HeatmapPoint struct {
    Latitude  float64 // Grid point latitude
    Longitude float64 // Grid point longitude
    RiskValue float64 // Calculated risk (0.0 to 1.0)
}
```

**Risk Value Interpretation:**
- 0.0 - 0.3: Low risk
- 0.3 - 0.6: Medium risk
- 0.6 - 1.0: High risk
Status Code: `403 Forbidden`
```json
{
  "error": "Access denied. Petugas role required"
}
```

**cURL Example:**
```bash
curl -X GET http://localhost:3000/api/v1/auth/users \
  -H "Authorization: Bearer v2.local.Gdh5kiOTyyaQ3..."
```

**Notes:**
- Only users with `petugas` role can access this endpoint
- Passwords are never included in the response
- Results are ordered by creation date (newest first)

---

## Reports

### 1. Create Report (🔒 Protected)

Creates a new mosquito larvae report with geolocation data. The location is stored as a PostGIS geometry point.

**Endpoint:** `POST /api/v1/reports`

**Headers:**
```
Authorization: Bearer <token>
Content-Type: application/json
```

**Request Body:**
```json
{
  "location_name": "string (required)",
  "latitude": "float64 (required, -90 to 90)",
  "longitude": "float64 (required, -180 to 180)",
  "description": "string (optional)"
}
```

**Example Request:**
```json
{
  "location_name": "Jl. Merdeka No. 10, Panusupan",
  "latitude": -7.250445,
  "longitude": 112.768845,
  "description": "Found mosquito larvae in water container near the house"
}
```

**Success Response:**

Status Code: `201 Created`

```json
{
  "id": "uuid",
  "user_id": "user-uuid",
  "location_name": "Jl. Merdeka No. 10, Panusupan",
  "latitude": -7.250445,
  "longitude": 112.768845,
  "description": "Found mosquito larvae in water container near the house",
  "status": "pending",
  "created_at": "2025-12-15T09:00:00Z",
  "updated_at": "2025-12-15T09:00:00Z"
}
```

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
  "error": "Unauthorized"
}
```

**cURL Example:**
```bash
curl -X POST http://localhost:3000/api/v1/reports \
  -H "Authorization: Bearer v2.local.Gdh5kiOTyyaQ3..." \
  -H "Content-Type: application/json" \
  -d '{
    "location_name": "Jl. Merdeka No. 10",
    "latitude": -7.250445,
    "longitude": 112.768845,
    "description": "Found larvae"
  }'
```

**Notes:**
- User ID is automatically extracted from the authentication token
- Initial status is always `pending`
- Location is stored as PostGIS geometry using ST_MakePoint

---

### 2. Get Reports (🔒 Protected - Role-Based)

Retrieves a paginated list of reports with role-based filtering:
- **Kader**: Can only view their own reports
- **Petugas**: Can view all reports

**Endpoint:** `GET /api/v1/reports`

**Headers:**
```
Authorization: Bearer <token>
```

**Query Parameters:**

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| page | integer | No | 1 | Page number (starts from 1) |
| limit | integer | No | 10 | Items per page (max 100) |

**Example Request:**
```
GET /api/v1/reports?page=1&limit=20
```

**Success Response:**

Status Code: `200 OK`

```json
{
  "data": [
    {
      "id": "uuid",
      "user_id": "user-uuid",
      "location_name": "Jl. Merdeka No. 10",
      "latitude": -7.250445,
      "longitude": 112.768845,
      "description": "Found larvae",
      "status": "pending",
      "verified_by": null,
      "verified_at": null,
      "created_at": "2025-12-15T09:00:00Z",
      "updated_at": "2025-12-15T09:00:00Z"
    },
    {
      "id": "uuid",
      "user_id": "user-uuid",
      "location_name": "Jl. Sudirman No. 5",
      "latitude": -7.251234,
      "longitude": 112.769876,
      "description": "Multiple larvae found",
      "status": "verified",
      "verified_by": "petugas-uuid",
      "verified_at": "2025-12-15T10:00:00Z",
      "created_at": "2025-12-15T08:00:00Z",
      "updated_at": "2025-12-15T10:00:00Z"
    }
  ],
  "meta": {
    "page": 1,
    "limit": 20,
    "total": 45
  }
}
```

**Error Responses:**

Status Code: `401 Unauthorized`
```json
{
  "error": "Unauthorized"
}
```

**cURL Example:**
```bash
curl -X GET "http://localhost:3000/api/v1/reports?page=1&limit=10" \
  -H "Authorization: Bearer v2.local.Gdh5kiOTyyaQ3..."
```

**Notes:**
- Kaders can only see reports they created
- Petugas can see all reports in the system
- Results are ordered by creation date (newest first)
- Location coordinates are extracted from PostGIS geometry using ST_X and ST_Y

---

### 3. Validate Report (🔒 Protected)

Validates a report by changing its status from `pending` to either `verified` or `rejected`. The verifier's user ID and timestamp are automatically recorded.

**Endpoint:** `PATCH /api/v1/reports/:id/validate`

**Headers:**
```
Authorization: Bearer <token>
Content-Type: application/json
```

**URL Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| id | string (UUID) | Report ID to validate |

**Request Body:**
```json
{
  "status": "string (required, 'verified' or 'rejected')"
}
```

**Example Request:**
```json
{
  "status": "verified"
}
```

**Success Response:**

Status Code: `200 OK`

```json
{
  "id": "uuid",
  "user_id": "user-uuid",
  "location_name": "Jl. Merdeka No. 10",
  "latitude": -7.250445,
  "longitude": 112.768845,
  "description": "Found larvae",
  "status": "verified",
  "verified_by": "verifier-uuid",
  "verified_at": "2025-12-15T10:00:00Z",
  "created_at": "2025-12-15T09:00:00Z",
  "updated_at": "2025-12-15T10:00:00Z"
}
```

**Error Responses:**

Status Code: `400 Bad Request`
```json
{
  "error": "Invalid status. Must be 'verified' or 'rejected'"
}
```

or

```json
{
  "error": "Report already validated"
}
```

Status Code: `404 Not Found`
```json
{
  "error": "Report not found"
}
```

**cURL Example:**
```bash
curl -X PATCH http://localhost:3000/api/v1/reports/uuid-here/validate \
  -H "Authorization: Bearer v2.local.Gdh5kiOTyyaQ3..." \
  -H "Content-Type: application/json" \
  -d '{"status": "verified"}'
```

**Status Transition Rules:**
- Can only validate reports with `pending` status
- Once validated (verified/rejected), status cannot be changed
- Verifier ID and timestamp are automatically set
- Invalid status transitions return error

---

### 4. Get Heatmap Data (🔒 Protected)

Generates risk estimation heatmap using Inverse Distance Weighting (IDW) algorithm based on validated reports. Returns a grid of points with calculated risk values.

**Endpoint:** `GET /api/v1/reports/heatmap`

**Headers:**
```
Authorization: Bearer <token>
```

**Query Parameters:**

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| res | integer | No | 50 | Grid resolution (points per dimension) |
| p | float | No | 2 | IDW power parameter (higher = more local influence) |

**Example Request:**
```
GET /api/v1/reports/heatmap?res=50&p=2
```

**Success Response:**

Status Code: `200 OK`

```json
[
  {
    "latitude": -7.250000,
    "longitude": 112.768000,
    "risk_value": 0.85
  },
  {
    "latitude": -7.250100,
    "longitude": 112.768100,
    "risk_value": 0.72
  },
  {
    "latitude": -7.250200,
    "longitude": 112.768200,
    "risk_value": 0.45
  }
]
```

**Response Fields:**
- `latitude`: Grid point latitude
- `longitude`: Grid point longitude
- `risk_value`: Calculated risk value (0.0 to 1.0)

**Error Responses:**

Status Code: `400 Bad Request`
```json
{
  "error": "Invalid parameters"
}
```

Status Code: `401 Unauthorized`
```json
{
  "error": "Unauthorized"
}
```

**cURL Example:**
```bash
curl -X GET "http://localhost:3000/api/v1/reports/heatmap?res=50&p=2" \
  -H "Authorization: Bearer v2.local.Gdh5kiOTyyaQ3..."
```

**Algorithm Details:**

The IDW (Inverse Distance Weighting) algorithm calculates risk values using:

$$
\text{risk}(x, y) = \frac{\sum_{i=1}^{n} \frac{1}{d_i^p}}{\sum_{i=1}^{n} \frac{1}{d_i^p}}
$$

Where:
- $d_i$ = distance from grid point to report $i$
- $p$ = power parameter (controls distance decay)
- $n$ = number of verified reports

**Parameter Guidelines:**
- **res (resolution)**: 
  - Lower (20-30): Faster, less detailed
  - Medium (50): Balanced performance and detail
  - Higher (100+): Slower, more detailed
- **p (power)**:
  - Lower (1-1.5): Smoother, wider influence
  - Medium (2): Standard interpolation
  - Higher (3-4): Sharper, more localized

**Notes:**
- Only includes reports with `verified` status
- Returns all users' reports (no role-based filtering)
- Boundary box calculated from all verified report locations
- Grid generated within boundary box with specified resolution
- Empty array returned if no verified reports exist

---

## Areas

### 1. Get All Areas (🔒 Protected)

Retrieves all area boundaries in standard GeoJSON FeatureCollection format. Areas are stored as PostGIS MultiPolygon geometries.

**Endpoint:** `GET /api/v1/areas`

**Headers:**
```
Authorization: Bearer <token>
```

**Success Response:**

Status Code: `200 OK`

```json
{
  "type": "FeatureCollection",
  "features": [
    {
      "type": "Feature",
      "geometry": {
        "type": "MultiPolygon",
        "coordinates": [
          [
            [
              [112.768845, -7.250445],
              [112.769845, -7.250445],
              [112.769845, -7.251445],
              [112.768845, -7.251445],
              [112.768845, -7.250445]
            ]
          ]
        ]
      },
      "properties": {
        "id": "uuid",
        "name": "Desa Panusupan",
        "created_at": "2025-12-14T23:00:00Z",
        "updated_at": "2025-12-14T23:00:00Z"
      }
    }
  ]
}
```

**Error Responses:**

Status Code: `401 Unauthorized`
```json
{
  "error": "Unauthorized"
}
```

Status Code: `500 Internal Server Error`
```json
{
  "error": "Failed to fetch areas"
}
```

**cURL Example:**
```bash
curl -X GET http://localhost:3000/api/v1/areas \
  -H "Authorization: Bearer v2.local.Gdh5kiOTyyaQ3..."
```

**Notes:**
- Returns standard GeoJSON format compatible with Leaflet, Mapbox, etc.
- Geometries are converted from PostGIS using ST_AsGeoJSON
- Coordinates follow GeoJSON order: [longitude, latitude]
- Areas can be directly rendered on web maps without client-side parsing

**Frontend Integration Example (Leaflet):**
```javascript
fetch('http://localhost:3000/api/v1/areas', {
  headers: {
    'Authorization': `Bearer ${token}`
  }
})
  .then(res => res.json())
  .then(data => {
    L.geoJSON(data, {
      style: { color: '#3388ff' }
    }).addTo(map);
  });
```

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
#### Register Endpoint

| Field | Type | Required | Rules |
|-------|------|----------|-------|
| name | string | Yes | Non-empty |
| username | string | Yes | Non-empty, unique |
| password | string | Yes | Minimum 6 characters |

#### Login Endpoint

| Field | Type | Required | Rules |
|-------|------|----------|-------|
| username | string | Yes | Non-empty |
| password | string | Yes | Non-empty |

#### Create Report Endpoint

| Field | Type | Required | Rules |
|-------|------|----------|-------|
| location_name | string | Yes | Non-empty |
| latitude | float64 | Yes | -90 to 90 |
| longitude | float64 | Yes | -180 to 180 |
| description | string | No | Optional text |

#### Validate Report Endpoint

| Field | Type | Required | Rules |
|-------|------|----------|-------|
| status | string | Yes | Must be "verified" or "rejected" |

#### Get Heatmap Endpoint

| Parameter | Type | Required | Rules |
|-----------|------|----------|-------|
| res | integer | No | Positive integer, recommended 20-100 |
| p | float | No | Positive number, recommended 1-4 |
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

## Workflow Examples

### Complete User Journey

#### 1. User Registration & Authentication
```bash
# Register as kader
curl -X POST http://localhost:3000/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"name":"John Doe","username":"johndoe","password":"pass123"}'

# Login to get token
curl -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"johndoe","password":"pass123"}'
# Response: {"token": "v2.local..."}
```

#### 2. Create Report (Kader)
```bash
curl -X POST http://localhost:3000/api/v1/reports \
  -H "Authorization: Bearer v2.local..." \
  -H "Content-Type: application/json" \
  -d '{
    "location_name":"Jl. Merdeka No. 10",
    "latitude":-7.250445,
    "longitude":112.768845,
    "description":"Found larvae in water container"
  }'
```

#### 3. View Own Reports (Kader)
```bash
curl -X GET http://localhost:3000/api/v1/reports?page=1&limit=10 \
  -H "Authorization: Bearer v2.local..."
```

#### 4. Validate Report (Petugas)
```bash
# Login as petugas
curl -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"petugas1","password":"petugaspass"}'

# Validate report
curl -X PATCH http://localhost:3000/api/v1/reports/{report-id}/validate \
  -H "Authorization: Bearer v2.local..." \
  -H "Content-Type: application/json" \
  -d '{"status":"verified"}'
```

#### 5. View Heatmap
```bash
curl -X GET "http://localhost:3000/api/v1/reports/heatmap?res=50&p=2" \
  -H "Authorization: Bearer v2.local..."
```

#### 6. View Areas
```bash
curl -X GET http://localhost:3000/api/v1/areas \
  -H "Authorization: Bearer v2.local..."
```

---

## Role-Based Access Summary

| Endpoint | Kader | Petugas | Public |
|----------|-------|---------|--------|
| POST /auth/register | ✅ | ✅ | ✅ |
| POST /auth/login | ✅ | ✅ | ✅ |
| POST /auth/logout | ✅ | ✅ | ✅ |
| GET /auth/users | ❌ | ✅ | ❌ |
| POST /reports | ✅ | ✅ | ❌ |
| GET /reports | ✅ (own) | ✅ (all) | ❌ |
| PATCH /reports/:id/validate | ✅ | ✅ | ❌ |
| GET /reports/heatmap | ✅ | ✅ | ❌ |
| GET /areas | ✅ | ✅ | ❌ |

---

**Last Updated:** December 15

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
