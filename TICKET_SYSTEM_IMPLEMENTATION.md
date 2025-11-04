# Ticket System Implementation

## Overview
This document describes the implementation of a comprehensive ticket support system with security features, rate limiting, and bilingual support.

## Backend Implementation

### 1. Data Model (`model/ticket.go`)
- **Ticket Structure:**
  - `id`: Primary key
  - `user_id`: Creator ID (indexed, references users table)
  - `title`: Ticket title (VARCHAR 200)
  - `content`: Ticket content (TEXT)
  - `status`: Status (pending/processing/resolved/closed, indexed)
  - `priority`: Priority (low/medium/high/urgent)
  - `category`: Category (technical/account/feature/other)
  - `attachments`: JSON array of attachments (optional)
  - `admin_reply`: Admin reply content
  - `replied_by`: Admin ID who replied
  - `replied_at`: Reply timestamp
  - `created_at`, `updated_at`: Timestamps (created_at indexed)
  - `deleted_at`: Soft delete support

- **Database Functions:**
  - `Insert()`: Create new ticket
  - `Update()`: Update ticket
  - `Delete()`: Soft delete ticket
  - `GetTicketById()`: Get ticket with username loaded
  - `GetTicketsByUserId()`: Paginated user tickets
  - `GetAllTickets()`: Paginated all tickets (admin)
  - `UpdateTicketStatus()`: Update ticket status
  - `UpdateTicketReply()`: Add admin reply

### 2. Validation Layer (`common/ticket_validation.go`)
- **Input Validation:**
  - `ValidateTicketTitle()`: 1-200 characters
  - `ValidateTicketContent()`: 1-5000 characters
  - `ValidateTicketStatus()`: Valid status values
  - `ValidateTicketPriority()`: Valid priority values
  - `ValidateTicketCategory()`: Valid category values

- **Security Functions:**
  - `SanitizeInput()`: XSS prevention
    - HTML entity escaping
    - Script tag removal
    - Event handler removal
    - JavaScript protocol removal

### 3. Service Layer (`service/ticket.go`)
- **Rate Limiting:**
  - Redis-based daily limit tracking
  - Key format: `ticket:daily:{user_id}:{date}`
  - 24-hour expiration
  - Configurable limit (default: 5 tickets/day)
  - Graceful degradation when Redis unavailable

- **Business Logic:**
  - `CreateTicket()`: Validation + rate limiting + sanitization
  - `GetUserTickets()`: User's own tickets (paginated)
  - `GetAllTicketsForAdmin()`: All tickets with filtering (admin only)
  - `GetTicketDetail()`: Permission-checked detail view
  - `UpdateTicketStatus()`: Status updates with permission check
  - `ReplyToTicket()`: Admin reply with auto-status update
  - `DeleteTicket()`: Soft delete with permission check

### 4. Controller Layer (`controller/ticket.go`)
- **Endpoints:**
  - `POST /api/ticket/create`: Create ticket (auth required)
  - `GET /api/ticket/list`: List tickets (auth, pagination, filtering)
  - `GET /api/ticket/:id`: Get ticket detail (auth, permission check)
  - `PUT /api/ticket/:id/reply`: Admin reply (admin auth)
  - `PUT /api/ticket/:id/status`: Update status (auth, permission check)
  - `DELETE /api/ticket/:id`: Delete ticket (auth, permission check)

- **Permission Control:**
  - Regular users: Only their own tickets
  - Admins: All tickets + reply + status management

### 5. Router Configuration (`router/api-router.go`)
- Ticket routes under `/api/ticket` group
- UserAuth middleware for all routes
- AdminAuth middleware for reply endpoint
- Nested route groups for clean organization

### 6. Database Migration (`model/main.go`)
- Added `Ticket{}` to `migrateDB()`
- Added to `migrateDBFast()` for parallel migration
- Auto-creates table on first run
- Indexes automatically created on:
  - `user_id`
  - `status`
  - `created_at`

### 7. System Configuration (`model/option.go`)
- Added `MaxTicketsPerUserPerDay` option
- Default value: 5
- Configurable via admin interface
- Synced to `common.OptionMap`

## Frontend Implementation

### 1. Page Components (`web/src/pages/Ticket/`)

#### `index.jsx`
- Root ticket component
- Exports TicketList as default

#### `TicketList.jsx`
- Displays paginated ticket table
- Features:
  - Status filtering (all/pending/processing/resolved/closed)
  - Color-coded status and priority tags
  - Create ticket button
  - Admin view: Shows all users' tickets
  - User view: Shows only own tickets
  - Actions: View, Close (admin), Delete
  - Click title to view details

#### `TicketCreate.jsx`
- Ticket creation form using Semi Design Form
- Fields:
  - Title (required, max 200 chars)
  - Category (select: technical/account/feature/other)
  - Priority (select: low/medium/high/urgent)
  - Content (textarea, required, max 5000 chars)
- Client-side validation
- XSS prevention (trim, length checks)
- Success redirect to ticket list

#### `TicketDetail.jsx`
- Full ticket information display
- Admin features:
  - Status change buttons (pending/processing/resolved/closed)
  - Reply form (textarea + submit)
  - Auto-updates ticket after reply
- User features:
  - View ticket details
  - See admin reply if present
  - Delete own ticket
- Responsive layout

#### `Ticket.css`
- Modern card-based design
- Color-coded status tags:
  - Pending: Amber
  - Processing: Blue
  - Resolved: Green
  - Closed: Grey
- Color-coded priority tags:
  - Urgent: Red
  - High: Orange
  - Medium: Amber
  - Low: Green
- Responsive design for mobile
- Admin reply section with distinct styling

### 2. Routing (`web/src/App.jsx`)
- Added ticket component imports
- Routes:
  - `/ticket` → TicketList (PrivateRoute)
  - `/ticket/create` → TicketCreate (PrivateRoute)
  - `/ticket/:id` → TicketDetail (PrivateRoute)
- All routes require authentication

### 3. Navigation (`web/src/components/layout/SiderBar.jsx`)
- Added ticket icon mapping (HelpCircle from lucide-react)
- Added to `routerMap`: `ticket: '/ticket'`
- Added to `financeItems` (personal section):
  - Text: "工单系统"
  - ItemKey: "ticket"
  - Route: "/ticket"
- Supports sidebar visibility configuration

### 4. Icon System (`web/src/helpers/render.jsx`)
- Imported `HelpCircle` from lucide-react
- Added icon case for 'ticket' and 'support' keys
- Returns HelpCircle icon with proper styling

### 5. Internationalization

#### Chinese (`web/src/i18n/locales/zh.json`)
Added 60+ translations including:
- UI labels: 工单系统, 创建工单, 工单列表, etc.
- Status: 待处理, 处理中, 已解决, 已关闭
- Priority: 低, 中, 高, 紧急
- Categories: 技术支持, 账号问题, 功能建议, 其他
- Actions: 提交, 取消, 查看, 删除, etc.
- Messages: Success/error messages

#### English (`web/src/i18n/locales/en.json`)
Complete English translations:
- UI labels: Ticket System, Create Ticket, Ticket List, etc.
- Status: Pending, Processing, Resolved, Closed
- Priority: Low, Medium, High, Urgent
- Categories: Technical Support, Account Issue, Feature Request, Other
- Full message translations

## Security Features

### 1. SQL Injection Prevention
- GORM parameterized queries throughout
- No raw SQL with user input
- Automatic escaping by ORM

### 2. XSS Prevention
- `html.EscapeString()` on all user inputs
- Regex-based removal of:
  - `<script>` tags
  - Event handlers (onclick, onerror, etc.)
  - `javascript:` protocol
- Applied before database storage

### 3. Rate Limiting
- Redis-based distributed rate limiting
- Per-user daily quota tracking
- Key format: `ticket:daily:{user_id}:{YYYY-MM-DD}`
- 24-hour TTL (auto-resets)
- Configurable limit (default: 5/day)
- Graceful fallback if Redis unavailable

### 4. Permission Control
- Authentication required for all endpoints
- User isolation: Can only access own tickets
- Admin escalation: Can access all tickets + reply
- Role-based UI rendering
- Permission checks in both controller and service layers

### 5. Input Validation
- Length limits enforced:
  - Title: 1-200 characters
  - Content: 1-5000 characters
- Required field validation
- Enum validation for status/priority/category
- Client-side + server-side validation

### 6. CSRF Protection
- Leverages existing Gin session middleware
- Same-site cookie policy
- Session-based authentication

## Testing Checklist

### Functionality Tests
- ✅ User can create ticket with valid data
- ✅ Rate limiting blocks excess tickets
- ✅ User can view only own tickets
- ✅ Admin can view all tickets
- ✅ Admin can reply to tickets
- ✅ Ticket status updates correctly
- ✅ Pagination works correctly
- ✅ Filtering by status works
- ✅ Soft delete preserves data

### Security Tests
- [ ] SQL Injection: Test with `' OR '1'='1`
- [ ] XSS: Test with `<script>alert('xss')</script>`
- [ ] XSS: Test with `<img src=x onerror=alert(1)>`
- [ ] Rate Limiting: Create 6 tickets rapidly
- [ ] Rate Limiting: Verify reset after 24 hours
- [ ] Permission: User A tries to access User B's ticket ID
- [ ] Permission: Non-admin tries to reply
- [ ] Input Validation: Empty title/content
- [ ] Input Validation: Oversized title (>200 chars)
- [ ] Input Validation: Oversized content (>5000 chars)

### UI Tests
- [ ] Responsive design on mobile
- [ ] Chinese/English language toggle
- [ ] Status color tags display correctly
- [ ] Priority color tags display correctly
- [ ] Form validation messages appear
- [ ] Success/error toasts display
- [ ] Modal confirmations work
- [ ] Sidebar navigation works

## Configuration

### Environment Variables
- `SQL_DSN`: Database connection string (required)
- `REDIS_CONN_STRING`: Redis connection string (for rate limiting)
- `SYNC_FREQUENCY`: Option sync frequency (seconds)

### System Options
- `MaxTicketsPerUserPerDay`: Max tickets per user per day (default: 5)
  - Configurable via admin settings
  - Stored in options table
  - Synced to memory cache

### Redis Keys
- `ticket:daily:{user_id}:{YYYY-MM-DD}`: Daily ticket count
  - Expires: 24 hours
  - Incremented on each ticket creation
  - Used for rate limiting

## API Reference

### Create Ticket
```http
POST /api/ticket/create
Authorization: Session
Content-Type: application/json

{
  "title": "Issue title",
  "content": "Detailed description",
  "priority": "medium",
  "category": "technical"
}
```

### List Tickets
```http
GET /api/ticket/list?page=1&page_size=10&status=pending
Authorization: Session
```

### Get Ticket Detail
```http
GET /api/ticket/:id
Authorization: Session
```

### Admin Reply
```http
PUT /api/ticket/:id/reply
Authorization: Session (Admin)
Content-Type: application/json

{
  "reply": "Response from admin"
}
```

### Update Status
```http
PUT /api/ticket/:id/status
Authorization: Session
Content-Type: application/json

{
  "status": "resolved"
}
```

### Delete Ticket
```http
DELETE /api/ticket/:id
Authorization: Session
```

## Database Schema

```sql
CREATE TABLE tickets (
  id INT PRIMARY KEY AUTO_INCREMENT,
  user_id INT NOT NULL,
  title VARCHAR(200) NOT NULL,
  content TEXT NOT NULL,
  status VARCHAR(20) DEFAULT 'pending',
  priority VARCHAR(20) DEFAULT 'medium',
  category VARCHAR(50) DEFAULT 'other',
  attachments TEXT,
  admin_reply TEXT,
  replied_by INT,
  replied_at DATETIME,
  created_at DATETIME,
  updated_at DATETIME,
  deleted_at DATETIME,
  
  INDEX idx_user_id (user_id),
  INDEX idx_status (status),
  INDEX idx_created_at (created_at),
  INDEX idx_deleted_at (deleted_at)
);
```

## Future Enhancements

### Planned Features
1. File attachment support
2. Email notifications on ticket updates
3. Ticket categories customization
4. SLA tracking and metrics
5. Ticket assignment to specific admins
6. Ticket priority auto-escalation
7. Canned responses for admins
8. Ticket search functionality
9. Export tickets to CSV
10. Ticket statistics dashboard

### Performance Optimizations
1. Redis caching for frequently accessed tickets
2. Database query optimization with proper indexes
3. Pagination improvements for large datasets
4. Lazy loading of ticket content

## Maintenance

### Regular Tasks
- Monitor Redis memory usage
- Review and adjust rate limits based on usage
- Archive old resolved/closed tickets
- Update XSS filter patterns as needed
- Review and update security keywords

### Monitoring
- Track ticket creation rates
- Monitor rate limit hits
- Log security violations
- Track admin response times
- Monitor database performance

## Compliance

### Data Protection
- Soft delete preserves audit trail
- User data isolated by permission checks
- Admin actions logged (via replied_by field)
- No PII in logs

### Security Standards
- OWASP Top 10 mitigations applied
- Input validation on all user data
- Output encoding prevents XSS
- Parameterized queries prevent SQL injection
- Rate limiting prevents abuse

## Support

For issues or questions regarding the ticket system implementation:
1. Check logs in `logs/` directory
2. Verify Redis connection if rate limiting fails
3. Check database migrations completed successfully
4. Verify option `MaxTicketsPerUserPerDay` is set
5. Review permission settings for admin users

## Changelog

### Version 1.0.0 (Initial Implementation)
- ✅ Core ticket CRUD operations
- ✅ Admin reply functionality
- ✅ Redis-based rate limiting
- ✅ XSS and SQL injection protection
- ✅ Permission-based access control
- ✅ Bilingual support (Chinese/English)
- ✅ Responsive UI with Semi Design
- ✅ Status and priority management
- ✅ Soft delete support
- ✅ Pagination and filtering
