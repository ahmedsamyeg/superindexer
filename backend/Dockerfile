# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /indexer-api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -o /indexer-worker ./cmd/worker
RUN CGO_ENABLED=0 GOOS=linux go build -o /indexer-migrate ./cmd/migrate

# Runtime stage
FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /app

# Copy binaries from builder
COPY --from=builder /indexer-api .
COPY --from=builder /indexer-worker .
COPY --from=builder /indexer-migrate .

# Copy migrations
COPY --from=builder /app/internal/database/migrations ./migrations

# Set ownership
RUN chown -R appuser:appgroup /app

USER appuser

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

ENTRYPOINT ["/app/indexer-api"]
