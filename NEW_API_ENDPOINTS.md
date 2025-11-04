# New API Endpoints Reference

## Group Rate Scheduling

### Create Schedule Rule
```
POST /api/admin/group-rate-schedule
Authorization: Bearer {admin_token}
Content-Type: application/json

{
  "group_name": "default",
  "time_start": "09:00",
  "time_end": "18:00",
  "rate_multiplier": 1.5,
  "enabled": true
}
```

### Get Schedules for Group
```
GET /api/admin/group-rate-schedule/group?group_name=default
Authorization: Bearer {admin_token}
```

### Get Current Effective Rate
```
GET /api/admin/group-rate-schedule/current?group_name=default
Authorization: Bearer {admin_token}
```

### Update Schedule Rule
```
PUT /api/admin/group-rate-schedule
Authorization: Bearer {admin_token}
Content-Type: application/json

{
  "id": 1,
  "time_start": "08:00",
  "time_end": "20:00",
  "rate_multiplier": 2.0,
  "enabled": true
}
```

### Delete Schedule Rule
```
DELETE /api/admin/group-rate-schedule/:id
Authorization: Bearer {admin_token}
```

### Get All Schedules (Paginated)
```
GET /api/admin/group-rate-schedule?page=1&page_size=20
Authorization: Bearer {admin_token}
```

### Force Update Group Rates
```
POST /api/admin/group-rate-schedule/update
Authorization: Bearer {admin_token}
```

## IP Protection

### Add IP to Blacklist
```
POST /api/admin/ip-protection/blacklist
Authorization: Bearer {admin_token}
Content-Type: application/json

{
  "ip": "192.168.1.100",
  "reason": "Suspicious activity",
  "scope": "global",
  "expires_at": "2024-12-31T23:59:59Z"  // optional
}
```

### Add IP to Whitelist
```
POST /api/admin/ip-protection/whitelist
Authorization: Bearer {admin_token}
Content-Type: application/json

{
  "ip": "10.0.0.0/24",  // Supports CIDR notation
  "reason": "Internal network",
  "scope": "global"
}
```

### Remove IP from List
```
DELETE /api/admin/ip-protection/list/:id
Authorization: Bearer {admin_token}
```

### Get IP Lists
```
GET /api/admin/ip-protection/list?list_type=blacklist&page=1&page_size=20
Authorization: Bearer {admin_token}

list_type: "blacklist" or "whitelist" (optional, returns both if not specified)
```

### Create Rate Limit Rule
```
POST /api/admin/ip-protection/rate-limit
Authorization: Bearer {admin_token}
Content-Type: application/json

{
  "name": "Aggressive IP limit",
  "ip": "192.168.1.100",
  "max_requests": 100,
  "time_window": 60,  // seconds
  "action": "reject",  // "reject", "warn", or "ban"
  "enabled": true
}
```

### Update Rate Limit Rule
```
PUT /api/admin/ip-protection/rate-limit
Authorization: Bearer {admin_token}
Content-Type: application/json

{
  "id": 1,
  "max_requests": 200,
  "enabled": false
}
```

### Delete Rate Limit Rule
```
DELETE /api/admin/ip-protection/rate-limit/:id
Authorization: Bearer {admin_token}
```

### Get Rate Limit Rules
```
GET /api/admin/ip-protection/rate-limit?page=1&page_size=20
Authorization: Bearer {admin_token}
```

### Ban IP
```
POST /api/admin/ip-protection/ban
Authorization: Bearer {admin_token}
Content-Type: application/json

{
  "ip": "192.168.1.100",
  "reason": "Rate limit exceeded multiple times",
  "ban_type": "temporary",  // "temporary" or "permanent"
  "duration": 24  // hours (for temporary bans)
}
```

### Unban IP
```
POST /api/admin/ip-protection/unban
Authorization: Bearer {admin_token}
Content-Type: application/json

{
  "ip": "192.168.1.100"
}
```

### Get Banned IPs
```
GET /api/admin/ip-protection/banned?page=1&page_size=20&include_expired=false
Authorization: Bearer {admin_token}
```

### Get IP Protection Statistics
```
GET /api/admin/ip-protection/stats
Authorization: Bearer {admin_token}

Response:
{
  "success": true,
  "data": {
    "active_bans": 5,
    "blacklist_count": 10,
    "whitelist_count": 3,
    "rate_limit_rules": 7,
    "blacklist": [...],
    "whitelist": [...],
    "limits": [...]
  }
}
```

## IP Usage Tracking

### Get Token IP Usage
```
GET /api/log/ip-usage/token/:id?since=2024-01-01T00:00:00Z
Authorization: Bearer {admin_token}

since: RFC3339 timestamp (optional)

Response:
{
  "success": true,
  "data": {
    "usages": [
      {
        "ip": "192.168.1.100",
        "first_seen_at": "2024-01-01T10:00:00Z",
        "last_seen_at": "2024-01-15T15:30:00Z",
        "request_count": 1250
      }
    ],
    "total_requests": 1250
  }
}
```

### Get User IP Usage
```
GET /api/log/ip-usage/user/:id?since=2024-01-01T00:00:00Z
Authorization: Bearer {admin_token}

since: RFC3339 timestamp (optional)
```

## Check-in System

### Perform Check-in
```
POST /api/user/checkin
Authorization: Bearer {user_token}

Response:
{
  "success": true,
  "message": "签到成功",
  "data": {
    "quota_awarded": 100000,
    "consecutive_days": 5
  }
}
```

### Get Check-in Status
```
GET /api/user/checkin/status
Authorization: Bearer {user_token}

Response:
{
  "success": true,
  "data": {
    "has_checked_in": true,
    "today_record": {
      "id": 123,
      "user_id": 1,
      "check_in_date": "2024-01-15",
      "quota_awarded": 100000,
      "consecutive_days": 5
    },
    "consecutive_days": 5
  }
}
```

### Get Check-in History
```
GET /api/user/checkin/history?page=1&page_size=30
Authorization: Bearer {user_token}

Response:
{
  "success": true,
  "data": {
    "items": [...],
    "total": 100,
    "page": 1,
    "page_size": 30
  }
}
```

### Get Check-in Reward Configuration
```
GET /api/user/checkin/rewards
Authorization: Bearer {user_token}

Response:
{
  "success": true,
  "data": [
    {
      "day_range": "1",
      "reward_quota": 100000,
      "description": "第1天",
      "is_bonus": false
    },
    {
      "day_range": "7-13",
      "reward_quota": 200000,
      "description": "第7-13天",
      "is_bonus": false
    },
    ...
  ]
}
```

## Playground Statistics

### Get Playground IP Statistics
```
GET /api/playground/ip-stats?hours=24
Authorization: Bearer {admin_token}

hours: Time range in hours (1-720, default: 24)

Response:
{
  "success": true,
  "data": {
    "total_ips": 45,
    "suspicious_ips": 3,  // IPs with >5 users
    "time_range": 24,
    "ip_stats": [
      {
        "ip": "192.168.1.100",
        "user_count": 8,
        "usernames": ["user1", "user2", "user3", ...],
        "last_active_at": "2024-01-15T15:30:00Z",
        "total_requests": 1250
      },
      ...
    ]
  }
}
```

## Response Format

All endpoints follow this response format:

### Success Response
```json
{
  "success": true,
  "message": "Operation completed successfully",
  "data": {
    // Response data
  }
}
```

### Error Response
```json
{
  "success": false,
  "message": "Error description"
}
```

### Paginated Response
```json
{
  "success": true,
  "message": "",
  "data": {
    "items": [...],
    "total": 100,
    "page": 1,
    "page_size": 20,
    "total_page": 5
  }
}
```

## Authentication

All endpoints require authentication via JWT token in the Authorization header:
```
Authorization: Bearer {token}
```

- Admin endpoints require admin role or higher
- User endpoints require any authenticated user
- Token can be session-based or access token

## Time Format

All timestamps use RFC3339 format:
```
2024-01-15T15:30:00Z
```

## IP Address Format

Supports both single IPs and CIDR notation:
- Single IP: `192.168.1.100`
- IPv6: `2001:0db8:85a3:0000:0000:8a2e:0370:7334`
- CIDR: `10.0.0.0/24`
- IPv6 CIDR: `2001:0db8::/32`

## Rate Limiting Headers

When rate limiting is active, responses include these headers:
```
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 2024-01-15T16:00:00Z
```

## Auto-ban Thresholds

IP protection includes automatic banning based on violations:
- 20 violations → 1 hour temporary ban
- 50 violations → 24 hour temporary ban
- 100 violations → Permanent ban

Violations are tracked for 1 hour windows.
