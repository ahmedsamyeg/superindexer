# Deployment Guide

## Quick Start (Development)

```bash
# Start all services
make run-dev

# Or with Docker Compose directly
cd deployments/docker
docker-compose up -d
```

Services available at:
- **API**: http://localhost:8080
- **Frontend**: http://localhost:3000
- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3001 (admin/admin)

## Production Deployment

### 1. Prepare Secrets

```bash
cd deployments/docker
cp .env.production.example .env.production
# Edit .env.production with real values

# Generate secrets
openssl rand -hex 32 > secrets/jwt_secret.txt
openssl rand -hex 32 > secrets/session_secret.txt
openssl rand -hex 32 > secrets/db_password.txt
openssl rand -hex 32 > secrets/grafana_password.txt
```

### 2. Run Production Stack

```bash
docker-compose -f docker-compose.prod.yml up -d
```

### 3. Verify

```bash
# Check health
curl http://localhost:8080/health
curl http://localhost:8080/ready

# Check metrics
curl http://localhost:9090/metrics
```

## Kubernetes Deployment

### Prerequisites
- kubectl configured
- kustomize installed

### Deploy to Development

```bash
kubectl apply -k deployments/kubernetes/overlays/development
kubectl rollout status deployment/indexer-api -n indexer-dev --timeout=300s
kubectl rollout status deployment/indexer-worker -n indexer-dev --timeout=300s
kubectl rollout status deployment/indexer-frontend -n indexer-dev --timeout=300s
```

### Deploy to Production

```bash
kubectl apply -k deployments/kubernetes/overlays/production
kubectl rollout status deployment/indexer-api -n indexer-prod --timeout=300s
```

## Monitoring

### Grafana Dashboards

Grafana auto-provisions dashboards from `grafana/dashboards/`. Default dashboard:
- **Usenet Indexer - Application Overview** (auto-loaded as home dashboard)

### Prometheus Alerts

Alerts are defined in `prometheus-alerting.yml`. Key alerts:
- `IndexerAppUnhealthy` - Application health check fails
- `IndexerDatabaseConnectionsHigh` - DB connections > 20
- `IndexerProcessingErrorsHigh` - > 10 errors in 5 minutes
- `IndexerJobsStuck` - Active jobs with failures for 30+ minutes
- `IndexerNNTPConnectionsLow` - No NNTP connections for 10 minutes

### Metrics

Available Prometheus metrics:
- `indexer_articles_processed_total` - Total articles indexed
- `indexer_spotnet_posts_processed_total` - Total Spotnet posts
- `indexer_processing_errors_total` - Processing errors
- `indexer_jobs_active` - Active indexing jobs
- `indexer_jobs_failed_total` - Failed jobs counter
- `indexer_worker_utilization` - Worker utilization
- `indexer_nntp_connections_active` - NNTP connections
- `indexer_ingestion_duration_seconds` - Ingestion latency
- `indexer_search_duration_seconds` - Search latency
- `indexer_database_connections_active` - DB connections
- `indexer_app_healthy` - Application health (1/0)

## Backup & Recovery

### Manual Backup

```bash
# Create backup
docker exec indexer-postgres-prod pg_dump -U indexer indexer > backups/backup_$(date +%Y%m%d).sql

# Restore from backup
docker exec -i indexer-postgres-prod psql -U indexer indexer < backups/backup_YYYYMMDD.sql
```

### Automated Backup

Set `BACKUP_ENABLED=true` in `.env.production` and configure `BACKUP_DIR` and `BACKUP_RETENTION`.

## Troubleshooting

### Services won't start

```bash
# Check logs
docker-compose logs -f api
docker-compose logs -f postgres

# Check health
docker inspect --format='{{.State.Health.Status}}' indexer-postgres-prod
docker inspect --format='{{.State.Health.Status}}' indexer-api-prod
```

### Database connection issues

```bash
# Check DB is healthy
docker exec indexer-postgres-prod pg_isready -U indexer

# Check connection pool
curl http://localhost:8080/api/v1/admin/db-stats
```

### Migration errors

```bash
# Run migrations manually
docker-compose run --rm migrate /app/indexer-migrate up

# Rollback
docker-compose run --rm migrate /app/indexer-migrate down 1
```
