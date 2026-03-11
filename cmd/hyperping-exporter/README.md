# Hyperping Exporter

Prometheus exporter for [Hyperping](https://hyperping.io/) monitoring metrics. Exposes monitor status, healthcheck state, SSL expiration, and more as Prometheus-compatible metrics.

Built on top of the [terraform-provider-hyperping](https://github.com/develeap/terraform-provider-hyperping) API client with retry logic, circuit breaker, and rate limit handling.

## Quick Start

```bash
# Install
go install github.com/develeap/terraform-provider-hyperping/cmd/hyperping-exporter@latest

# Run
export HYPERPING_API_KEY="sk_your_api_key"
hyperping-exporter
# → Metrics at http://localhost:9312/metrics
```

## Configuration

| Flag | Env Var | Default | Description |
|------|---------|---------|-------------|
| `--api-key` | `HYPERPING_API_KEY` | (required) | Hyperping API key |
| `--listen-address` | | `:9312` | Address to listen on |
| `--metrics-path` | | `/metrics` | Metrics endpoint path |
| `--cache-ttl` | | `60s` | API data refresh interval |
| `--log-level` | | `info` | Log level (debug, info, warn, error) |
| `--log-format` | | `text` | Log format (text, json) |

## Endpoints

| Path | Description |
|------|-------------|
| `/metrics` | Prometheus metrics |
| `/healthz` | Liveness probe (always 200) |
| `/readyz` | Readiness probe (200 after first successful scrape) |
| `/` | Landing page with links |

## Metrics

### Monitor Metrics

Labels: `uuid`, `name`

| Metric | Type | Description |
|--------|------|-------------|
| `hyperping_monitor_up` | gauge | 1 if monitor is up, 0 if down |
| `hyperping_monitor_paused` | gauge | 1 if monitor is paused, 0 if active |
| `hyperping_monitor_ssl_expiration_days` | gauge | Days until SSL certificate expiration (HTTPS only) |
| `hyperping_monitor_check_interval_seconds` | gauge | Check frequency in seconds |

#### Monitor Info Metric

Labels: `uuid`, `name`, `protocol`, `url`, `project_uuid`, `http_method`

| Metric | Type | Description |
|--------|------|-------------|
| `hyperping_monitor_info` | gauge | Always 1; carries metadata as labels |

### Healthcheck Metrics

Labels: `uuid`, `name`

| Metric | Type | Description |
|--------|------|-------------|
| `hyperping_healthcheck_up` | gauge | 1 if healthcheck is up, 0 if down |
| `hyperping_healthcheck_paused` | gauge | 1 if paused, 0 if active |
| `hyperping_healthcheck_period_seconds` | gauge | Expected ping period in seconds |

### Summary Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `hyperping_monitors` | gauge | Total number of monitors |
| `hyperping_healthchecks` | gauge | Total number of healthchecks |

### Exporter Self-Monitoring

| Metric | Type | Description |
|--------|------|-------------|
| `hyperping_scrape_duration_seconds` | gauge | Duration of last API scrape |
| `hyperping_scrape_success` | gauge | 1 if last scrape succeeded, 0 if failed |

Standard Go runtime and process metrics are also exposed (`go_*`, `process_*`).

## Architecture

The exporter uses a **background cache refresh** pattern:

1. A background goroutine calls the Hyperping API at the configured interval (`--cache-ttl`)
2. Monitor and healthcheck data are fetched **in parallel** and cached in memory
3. Prometheus scrapes read from the cache — fast, predictable, no API hammering
4. If an API call fails, the previous cached data is preserved

This design prevents multiple Prometheus scrapers from multiplying API calls and avoids scrape timeout issues.

## Docker

```bash
docker run -e HYPERPING_API_KEY=sk_xxx -p 9312:9312 \
  ghcr.io/develeap/hyperping-exporter:latest
```

## Prometheus Configuration

```yaml
scrape_configs:
  - job_name: hyperping
    scrape_interval: 60s
    static_configs:
      - targets: ['localhost:9312']
```

Set `scrape_interval` to match or exceed `--cache-ttl` for efficient operation.

## Alerting Examples

```yaml
groups:
  - name: hyperping
    rules:
      - alert: MonitorDown
        expr: hyperping_monitor_up == 0 and hyperping_monitor_paused == 0
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "Monitor {{ $labels.name }} is down"

      - alert: SSLExpiringSoon
        expr: hyperping_monitor_ssl_expiration_days < 14
        for: 1h
        labels:
          severity: warning
        annotations:
          summary: "SSL cert for {{ $labels.name }} expires in {{ $value }} days"

      - alert: HealthcheckDown
        expr: hyperping_healthcheck_up == 0 and hyperping_healthcheck_paused == 0
        for: 10m
        labels:
          severity: critical
        annotations:
          summary: "Healthcheck {{ $labels.name }} is down"

      - alert: ExporterScrapeFailure
        expr: hyperping_scrape_success == 0
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Hyperping exporter failed to scrape API"
```

## Building

```bash
# From repository root
go build -o hyperping-exporter ./cmd/hyperping-exporter/

# With version info
go build -ldflags "-X main.version=1.0.0" -o hyperping-exporter ./cmd/hyperping-exporter/
```
