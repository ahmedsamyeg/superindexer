# API Integration Tests

This directory contains integration tests for the HTTP API endpoints.

## Test Structure

Tests are organized by component:
- `auth_test.go` - Authentication endpoint tests
- `user_test.go` - User management endpoint tests  
- `search_test.go` - Search endpoint tests
- `spotnet_test.go` - Spotnet source management tests
- `nntp_test.go` - NNTP server management tests

## Running Tests

```bash
# Run all integration tests
go test ./internal/http/test/...

# Run specific test suite
go test ./internal/http/test/ -run TestAuth

# Run with coverage
go test ./internal/http/test/ -coverprofile=coverage.out
```

## Test Coverage

Tests cover:
- Authentication and authorization flows
- User CRUD operations with proper permissions
- Search functionality with various filters
- Spotnet and NNTP source management
- Error handling and validation
- API response format consistency

## Test Database Setup

Integration tests use a dedicated test database:
- `postgres://test:test@localhost:5432/test_indexer`
- Database is cleaned and migrated before each test run
- Test data is isolated in separate schema or tables

## Example Test Pattern

```go
func TestAuthLogin(t *testing.T) {
    // Setup test database and fixtures
    db := setupTestDatabase()
    
    // Create test user
    user := &models.User{
        Username: "testuser",
        Email:    "test@example.com",
        PasswordHash: "hashed_password",
        Role: "user",
    }
    
    // Test login endpoint
    req := httptest.NewRequest("POST", "/auth/login", strings.NewReader(`{
        "username": "testuser",
        "password": "password"
    }`))
    
    // Execute and assert results
    rec := httptest.NewRecorder()
    // ... actual test logic
    
    assert.Equal(t, http.StatusOK, rec.Code)
}
```