# Deployment Guide - Radar Jentik API

This guide covers different deployment strategies for the Radar Jentik API.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Environment Configuration](#environment-configuration)
- [Deployment Options](#deployment-options)
  - [Docker Deployment](#docker-deployment)
  - [Bare Metal Deployment](#bare-metal-deployment)
  - [Cloud Deployment (VPS)](#cloud-deployment-vps)
- [Database Setup](#database-setup)
- [Security Considerations](#security-considerations)
- [Monitoring and Logging](#monitoring-and-logging)
- [Backup and Recovery](#backup-and-recovery)
- [Troubleshooting](#troubleshooting)

## Prerequisites

- Server with Ubuntu 20.04+ or similar Linux distribution
- Minimum 2GB RAM, 2 CPU cores, 20GB storage (for PostGIS support)
- PostgreSQL 16 with PostGIS 3.5 extension
- Domain name (optional, for HTTPS)
- SSH access to server

## Environment Configuration

### Production Environment Variables

Create a `.env` file for production with secure values:

```env
# Application Settings
APP_PORT=:3000
APP_ENV=production

# Database Configuration (PostgreSQL with PostGIS)
DB_HOST=localhost
DB_PORT=5432
DB_NAME=radar_jentik_prod
DB_USERNAME=rj_user
DB_PASSWORD=STRONG_PASSWORD_HERE
DB_TIMEZONE=Asia/Jakarta
DB_SSL_MODE=require

# PASETO Token Configuration
PASETO_SECRET_KEY=GENERATE_32_CHAR_RANDOM_KEY_HERE
PASETO_EXP_DURATION=24h
PASETO_AUDIENCE=radar-jentik-app
PASETO_ISSUER=radar-jentik-api
```

**Important Notes:**
- Database must have PostGIS extension enabled
- Use `postgis/postgis:16-3.5-alpine` Docker image for PostGIS support
- Ensure sufficient storage for spatial data and heatmap calculations

### Generating Secure Keys

**PASETO Secret Key (32 characters):**
```bash
openssl rand -base64 32 | head -c 32
```

**Database Password:**
```bash
openssl rand -base64 24
```

## Deployment Options

### Docker Deployment

#### Option 1: Docker Compose (Recommended for Small/Medium Scale)

1. **Prepare the server:**

```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Install Docker Compose
sudo apt install docker-compose -y

# Add user to docker group
sudo usermod -aG docker $USER
newgrp docker
```

2. **Clone the repository:**

```bash
git clone https://github.com/ipincamp/radar-jentik-api.git
cd radar-jentik-api
```

3. **Configure environment:**

```bash
cp .env.example .env
nano .env  # Edit with production values
```

4. **Create production docker-compose:**

Create `docker-compose.prod.yml`:

```yaml
version: '3.8'

services:
  api:
    build:
      context: .
      dockerfile: Dockerfile
    container_name: rj_api
    restart: unless-stopped
    ports:
      - "3000:3000"
    environment:
      - APP_PORT=:3000
      - APP_ENV=production
      - DB_HOST=db
      - DB_PORT=5432
      - DB_NAME=${DB_NAME}
      - DB_USERNAME=${DB_USERNAME}
      - DB_PASSWORD=${DB_PASSWORD}
      - DB_TIMEZONE=${DB_TIMEZONE}
      - DB_SSL_MODE=${DB_SSL_MODE}
      - PASETO_SECRET_KEY=${PASETO_SECRET_KEY}
      - PASETO_EXP_DURATION=${PASETO_EXP_DURATION}
      - PASETO_AUDIENCE=${PASETO_AUDIENCE}
      - PASETO_ISSUER=${PASETO_ISSUER}
    depends_on:
      db:
        conditiois/postgis:16-3.5-alpine
    container_name: rj_postgres
    restart: unless-stopped
    environment:
      POSTGRES_DB: ${DB_NAME}
      POSTGRES_USER: ${DB_USERNAME}
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - rj_data:/var/lib/postgresql/data
    networks:
      - rj_network
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${DB_USERNAME} -d ${DB_NAME}"]
      interval: 10s
      timeout: 5s
      retries: 5
    # PostGIS-specific configuration
    command: postgres -c shared_preload_libraries=postgis-3rk
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${DB_USERNAME} -d ${DB_NAME}"]
      interval: 10s
      timeout: 5s
      retries: 5

networks:
  rj_network:
    driver: bridge

volumes:
  rj_data:
```

5. **Create Dockerfile:**

```dockerfile
# Build stage
FROM golang:1.25.5-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/api

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/main .

EXPOSE 3000

CMD ["./main"]
```

6. **Deploy:**go run cmd/migrate/main.go

# (Optional) Seed area data
docker-compose -f docker-compose.prod.yml exec api go run cmd/seeder/main.go /path/to/geojson.txt

# Check logs
docker-compose -f docker-compose.prod.yml logs -f
```

**PostGIS Verification:**

```bash
# Verify PostGIS extension
docker-compose -f docker-compose.prod.yml exec db psql -U ${DB_USERNAME} -d ${DB_NAME} -c "SELECT PostGIS_version();"
docker-compose -f docker-compose.prod.yml up -d

# Run migrations
docker-compose -f docker-compose.prod.yml exec api ./main migrate

# Check logs
docker-compose -f docker-compose.prod.yml logs -f
```

#### Option 2: Docker Swarm (For High Availability)

```bash
# Initialize swarm
docker swarm init

# Deploy stack
docker stack deploy  with PostGIS
sudo apt install postgresql postgresql-contrib postgis postgresql-16-postgis-3
# Check services
docker service ls

# Scale API service
docker service scale radar_jentik_api=3
```

### Bare Metal Deployment

#### 1. Install Dependencies

```bash
# Update system
sudo apt update && sud with PostGIS

```bash
# Switch to postgres user
sudo -u postgres psql

# In PostgreSQL shell:
CREATE DATABASE radar_jentik_prod;
CREATE USER rj_user WITH PASSWORD 'your_secure_password';
GRANT ALL PRIVILEGES ON DATABASE radar_jentik_prod TO rj_user;

# Connect to the database
\c radar_jentik_prod

# Enable PostGIS extension
CREATE EXTENSION IF NOT EXISTS postgis;
CREATE EXTENSION IF NOT EXISTS postgis_topology;

# Verify PostGIS installation
SELECT PostGIS_version();

# Grant necessary permissions
GRANT ALL ON geometry_columns TO rj_user;
GRANT ALL ON spatial_ref_sys TO rj_user;

\q
```

**PostGIS Extension Verification:**
```bash
# Check PostGIS functions
suBuild migration tool
go build -o migrate ./cmd/migrate/main.go

# Build seeder tool
go build -o seeder ./cmd/seeder/main.go

# Run migrations
./migrate

# (Optional) Seed area data
./seeder /path/to/geojson.txt
#### 2. Setup Database

```bash
# Switch to postgres user
sudo -u postgres psql

# In PostgreSQL shell:
CREATE DATABASE radar_jentik_prod;
CREATE USER rj_user WITH PASSWORD 'your_secure_password';
GRANT ALL PRIVILEGES ON DATABASE radar_jentik_prod TO rj_user;
\q
```

#### 3. Deploy Application

```bash
# Create app directory
sudo mkdir -p /opt/radar-jentik-api
sudo chown $USER:$USER /opt/radar-jentik-api

# Clone repository
cd /opt/radar-jentik-api
git clone https://github.com/ipincamp/radar-jentik-api.git .

# Configure environment
cp .env.example .env
nano .env  # Edit with production values

# Build application
go build -o radar-jentik-api ./cmd/api/main.go

# Run migrations
go run ./cmd/migrate/main.go
```

#### 4. Create Systemd Service

Create `/etc/systemd/system/radar-jentik-api.service`:

```ini
[Unit]
Description=Radar Jentik API
After=network.target postgresql.service
Wants=postgresql.service

[Service]
Type=simple
User=www-data
Group=www-data
WorkingDirectory=/opt/radar-jentik-api
ExecStart=/opt/radar-jentik-api/radar-jentik-api
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

# Security
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/opt/radar-jentik-api

[Install]
WantedBy=multi-user.target
```

Enable and start the service:

```bash
sudo systemctl daemon-reload
sudo systemctl enable radar-jentik-api
sudo systemctl start radar-jentik-api
sudo systemctl status radar-jentik-api
```

### Cloud Deployment (VPS)

#### AWS EC26/main/postgresql.conf

# Recommended settings for PostGIS:
max_connections = 100
shared_buffers = 512MB          # Increased for spatial data
effective_cache_size = 2GB       # Increased for better query performance
maintenance_work_mem = 128MB     # Increased for spatial indexing
checkpoint_completion_target = 0.9
wal_buffers = 16MB
default_statistics_target = 100
random_page_cost = 1.1
effective_io_concurrency = 200
work_mem = 4MB                   # Increased for spatial operations
min_wal_size = 1GB
max_wal_size = 4GB

# PostGIS-specific settings
shared_preload_libraries = 'postgis-3'
```

**Spatial Index Optimization:**

```sql
-- Create spatial indexes for better performance
CREATE INDEX IF NOT EXISTS idx_reports_location ON reports USING GIST (location);
CREATE INDEX IF NOT EXISTS idx_areas_geometry ON areas USING GIST (geometry);

-- Vacuum and analyze for spatial data
VACUUM ANALYZE reports;
VACUUM ANALYZE areas;
#### DigitalOcean Droplet

1. **Create Droplet:**
   - Ubuntu 20.04
   - Basic ($6/month or better)
   - Enable monitoring

2. **Deploy using Docker:**

```bash
ssh root@your-droplet-ip

# Follow "Docker Deployment" steps above
```

#### Google Cloud Platform

1. **Create Compute Engine instance**
2. **Setup firewall rules**
3. **Deploy using preferred method**

## Database Setup

### Production Database Configuration

```bash
# Edit PostgreSQL config
sudo nano /etc/postgresql/14/main/postgresql.conf

# Recommended settings:
max_connections = 100
shared_buffers = 256MB
effective_cache_size = 1GB
maintenance_work_mem = 64MB
checkpoint_completion_target = 0.9
wal_buffers = 16MB
default_statistics_target = 100
random_page_cost = 1.1
effective_io_concurrency = 200
work_mem = 2621kB
min_wal_size = 1GB
max_wal_size = 4GB
```

### Backup Strategy

**Automated backup script** (`/opt/scripts/backup-db.sh`):

```bash
#!/bin/bash
BACKUP_DIR="/var/backups/radar-jentik"
DATE=$(date +%Y%m%d_%H%M%S)
DB_NAME="radar_jentik_prod"
DB_USER="rj_user"

mkdir -p $BACKUP_DIR

# Backup database
PGPASSWORD=$DB_PASSWORD pg_dump -U $DB_USER -h localhost $DB_NAME | gzip > $BACKUP_DIR/backup_$DATE.sql.gz

# Keep only last 7 days
find $BACKUP_DIR -name "backup_*.sql.gz" -mtime +7 -delete

echo "Backup completed: backup_$DATE.sql.gz"
```

Make executable and schedule:

```bash
chmod +x /opt/scripts/backup-db.sh

# Add to crontab (daily at 2 AM)
crontab -e
0 2 * * * /opt/scripts/backup-db.sh
```

## Security Considerations

### 1. Firewall Setup

```bash
# Enable UFW
sudo ufw enable

# Allow SSH
sudo ufw allow 22/tcp

# Allow API port (if direct access needed)
sudo ufw allow 3000/tcp

# Check status
sudo ufw status
```

### 2. Nginx Reverse Proxy with SSL

**Install Nginx and Certbot:**

```bash
sudo apt install nginx certbot python3-certbot-nginx -y
```

**Nginx configuration** (`/etc/nginx/sites-available/radar-jentik-api`):

```nginx
server {
    listen 80;
    server_name api.yourdomain.com;

    location / {
        proxy_pass http://localhost:3000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
    }
}
```

**Enable site and get SSL:**

```bash
sudo ln -s /etc/nginx/sites-available/radar-jentik-api /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl restart nginx

# Get SSL certificate
sudo certbot --nginx -d api.yourdomain.com
```

### 3. Rate Limiting (Future)

Will be implemented in application layer and/or Nginx:

```nginx
limit_req_zone $binary_remote_addr zone=api_limit:10m rate=10r/s;

location /api/v1/auth/login {
    limit_req zone=api_limit burst=5 nodelay;
    proxy_pass http://localhost:3000;
}
```

## Monitoring and Logging

### Application Logs

```bash
# View systemd logs
sudo journalctl -u radar-jentik-api -f

# View Docker logs
docker-compose logs -f api
```

### Health Check Endpoint (Future)

Will implement `/health` endpoint:

```bash
curl http://localhost:3000/health
```

### Monitoring Tools (Recommended)

- **Prometheus** + **Grafana**: Metrics and dashboards
- **Uptime Kuma**: Uptime monitoring
- **Sentry**: Error tracking

## Backup and Recovery

### Database Restore

```bash
# Restore from backup
gunzip < /var/backups/radar-jentik/backup_20251214_020000.sql.gz | \
  PGPASSWORD=$DB_PASSWORD psql -U rj_user -h localhost radar_jentik_prod
```

### Application Rollback

```bash
# Docker
docker-compose down
git checkout previous-version
docker-compose up -d

# Systemd
sudo systemctl stop radar-jentik-api
cd /opt/radar-jentik-api
git checkout previous-version
go build -o radar-jentik-api ./cmd/api/main.go
sudo systemctl start radar-jentik-api
```
User indexes
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);

-- Report indexes
CREATE INDEX IF NOT EXISTS idx_reports_user_id ON reports(user_id);
CREATE INDEX IF NOT EXISTS idx_reports_status ON reports(status);
CREATE INDEX IF NOT EXISTS idx_reports_created_at ON reports(created_at);

-- Spatial indexes (GIST for geometry types)
CREATE INDEX IF NOT EXISTS idx_reports_location ON reports USING GIST (location);
CREATE INDEX IF NOT EXISTS idx_areas_geometry ON areas USING GIST (geometry);

-- Composite indexes for common queries
CREATE INDEX IF NOT EXISTS idx_reports_status_user ON reports(status, user_id);

-- Analyze tables for query planner
ANALYZE users;
ANALYZE reports;
ANALYZE areas;
```

**Spatial Query Performance Tips:**
- Always use spatial indexes (GIST) for geometry columns
- Use `ST_DWithin` instead of `ST_Distance` for distance queries
- Pre-filter with bounding box queries before complex spatial operations
- Consider using simplified geometries for large-scale visualizationsheck logs
sudo journalctl -u radar-jentik-api -n 50

# Common issues:
# - Database connection: Check DB credentials in .env
# - Port in use: lsof -i :3000
# - Permissions: Check fil - important for spatial data
vacuumdb -U rj_user -d radar_jentik_prod -z -v

# Reindex (monthly) - includes spatial indexes
reindexdb -U rj_user -d radar_jentik_prod

# Update spatial statistics (monthly)
sudo -u postgres psql -d radar_jentik_prod <<EOF
VACUUM ANALYZE reports;
VACUUM ANALYZE areas;
SELECT UpdateGeometrySRID('reports', 'location', 4326);
SELECT UpdateGeometrySRID('areas', 'geometry', 4326);
EOF
```

**PostGIS-Specific Maintenance:**
```bash
# Check spatial reference systems
psql -U rj_user -d radar_jentik_prod -c "SELECT * FROM spatial_ref_sys WHERE srid=4326;"

# Validate geometries
psql -U rj_user -d radar_jentik_prod -c "SELECT id, ST_IsValid(location) FROM reports WHERE NOT ST_IsValid(location);"
```bash
# Test PostgreSQL connection
psql -U rj_user -h localhost -d radar_jentik_prod

# Check PostgreSQL status
sudo systemctl status postgresql

# Check PostgreSQL logs
sudo tail -f /var/log/postgresql/postgresql-14-main.log
```

### High Memory Usage

```bash
# Check memory
free -h

# Check application memory
ps aux | grep radar-jentik-api

# If using Docker, set memory limits in docker-compose
```

### SSL Certificate Issues

```bash
# Renew certificate
sudo certbot renew

# Test renewal
sudo certbot renew --dry-run
```

## Performance Tuning

### Database Optimization

```sql
-- Create indexes
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_created_at ON users(created_at);

-- Analyze tables
ANALYZE users;
```

### Application Optimization

- Use connection pooling (GORM default)
- Implement caching (Redis - future)
- Enable GZIP compression in Nginx
- Use CDN for static assets (if any)

## Maintenance

### Regular Updates

```bash
# Update application
cd /opt/radar-jentik-api
git pull origin main
go build -o radar-jentik-api ./cmd/api/main.go
sudo systemctl restart radar-jentik-api

# Update system packages
sudo apt update && sudo apt upgrade -y
```

### Database Maintenance

```bash
# Vacuum database (weekly)
vacuumdb -U rj_user -d radar_jentik_prod -z -v

# Reindex (monthly)
reindexdb -U rj_user -d radar_jentik_prod
```

---

**Need Help?** Open an issue on GitHub or contact the maintainers.
