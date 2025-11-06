# Security APIs and Heimdall Auto-Enforcement

This document describes the enhanced security APIs, automated enforcement system, and Heimdall integration for real-time blocking and protection.

## Overview

The security system now includes:
- **Device Fingerprinting** - Track and cluster suspicious devices
- **IP Clustering** - Identify and monitor suspicious IP addresses
- **Anomaly Detection** - Detect and track security anomalies with severity levels
- **Automated Enforcement** - Auto-trigger defensive actions (block/redirect/ban)
- **Heimdall Middleware** - Real-time enforcement of security directives
- **Manual Review** - Approve or ignore anomalies with audit trail

## Architecture

```
┌─────────────────┐
│  User Request   │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ TokenAuth       │ ◄─── Authenticate user
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Heimdall        │ ◄─── Check blocklist (Redis)
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Governance      │ ◄─── Detect violations
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Create Anomaly  │ ◄─── Log security event
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Auto-Enforce    │ ◄─── Trigger action if high severity
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Update Redis    │ ◄─── Publish directive to Heimdall
└─────────────────┘
```

## API Endpoints

### Dashboard

**GET /api/security/dashboard**

Returns comprehensive security metrics including:
- Total violations
- Unique users with violations
- Today's violation count
- Top keywords matched
- Daily trends
- Device clusters (top 10 suspicious)
- Suspicious IPs (top 10)
- Anomaly counts by severity
- Anomaly trends

Query Parameters:
- `start_time` (optional): RFC3339 timestamp (default: 7 days ago)
- `end_time` (optional): RFC3339 timestamp (default: now)

Response:
```json
{
  "success": true,
  "data": {
    "total_count": 150,
    "unique_users": 25,
    "today_count": 10,
    "top_keywords": [...],
    "daily_trend": [...],
    "device_clusters": [...],
    "suspicious_ips": [...],
    "anomaly_counts": {
      "malicious": 15,
      "violation": 135
    },
    "anomaly_trends": [...]
  }
}
```

### Device Clusters

**GET /api/security/devices**

Returns device fingerprint clusters with risk scores.

Query Parameters:
- `page` (default: 1)
- `page_size` (default: 10)
- `blocked` (optional): "true" or "false"

Response:
```json
{
  "success": true,
  "data": {
    "devices": [
      {
        "id": 1,
        "user_id": 123,
        "fingerprint": "abc123...",
        "user_agent": "Mozilla/5.0...",
        "ip_address": "192.168.1.1",
        "first_seen_at": "2024-01-01T00:00:00Z",
        "last_seen_at": "2024-01-02T00:00:00Z",
        "request_count": 150,
        "is_blocked": false,
        "risk_score": 75
      }
    ],
    "total": 100,
    "page": 1,
    "page_size": 10
  }
}
```

### IP Clusters

**GET /api/security/ip-clusters**

Returns IP clusters with suspicious activity metrics.

Query Parameters:
- `page` (default: 1)
- `page_size` (default: 10)
- `blocked` (optional): "true" or "false"
- `min_risk_score` (default: 0)

Response:
```json
{
  "success": true,
  "data": {
    "clusters": [
      {
        "id": 1,
        "ip_address": "192.168.1.1",
        "country": "US",
        "city": "San Francisco",
        "unique_users": 50,
        "total_requests": 1000,
        "violation_count": 10,
        "is_blocked": false,
        "risk_score": 60,
        "first_seen_at": "2024-01-01T00:00:00Z",
        "last_seen_at": "2024-01-02T00:00:00Z"
      }
    ],
    "total": 50,
    "page": 1,
    "page_size": 10
  }
}
```

### Security Anomalies

**GET /api/security/anomalies**

Returns detected anomalies with filtering options.

Query Parameters:
- `page` (default: 1)
- `page_size` (default: 10)
- `user_id` (optional): Filter by user
- `severity` (optional): "malicious" or "violation"
- `anomaly_type` (optional): Type of anomaly
- `status` (optional): "pending", "actioned", "approved", "ignored"
- `start_time` (optional): RFC3339 timestamp
- `end_time` (optional): RFC3339 timestamp

Response:
```json
{
  "success": true,
  "data": {
    "anomalies": [
      {
        "id": 1,
        "user_id": 123,
        "token_id": 456,
        "detected_at": "2024-01-01T12:00:00Z",
        "anomaly_type": "high_rpm",
        "severity": "malicious",
        "description": "Excessive request rate detected",
        "metadata": "{\"rpm\":150}",
        "ip_address": "192.168.1.1",
        "device_id": "device-fp-123",
        "risk_score": 85,
        "action_taken": "ban",
        "actioned_at": "2024-01-01T12:00:01Z",
        "status": "actioned",
        "reviewed_by": null,
        "reviewed_at": null,
        "review_decision": "",
        "review_rationale": ""
      }
    ],
    "total": 100,
    "page": 1,
    "page_size": 10
  }
}
```

### Manual Override

**POST /api/security/anomalies/:id/approve**

Approve an anomaly (mark as legitimate activity).

Request Body:
```json
{
  "rationale": "Verified user, false positive"
}
```

Response:
```json
{
  "success": true,
  "message": "Anomaly approved successfully"
}
```

**POST /api/security/anomalies/:id/ignore**

Ignore an anomaly and rollback any actions taken.

Request Body:
```json
{
  "rationale": "Testing activity, can be ignored"
}
```

Response:
```json
{
  "success": true,
  "message": "Anomaly ignored successfully"
}
```

## Automated Enforcement

### Configuration

Update security settings to enable auto-enforcement:

**PUT /api/security/settings**

```json
{
  "auto_enforcement_enabled": true,
  "auto_ban_enabled": true,
  "auto_block_enabled": true,
  "auto_ban_threshold": 10,
  "violation_redirect_model": "gpt-3.5-turbo"
}
```

### Enforcement Actions

Based on severity and risk score, the system automatically takes actions:

| Risk Score | Action     | Description                                  |
|------------|------------|----------------------------------------------|
| > 80       | Ban        | User banned, all requests blocked            |
| 51-80      | Block      | Temporary block via Heimdall                 |
| 31-50      | Redirect   | Requests redirected to safe model            |
| ≤ 30       | Log        | Only logged, no enforcement                  |

### Heimdall Integration

Heimdall middleware enforces security directives in real-time:

1. **Detection**: Governance middleware detects violation
2. **Anomaly Creation**: Security anomaly is created with severity
3. **Auto-Enforcement**: If severity is "malicious", trigger action
4. **Redis Directive**: Directive published to Redis
5. **Heimdall Check**: Next request checked against blocklist
6. **Block/Allow**: Request either blocked or allowed to proceed

Redis Keys:
- `heimdall:directive:{user_id}` - User-specific directive
- Channel: `heimdall:directives` - Pub/sub for real-time updates

## Database Schema

### device_fingerprints

```sql
CREATE TABLE device_fingerprints (
  id INT PRIMARY KEY AUTO_INCREMENT,
  user_id INT NOT NULL,
  fingerprint VARCHAR(256) NOT NULL,
  user_agent TEXT,
  ip_address VARCHAR(45),
  first_seen_at TIMESTAMP NOT NULL,
  last_seen_at TIMESTAMP NOT NULL,
  request_count INT DEFAULT 0,
  is_blocked BOOLEAN DEFAULT FALSE,
  risk_score INT DEFAULT 0,
  INDEX idx_user_id (user_id),
  INDEX idx_fingerprint (fingerprint),
  INDEX idx_ip_address (ip_address),
  INDEX idx_is_blocked (is_blocked)
);
```

### ip_clusters

```sql
CREATE TABLE ip_clusters (
  id INT PRIMARY KEY AUTO_INCREMENT,
  ip_address VARCHAR(45) UNIQUE NOT NULL,
  country VARCHAR(2),
  city VARCHAR(100),
  unique_users INT DEFAULT 0,
  total_requests INT DEFAULT 0,
  violation_count INT DEFAULT 0,
  is_blocked BOOLEAN DEFAULT FALSE,
  risk_score INT DEFAULT 0,
  first_seen_at TIMESTAMP NOT NULL,
  last_seen_at TIMESTAMP NOT NULL,
  INDEX idx_is_blocked (is_blocked)
);
```

### security_anomalies

```sql
CREATE TABLE security_anomalies (
  id INT PRIMARY KEY AUTO_INCREMENT,
  user_id INT NOT NULL,
  token_id INT,
  detected_at TIMESTAMP NOT NULL,
  anomaly_type VARCHAR(50),
  severity VARCHAR(20),
  description TEXT,
  metadata TEXT,
  ip_address VARCHAR(45),
  device_id VARCHAR(256),
  risk_score INT DEFAULT 0,
  action_taken VARCHAR(50),
  actioned_at TIMESTAMP,
  status VARCHAR(20) DEFAULT 'pending',
  reviewed_by INT,
  reviewed_at TIMESTAMP,
  review_decision VARCHAR(20),
  review_rationale TEXT,
  INDEX idx_user_id (user_id),
  INDEX idx_token_id (token_id),
  INDEX idx_detected_at (detected_at),
  INDEX idx_severity (severity),
  INDEX idx_status (status),
  INDEX idx_ip_address (ip_address),
  INDEX idx_device_id (device_id)
);
```

## Testing

### Unit Tests

Run security enforcement tests:
```bash
go test ./service -run TestSecurity
go test ./controller -run TestSecurity
```

### Integration Tests

Test the full workflow:
```bash
go test ./middleware -run TestHeimdall
```

### Manual Testing

1. **Create an anomaly**:
```bash
curl -X POST http://localhost:3000/api/security/anomalies \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 123,
    "anomaly_type": "test",
    "severity": "malicious",
    "description": "Test anomaly",
    "risk_score": 90
  }'
```

2. **Verify enforcement**: Try making a request as that user - should be blocked

3. **Check dashboard**:
```bash
curl http://localhost:3000/api/security/dashboard \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

4. **Review and ignore**:
```bash
curl -X POST http://localhost:3000/api/security/anomalies/1/ignore \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"rationale": "False positive"}'
```

## Monitoring

### Key Metrics

Monitor these Redis keys:
- `heimdall:directive:*` - Active enforcement directives
- Channel `heimdall:directives` - Real-time directive stream

### Logs

Check logs for:
- `heimdall block` - User blocked by Heimdall
- `governance flag triggered` - Violation detected
- `security action taken` - Enforcement action executed

## Best Practices

1. **Tune Risk Scores**: Adjust risk score thresholds based on your use case
2. **Review Regularly**: Check pending anomalies and approve/ignore as needed
3. **Monitor False Positives**: Track approval/ignore ratios
4. **Configure Redis**: Ensure Redis is enabled for real-time enforcement
5. **Audit Trail**: Review `reviewed_by` and `review_rationale` fields
6. **Alert on High Severity**: Set up notifications for malicious anomalies

## Troubleshooting

### Heimdall not blocking users

1. Check Redis is enabled: `REDIS_CONN_STRING` environment variable
2. Verify directive in Redis: `redis-cli GET heimdall:directive:{user_id}`
3. Check middleware is loaded in relay router

### Auto-enforcement not triggering

1. Verify settings: `GET /api/security/settings`
2. Check anomaly severity is "malicious"
3. Ensure `auto_enforcement_enabled` is true

### False positives

1. Review anomaly metadata for detection details
2. Adjust governance keyword policies
3. Use manual ignore to prevent future similar blocks
4. Consider increasing risk score thresholds
