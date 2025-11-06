# Heimdall Telemetry System

Heimdall is a comprehensive telemetry and analytics system for monitoring API requests, collecting metadata, and providing insights for security and performance analysis.

## Features

### 🔍 Metadata Extraction
- **IP Address Extraction**: Parses `X-Forwarded-For`, `Forwarded`, `X-Real-IP`, `CF-Connecting-IP` headers
- **Client Information**: Extracts `User-Agent`, `X-Device-Id`, and other client metadata
- **Header Validation**: Whitelist-based approach with sanitization against XSS and injection attacks
- **IP Normalization**: Validates and normalizes IP addresses with private IP detection

### 📊 Request Logging
- **Comprehensive Logging**: Captures request/response metadata, latency, status codes, payload sizes
- **Parameter Digests**: Creates hashed digests of request parameters for anomaly detection
- **Cookie Sanitization**: Automatically redacts sensitive cookie values (sessions, tokens, auth)
- **Geolocation Support**: Optional geolocation data based on client IP (configurable)

### ⚡ High-Performance Architecture
- **Async Processing**: Non-blocking telemetry collection using buffered channels
- **Worker Pool**: Configurable number of worker goroutines for processing
- **Backpressure Handling**: Graceful degradation when database is unavailable
- **Disk Queueing**: Persistent fallback queue for zero-loss guarantee

### 📈 Analytics & Metrics
- **Frequency Counters**: Redis-based counters for URLs, tokens, and users
- **Real-time Analytics**: Hourly rollups and aggregation
- **Anomaly Detection**: Parameter digest analysis and usage pattern monitoring
- **Performance Metrics**: Latency tracking, error rates, and throughput analysis

## Configuration

### Environment Variables

```bash
# Enable/disable Heimdall telemetry
HEIMDALL_TELEMETRY_ENABLED=true

# Enable geolocation lookup
HEIMDALL_GEOLOCATION_ENABLED=false

# Buffer configuration
HEIMDALL_BUFFER_SIZE=10000
HEIMDALL_WORKER_COUNT=5

# Retry configuration
HEIMDALL_RETRY_ATTEMPTS=3
HEIMDALL_RETRY_DELAY_MS=1000

# Disk queue configuration
HEIMDALL_DISK_QUEUE_ENABLED=true
HEIMDALL_DISK_QUEUE_PATH=/tmp/heimdall_queue

# Flush configuration
HEIMDALL_FLUSH_INTERVAL_MS=5000
```

## Database Schema

### HeimdallRequestLog Table

| Column | Type | Description | Indexed |
|--------|------|-------------|----------|
| id | int | Primary key | ✓ |
| request_id | string(64) | Unique request identifier | ✓ |
| occurred_at | timestamp | Request timestamp | ✓ |
| auth_key_fingerprint | string(128) | Hashed authorization key | ✓ |
| user_id | int | User ID (nullable) | ✓ |
| token_id | int | Token ID (nullable) | ✓ |
| normalized_url | string(512) | Normalized request URL | ✓ |
| http_method | string(16) | HTTP method | ✓ |
| http_status | int | HTTP status code | ✓ |
| latency_ms | bigint | Request latency in milliseconds | ✓ |
| client_ip | string(64) | Client IP address | ✓ |
| client_user_agent | string(512) | Sanitized user agent | |
| client_device_id | string(128) | Client device identifier | ✓ |
| request_size_bytes | bigint | Request payload size | |
| response_size_bytes | bigint | Response payload size | |
| param_digest | string(128) | Hash of request parameters | ✓ |
| sanitized_cookies | text | Sanitized cookie string | |
| country_code | string(8) | Country code (if geolocation enabled) | ✓ |
| region | string(64) | Region/State | |
| city | string(128) | City name | |
| processing_time_ms | bigint | Internal processing time | |
| upstream_provider | string(128) | Upstream service provider | ✓ |
| model_name | string(128) | AI model name | ✓ |
| error_message | text | Error message (if any) | |
| error_type | string(64) | Error category | ✓ |
| created_at | timestamp | Record creation time | |
| updated_at | timestamp | Record update time | |

## API Endpoints

### Authentication Required
All endpoints require user authentication. Admin endpoints require admin privileges.

#### Telemetry Statistics
```http
GET /heimdall/stats
```
Returns current telemetry worker statistics and configuration.

#### Configuration
```http
GET /heimdall/config
PUT /heimdall/config
```
View or update Heimdall configuration.

#### Metrics
```http
GET /heimdall/metrics/urls?time_window=1h
GET /heimdall/metrics/tokens?time_window=24h
GET /heimdall/metrics/users?time_window=7d
```
Retrieve frequency metrics for URLs, tokens, or users.

#### Anomaly Detection
```http
GET /heimdall/metrics/anomaly?time_window=1h
```
Get data for anomaly detection analysis.

#### Dashboard
```http
GET /heimdall/dashboard?time_window=24h
```
Comprehensive dashboard with all metrics.

#### Admin Only
```http
POST /heimdall/admin/cleanup
POST /heimdall/admin/rollups
```
Trigger cleanup or generate hourly rollups.

## Redis Keys

Heimdall uses Redis for real-time metrics:

```
heimdall:url:{url}:count          - URL request counter
heimdall:token:{token_id}:count   - Token request counter
heimdall:user:{user_id}:count     - User request counter
```

All keys have a 1-hour TTL for automatic cleanup.

## Security Considerations

### Data Sanitization
- **User Agent**: XSS characters replaced with HTML entities
- **Cookies**: Sensitive values (session, token, auth) replaced with `***`
- **IP Validation**: Invalid IP addresses are rejected
- **Header Filtering**: Only whitelisted headers are processed

### Privacy Protection
- **Authorization Hashing**: Raw auth keys are never stored, only fingerprints
- **Configurable Geolocation**: Disabled by default, requires explicit enablement
- **Data Retention**: Redis keys auto-expire, database cleanup configurable

### Access Control
- **Authentication**: All endpoints require valid user session
- **Authorization**: Admin endpoints require admin privileges
- **Rate Limiting**: Respects existing rate limiting middleware

## Performance Impact

### Minimal Overhead
- **Async Processing**: Main request path is not blocked
- **Buffered Channels**: 10,000 entry buffer by default
- **Efficient Indexing**: Optimized database indexes for common queries
- **Connection Pooling**: Reuses database connections efficiently

### Resource Usage
- **Memory**: ~50MB for 10,000 buffered entries
- **CPU**: ~1% overhead per request for metadata extraction
- **Storage**: ~200 bytes per request in database
- **Network**: Minimal additional Redis operations

## Monitoring

### Health Checks
Monitor these metrics for system health:

```bash
# Worker status
GET /heimdall/stats

# Buffer utilization
# Check "buffer_length" vs "buffer_capacity"

# Disk queue size
# Check "disk_queue_size" if enabled

# Error rates
# Monitor "error_rate" in metrics endpoints
```

### Alerting
Set up alerts for:
- Buffer utilization > 80%
- Disk queue size growing continuously
- Error rate > 5%
- Worker not running

## Troubleshooting

### Common Issues

#### High Buffer Utilization
- Increase `HEIMDALL_BUFFER_SIZE`
- Increase `HEIMDALL_WORKER_COUNT`
- Check database performance

#### Disk Queue Growing
- Database connectivity issues
- Insufficient worker capacity
- Disk space constraints

#### Missing Data
- Check `HEIMDALL_TELEMETRY_ENABLED=true`
- Verify middleware is loaded
- Check Redis connectivity

### Debug Mode
Enable debug logging:
```bash
export DEBUG_ENABLED=true
```

## Integration Examples

### Custom Analytics
```go
// Get URL metrics for last hour
analytics := service.GlobalHeimdallAnalyticsService
metrics, err := analytics.GetURLFrequencyMetrics(ctx, time.Hour)

// Get anomaly detection data
data, err := analytics.GetAnomalyDetectionData(ctx, 24*time.Hour)
```

### Custom Middleware
```go
// Add custom metadata extraction
func customMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Your custom logic here
        c.Next()
    }
}
```

## Development

### Running Tests
```bash
# Unit tests
go test ./model/heimdall_request_log_test.go

# Integration tests
go test ./model/heimdall_integration_test.go

# Middleware tests
go test ./middleware/heimdall_telemetry_test.go
```

### Benchmarking
```bash
# Performance benchmarks
go test -bench=. ./middleware/
go test -bench=. ./model/
```

## Future Enhancements

### Planned Features
- **Machine Learning**: Anomaly detection using ML models
- **Real-time Alerts**: Webhook notifications for anomalies
- **Advanced Geolocation**: ISP and organization detection
- **Custom Dashboards**: User-configurable dashboard layouts
- **Data Export**: CSV/JSON export for external analysis

### Extensibility
- **Plugin System**: Custom metadata extractors
- **Storage Backends**: Alternative storage systems
- **Analytics Extensions**: Custom metric calculations

## License

This telemetry system is part of the NebulaGate project and follows the same licensing terms.
