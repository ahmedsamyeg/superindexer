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
- PostgreSQL-powered full-text search with relevance ranking
- Advanced search filters and faceted results
- Background indexing workers with concurrent processing
- User authentication and role-based access control
- Administrative dashboard
- Docker and Kubernetes deployment support
- Prometheus metrics and Grafana dashboards

## Quick Start

### Prerequisites

- Docker and Docker Compose
- Go 1.22+ (for local development)
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

**Services:**
- API: http://localhost:8080
- Frontend: http://localhost:3000
- Prometheus: http://localhost:9090
- Grafana: http://localhost:3001 (admin/admin)

### Local Development

#### Backend

```bash
cd backend

# Install dependencies
go mod tidy

# Run migrations
go run cmd/migrate/main.go up

# Start API server
go run cmd/api/main.go

# Start worker
go run cmd/worker/main.go
```

#### Frontend

```bash
cd frontend

# Install dependencies
npm install

# Start development server
npm run dev
```

## Project Structure

```
.
├── backend/
│   ├── cmd/
│   │   ├── api/           # API service entry point
│   │   ├── worker/        # Worker service entry point
│   │   └── migrate/       # Migration tool entry point
│   ├── internal/
│   │   ├── config/        # Configuration
│   │   ├── database/
│   │   │   ├── migrations/ # SQL migrations
│   │   │   ├── models/    # Bun ORM models
│   │   │   └── repositories/ # Data access layer
│   │   ├── http/
│   │   │   ├── handlers/  # HTTP handlers
│   │   │   └── middleware/ # HTTP middleware
│   │   ├── auth/          # Authentication
│   │   ├── usenet/        # NNTP integration
│   │   ├── spotnet/       # Spotnet integration
│   │   ├── indexing/      # Indexing engine
│   │   ├── search/        # Search service
│   │   └── observability/ # Metrics and tracing
│   └── pkg/               # Shared packages
├── frontend/
│   ├── src/
│   │   ├── components/    # React components
│   │   ├── pages/         # Page components
│   │   ├── hooks/         # Custom React hooks
│   │   ├── lib/           # Utilities
│   │   └── services/      # API client
│   └── nginx/             # Production nginx config
├── deployments/
│   ├── docker/
│   │   ├── docker-compose.yml
│   │   ├── prometheus.yml
│   │   └── grafana/
│   └── kubernetes/
│       ├── base/
│       └── overlays/
└── .github/
    └── workflows/
        └── ci.yml
```

## API Endpoints

### Authentication
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/register` - User registration
- `POST /api/v1/auth/logout` - User logout
- `GET /api/v1/auth/me` - Get current user

### Search
- `GET /api/v1/search?q=term` - Full-text search
- `GET /api/v1/search/articles?q=term` - Search articles
- `GET /api/v1/search/spotnet?q=term` - Search Spotnet posts
- `GET /api/v1/search/facets?q=term` - Get search facets

### NNTP Servers
- `GET /api/v1/servers` - List servers
- `POST /api/v1/servers` - Create server
- `PUT /api/v1/servers/:id` - Update server
- `DELETE /api/v1/servers/:id` - Delete server
- `POST /api/v1/servers/:id/test` - Test connection

### Newsgroups
- `GET /api/v1/groups` - List newsgroups
- `GET /api/v1/groups/:id` - Get newsgroup
- `POST /api/v1/groups` - Create newsgroup
- `PUT /api/v1/groups/:id` - Update newsgroup
- `DELETE /api/v1/groups/:id` - Delete newsgroup
- `GET /api/v1/groups/:id/discover` - Discover groups
- `POST /api/v1/groups/:id/index` - Manual index

### Categories
- `GET /api/v1/categories` - List categories
- `POST /api/v1/categories` - Create category
- `PUT /api/v1/categories/:id` - Update category
- `DELETE /api/v1/categories/:id` - Delete category

### Indexing Jobs
- `GET /api/v1/indexing/jobs` - List jobs
- `POST /api/v1/indexing/jobs` - Create job
- `POST /api/v1/indexing/jobs/:id/pause` - Pause job
- `POST /api/v1/indexing/jobs/:id/resume` - Resume job
- `POST /api/v1/indexing/jobs/:id/retry` - Retry job
- `DELETE /api/v1/indexing/jobs/:id` - Cancel job

### Admin
- `GET /api/v1/admin/dashboard` - Dashboard stats
- `GET /api/v1/admin/stats` - System statistics
- `GET /api/v1/admin/system` - System info
- `GET /api/v1/admin/db/stats` - Database stats

## Database Schema

The application uses PostgreSQL with the following core tables:

- `users` - User accounts
- `nntp_servers` - NNTP server configurations
- `newsgroups` - Usenet newsgroups
- `articles` - Indexed Usenet articles
- `spotnet_sources` - Spotnet source configurations
- `spotnet_posts` - Indexed Spotnet posts
- `categories` - Content categories
- `indexing_jobs` - Background indexing jobs
- `indexing_checkpoints` - Processing checkpoints
- `user_search_history` - User search history
- `bookmarks` - User bookmarks
- `system_settings` - System configuration

## Search

The search engine uses PostgreSQL native Full-Text Search with:

- `tsvector` / `tsquery` for text search
- `websearch_to_tsquery` for natural language queries
- `ts_rank_cd` for relevance ranking
- GIN indexes for fast searching
- `unaccent` extension for accent-insensitive search

### Search Weights

| Field | Weight |
|-------|--------|
| Title | A (highest) |
| Category | B |
| Description | C |
| Additional metadata | D (lowest) |

## Kubernetes Deployment

```bash
# Development
kubectl apply -k deployments/kubernetes/overlays/development

# Production
kubectl apply -k deployments/kubernetes/overlays/production
```

## Development Phases

| Phase | Status | Description |
|-------|--------|-------------|
| 0 | In Progress | Architecture and Project Foundation |
| 1 | Pending | Database Foundation |
| 2 | Pending | Backend API Foundation |
| 3 | Pending | Authentication and Authorization |
| 4 | Pending | Usenet Integration |
| 5 | Pending | Spotnet Integration |
| 6 | Pending | Metadata and Indexing Engine |
| 7 | Pending | PostgreSQL Search Engine |
| 8 | Pending | Frontend Foundation |
| 9 | Pending | Search and Discovery Interface |
| 10 | Pending | Administration Interface |
| 11 | Pending | Docker and Kubernetes |
| 12 | Pending | Observability and Performance |
| 13 | Pending | Security, Hardening and Release |

## License

MIT
