# Heimdall Reverse Proxy

A lightweight reverse proxy with pluggable load-balancing strategies and active health checking, written in Go.

## Features

- **Round-Robin Load Balancing** — distributes requests evenly across healthy backends
- **Active Health Checks** — periodically probes each backend and removes unhealthy ones from rotation
- **Pluggable Strategy Interface** — add new balancing algorithms by implementing a single interface
- **YAML Configuration** — single `app.yaml` file for all settings
- **Graceful Shutdown** — health-check goroutines are cancelled via context on stop

## Architecture

```
                         ┌──────────────────────┐
                         │      Client          │
                         └──────────┬───────────┘
                                    │
                                    ▼
                         ┌──────────────────────┐
                         │   Heimdall Proxy      │
                         │   :8080               │
                         │                      │
                         │  ┌────────────────┐  │
                         │  │ LoadBalancer    │  │
                         │  │                │  │
                         │  │ Strategy ──────┤  │
                         │  │ (Round-Robin)   │  │
                         │  └──┬────┬────┬───┘  │
                         └─────┼────┼────┼──────┘
                               │    │    │
                    ┌──────────┘    │    └──────────┐
                    ▼               ▼               ▼
              ┌──────────┐   ┌──────────┐   ┌──────────┐
              │ Backend 0│   │ Backend 1│   │ Backend 2│
              │ :8081    │   │ :8082    │   │ :8083    │
              └──────────┘   └──────────┘   └──────────┘

        Health checks run in background goroutines per backend.
        Unhealthy backends are skipped during routing.
```

## Quick Start

### From source

```bash
go build -o heimdall .
./heimdall
```

### With Docker

```bash
docker build -t heimdall-reverse-proxy .
docker run --rm -p 8080:8080 heimdall-reverse-proxy
```

### With Make

```bash
make build    # lint + compile
make test     # run unit tests
make docker   # build Docker image
```

## Configuration

All configuration lives in `app.yaml`:

```yaml
server:
  port: "8080"                    # port the proxy listens on

load_balancer:
  enabled: true                   # enable/disable the load balancer
  strategy: round-robin           # balancing algorithm (see below)
  health:
    path: "/health"               # endpoint to probe on each backend
    interval: 2s                  # time between health checks
    timeout: 10s                  # per-check HTTP timeout
  resources:                      # list of backend URLs
    - "http://localhost:8081"
    - "http://localhost:8082"
    - "http://localhost:8083"
```

| Field | Type | Description |
|---|---|---|
| `server.port` | string | Listen port for the proxy |
| `load_balancer.enabled` | bool | Toggle load balancing on/off |
| `load_balancer.strategy` | string | Algorithm name (`round-robin`) |
| `load_balancer.health.path` | string | Health endpoint path on backends |
| `load_balancer.health.interval` | duration | How often to run health checks |
| `load_balancer.health.timeout` | duration | HTTP timeout per health check |
| `load_balancer.resources` | []string | Backend server URLs |

## Health Checks

Each backend runs a health-check goroutine that:

1. Sends `GET <backend_url><health_path>` at the configured interval
2. Expects HTTP 200 — any other status or connection error marks the backend as **dead**
3. A successful probe marks the backend as **alive**

The load balancer skips dead backends when selecting the next target. If all backends are down, requests receive a `NoHealthyBackendsError`.

## Adding New Strategies

Implement the `StrategyAlgorithm` interface:

```go
type StrategyAlgorithm interface {
    NextBackend(backends []*Backend) (*Backend, error)
}
```

Then register it in `load_balancer/mapper.go`:

```go
func ToLoadBalancerStrategy(strategyStr string) (StrategyAlgorithm, error) {
    formatted := strings.ToUpper(strings.ReplaceAll(strategyStr, "-", "_"))
    switch formatted {
    case "ROUND_ROBIN":
        return &RoundRobinStrategy{}, nil
    case "LEAST_CONNECTIONS":          // add your case
        return &LeastConnectionsStrategy{}, nil
    default:
        return nil, &ConfigError{...}
    }
}
```

## Testing

**Unit tests** — test error handling and configuration validation:

```bash
go test ./load_balancer/test/ -count=1
```

**Integration tests** — use [testcontainers-go](https://github.com/testcontainers/testcontainers-go) to spin up real Nginx backends in Docker:

```bash
go test ./load_balancer/test/ -run='TestRoundRobinDistribution|TestSkipsUnhealthyBackend|TestHealthCheckMarksBackendAlive' -count=1
```

**Benchmarks** — measure strategy throughput and end-to-end proxy performance:

```bash
go test ./load_balancer/test/ -bench=. -benchmem
```

## Project Structure

```
.
├── reverse_proxy.go              # entrypoint — HTTP server setup
├── config.go                     # YAML config loading
├── app.yaml                      # default configuration
├── load_balancer/
│   ├── load_balancer.go          # LoadBalancer struct, creation, routing
│   ├── load_balancer_strategy.go # StrategyAlgorithm interface + RoundRobin
│   ├── server.go                 # Backend struct, health checks, reverse proxy
│   ├── mapper.go                 # strategy name → implementation mapping
│   ├── errors.go                 # typed errors
│   └── test/
│       ├── load_balancer_test.go       # unit + integration tests
│       ├── load_balancer_bench_test.go # benchmarks
│       └── testutil_test.go            # testcontainers helpers
├── Dockerfile                    # multi-stage build
├── Makefile                      # build, lint, test, docker targets
└── go.mod
```

## License

This project is licensed under the MIT License — see the [LICENSE](LICENSE) file for details.
