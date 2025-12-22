# Verify /metrics Endpoint Setup

## Current Setup

The `/metrics` endpoint is already configured in your code:

1. **In `api/server.go` (line 56):**
   ```go
   // Setup Prometheus /metrics endpoint
   observability.SetupPrometheusEndpoint(router)
   ```

2. **In `observability/prometheus.go`:**
   ```go
   func SetupPrometheusEndpoint(router *gin.Engine) {
       // Add the /metrics endpoint
       router.GET("/metrics", gin.WrapH(promhttp.Handler()))
       slog.Info("Prometheus metrics endpoint configured at /metrics")
   }
   ```

## Testing the Endpoint

### 1. Build and Run the Service
```bash
# Build the Docker image
docker build -t inventory-service:1.0.0 .

# Start the service
docker-compose up -d inventory-service

# Or run locally (if you have Go installed)
go run main.go
```

### 2. Test the /metrics Endpoint
```bash
# Test if the endpoint is accessible
curl http://localhost:13740/metrics

# You should see Prometheus metrics output like:
# HELP go_gc_duration_seconds A summary of the pause duration of garbage collection cycles.
# TYPE go_gc_duration_seconds summary
# go_gc_duration_seconds{quantile="0"} 0
# ...
# HELP http_requests_total Total number of HTTP requests
# TYPE http_requests_total counter
# http_requests_total{endpoint="/metrics",method="GET",status_code="200"} 1
```

### 3. Check for Your Custom Metrics
```bash
# Filter for your custom HTTP status metrics
curl http://localhost:13740/metrics | grep http_response_status_total

# Filter for inventory metrics
curl http://localhost:13740/metrics | grep inventory_

# Filter for database metrics
curl http://localhost:13740/metrics | grep database_
```

### 4. Generate Some Traffic to See Metrics
```bash
# Make some requests to generate metrics
curl http://localhost:13740/health/healthz
curl http://localhost:13740/v1/inventory/list
curl http://localhost:13740/v1/inventory/999  # This should generate a 4xx error

# Then check metrics again
curl http://localhost:13740/metrics | grep http_response_status_total
```

## Troubleshooting

If the `/metrics` endpoint is not working:

### 1. Check if the service is running
```bash
docker ps | grep inventory-service
# or
curl http://localhost:13740/health/healthz
```

### 2. Check the logs
```bash
docker logs inventory-service
# Look for: "Prometheus metrics endpoint configured at /metrics"
```

### 3. Check port binding
```bash
# Make sure port 13740 is exposed
netstat -tlnp | grep 13740
# or
lsof -i :13740
```

### 4. Test with verbose curl
```bash
curl -v http://localhost:13740/metrics
```

## Expected Output

When working correctly, you should see metrics like:
```
# HELP http_response_status_total Total number of HTTP responses by status class
# TYPE http_response_status_total counter
http_response_status_total{endpoint="/health/healthz",method="GET",status_class="2xx"} 1
http_response_status_total{endpoint="/v1/inventory/:id",method="GET",status_class="4xx"} 1

# HELP http_requests_total Total number of HTTP requests
# TYPE http_requests_total counter
http_requests_total{endpoint="/health/healthz",method="GET",status_code="200"} 1
http_requests_total{endpoint="/v1/inventory/:id",method="GET",status_code="400"} 1
```

The `/metrics` endpoint is already properly configured in your code!