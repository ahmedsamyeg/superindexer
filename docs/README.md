# Usenet + Spotnet Indexer Documentation

This documentation provides comprehensive information about the Usenet + Spotnet Indexer project, including architecture, components, deployment, and usage instructions.

## Table of Contents

1. [Project Overview](#project-overview)
2. [Architecture](#architecture)
3. [Components](#components)
4. [API Documentation](#api-documentation)
5. [Deployment](#deployment)
6. [Testing Strategy](#testing-strategy)
7. [Configuration](#configuration)
8. [Troubleshooting](#troubleshooting)

## Project Overview

The Usenet + Spotnet Indexer is a modern, scalable, self-hosted platform for indexing and searching content from Usenet NNTP servers and Spotnet sources.

### Key Features

- **NNTP Server Management**: Connect to and manage multiple Usenet NNTP servers
- **Newsgroup Discovery**: Automatically discover available newsgroups on servers
- **Usenet Article Ingestion**: Retrieve and index article headers from newsgroups  
- **Spotnet Integration**: Connect to and index content from Spotnet sources
- **Full-Text Search**: PostgreSQL-powered search with relevance ranking
- **Authentication & Authorization**: Secure user management with JWT tokens
- **Admin Dashboard**: Comprehensive system monitoring and management
- **Containerized Deployment**: Docker Compose and Kubernetes support

## Architecture

The application follows a modular monolith architecture with:

### Backend Components

- **API Service**: Main HTTP API built with Echo framework
- **Database Layer**: PostgreSQL with Bun ORM and pgx driver  
- **Indexing Engine**: Concurrent processing workers for content ingestion
- **Search Engine**: PostgreSQL native full-text search with GIN indexes
- **Authentication Service**: JWT-based authentication and authorization
- **Monitoring & Observability**: Prometheus metrics and Grafana dashboards

### Frontend Components

- **React Application**: Modern web interface built with Vite and TypeScript
- **UI Components**: shadcn/ui component library for consistent design
- **Search Interface**: Complete search experience with filtering and results display

## Components

### Backend Services

1. **API Service** (`cmd/api`)
   - Main HTTP API server with Echo framework
   - REST endpoints for all application functionality
   - Authentication and authorization middleware

2. **Worker Service** (`cmd/worker`)
   - Background indexing workers for content ingestion
   - Concurrent processing with bounded worker pools
   - Job coordination and scheduling

3. **Migration Service** (`cmd/migrate`)
   - Database schema migrations with golang-migrate
   - Migration scripts for all database changes

### Database Schema

The application uses PostgreSQL with the following core tables:

- `users`: User accounts with roles and authentication
- `nntp_servers`: Configured NNTP server connections  
- `newsgroups`: Usenet newsgroups with article ranges
- `articles`: Indexed Usenet articles with search vectors
- `spotnet_sources`: Configured Spotnet source connections
- `spotnet_posts`: Indexed Spotnet posts with metadata
- `categories`: Content categories for organization
- `indexing_jobs`: Background indexing job tracking

## API Documentation

The API is documented using OpenAPI 3.1 specification and can be accessed at `/api/v1/docs` when running the application.

### Authentication

All protected endpoints require a valid JWT token in the Authorization header:

```
Authorization: Bearer <your-jwt-token>
```

### Endpoints

#### Authentication (`/api/v1/auth`)

- `POST /login` - Authenticate and return JWT token
- `POST /register` - Register new user account  
- `POST /logout` - Logout (client-side token removal)
- `POST /password/reset` - Request password reset
- `POST /password/reset/confirm` - Confirm password reset

#### Users (`/api/v1/users`)

- `GET /` - List all users (admin only)
- `GET /:id` - Get user by ID (admin only)  
- `PUT /:id` - Update user (admin only)
- `DELETE /:id` - Delete user (admin only)
- `POST /:id/toggle-active` - Toggle user active status (admin only)

#### Search (`/api/v1/search`)

- `GET /` - Search articles and posts with filters
- `GET /categories` - Get search categories

#### NNTP Servers (`/api/v1/servers`)

- `GET /` - List all configured NNTP servers
- `GET /:id` - Get server by ID
- `POST /` - Create new NNTP server  
- `PUT /:id` - Update server configuration
- `DELETE /:id` - Delete server

#### Newsgroups (`/api/v1/groups`)

- `GET /` - List all newsgroups
- `GET /:id` - Get newsgroup by ID
- `POST /` - Create new newsgroup (admin only)
- `PUT /:id` - Update newsgroup (admin only)
- `DELETE /:id` - Delete newsgroup (admin only)

#### Spotnet Sources (`/api/v1/spotnet/sources`)

- `GET /` - List all configured Spotnet sources
- `GET /:id` - Get source by ID  
- `POST /` - Create new Spotnet source
- `PUT /:id` - Update source configuration
- `DELETE /:id` - Delete source

#### Indexing Jobs (`/api/v1/indexing/jobs`)

- `GET /` - List all indexing jobs
- `GET /:id` - Get job by ID  
- `POST /` - Create new indexing job
- `PUT /:id/retry` - Retry failed job
- `PUT /:id/cancel` - Cancel pending job

## Deployment

### Docker Compose Deployment

The application can be deployed using Docker Compose:

```bash
# Copy environment file
cp .env.example .env

# Start all services  
docker compose -f deployments/docker/docker-compose.yml up -d

# Run database migrations
docker compose -f deployments/docker/docker-compose.yml run --rm migrate up

# View logs
docker compose -f deployments/docker/docker-compose.yml logs -f api
```

### Kubernetes Deployment

For production deployment, the application can be deployed using Kubernetes:

```bash
# Apply configuration to cluster
kubectl apply -k deployments/kubernetes/overlays/production

# Or use the provided deployment scripts
./deployments/kubernetes/deploy.sh production
```

## Testing Strategy

The project follows a comprehensive testing strategy with multiple layers:

### Backend Unit Tests

Tests for individual components and services:

- Authentication service (password hashing, JWT generation)
- User management service  
- Search query building and execution
- Database repository operations
- Indexing job processing logic

### Integration Tests

Tests for end-to-end API functionality:

- Authentication and authorization flows
- User CRUD operations with proper permissions
- Search functionality with various filters  
- NNTP and Spotnet source management
- Error handling and validation

### Database Integration Tests

Tests for database operations and migrations:

- Database schema migrations
- Bun repository operations (Create, Read, Update, Delete)
- pgx COPY ingestion performance tests
- Bulk inserts and duplicate detection  
- Full-text search functionality
- Search ranking and filtering

### Frontend Tests

Tests for React components and user interface:

- Component rendering and state management
- User interaction flows (search, login, admin actions)
- Responsive design and mobile compatibility
- Form validation and error handling

## Configuration

The application uses environment variables for configuration:

### Database Configuration

```
DATABASE_DSN=postgres://indexer:indexer@localhost:5432/indexer?sslmode=disable
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5  
DB_CONN_MAX_LIFETIME=5m
DB_CONN_MAX_IDLE_TIME=2m
```

### Authentication Configuration

```
JWT_SECRET=change-me-in-production
JWT_EXPIRY=24h
SESSION_SECRET=change-me-in-production  
SESSION_EXPIRY=7d
```

### Server Configuration

```
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
SERVER_READ_TIMEOUT=15s
SERVER_WRITE_TIMEOUT=15s
SERVER_SHUTDOWN_TIMEOUT=30s
```

### NNTP Configuration

```
NNTP_DEFAULT_MAX_CONNECTIONS=10
NNTP_DEFAULT_TIMEOUT=30s
NNTP_DEFAULT_RETRY_ATTEMPTS=3
NNTP_DEFAULT_RETRY_DELAY=2s
```

### Spotnet Configuration

```
SPOTNET_DEFAULT_TIMEOUT=30s
SPOTNET_DEFAULT_RETRY_ATTEMPTS=3
SPOTNET_DEFAULT_RETRY_DELAY=2s
```

## Troubleshooting

### Common Issues

1. **Database Connection Failures**
   - Check that PostgreSQL is running and accessible
   - Verify DATABASE_DSN environment variable format
   - Ensure proper SSL configuration (disable for local dev)

2. **Authentication Issues**
   - Verify JWT_SECRET is properly configured
   - Check that tokens are being sent with correct format
   - Confirm expiration times are reasonable

3. **Search Performance Issues**
   - Ensure GIN indexes are properly created on search fields
   - Check PostgreSQL configuration for optimal search performance
   - Monitor database connection pool usage

### Logging and Monitoring

The application uses structured logging with:

- Log level configuration (debug, info, warn, error)
- JSON formatted logs for easy parsing
- Context-aware logging with request IDs
- Health check endpoints (`/health`, `/ready`)

### Backup and Recovery

The application supports:

- Automated PostgreSQL backups with retention policies
- Point-in-time recovery capabilities  
- Database restore procedures documented in deployment guide
- Configuration backup and management

## Contributing

Contributions are welcome! Please follow these steps:

1. Fork the repository
2. Create a feature branch  
3. Make your changes
4. Write tests for new functionality
5. Update documentation as needed
6. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.