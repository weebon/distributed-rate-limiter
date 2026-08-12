# Distributed Rate Limiter

A distributed API rate limiter and reverse proxy gateway, built in Go, using Redis for shared state across multiple gateway instances. Verified to correctly enforce a single shared limit under concurrent load exceeding 20,000 req/sec per instance — see [RESULTS.md](./RESULTS.md).

## What it does

- Sits in front of a backend service as a reverse proxy
- Enforces rate limits per client (by IP) and per route
- Two selectable algorithms: **token bucket** and **sliding window log**
- Limit state lives in Redis, shared atomically (via Lua scripting) across any number of gateway instances — not per-process, genuinely distributed
- Per-route configuration: different endpoints can have different limits

## Why this project

Rate limiting looks trivial on a single machine — a counter and a mutex will do. The moment you run more than one instance of a service (which any real production system does, for availability and scale), a per-process counter stops meaning anything: each instance enforces its own separate limit, and the *actual* combined limit becomes instance count × intended limit.

This project exists to solve that properly: move the limit state out of the process and into shared storage (Redis), and make the check-and-decrement atomic so concurrent requests across many instances can't race past the limit. That's the core distributed-systems problem here — not the rate limiting math itself, but coordinating shared state correctly under concurrency.

**Why Lua scripts, not just Redis commands:** a naive `GET` then `SET` from the application code is not atomic — two gateway instances could both read "1 token left" and both allow the request. Redis executes Lua scripts as a single atomic operation, so the entire "read state, compute new state, write state" sequence happens indivisibly, no matter how many gateway instances are hitting Redis concurrently.

**Why two algorithms:** token bucket allows bursts (good for clients with occasional spiky-but-legitimate traffic) but can let through up to double the intended rate right at a refill boundary. Sliding window log is more precise (no boundary burst) but costs more memory per client (one entry per request in the window, vs. two numbers for token bucket). Implementing both was a chance to compare a practical trade-off rather than just picking one.

## Architecture

Client → Gateway (rate limit check) → Backend
↓
Redis (shared limit state)


Multiple gateway instances can run behind a load balancer, all checking the same Redis-backed limiter, enforcing one consistent limit regardless of which instance handles a given request.

## Algorithms

**Token bucket** — each client gets a bucket that refills at a fixed rate and drains on each request. Allows short bursts up to bucket capacity.

**Sliding window log** — counts requests within a rolling time window using a Redis sorted set. More accurate than fixed windows, no boundary burst issue.

Both are implemented as atomic Redis Lua scripts, so concurrent requests across multiple gateway instances can't race past the limit.

## Tech stack

- Go (stdlib `net/http`, `net/http/httputil` for the reverse proxy)
- Redis (shared state + Lua scripting for atomicity)
- Tested with Go's built-in `-race` detector and load tested with [`hey`](https://github.com/rakyll/hey)

## Prerequisites

- **Go** 1.21+ — [install instructions](https://go.dev/doc/install)
- **Redis** — on Debian/Ubuntu (including WSL): `sudo apt install redis-server`, then start it with `sudo service redis-server start`
- **(Optional) `hey`** for load testing — `sudo apt install hey`, or `go install github.com/rakyll/hey@latest`

## Running locally

**1. Start Redis** (if not already running):
```bash
sudo service redis-server start
redis-cli ping   # should reply PONG
```

**2. Start the backend (dummy service being protected):**
```bash
cd cmd/backend
go run main.go
```

**3. Start the gateway** (in a new terminal):
```bash
cd cmd/gateway
go run main.go
```
Defaults to token bucket algorithm on port 8080. Override with env vars:
```bash
ALGO=sliding_window PORT=8081 go run main.go
```

**4. Test it:**
```bash
curl http://localhost:8080/api/test
```

## Per-route limits

Configured in `internal/config/config.go`:

| Route | Capacity | Refill rate |
|-------|----------|--------------|
| `/api/test` (default) | 5 | 1/sec |
| `/api/search` | 20 | 5/sec |
| `/api/upload` | 2 | 0.1/sec |

## Testing

```bash
go test -race ./...
```

## Results

See [RESULTS.md](./RESULTS.md) for load test data proving distributed correctness — combined allowed requests across two concurrent gateway instances matched the single-instance baseline rather than doubling, confirming the shared limit holds under real concurrent load.

## Status

Core functionality complete: distributed rate limiting, two algorithms, per-route config, verified under load. Optional next steps: live metrics dashboard, dynamic config reload, cluster membership for gateway auto-discovery.