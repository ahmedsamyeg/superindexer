# Usenet + Spotnet Indexer

A modern, scalable, self-hosted Usenet and Spotnet indexing platform with PostgreSQL-powered full-text search.

## Architecture

- **Backend**: Go (Echo, Bun ORM, pgx)
- **Database**: PostgreSQL with native Full-Text Search
- **Frontend**: React, TypeScript, Vite, shadcn/ui
- **Infrastructure**: Docker, Docker Compose, Kubernetes

## Features

- NNTP server connectivity and management
- Newsgroup discovery and management
- Usenet article header ingestion
- Spotnet source integration and post ingestion
- Metadata extraction and normalization
- Duplicate detection
- PostgreSQL-powered full-text search
- Advanced search filters and relevance ranking
- Background indexing workers with concurrent processing
- User authentication and role-based access control
- Administrative dashboard
- Docker and Kubernetes deployment support
- Prometheus metrics and Grafana dashboards

## Quick Start

### Prerequisites

- Docker and Docker Compose
- Go 1.26+ (for local development)
- Node.js 20+ (for frontend development)

### Docker Compose (Recommended)

```bash
# Copy environment file
cp .env.example .env

# Start all services
docker compose -f deployments/docker/docker-compose.yml up -d

# Run database migrations
docker compose -f deployments/docker/docker-compose.yml run --rm migrate up

# View logs
docker compose -f deployments/docker/docker-compose.yml logs -f api
docker compose -f deployments/docker/docker-compose.yml logs -f worker

# Stop all services
docker compose -f deployments/docker/docker-compose.yml down
```

### Local Development (Backend)

```bash
# Clone the repository
git clone https://github.com/superindexer/indexer.git
cd indexer

# Install Go dependencies
cd backend && go mod tidy

# Run database migrations locally (requires PostgreSQL)
# First ensure PostgreSQL is running:
# docker run -d --name postgres-indexer -e POSTGRES_PASSWORD=indexer -p 5432:5432 postgres:16

# Run migrations
cd backend && go run cmd/migrate/main.go -database "postgres://indexer:indexer@localhost:5432/indexer?sslmode=disable" -path ./internal/database/migrations up

# Start API server
go run cmd/api/main.go
```

### Local Development (Frontend)

```bash
# Install Node dependencies
cd frontend && npm install

# Start development server
npm run dev

# Build for production
npm run build
```

### Configuration

The application uses environment variables for configuration. Create a `.env` file:

```bash
# Database configuration  
DATABASE_DSN=postgres://indexer:indexer@localhost:5432/indexer?sslmode=disable
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
DB_CONN_MAX_LIFETIME=5m
DB_CONN_MAX_IDLE_TIME=2m

# Server configuration
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
SERVER_READ_TIMEOUT=15s
SERVER_WRITE_TIMEOUT=15s

# Authentication configuration
JWT_SECRET=change-me-in-production
JWT_EXPIRY=24h
SESSION_SECRET=change-me-in-production
SESSION_EXPIRY=7d

# NNTP configuration
NNTP_DEFAULT_MAX_CONNECTIONS=10
NNTP_DEFAULT_TIMEOUT=30s
NNTP_DEFAULT_RETRY_ATTEMPTS=3
NNTP_DEFAULT_RETRY_DELAY=2s

# Spotnet configuration  
SPOTNET_DEFAULT_TIMEOUT=30s
SPOTNET_DEFAULT_RETRY_ATTEMPTS=3
SPOTNET_DEFAULT_RETRY_DELAY=2s

# Observability configuration
PROMETHEUS_ENABLED=true
OTEL_ENABLED=false
LOG_LEVEL=info
LOG_FORMAT=json
```

## Project Structure

```
.
├── backend/                 # Go backend services
│   ├── cmd/                 # Application entry points
│   │   ├── api/             # API server
│   │   ├── worker/          # Background worker  
│   │   └── migrate/         # Database migrations
│   ├── internal/            # Internal packages
│   │   ├── config/          # Configuration management
│   │   ├── database/        # Database integration and migrations
│   │   ├── http/            # HTTP API and handlers
│   │   └── auth/            # Authentication services
│   └── go.mod               # Go module dependencies
├── frontend/                # React frontend application
│   ├── src/                 # Source code
│   └── package.json         # Frontend dependencies
├── deployments/             # Deployment configurations  
│   ├── docker/              # Docker Compose setup
│   └── kubernetes/          # Kubernetes manifests
└── docs/                    # Documentation
```

## Development Setup

### Backend Development

1. **Database Setup**:
```bash
# Start PostgreSQL locally (or use Docker)
docker run -d --name postgres-indexer -e POSTGRES_PASSWORD=indexer -p 5432:5432 postgres:16
```

2. **Run Migrations**:
```bash
# Run database migrations  
cd backend && go run cmd/migrate/main.go -database "postgres://indexer:indexer@localhost:5432/indexer?sslmode=disable" -path ./internal/database/migrations up
```

3. **Start API Server**:
```bash
cd backend && go run cmd/api/main.go
```

### Frontend Development

1. **Install Dependencies**:
```bash
cd frontend && npm install
```

2. **Start Development Server**:
```bash
cd frontend && npm run dev
```

### Testing

1. **Run Backend Tests**:
```bash
cd backend && go test ./...
```

2. **Run Frontend Tests**:
```bash
cd frontend && npm run test
```

## API Endpoints

### Authentication (`/api/v1/auth`)

- `POST /login` - Authenticate and return JWT token
- `POST /register` - Register new user account  
- `POST /logout` - Logout (client-side token removal)
- `POST /password/reset` - Request password reset
- `POST /password/reset/confirm` - Confirm password reset

### Users (`/api/v1/users`)

- `GET /` - List all users (admin only)
- `GET /:id` - Get user by ID (admin only)  
- `POST /` - Create new user (admin only)
- `PUT /:id` - Update user (admin only)
- `DELETE /:id` - Delete user (admin only)
- `POST /:id/toggle-active` - Toggle user active status (admin only)

### Search (`/api/v1/search`)

- `GET /` - Search articles and posts with filters
- `GET /categories` - Get search categories

### NNTP Servers (`/api/v1/servers`)

- `GET /` - List all configured NNTP servers
- `GET /:id` - Get server by ID
- `POST /` - Create new NNTP server  
- `PUT /:id` - Update server configuration
- `DELETE /:id` - Delete server

### Newsgroups (`/api/v1/groups`)

- `GET /` - List all newsgroups
- `GET /:id` - Get newsgroup by ID
- `POST /` - Create new newsgroup (admin only)
- `PUT /:id` - Update newsgroup (admin only)
- `DELETE /:id` - Delete newsgroup (admin only)

### Spotnet Sources (`/api/v1/spotnet/sources`)

- `GET /` - List all configured Spotnet sources
- `GET /:id` - Get source by ID  
- `POST /` - Create new Spotnet source
- `PUT /:id` - Update source configuration
- `DELETE /:id` - Delete source

### Indexing Jobs (`/api/v1/indexing/jobs`)

- `GET /` - List all indexing jobs
- `GET /:id` - Get job by ID  
- `POST /` - Create new indexing job
- `PUT /:id/retry` - Retry failed job
- `PUT /:id/cancel` - Cancel pending job

## Monitoring and Metrics

The application exposes Prometheus metrics:

- `/metrics` - Prometheus metrics endpoint
- `indexer_api_requests_total` - API request counters  
- `indexer_search_duration_seconds` - Search query duration
- `indexer_database_connections_active` - Active database connections

## Security Considerations

### Authentication and Authorization
- JWT tokens for stateless session management  
- Password hashing with Argon2id
- Role-based access control (admin, user roles)
- Secure session management

### Database Security
- Environment-specific database connection strings
- Proper parameterized queries to prevent SQL injection
- Database schema with proper constraints and validation

### Deployment Security
- HTTPS support (when configured)
- API rate limiting (basic implementation)
- Secure header configuration

## Troubleshooting

### Database Connection Issues
1. **Check PostgreSQL is running**: `docker ps | grep postgres`
2. **Verify connection string**: Ensure DATABASE_DSN is correct
3. **Check database permissions**: User should have proper access rights

### API Server Issues
1. **Port conflicts**: Check if port 8080 is in use
2. **Missing environment variables**: Ensure all required env vars are set
3. **Database migrations**: Run `go run cmd/migrate/main.go` if migrations failed

### Frontend Issues
1. **Node.js version**: Ensure Node.js 20+ is installed
2. **Dependencies**: Run `npm install` in frontend directory
3. **API connectivity**: Ensure backend is running at correct URL

## Contributing

1. Fork the repository
2. Create a feature branch  
3. Make your changes
4. Write tests for new functionality
5. Update documentation as needed
6. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.