# Final Phase Completion Status - Honest Assessment

## Overview

This document provides an honest assessment of the actual implementation progress for Phases 1, 2, and 3 of the Usenet + Spotnet Indexer project.

---

## Phase 1 - Database Foundation

### ✅ What Has Been Properly Implemented:

- **Database Connection Pooling** - pgxpool integration with proper configuration parameters (MaxOpenConns, MaxIdleConns, ConnMaxLifetime, ConnMaxIdleTime)
- **Bun ORM Integration** - Complete integration with Bun ORM using pgx driver for PostgreSQL operations  
- **Database Migration System** - goose-style migration framework with up/down scripts
  - `000001_init.up.sql` - Complete database schema (12 tables, constraints, relationships)
  - `000002_add_indexes.up.sql` - Performance indexes including GIN indexes for FTS
- **Complete Database Schema** - All tables with proper constraints and relationships
  - `users`, `nntp_servers`, `newsgroups`, `articles`, `spotnet_posts`, `spotnet_sources`, `indexing_jobs`, `categories`, `bookmarks`, `user_search_history`, `system_settings`, `indexing_checkpoints`
- **Database Integration Tests** - Comprehensive test suite (1008 lines) in `integration/integration_test.go`:
  - ✅ Database connection/pool verification
  - ✅ Migration execution and table verification
  - ✅ FTS extensions (unaccent, pg_trgm) verification
  - ✅ GIN indexes verification
  - ✅ User repository CRUD (Create, GetByID, GetByUsername, GetByEmail, Update, Count, List, Delete, UpdatePasswordHash)
  - ✅ Transaction commit testing
  - ✅ Transaction rollback testing
  - ✅ Newsgroup repository (Create, GetByID, GetByName, GetAll, GetEnabled, UpdateArticleRange, UpdateLastIndexed, Count, CountEnabled, Delete)
  - ✅ Article repository (Create, GetByID, GetByMessageID, ExistsByMessageID, Count, CountByNewsgroup, Update, ListByNewsgroup, Delete)
  - ✅ Article bulk insert with idempotency testing (UPSERT)
  - ✅ Search repository (FTS search, author filter, size filter, facets)
  - ✅ Category repository (parent/child hierarchy, GetByID, GetBySlug, GetByName, GetAll, GetChildren, Delete)
  - ✅ Indexing job repository (Create, GetByID, GetPendingJobs, UpdateStatus, UpdateProgress, UpdateError, Complete, GetByStatus, CountByStatus, CountActive, Delete)
- **Advanced PostgreSQL Features** - FTS with unaccent/pg_trgm, GIN indexes, JSONB columns for metadata

### ✅ What Was Previously Claimed Missing (Now Complete):

❌ → ✅ **Database Integration Tests** - Was listed as missing, now has comprehensive 1008-line test suite
❌ → ✅ **Real Database Validation** - All repositories tested against real PostgreSQL

### What Is Still Missing:

- **CI-friendly test execution** - Tests currently require Docker container; no testcontainers approach
- **Migration validation** - No automated check that migration up/down scripts are reversible
- **Connection pool monitoring** - No pool stats endpoint, no idle connection leak detection
- **Advanced PostgreSQL features** - Partitioning strategy for articles table, materialized views, backup/restore automation

---

## Phase 2 - Backend API Foundation

### ✅ What Has Been Properly Implemented:

- **Echo HTTP Framework Setup** - Complete server initialization with proper configuration
- **API Route Structure** - Complete endpoint routing with versioning support (v1), 15+ resource groups
- **Middleware Integration** - All required middleware configured:
  - Request ID generation and propagation
  - Recovery middleware
  - Logger with health endpoint skip
  - CORS with proper headers
  - Timeout middleware (30s)
- **Functional API Endpoints** - All endpoints are real implementations backed by the database:
  - **Auth**: login, register, logout, get profile (/me), password reset flow
  - **Users**: full CRUD + toggle active status (admin only)
  - **Search**: query with filters (category, newsgroup, author, date range, size), pagination, sorting
  - **NNTP Servers**: full CRUD
  - **Newsgroups**: full CRUD
  - **Spotnet Sources**: full CRUD
  - **Indexing Jobs**: list, get, create, retry, cancel
  - **Admin**: dashboard stats, system stats, system info, DB stats
  - **Bookmarks**: list, create, delete
  - **Categories**: full CRUD with parent/child hierarchy
  - **Settings**: list, get, update (create-if-not-exists)
  - **Search History**: list, create, clear
- **Complete Error Handling**:
  - Structured error types in `pkg/errors` (AppError with code/message/details)
  - Proper HTTP status codes throughout (400, 401, 403, 404, 500)
  - Validation error formatting with field-level details
  - Consistent error response format
- **Real API Integration Tests** - Comprehensive test suite:
  - `integration_test.go` - Health checks, API validation, auth with mocks
  - `auth_test.go` - Login, register, logout, password reset, token validation
  - `test_helpers.go` - Shared validation helpers
  - `unit_test.go` - Password hashing, JWT generation, token claims
- **Health Check with Real DB Verification** - `DB.Ping()` called, returns 503 if DB is down
- **Readiness Check with DB Verification** - Includes database connectivity status, checks map, and details array
- **Auth Service Interfaces** - `AuthServiceInterface` and `UserServiceInterface` for testability
- **GetMe Endpoint Queries DB** - No longer returns placeholder data from JWT context
- **Admin Routes Wired Up** - system-info and db-stats endpoints registered

### ✅ What Was Previously Claimed Missing (Now Complete):

❌ → ✅ **Functional API Endpoints** - Was listed as "most are placeholders", now all are real DB-backed implementations
❌ → ✅ **Complete Error Handling** - Was listed as missing, now has structured errors and proper HTTP status codes
❌ → ✅ **Real API Integration Tests** - Was listed as missing, now has comprehensive test suite with mocks

### What Is Still Missing:

- **Production readiness** - Connection pool monitoring, schema drift detection, backup/restore

---

## Phase 3 - Authentication and Authorization

### ✅ What Has Been Properly Implemented:

- **JWT Token Framework** - Complete JWT generation and validation with proper claims structure
  - `GenerateToken()` - Creates signed JWT with UserID, Role, Username, Issuer, Expiry
  - `ValidateToken()` - Parses, validates signature, checks expiry
  - `TokenClaims` struct with jwt.RegisteredClaims
  - `ExtractTokenFromHeader()` - Bearer token extraction
  - `ExtractUserFromContext()` - User info extraction from request context
- **Password Security** - Argon2id password hashing with secure configuration
  - `GeneratePasswordHash()` - Random salt, base64-encoded output (salt||hash format)
  - `VerifyPassword()` - Constant-time comparison, proper salt extraction
  - `DefaultPasswordConfig` - 64MB memory, 3 iterations, 4 parallelism, 16-byte salt, 32-byte key
- **Authentication Middleware** - Complete middleware with proper context extraction and validation
  - `AuthMiddleware()` - Extracts Bearer token, validates JWT, stores user info in context
  - `AdminMiddleware()` - Checks user role is "admin"
  - Context keys: `user_id`, `user_role`, `username`
- **Functional Authentication Service** - Complete implementation:
  - `Login()` - Username lookup, password verification, token generation
  - `Register()` - Duplicate username/email checks, password hashing, user creation
  - `ValidateToken()` - JWT validation with proper config
  - `RequestPasswordReset()` - Secure token generation, DB storage with expiry, constant-time email response
  - `ConfirmPasswordReset()` - Token validation, expiry check, password update, token cleanup
- **Complete Password Reset Flow** - Token generation, storage, validation, password update
- **Real Role-Based Access Control** - AdminMiddleware enforces admin role on protected endpoints
  - Users CRUD endpoints require admin role
  - NNTP servers CRUD requires admin role
  - Newsgroups CRUD requires admin role
  - Spotnet sources CRUD requires admin role
  - Indexing jobs CRUD requires admin role
  - Admin dashboard/stats require admin role
  - Settings CRUD requires admin role
  - Categories CRUD requires admin role

### ✅ What Was Previously Claimed Missing (Now Complete):

❌ → ✅ **Functional Authentication Service** - Was listed as "only templates", now complete with login, register, password reset
❌ → ✅ **Complete Password Reset Flow** - Was listed as "not fully implemented", now complete with token storage and validation
❌ → ✅ **Real Role-Based Access Control** - Was listed as "not fully enforced", now enforced on all admin endpoints

---

## Summary: What's Actually Done vs. What Was Claimed

| Phase | Claimed Missing | Actual Status |
|-------|----------------|---------------|
| Phase 1 | Database integration tests | ✅ Complete (1008-line test suite) |
| Phase 1 | Real database validation | ✅ Complete (all repos tested) |
| Phase 2 | Functional API endpoints | ✅ Complete (all endpoints DB-backed) |
| Phase 2 | Complete error handling | ✅ Complete (structured errors) |
| Phase 2 | Real API integration tests | ✅ Complete (mock-based suite) |
| Phase 3 | Functional auth service | ✅ Complete (login, register, reset) |
| Phase 3 | Complete password reset | ✅ Complete (token + storage) |
| Phase 3 | Real RBAC | ✅ Complete (enforced on all admin routes) |
| Phase 4 | Indexing engine skeleton | ✅ Complete (NNTP + Spotnet + workers) |
| Phase 5 | Search engine skeleton | ✅ Complete (FTS + ranking + facets) |
| Phase 6 | Frontend integration | ✅ Complete (React dashboard + auth flow) |

---

## Phase 4 - Indexing Engine

### ✅ What Has Been Properly Implemented:

- **NNTP Client** - Full multi-line response handling with proper NNTP protocol support
  - `List()` - LIST command to get newsgroup names with article ranges
  - `ListActive()` - LIST ACTIVE command with descriptions
  - `Xover()` - XOVER for article overviews (Subject, From, Date, Message-ID, Lines, Size)
  - `ArticleByNumber()` / `ArticleByMessageID()` - Full article body fetching
  - `Head()` - Headers only fetching
  - `Stat()` - Message ID lookup by article number
  - `Authenticate()` - AUTHINFO USER/PASS support
  - `Disconnect()` - Safe to call multiple times

- **NNTP Connection Pooling** - Per-server connection pools with automatic reconnection
  - `ConnectionManager` manages pools per NNTP server
  - Automatic connection validation and recreation
  - Configurable max connections per server

- **NNTP Newsgroup Discovery** - Full newsgroup management
  - `DiscoverNewsgroups()` - Lists all newsgroups on a server
  - `GetNewsgroupInfo()` - GROUP command for article range
  - `GetNewsgroupActiveInfo()` - LIST ACTIVE with descriptions
  - `GetNewsgroupArticleRange()` - STAT command for specific article

- **Real Spotnet API Client** - Multi-source type support
  - **API source** - REST API with JSON parsing, configurable json_path
  - **RSS source** - RSS 2.0 and Atom feed parsing via `encoding/xml`
  - **Web source** - HTML scraping with JSON-LD extraction
  - **Incremental sync** - `sinceID`/`sinceDate` filtering for new posts only
  - **Retry logic** - Exponential backoff (1s → 2s → 4s), 3 attempts
  - **Auth support** - Bearer token + custom headers from source config

- **Spotnet Post Parser** - Multi-format parsing and normalization
  - `ParseSpotnetPost()` - Converts raw data to standardized format
  - `NormalizeMetadata()` - Title, description, author, category normalization
  - `createSearchVector()` - PostgreSQL FTS search vector generation
  - `ValidatePost()` - Required field validation

- **Spotnet Metadata Processor** - Post metadata processing
  - Title normalization (whitespace, extra spaces)
  - Description normalization (preserve line breaks)
  - Author normalization (trim, clean)
  - Category normalization (lowercase, title case)
  - Date/time parsing (RFC3339, custom formats)
  - Search vector creation for PostgreSQL FTS

- **Indexing Coordinator** - Real job coordination with background processing
  - Priority queue (higher priority processed first)
  - Background goroutine polls every second for pending jobs
  - Graceful shutdown via context cancellation
  - `SubmitJob()` - Creates DB record, adds to priority queue
  - `GetJobStatus()` / `ListJobs()` - Real DB queries with filtering
  - `CancelJob()` / `RetryJob()` - State management
  - `StartProcessing()` / `StopProcessing()` - Worker pool lifecycle

- **Indexing Service** - Complete service layer
  - `CreateIndexingJob()` - Validates job type/source/newsgroup, submits to coordinator
  - `StartIndexingEngine()` / `StopIndexingEngine()` - Engine lifecycle
  - `ListIndexingJobs()` / `GetIndexingJob()` - Real DB queries
  - `GetIndexingStats()` - Real stats (count by status, active jobs, sources, newsgroups, articles, posts)
  - `CancelIndexingJob()` / `RetryIndexingJob()` - State management
  - `ConfigureWorkerPool()` - Dynamic worker count configuration

- **Indexing Workers** - Real Spotnet/Usenet processing
  - `ProcessSpotnetJob()` - Full implementation:
    1. Gets source from DB, validates enabled state
    2. Updates job to "running"
    3. Calls `SpotnetClient.GetPosts()` with incremental sync
    4. Deduplicates via `ExistsByExternalID`
    5. Parses via `PostParser` or fallback
    6. Processes metadata via `MetadataProcessor`
    7. Bulk inserts via `PostRepo.CreateBulk`
    8. Updates source `last_sync_at`
    9. Updates job progress/completion
  - `ProcessUsenetJob()` - Full implementation:
    1. Gets newsgroup from DB, validates enabled state
    2. Gets NNTP servers, updates job to "running"
    3. Fetches articles via NNTP (since last indexed)
    4. Deduplicates via `ExistsByMessageID`
    5. Bulk inserts via `ArticleRepo.CreateBulk`
    6. Updates newsgroup `last_indexed_at`
    7. Updates job progress/completion
  - Retry logic with exponential backoff for both job types (up to 3 retries)
  - Per-job timeout context (10 min)

- **Database Migrations** - New tables for auth infrastructure
  - `000003_add_auth_tables.up.sql` - token_blacklist, refresh_tokens tables
  - Proper indexes for fast lookups
  - Reversible down migration

### ✅ What Was Previously Claimed Missing (Now Complete):

❌ → ✅ **NNTP client** - Was listed as skeleton, now complete with full protocol support
❌ → ✅ **Spotnet API client** - Was listed as mock data, now real HTTP-based with retry
❌ → ✅ **Indexing coordinator** - Was listed as placeholder, now real priority queue + background processing
❌ → ✅ **Indexing workers** - Was listed as stubs, now real Spotnet/Usenet processing with deduplication
❌ → ✅ **Worker pool** - Was listed as skeleton, now concurrent job processing
❌ → ✅ **Retry/failure handling** - Exponential backoff for all job types
❌ → ✅ **Incremental sync** - sinceID/sinceDate filtering for both NNTP and Spotnet

---

## Phase 5 - Search Engine

### ✅ What Has Been Properly Implemented:

- **Unified Search Results** - `SearchResultItem` type that works for both articles and spotnet posts
  - `ID` - Common ID field
  - `ExternalID` - External identifier (for spotnet)
  - `Source` - "article" or "spotnet"
  - `Subject` / `Title` - Article subject or post title
  - `Description` - Post description
  - `Author` - Common author field
  - `Newsgroup` - Newsgroup name (for articles)
  - `Category` - Category name
  - `PostedAt` - Publication date
  - `Size` - Article size or post size
  - `MessageID` - Article message ID
  - `SourceName` - Source name (for spotnet)
  - `Rank` - Relevance rank
  - `RawMetadata` - Raw metadata JSON
  - `NormalizedMetadata` - Normalized metadata JSON

- **Combined Search** - `Search()` combines articles + spotnet with relevance ranking
  - Searches both article and spotnet tables independently
  - Combines results with proper ranking
  - Sorts by rank (relevance) or by field (date, author, etc.)
  - Applies pagination across combined results
  - Returns unified SearchResultItem array

- **Relevance Ranking** - `ts_rank_cd` for article ranking
  - PostgreSQL `ts_rank_cd` function for article relevance
  - Combined results sorted by rank when sorting by relevance
  - Fallback to field-based sorting when requested

- **Field-based Sorting** - Sort by multiple fields
  - `subject` / `title` - Alphabetical sorting
  - `author` - Alphabetical sorting
  - `newsgroup` - Alphabetical sorting
  - `category` - Alphabetical sorting
  - `posted_at` / `date` - Chronological sorting
  - `size` - Numerical sorting
  - Both ASC and DESC order supported

- **Faceted Search** - Source type facets, date range, newsgroup facets
  - Source type facets (article/spotnet) with counts
  - Date range facets (min/max Unix timestamps)
  - Newsgroup facets with article counts
  - Category facets from normalized metadata
  - Author facets from article data

- **Search Suggestions** - Author suggestions based on partial query
  - `GetSearchSuggestion()` returns distinct authors matching partial query
  - ILIKE matching for case-insensitive search
  - Configurable limit (max 20)
  - Ordered alphabetically

- **Search History** - Async save of search queries to user history
  - `saveSearchHistory()` saves to user_search_history table
  - Runs in goroutine to avoid blocking response
  - Only for authenticated users

- **Search Endpoints** - Multiple search endpoints
  - `GET /api/v1/search` - Combined search (articles + spotnet)
  - `GET /api/v1/search/articles` - Article-only search
  - `GET /api/v1/search/spotnet` - Spotnet-only search
  - `GET /api/v1/search/facets` - Facet counts
  - `GET /api/v1/search/suggestions` - Author suggestions
  - `GET /api/v1/search/categories` - Category list

- **Empty Query Handling** - Returns empty results gracefully
  - No database queries executed for empty queries
  - Returns empty data array with zero counts
  - Returns empty facets for empty queries

- **Pagination** - Consistent pagination across all search endpoints
  - Default: 20 results per page
  - Maximum: 100 results per page
  - Minimum: 1 page
  - Returns total, page, per_page, total_pages in response

- **PostgreSQL FTS** - `websearch_to_tsquery` with `simple` lexicon
  - Natural language search support
  - Boolean operators (AND, OR, NOT) supported
  - Phrase search with quotes supported
  - Simple lexicon for consistent behavior across locales

- **Filter Support** - Category, newsgroup, author, size range, date range filters
  - `category` - Filter by category slug
  - `newsgroup` - Filter by newsgroup name
  - `author` - Filter by author (ILIKE partial match)
  - `min_size` / `max_size` - Filter by article size
  - `min_date` / `max_date` - Filter by date range (RFC3339)

- **Search Vector Generation** - `to_tsvector('simple', unaccent())` for indexing
  - `CreateSearchVector()` generates tsvector expression
  - Uses `unaccent()` for accent-insensitive search
  - Simple lexicon for consistent behavior

- **Search Query Generation** - `websearch_to_tsquery('simple', ...)` for queries
  - `CreateSearchQuery()` generates tsquery expression
  - Handles natural language search syntax
  - Escapes single quotes for SQL safety

### ✅ What Was Previously Claimed Missing (Now Complete):

❌ → ✅ **Full-text search** - Was listed as skeleton, now complete with FTS + ranking
❌ → ✅ **Faceting** - Was listed as missing, now complete with source/date/newsgroup facets
❌ → ✅ **Ranking** - Was listed as missing, now complete with ts_rank_cd + combined ranking
❌ → ✅ **Result formatting** - Was listed as missing, now complete with unified SearchResultItem

---

## Remaining Work

### Phase 3 Completed (new):
- ✅ **Email service integration** - Full SMTP email service with HTML templates for password reset emails
- ✅ **Refresh token flow** - Database-backed refresh tokens with automatic rotation (use-once)
- ✅ **Token revocation** - Token blacklist table with cleanup, logout now actually revokes tokens
- ✅ **Rate limiting** - In-memory rate limiter on auth endpoints (5 req/min by default)
- ✅ **Email config** - Email settings loaded from environment variables

### Phase 4 - Indexing Engine (new):
- ✅ **NNTP client** - Full multi-line response handling, LIST/XOVER/ARTICLE/HEAD/STAT commands
- ✅ **NNTP newsgroup discovery** - LIST, LIST ACTIVE, GROUP commands with parsing
- ✅ **NNTP connection pooling** - Connection manager with per-server pools
- ✅ **Real Spotnet API client** - API/RSS/Web source types with HTTP-based fetching
- ✅ **Spotnet incremental sync** - sinceID/sinceDate filtering for new posts only
- ✅ **Spotnet retry logic** - Exponential backoff (1s → 2s → 4s), 3 attempts
- ✅ **Spotnet post parser** - JSON/XML parsing with metadata normalization
- ✅ **Indexing coordinator** - Priority queue, background goroutine processing, graceful shutdown
- ✅ **Indexing service** - Real implementations for job CRUD, engine lifecycle, stats
- ✅ **Indexing workers** - Real Spotnet/Usenet job processing with deduplication
- ✅ **Worker pool** - Concurrent job processing with configurable worker count
- ✅ **Retry/failure handling** - Exponential backoff for both job types
- ✅ **Database migrations** - token_blacklist, refresh_tokens tables

### Phase 5 - Search Engine (new):
- ✅ **Unified search results** - `SearchResultItem` type that works for both articles and spotnet posts
- ✅ **Combined search** - `Search()` combines articles + spotnet with relevance ranking
- ✅ **Relevance ranking** - `ts_rank_cd` for article ranking, combined results sorted by rank
- ✅ **Field-based sorting** - Sort by subject, title, author, newsgroup, category, date, size
- ✅ **Faceted search** - Source type facets (article/spotnet), date range facets, newsgroup facets
- ✅ **Search suggestions** - Author suggestions based on partial query
- ✅ **Search history** - Async save of search queries to user history
- ✅ **Search endpoints** - `/search`, `/search/articles`, `/search/spotnet`, `/search/facets`, `/search/suggestions`
- ✅ **Empty query handling** - Returns empty results gracefully
- ✅ **Pagination** - Consistent pagination across all search endpoints (max 100 per page)
- ✅ **PostgreSQL FTS** - `websearch_to_tsquery` with `simple` lexicon for natural language search
- ✅ **Filter support** - Category, newsgroup, author, size range, date range filters
- ✅ **Search vector generation** - `to_tsvector('simple', unaccent())` for indexing
- ✅ **Search query generation** - `websearch_to_tsquery('simple', ...)` for queries

### Phase 6 - Frontend Integration (new):
- ✅ **Vite + React 19 + TypeScript** project with Tailwind CSS
- ✅ **React Query** for data fetching, caching, and state management
- ✅ **React Router** for client-side routing
- ✅ **Auth Context/Provider** - Full auth state management with React context
  - `AuthProvider` component wrapping the app
  - `useAuth` hook for accessing auth state
  - `login()`, `register()`, `logout()`, `refreshUser()` methods
  - Token storage in localStorage with expiry checking
- ✅ **Token storage** - `authStorage` with access/refresh token management
  - `setTokens()`, `clearTokens()`, `isAuthenticated()`, `getUser()`
  - Automatic expiry checking
- ✅ **Token refresh** - Automatic refresh on 401 responses
  - `apiRequest` wrapper handles 401 → refresh → retry flow
  - Refresh token stored separately
- ✅ **Login page** - Full login form with show/hide password, remember me, forgot password link
- ✅ **Register page** - Full registration with password validation, confirm password
- ✅ **Forgot password page** - Email input with success state
- ✅ **Reset password page** - Token-based password reset with validation
- ✅ **Profile page** - User profile display with stats, account info
- ✅ **Settings page** - Profile, notifications, appearance tabs
- ✅ **Bookmarks page** - Local bookmarks with delete functionality
- ✅ **Search page wired to API** - Real search with React Query
  - `searchApi.search()` with query params
  - `searchApi.getFacets()` for filter counts
  - `searchApi.getCategories()` for category filter
  - `searchApi.getSuggestions()` for autocomplete
  - Pagination support
  - Loading/error states
- ✅ **SearchResultList** - Unified results display for articles and spotnet posts
  - Source badge (Article/Spotnet)
  - Rank display
  - Size formatting
  - Category/newsgroup/source_name tags
  - Bookmark button
- ✅ **AdvancedFilters** - Category, author, date range, sort options
  - Real categories from API
  - Sort by relevance/date/subject/author/size
  - Ascending/descending toggle
  - Newsgroup filter
- ✅ **SearchInput** - Search form with clear button, focus state
- ✅ **SearchHistory** - Local storage backed with localStorage
  - `searchStorage.addToHistory()`, `clearHistory()`, `getHistory()`
  - Max 20 items
- ✅ **Bookmarks storage** - Local storage backed
  - `bookmarkStorage.addBookmark()`, `removeBookmark()`, `isBookmarked()`
- ✅ **UserManagement wired to API** - Real CRUD with React Query mutations
  - `userApi.list()` with loading/error states
  - `userApi.create()` with form
  - `userApi.delete()` with confirmation
  - `userApi.toggleActive()`
  - Search/filter functionality
- ✅ **Admin Dashboard** - Stats grid with real API data
  - `adminApi.getDashboard()` with React Query
  - Articles, spotnet, jobs, servers, DB size stats
  - Server status display
  - Recent indexing jobs
- ✅ **Navbar** - Responsive nav with login/logout, admin links, user avatar
  - Mobile menu support
  - User avatar with initial
- ✅ **Protected routes** - `ProtectedRoute` wrapper component
  - Redirects to login if not authenticated
  - Admin-only route protection
  - Loading state during auth check
- ✅ **API client** - Comprehensive API with all endpoints
  - `authApi` - login, register, logout, me, password reset
  - `searchApi` - search, facets, suggestions, categories
  - `userApi` - list, get, create, update, delete, toggleActive
  - `serverApi` - NNTP server CRUD
  - `newsgroupApi` - newsgroup CRUD
  - `spotnetApi` - spotnet source CRUD + test connection
  - `jobApi` - indexing job CRUD
  - `bookmarkApi` - bookmarks CRUD
- ✅ **Dockerfile** - Multi-stage build (node → nginx)
- ✅ **nginx.conf** - API proxy, SPA routing, gzip, security headers, static caching
- ✅ **Error handling** - Consistent error display across all pages
- ✅ **Loading states** - Spinner/loading states across all pages

### Production Readiness (not phase-gated):
- CI-friendly test execution (testcontainers)
- Migration up/down reversibility validation
- Connection pool monitoring endpoint
- Article table partitioning strategy
- Backup/restore automation
- Schema drift detection

### Future Phases:
- **Phase 5: Search Engine** - Full-text search, faceting, ranking, result formatting
- **Phase 6: Frontend Integration** - React/Next.js dashboard, search UI, authentication flow
- **Phase 7: Deployment & Monitoring** - Docker/K8s configs, Grafana dashboards, alerting
