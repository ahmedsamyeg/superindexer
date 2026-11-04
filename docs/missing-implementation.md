# Missing Implementation Checklist

The following table summarizes the incomplete or placeholder code found in the current repository. These gaps prevent the worker service from running, limit observability cleanup, and leave error handling in an unsafe state.

| File | Line(s) | Issue | Impact |
|------|---------|-------|--------|
| `backend/cmd/worker/main.go` | 51‑54 | **TODO** – worker, NNTP ingestion, Spotnet ingestion, job scheduler not started. | The worker process starts and immediately blocks waiting for a shutdown signal; no indexing happens.
| `backend/internal/observability/observability.go` | 154 | **TODO** – tracer shutdown not implemented. | Traces may be lost on graceful exit.
| `backend/internal/database/postgres.go` | 119 | **panic(r)** left in migration error path. | Migration failures crash the process instead of returning an error.
| `backend/internal/http/middleware/timeout.go` | 50‑51 | Incomplete panic recovery comment. | Potential unhandled panics during request handling.

## Suggested Implementation Steps
1. **Worker Service** – Implement `workers.Init(ctx, db)` and start ingestion loops.
2. **Ingestion Services** – Add NNTP and Spotnet packages with `Start(ctx)` goroutines.
3. **Scheduler** – Use a ticker or cron to trigger periodic jobs.
4. **Observability Shutdown** – Call tracer provider `Shutdown` and flush spans in `observability.Stop()`.
5. **Migration Error Handling** – Replace `panic(r)` with proper error propagation.
6. **Timeout Middleware** – Ensure recovered panics are logged and handled safely.

---

This checklist can be used as a baseline for the next sprint of development.
