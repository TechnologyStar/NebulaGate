# Security Analytics - Anomaly Detection Engine

## Overview

The Anomaly Detection Engine is a comprehensive system for analyzing user behavior patterns and detecting security anomalies across the platform. It aggregates device/IP activity, conversation sessions, and applies intelligent detection rules to identify suspicious behavioral patterns.

## Architecture

### Components

1. **Data Aggregation Layer** (`service/securityanalytics/aggregation.go`)
   - Device Activity Aggregation
   - IP Activity Aggregation (with sliding time windows)
   - Conversation Linkage

2. **Anomaly Detection Rules** (`service/securityanalytics/detection.go`)
   - Quota Spike Detection
   - Abnormal Login Ratio Detection
   - High Request Ratio Detection
   - Unusual Device Activity Detection

3. **Detection Engine** (`service/securityanalytics/engine.go`)
   - Background processing loop
   - Concurrent user analysis
   - Deduplication and TTL management

4. **Service Integration** (`service/anomaly.go`)
   - High-level API for anomaly processing
   - Configuration management
   - Lifecycle management (start/stop)

5. **Data Models** (`model/security_anomaly.go`, `model/migrations/20250225_security_anomalies.go`)
   - `SecurityAnomaly`: Records detected anomalies
   - `AnomalyBaseline`: User behavior baselines

6. **REST API** (`controller/anomaly.go`)
   - User anomaly endpoints
   - Admin management endpoints

## Configuration

### Environment Variables

```bash
ANOMALY_DETECTION_ENABLED=true          # Enable anomaly detection
```

### Option Database Keys

Configure via the `/api/anomalies/admin/settings` endpoint or directly in the Option table:

| Key | Default | Description |
|-----|---------|-------------|
| `anomaly_detection_enabled` | false | Master enable/disable |
| `anomaly_detection_interval_seconds` | 3600 | Processing interval (1 hour) |
| `anomaly_detection_window_hours` | 24 | Analysis window (24 hours) |
| `anomaly_quota_spike_percent` | 150 | Quota spike threshold (%) |
| `anomaly_login_ratio_threshold` | 1000 | Max requests per login |
| `anomaly_request_ratio_threshold` | 500 | High request ratio threshold |
| `anomaly_new_device_requests` | 100 | New device activity threshold |
| `anomaly_ip_change_threshold` | 5 | Max IPs per device |

## Anomaly Rules

### 1. Quota Spike Detection

**Rule Type**: `quota_spike`

Detects sudden quota consumption without corresponding API calls.

**How it works**:
- Compares actual quota usage vs. expected quota based on request count
- Uses user baseline to calculate expected quota per request
- Triggers when actual > expected + tolerance

**Severity Calculation**:
- Low: < 100% deviation
- Medium: 100-200% deviation
- High: 200-400% deviation
- Critical: > 400% deviation

**Example Evidence**:
```json
{
  "expected_quota": 1000,
  "actual_quota": 2500,
  "request_count": 100,
  "deviation_percent": 150,
  "baseline_value": 10
}
```

### 2. Abnormal Login Ratio Detection

**Rule Type**: `abnormal_login_ratio`

Detects abnormal login frequency vs. API usage ratio.

**How it works**:
- Calculates requests-per-login ratio
- Compares against user baseline (default: 100 requests per login)
- Triggers when ratio exceeds 2x baseline

**Example Evidence**:
```json
{
  "request_count": 10000,
  "login_count": 1,
  "actual_ratio": 10000,
  "baseline_ratio": 100,
  "deviation_percent": 9900
}
```

### 3. High Request Ratio Detection

**Rule Type**: `high_request_ratio`

Detects unusually high requests-to-login ratio.

**How it works**:
- Checks if request count exceeds configurable threshold
- Special handling for sessions with no logins
- Configurable threshold (default: 500 requests per login)

**Example**:
- 1000 requests with 1 login = ratio 1000 (triggers if threshold is 500)
- 1000 requests with 0 logins = suspicious (triggers if > 1000 requests)

### 4. Unusual Device Activity Detection

**Rule Type**: `unusual_device_activity`

Detects suspicious device behavior patterns.

**How it works**:
- Identifies new devices with high activity (> 100 requests within 1 hour of first appearance)
- Detects devices switching between many IPs (> 5 different IPs)

**Example Evidence**:
```json
{
  "device_id": "new_device_abc123",
  "request_count": 150,
  "unique_ips": 6,
  "ips": ["192.168.1.1", "192.168.1.2", ...],
  "first_seen": "2025-02-25T10:00:00Z",
  "last_seen": "2025-02-25T10:30:00Z"
}
```

## Data Aggregation

### Device Aggregation

Groups logs by normalized device_id and user_id:

```go
type DeviceAggregationResult struct {
    NormalizedDeviceId string
    UserId             int
    RequestCount       int64
    UniqueIPs          []string
    UniqueModels       []string
    LastSeenAt         time.Time
    FirstSeenAt        time.Time
}
```

**Usage**:
```go
devices, err := AggregateDeviceActivity(userId, startTime, endTime)
```

### IP Aggregation

Aggregates activity per IP address with sliding time windows:

```go
type IPAggregationWindow struct {
    IP               string
    UserId           int
    RequestCount     int64
    UniqueDevices    int64
    UniqueModels     []string
    ASN              string
    Subnet           string
    LastActivityTime time.Time
}
```

**NAT Handling**:
- Extracts subnet information (e.g., 192.168.1.0/24 from 192.168.1.1)
- Groups users by ASN and subnet for NAT relaxation
- Supports both IPv4 and IPv6

**Usage**:
```go
ips, err := AggregateIPActivity(userId, startTime, endTime)
```

### Conversation Linkage

Links conversation sessions to request logs by grouping consecutive requests:

```go
type ConversationLinkageResult struct {
    ConversationId    string
    RequestIds        []string
    UserId            int
    StartTime         time.Time
    EndTime           time.Time
    RequestCount      int64
    QuotaUsed         int64
    Models            []string
}
```

**Session Detection**:
- Groups requests within 30-minute windows
- Creates separate sessions if gap > 30 minutes
- Tracks models and quota per session

**Usage**:
```go
sessions, err := LinkConversationsWithRequests(userId, startTime, endTime)
```

## API Endpoints

### User Endpoints

#### Get User Anomalies
```
GET /api/anomalies/
Authorization: Bearer <token>
Query Parameters:
  - page: 1
  - limit: 20
  - rule_type: (optional) quota_spike, abnormal_login_ratio, etc.
  - severity: (optional) low, medium, high, critical

Response:
{
  "data": [
    {
      "id": 1,
      "user_id": 123,
      "rule_type": "quota_spike",
      "severity": "high",
      "message": "Quota spike detected...",
      "evidence": {...},
      "detected_at": "2025-02-25T10:00:00Z",
      "is_resolved": false
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 5
  }
}
```

#### Get Anomaly Statistics
```
GET /api/anomalies/statistics
Authorization: Bearer <token>
Query Parameters:
  - days: 7 (default)

Response:
{
  "data": {
    "total_count": 10,
    "unique_users": 5,
    "by_rule_type": [
      {"rule_type": "quota_spike", "count": 6},
      {"rule_type": "high_request_ratio", "count": 4}
    ],
    "by_severity": [
      {"severity": "high", "count": 7},
      {"severity": "medium", "count": 3}
    ]
  },
  "period": {
    "start_time": "2025-02-18T00:00:00Z",
    "end_time": "2025-02-25T00:00:00Z",
    "days": 7
  }
}
```

#### Resolve Anomaly
```
POST /api/anomalies/:id/resolve
Authorization: Bearer <token>

Response:
{
  "message": "anomaly resolved"
}
```

### Admin Endpoints

#### Get All Anomalies
```
GET /api/anomalies/admin/
Authorization: Bearer <admin_token>
Query Parameters:
  - page: 1
  - limit: 20
  - user_id: (optional)
  - rule_type: (optional)
  - severity: (optional)
```

#### Get Anomaly Settings
```
GET /api/anomalies/admin/settings
Authorization: Bearer <admin_token>

Response:
{
  "data": {
    "enabled": true,
    "interval_seconds": "3600",
    "window_hours": "24",
    "quota_spike_percent": "150",
    "login_ratio_threshold": "1000",
    "request_ratio_threshold": "500",
    "new_device_requests": "100",
    "ip_change_threshold": "5"
  }
}
```

#### Update Anomaly Settings
```
PUT /api/anomalies/admin/settings
Authorization: Bearer <admin_token>
Content-Type: application/json

Request Body:
{
  "detection_enabled": true,
  "detection_interval_seconds": 3600,
  "quota_spike_percent": 150
}

Response:
{
  "message": "settings updated"
}
```

#### Process User Anomalies
```
POST /api/anomalies/admin/users/:user_id/process
Authorization: Bearer <admin_token>

Response:
{
  "user_id": 123,
  "anomalies": [...],
  "count": 3
}
```

## Baseline Management

### Automatic Baseline Updates

Baselines are calculated and stored automatically:

1. **On User Activity**: After analyzing a user
2. **Rolling Window**: 30-day historical data
3. **Metrics Tracked**:
   - `quota_usage`: Total quota consumed
   - `login_ratio`: Requests per login
   - `request_count`: Total requests

### Manual Baseline Update

```go
import "github.com/QuantumNous/new-api/service"

err := service.UpdateAnomalyBaseline(userId)
```

### Baseline Structure

```sql
CREATE TABLE anomaly_baselines (
  id INT PRIMARY KEY AUTO_INCREMENT,
  user_id INT NOT NULL,
  metric_type VARCHAR(100) NOT NULL,
  baseline_value FLOAT,
  standard_deviation FLOAT,
  window_size_seconds INT,
  sample_size INT,
  last_updated_at TIMESTAMP,
  created_at TIMESTAMP,
  UNIQUE KEY uk_user_metric (user_id, metric_type)
);
```

## Deduplication

### TTL-Based Deduplication

Anomalies are deduplicated based on:
- User ID
- Rule type
- Time window (1 hour)

### TTL Cleanup

Resolved anomalies are automatically cleaned up:
- Low severity: 30 days
- Medium severity: 14 days
- High severity: 7 days
- Critical severity: 24 hours

```go
// Manual cleanup
count, err := service.CleanupExpiredAnomalies()
```

## Background Processing

### Scheduling

The engine runs on a configurable interval (default: every hour) and:

1. Fetches active users with recent activity (last 7 days)
2. Analyzes each user with configured detectors
3. Creates anomaly records for detected anomalies
4. Manages baseline updates

### Concurrency Control

- Default concurrency: 5 workers
- Configurable via engine initialization
- Uses semaphore pattern for fair resource sharing

### Monitoring

Check engine status and statistics:

```bash
# Get anomaly statistics
curl -H "Authorization: Bearer <token>" \
  https://api.example.com/api/anomalies/statistics?days=7

# Admin: Get all anomalies
curl -H "Authorization: Bearer <admin_token>" \
  https://api.example.com/api/anomalies/admin/
```

## NAT Considerations

### Subnet-Based Clustering

The engine accounts for NAT by:

1. **Subnet Extraction**: Derives /24 subnet from IPv4 addresses
   - Example: 192.168.1.1 → 192.168.1.0/24

2. **IPv6 Support**: Derives /64 subnet from IPv6 addresses
   - Example: 2001:db8::1 → 2001:db8::/64

3. **Relaxation Heuristics**:
   - Same subnet = less suspicious
   - Different ASN but same device = may be legitimate (e.g., mobile + WiFi)
   - ASN information stored for future GeoIP enrichment

### Future Enhancements

- Integration with GeoIP/ASN lookup services
- ML-based behavioral baseline learning
- Configurable NAT relaxation rules

## Testing

### Unit Tests

Run unit tests for detectors:
```bash
cd service/securityanalytics
go test -v detection_test.go detection.go
```

### Integration Tests

Database-dependent tests (marked with `t.Skip()`):
```bash
cd service/securityanalytics
go test -v engine_integration_test.go -run TestAggregationQueryResults
```

### Manual Testing

1. Enable anomaly detection:
```bash
curl -X PUT https://api.example.com/api/anomalies/admin/settings \
  -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{"detection_enabled": true}'
```

2. Process specific user:
```bash
curl -X POST https://api.example.com/api/anomalies/admin/users/123/process \
  -H "Authorization: Bearer <admin_token>"
```

3. View detected anomalies:
```bash
curl https://api.example.com/api/anomalies/admin/ \
  -H "Authorization: Bearer <admin_token>"
```

## Performance Considerations

### Query Optimization

- Uses indexed fields: `user_id`, `created_at`, `rule_type`, `severity`
- Aggregation queries optimized with GROUP BY
- IP aggregation filters on non-empty IP field

### Memory Management

- Streaming log queries to avoid loading entire datasets
- Goroutine concurrency limits (default: 5)
- TTL cleanup runs on schedule

### Database Indexes

Required indexes for optimal performance:
```sql
-- In logs table
CREATE INDEX idx_logs_user_created ON logs(user_id, created_at);
CREATE INDEX idx_logs_ip ON logs(ip);

-- In security_anomalies table
CREATE INDEX idx_anomalies_user ON security_anomalies(user_id);
CREATE INDEX idx_anomalies_detected ON security_anomalies(detected_at);
CREATE INDEX idx_anomalies_rule ON security_anomalies(rule_type);
CREATE INDEX idx_anomalies_resolved ON security_anomalies(is_resolved);
```

## Troubleshooting

### No anomalies detected

1. Verify engine is enabled:
```bash
curl https://api.example.com/api/anomalies/admin/settings
```

2. Check logs:
```bash
grep "anomaly engine" /var/log/application.log
```

3. Ensure users have sufficient activity
4. Check baseline data exists

### High false positive rate

1. Adjust thresholds via settings API
2. Increase deviation tolerance
3. Update baselines manually for specific users

### Performance issues

1. Reduce concurrency in engine startup
2. Increase processing interval
3. Add database indexes
4. Monitor goroutine count

## Future Enhancements

- [ ] ML-based anomaly scoring
- [ ] GeoIP/ASN integration
- [ ] Real-time alerts via webhooks
- [ ] Custom rule creation UI
- [ ] Anomaly trend analysis
- [ ] Automated response actions (throttle, warn, block)
- [ ] Integration with SIEM systems
- [ ] Kafka-based event streaming
