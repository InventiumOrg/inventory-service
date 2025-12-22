# Testing HTTP Status Code Metrics

## Test the Status Code Metrics

After starting your service, you can test the different HTTP status codes to see the metrics in action:

### 1. Test 2xx (Success) Responses
```bash
# Test successful requests
curl http://localhost:13740/health/healthz
curl http://localhost:13740/v1/inventory/list
```

### 2. Test 4xx (Client Error) Responses
```bash
# Test invalid inventory ID (400 Bad Request)
curl http://localhost:13740/v1/inventory/invalid-id

# Test non-existent endpoint (404 Not Found)
curl http://localhost:13740/non-existent-endpoint
```

### 3. Test 5xx (Server Error) Responses
```bash
# These would typically happen due to database issues or internal errors
# You might need to simulate these by temporarily stopping your database
# or creating a test endpoint that returns 500 errors
```

### 4. View Metrics
```bash
# Check the metrics endpoint
curl http://localhost:13740/metrics | grep http_response_status_total

# You should see output like:
# http_response_status_total{endpoint="/health/healthz",method="GET",status_class="2xx"} 5
# http_response_status_total{endpoint="/v1/inventory/:id",method="GET",status_class="4xx"} 2
# http_response_status_total{endpoint="unknown",method="GET",status_class="4xx"} 1
```

### 5. Query in Prometheus
Once you have Prometheus running (http://localhost:9090), you can query:

```promql
# Total 2xx responses
sum(http_response_status_total{status_class="2xx"})

# Total 4xx responses  
sum(http_response_status_total{status_class="4xx"})

# Total 5xx responses
sum(http_response_status_total{status_class="5xx"})

# Error rate (4xx + 5xx) as percentage
(
  sum(rate(http_response_status_total{status_class="4xx"}[5m])) +
  sum(rate(http_response_status_total{status_class="5xx"}[5m]))
) / sum(rate(http_response_status_total[5m])) * 100
```

### 6. Grafana Dashboard Panels

Create panels in Grafana with these queries:

**Success Rate Panel:**
```promql
sum(rate(http_response_status_total{status_class="2xx"}[5m])) / sum(rate(http_response_status_total[5m])) * 100
```

**Error Rate Panel:**
```promql
(
  sum(rate(http_response_status_total{status_class="4xx"}[5m])) +
  sum(rate(http_response_status_total{status_class="5xx"}[5m]))
) / sum(rate(http_response_status_total[5m])) * 100
```

**Status Code Breakdown (Pie Chart):**
```promql
sum by (status_class) (rate(http_response_status_total[5m]))
```