# Prometheus Metrics Integration

This document explains how to use the Prometheus metrics integration in the inventory service.

## Overview

The inventory service now exposes Prometheus metrics alongside OpenTelemetry metrics for comprehensive observability:

- **Prometheus**: For metrics collection and alerting
- **OpenTelemetry**: For distributed tracing and logs to ELK stack
- **Grafana**: For metrics visualization
- **ELK Stack**: For logs and traces (your existing setup)

## Metrics Endpoint

The service exposes Prometheus metrics at:
```
http://localhost:13740/metrics
```

## Available Metrics

### HTTP Metrics
- `http_requests_total` - Total HTTP requests (labels: method, endpoint, status_code)
- `http_request_duration_seconds` - HTTP request duration histogram (labels: method, endpoint)
- `http_requests_in_flight` - Current number of requests being processed
- `http_response_status_total` - **NEW**: Total HTTP responses by status class (labels: method, endpoint, status_class)
  - Status classes: `2xx`, `3xx`, `4xx`, `5xx`, `1xx`

### Database Metrics
- `database_connections_active` - Number of active database connections
- `database_operation_duration_seconds` - Database operation duration (labels: operation, table)
- `database_operation_errors_total` - Database operation errors (labels: operation, table, error_type)

### Business Metrics
- `inventory_operations_total` - Total inventory operations (labels: operation, category, location)
- `inventory_items_active` - Current number of active inventory items
- `authentication_attempts_total` - Authentication attempts (labels: status, method)

### System Metrics (Automatic)
- `go_*` - Go runtime metrics (goroutines, memory, GC)
- `process_*` - Process metrics (CPU, memory, file descriptors)

## Setup Instructions

### 1. Build and Run with Docker Compose

```bash
# Build the service
docker build -t inventory-service:1.0.0 .

# Start all services (inventory + prometheus + grafana)
docker-compose up -d
```

### 2. Access Services

- **Inventory Service**: http://localhost:13740
- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000 (admin/admin)

### 3. Configure Grafana

1. Open Grafana at http://localhost:3000
2. Login with admin/admin
3. Add Prometheus as a data source:
   - URL: `http://prometheus:9090`
   - Access: Server (default)
4. Import or create dashboards for your metrics

### 4. Example Prometheus Queries

```promql
# Request rate
rate(http_requests_total[5m])

# Request duration 95th percentile
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))

# Database operation errors
rate(database_operation_errors_total[5m])

# Active inventory items
inventory_items_active

# Inventory operations by category
sum(rate(inventory_operations_total[5m])) by (category)

# NEW: HTTP Status Code Metrics
# 2xx success rate
rate(http_response_status_total{status_class="2xx"}[5m])

# 4xx client error rate
rate(http_response_status_total{status_class="4xx"}[5m])

# 5xx server error rate
rate(http_response_status_total{status_class="5xx"}[5m])

# Error rate percentage (4xx + 5xx / total)
(
  rate(http_response_status_total{status_class="4xx"}[5m]) +
  rate(http_response_status_total{status_class="5xx"}[5m])
) / 
rate(http_response_status_total[5m]) * 100

# Success rate percentage (2xx / total)
rate(http_response_status_total{status_class="2xx"}[5m]) / 
rate(http_response_status_total[5m]) * 100

# Status code breakdown by endpoint
sum(rate(http_response_status_total[5m])) by (endpoint, status_class)
```

## Integration with Existing ELK Stack

Your existing ELK stack integration remains unchanged:
- **Logs**: Continue shipping to ELK via OTEL SDK
- **Traces**: Continue shipping to ELK via OTEL SDK  
- **Metrics**: Now available in both Prometheus (for alerting) and OTEL (for correlation)

## Alerting

You can set up Prometheus alerting rules. Example alerts for HTTP status codes:

```yaml
# alerts.yml
groups:
- name: inventory-service-http
  rules:
  - alert: HighErrorRate
    expr: |
      (
        rate(http_response_status_total{status_class="4xx"}[5m]) +
        rate(http_response_status_total{status_class="5xx"}[5m])
      ) / rate(http_response_status_total[5m]) > 0.1
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "High HTTP error rate detected"
      description: "Error rate is {{ $value | humanizePercentage }} for the last 5 minutes"

  - alert: HighServerErrorRate
    expr: rate(http_response_status_total{status_class="5xx"}[5m]) / rate(http_response_status_total[5m]) > 0.05
    for: 2m
    labels:
      severity: critical
    annotations:
      summary: "High server error rate detected"
      description: "5xx error rate is {{ $value | humanizePercentage }} for the last 5 minutes"

  - alert: NoSuccessfulRequests
    expr: rate(http_response_status_total{status_class="2xx"}[10m]) == 0
    for: 5m
    labels:
      severity: critical
    annotations:
      summary: "No successful HTTP requests"
      description: "No 2xx responses in the last 10 minutes"
```

## Custom Metrics

To add custom metrics, use the `PrometheusMetrics` struct in your handlers:

```go
// Record a custom business metric
h.prometheusMetrics.RecordInventoryOperation("create", "electronics", "warehouse-a")

// Update a gauge
h.prometheusMetrics.UpdateInventoryCount(totalCount)

// Record database operation with automatic timing
err := h.prometheusMetrics.WithDBMetrics("select", "inventory", func() error {
    return h.queries.GetInventory(ctx, id)
})
```

## Monitoring Best Practices

1. **Use labels wisely** - Don't create high-cardinality labels
2. **Monitor the four golden signals**:
   - Latency: `http_request_duration_seconds`
   - Traffic: `rate(http_requests_total[5m])`
   - Errors: `rate(http_requests_total{status_code=~"5.."}[5m])`
   - Saturation: `database_connections_active`, `go_memstats_*`

3. **Set up alerts** for critical metrics
4. **Create dashboards** for different audiences (dev, ops, business)

## Troubleshooting

### Metrics not appearing
- Check if `/metrics` endpoint is accessible
- Verify Prometheus configuration in `prometheus.yml`
- Check Docker network connectivity

### High memory usage
- Reduce metric retention time in Prometheus config
- Review label cardinality
- Consider metric sampling for high-volume metrics