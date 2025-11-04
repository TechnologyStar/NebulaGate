# Security Center & UX Upgrade Implementation

## Overview
This document describes the implementation of the comprehensive security and UX upgrade for the new-api system, covering 5 major modules.

## Implemented Features

### Module 1: Security Center - Key IP Tracking ✅

**Database Models:**
- `TokenIPUsage` - Already existed, tracks IP usage per token
- `UserIPUsage` - Already existed, tracks IP usage per user

**Backend APIs:**
- `GET /api/log/ip-usage/token/:id` - Get IP usage for a specific token
- `GET /api/log/ip-usage/user/:id` - Get IP usage for a specific user
- Supports `since` parameter for time-range filtering (RFC3339 format)

**Features:**
- Automatic IP recording in TokenAuth middleware
- IP tracking with first seen, last seen, and request count
- Support for X-Forwarded-For, X-Real-IP headers for proxy scenarios

### Module 2: Group Rate Scheduling System ✅

**Database Models:**
- `GroupRateSchedule` - Time-based rate multiplier rules for groups

**Backend APIs:**
- `POST /api/admin/group-rate-schedule` - Create schedule rule
- `GET /api/admin/group-rate-schedule/group?group_name=xxx` - Get schedules for group
- `GET /api/admin/group-rate-schedule/current?group_name=xxx` - Get current effective rate
- `PUT /api/admin/group-rate-schedule` - Update schedule rule
- `DELETE /api/admin/group-rate-schedule/:id` - Delete schedule rule
- `GET /api/admin/group-rate-schedule` - Get all schedules (paginated)
- `POST /api/admin/group-rate-schedule/update` - Force update all group rates

**Features:**
- Time-based rate multipliers (HH:MM format)
- Support for cross-midnight time ranges (e.g., 22:00-02:00)
- Redis caching with 5-minute TTL
- Automatic rate updates every minute via background scheduler
- Rate multiplier retrieval service for billing integration

**Background Services:**
- `StartGroupRateScheduler()` - Updates rates every minute
- Cleanup of expired schedules

### Module 3: IP Protection & Rate Limiting ✅

**Database Models:**
- `IPList` - Blacklist/whitelist entries with expiration support
- `IPRateLimit` - Rate limiting rules per IP
- `IPBan` - IP ban records (temporary/permanent)

**Backend APIs:**
- Blacklist/Whitelist Management:
  - `POST /api/admin/ip-protection/blacklist` - Add IP to blacklist
  - `POST /api/admin/ip-protection/whitelist` - Add IP to whitelist
  - `DELETE /api/admin/ip-protection/list/:id` - Remove IP from list
  - `GET /api/admin/ip-protection/list?list_type=xxx` - Get IP lists

- Rate Limiting:
  - `POST /api/admin/ip-protection/rate-limit` - Create rate limit rule
  - `PUT /api/admin/ip-protection/rate-limit` - Update rate limit rule
  - `DELETE /api/admin/ip-protection/rate-limit/:id` - Delete rate limit rule
  - `GET /api/admin/ip-protection/rate-limit` - Get all rate limit rules

- IP Banning:
  - `POST /api/admin/ip-protection/ban` - Ban an IP (temporary/permanent)
  - `POST /api/admin/ip-protection/unban` - Unban an IP
  - `GET /api/admin/ip-protection/banned` - Get banned IPs list

- Statistics:
  - `GET /api/admin/ip-protection/stats` - Get IP protection statistics

**Features:**
- IP validation with CIDR support
- Automatic expiration of temporary bans and list entries
- Whitelist bypass for rate limiting
- Redis-based rate limiting with sliding window
- Auto-ban mechanism based on violation thresholds:
  - 20 violations → 1 hour ban
  - 50 violations → 24 hour ban
  - 100 violations → permanent ban
- Hourly cleanup of expired bans and list entries

**Middleware:**
- `IPProtection()` - Checks blacklist, whitelist, and bans
- `IPRateLimit()` - Enforces IP-based rate limiting
- `GetClientIP()` - Extracts real client IP from headers

### Module 4: Daily Check-in System ✅

**Database Models:**
- `CheckInRecord` - Already existed

**Backend APIs:**
- `POST /api/user/checkin` - Perform daily check-in
- `GET /api/user/checkin/status` - Get today's check-in status
- `GET /api/user/checkin/history` - Get check-in history (paginated)
- `GET /api/user/checkin/rewards` - Get reward configuration

**Features:**
- Daily check-in with quota rewards
- Consecutive day tracking (resets if skipped)
- Unique constraint prevents duplicate check-ins
- Transaction-based quota award
- Reward tiers:
  - Day 1: 100,000 quota
  - Days 2-6: 100,000 quota/day
  - Days 7-13: 200,000 quota/day
  - Days 14-29: 300,000 quota/day
  - Day 30+: 500,000 quota/day (bonus tier)
- Automatic log creation for check-in rewards

### Module 5: Playground Enhancements ✅

**Backend APIs:**
- `GET /api/playground/ip-stats?hours=24` - Get IP user statistics (admin only)

**Features:**
- IP-based user statistics for playground usage
- Configurable time range (1-720 hours)
- Groups users by IP address
- Identifies suspicious IPs (>5 users per IP)
- Returns:
  - Total IPs count
  - Suspicious IPs count
  - Per-IP statistics: user count, usernames, last active time, total requests
  - Results sorted by user count (descending)

## Technical Implementation

### Database Migrations
All new models are automatically migrated via GORM AutoMigrate in `model/main.go`:
- `GroupRateSchedule`
- `IPList`
- `IPRateLimit`
- `IPBan`

### Redis Caching
- Group rate multipliers: `group:rate:current:{group_name}` (5 min TTL)
- IP rate limits: `rate_limit:ip:{ip}` (custom TTL per rule)
- IP violations: `ip_violations:{ip}` (1 hour TTL)

### Background Services
Started in `main.go`:
- Group rate scheduler (1 minute interval)
- IP ban/list cleanup (1 hour interval)

### Security
- All admin endpoints require `AdminAuth()` middleware
- User endpoints require `UserAuth()` middleware
- IP extraction supports proxy headers (X-Forwarded-For, X-Real-IP, CF-Connecting-IP)
- Transaction-based operations for data consistency

## API Routes Summary

### Admin Routes
- `/api/admin/group-rate-schedule/*` - Group rate scheduling
- `/api/admin/ip-protection/*` - IP protection management
- `/api/log/ip-usage/*` - IP usage tracking
- `/api/playground/ip-stats` - Playground statistics

### User Routes
- `/api/user/checkin` - Check-in system
- `/api/user/checkin/*` - Check-in related queries

## Testing

To test the implementation:

1. **Group Rate Scheduling:**
```bash
# Create a schedule
curl -X POST http://localhost:3000/api/admin/group-rate-schedule \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "group_name": "default",
    "time_start": "09:00",
    "time_end": "18:00",
    "rate_multiplier": 1.5,
    "enabled": true
  }'

# Get current rate
curl http://localhost:3000/api/admin/group-rate-schedule/current?group_name=default \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN"
```

2. **IP Protection:**
```bash
# Add IP to blacklist
curl -X POST http://localhost:3000/api/admin/ip-protection/blacklist \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "ip": "192.168.1.100",
    "reason": "Suspicious activity",
    "scope": "global"
  }'

# Ban an IP
curl -X POST http://localhost:3000/api/admin/ip-protection/ban \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "ip": "192.168.1.100",
    "reason": "Rate limit exceeded",
    "ban_type": "temporary",
    "duration": 24
  }'
```

3. **Check-in:**
```bash
# Perform check-in
curl -X POST http://localhost:3000/api/user/checkin \
  -H "Authorization: Bearer YOUR_USER_TOKEN"

# Get check-in status
curl http://localhost:3000/api/user/checkin/status \
  -H "Authorization: Bearer YOUR_USER_TOKEN"
```

4. **Playground Stats:**
```bash
# Get IP statistics (admin only)
curl http://localhost:3000/api/playground/ip-stats?hours=24 \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN"
```

## Frontend Integration Notes

For frontend developers, the following endpoints are available:

### Security Center
- Display token/user IP usage in security dashboard
- Add search and filter by time range
- Show IP geolocation (optional, requires external service)

### Group Rate Management
- Time picker for schedule rules (HH:MM format)
- Visual timeline showing 24-hour rate distribution
- Real-time current rate display
- Support for cross-midnight ranges

### IP Protection Dashboard
- IP list management (add/remove with CIDR support)
- Rate limit rule configuration
- Ban management with auto-unban countdown
- Statistics dashboard with charts

### Check-in UI
- Large check-in button with animation
- Consecutive days counter
- Monthly calendar view with check-in markers
- Reward tier display
- Check-in reminder notification

### Playground
- Admin-only IP statistics panel
- Table showing IP → users mapping
- Highlight suspicious IPs (>5 users)
- Time range selector

## Notes

1. All date/time parameters use RFC3339 format for consistency
2. Pagination uses `page` and `page_size` query parameters
3. Redis is required for rate limiting and caching features
4. IP tracking happens automatically in TokenAuth middleware
5. Check-in dates use UTC timezone
6. All admin endpoints require admin authentication
7. Rate multiplier changes take effect within 1 minute via background scheduler
8. Temporary bans and expired list entries are cleaned up hourly

## Future Enhancements

Potential improvements for future iterations:
- IP geolocation service integration
- Advanced rate limiting algorithms (token bucket, leaky bucket)
- Machine learning-based anomaly detection
- Email notifications for security events
- Audit log for all security operations
- Custom check-in reward configuration via admin UI
- Check-in streak recovery mechanism
- Playground UI glassmorphism design implementation
- Real-time notifications via WebSocket
