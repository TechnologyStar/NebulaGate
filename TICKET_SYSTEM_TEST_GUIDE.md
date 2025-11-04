# Ticket System Testing Guide

## Quick Start

### Prerequisites
1. Start the application
2. Ensure Redis is running (for rate limiting)
3. Create a regular user account
4. Create an admin user account

### Basic Functionality Tests

#### Test 1: Create a Ticket (Regular User)
1. Log in as regular user
2. Click "工单系统" in sidebar
3. Click "创建工单" button
4. Fill in the form:
   - Title: "Test Ticket Title"
   - Category: "技术支持"
   - Priority: "中"
   - Content: "This is a test ticket content"
5. Click "提交"
6. Verify success message appears
7. Verify redirect to ticket list
8. Verify ticket appears in list

#### Test 2: View Ticket List (Regular User)
1. Log in as regular user
2. Navigate to /ticket
3. Verify only own tickets visible
4. Verify columns: ID, 标题, 状态, 优先级, 分类, 创建时间, 操作
5. Test pagination if >10 tickets exist
6. Test status filter dropdown

#### Test 3: View Ticket Detail (Regular User)
1. Click on a ticket title or "查看" button
2. Verify all ticket details displayed:
   - Ticket ID
   - Status (colored tag)
   - Priority (colored tag)
   - Category
   - Created time
   - Title
   - Content
3. Verify no admin reply form visible
4. Verify can delete own ticket

#### Test 4: Admin View All Tickets
1. Log in as admin user
2. Navigate to /ticket
3. Verify can see tickets from all users
4. Verify "创建者" column visible
5. Verify status filter works
6. Click on any ticket to view details

#### Test 5: Admin Reply to Ticket
1. Log in as admin
2. Open any pending ticket
3. Verify "添加回复" section visible
4. Enter reply text
5. Click "提交回复"
6. Verify success message
7. Verify ticket status changes to "处理中"
8. Verify reply appears in "管理员回复" section

#### Test 6: Update Ticket Status (Admin)
1. Log in as admin
2. Open any ticket
3. Click status buttons (待处理/处理中/已解决/已关闭)
4. Verify status updates successfully
5. Verify status tag color changes
6. Return to list, verify status reflected

## Security Tests

### SQL Injection Tests

#### Test 7: SQL Injection in Title
1. Create ticket with title: `Test' OR '1'='1`
2. Verify ticket created successfully
3. Verify title stored as-is (escaped)
4. Verify no SQL error
5. Verify no unauthorized data access

#### Test 8: SQL Injection in Content
1. Create ticket with content:
```
' UNION SELECT * FROM users --
```
2. Verify ticket created successfully
3. Verify content stored safely
4. Verify no data leakage

### XSS Tests

#### Test 9: XSS in Title
1. Create ticket with title:
```html
<script>alert('XSS')</script>
```
2. Verify ticket created
3. View ticket detail
4. Verify no alert popup
5. Verify script tags escaped/removed

#### Test 10: XSS with Image Tag
1. Create ticket with content:
```html
<img src=x onerror=alert('XSS')>
```
2. Verify ticket created
3. View ticket detail
4. Verify no alert popup
5. Verify img tag and onerror removed

#### Test 11: XSS with Event Handler
1. Create ticket with content:
```html
<div onclick="alert('XSS')">Click me</div>
```
2. Verify ticket created
3. View ticket detail
4. Click the content area
5. Verify no alert popup
6. Verify onclick attribute removed

#### Test 12: JavaScript Protocol
1. Create ticket with content:
```html
<a href="javascript:alert('XSS')">Click</a>
```
2. Verify ticket created
3. View ticket detail
4. Verify link sanitized
5. Verify no javascript execution

### Rate Limiting Tests

#### Test 13: Daily Limit Enforcement
1. Log in as regular user
2. Create 5 tickets rapidly (default limit)
3. Verify all 5 created successfully
4. Try to create 6th ticket
5. Verify error: "已达到每日工单创建上限"
6. Verify 6th ticket not created

#### Test 14: Rate Limit Reset
1. After hitting daily limit
2. Check Redis key: `ticket:daily:{user_id}:{today}`
3. Verify key has 24-hour TTL
4. Or manually delete Redis key
5. Try creating ticket again
6. Verify can create new ticket

#### Test 15: Rate Limit Configuration
1. Log in as root/admin
2. Navigate to system settings
3. Find "MaxTicketsPerUserPerDay" setting
4. Change from 5 to 10
5. Save settings
6. Log in as regular user
7. Verify new limit applies

### Permission Tests

#### Test 16: User Isolation
1. Create ticket as User A
2. Note the ticket ID
3. Log out and log in as User B
4. Try to access User A's ticket: `/ticket/{id}`
5. Verify error: "无权访问该工单"
6. Verify ticket details not displayed

#### Test 17: Non-Admin Cannot Reply
1. Log in as regular user (non-admin)
2. Open any ticket
3. Verify no "添加回复" section visible
4. Try direct API call:
```bash
curl -X PUT http://localhost:3000/api/ticket/1/reply \
  -H "Content-Type: application/json" \
  -d '{"reply":"Unauthorized reply"}'
```
5. Verify 403 Forbidden or similar error

#### Test 18: Non-Owner Cannot Delete Others' Tickets
1. Log in as User A
2. Create a ticket
3. Log in as User B (non-admin)
4. Try to delete User A's ticket via API
5. Verify error: "无权删除该工单"

### Input Validation Tests

#### Test 19: Empty Title
1. Try to create ticket with empty title
2. Verify client-side validation triggers
3. Verify error: "标题不能为空"
4. Verify form not submitted

#### Test 20: Oversized Title
1. Create ticket with 201-character title
2. Verify client-side validation (maxLength)
3. Or if submitted, verify server error
4. Verify error: "标题长度不能超过200字符"

#### Test 21: Empty Content
1. Try to create ticket with empty content
2. Verify client-side validation triggers
3. Verify error: "内容不能为空"
4. Verify form not submitted

#### Test 22: Oversized Content
1. Create ticket with 5001-character content
2. Verify client-side validation (maxLength)
3. Or if submitted, verify server error
4. Verify error: "内容长度不能超过5000字符"

#### Test 23: Invalid Status
1. Try to update ticket with invalid status via API:
```bash
curl -X PUT http://localhost:3000/api/ticket/1/status \
  -H "Content-Type: application/json" \
  -d '{"status":"invalid_status"}'
```
2. Verify error: "无效的状态"
3. Verify ticket status unchanged

#### Test 24: Invalid Priority
1. Try to create ticket with invalid priority via API
2. Verify error: "无效的优先级"
3. Verify ticket not created

## UI/UX Tests

#### Test 25: Responsive Design
1. Open ticket system on desktop (1920x1080)
2. Verify layout looks good
3. Resize to tablet (768x1024)
4. Verify layout adjusts correctly
5. Resize to mobile (375x667)
6. Verify all controls accessible
7. Verify table responsive or scrollable

#### Test 26: Language Toggle
1. Change language to Chinese
2. Navigate to ticket system
3. Verify all labels in Chinese
4. Change language to English
5. Navigate to ticket system
6. Verify all labels in English
7. Test both languages for:
   - Status labels
   - Priority labels
   - Category labels
   - Button text
   - Error messages

#### Test 27: Status Color Tags
1. Create tickets with each status
2. Verify colors:
   - Pending: Amber/Yellow
   - Processing: Blue
   - Resolved: Green
   - Closed: Grey
3. Verify colors visible in both list and detail views

#### Test 28: Priority Color Tags
1. Create tickets with each priority
2. Verify colors:
   - Urgent: Red
   - High: Orange
   - Medium: Amber/Yellow
   - Low: Green
3. Verify colors visible in both list and detail views

#### Test 29: Form Validation Messages
1. Try to submit form with missing title
2. Verify error message appears below field
3. Try to submit with missing content
4. Verify error message appears below field
5. Fill correctly and verify messages clear

#### Test 30: Toast Notifications
1. Create a ticket
2. Verify success toast appears
3. Try to create ticket over limit
4. Verify error toast appears
5. Verify toasts auto-dismiss after ~3 seconds
6. Verify toasts readable and styled correctly

#### Test 31: Modal Confirmations
1. Try to delete a ticket
2. Verify confirmation modal appears
3. Verify modal has title: "确认删除"
4. Verify modal has message: "确定要删除此工单吗？"
5. Click "取消" - verify modal closes, ticket not deleted
6. Click "确定" - verify modal closes, ticket deleted

#### Test 32: Sidebar Navigation
1. Verify "工单系统" appears in sidebar under "个人中心"
2. Click on "工单系统"
3. Verify navigates to /ticket
4. Verify menu item highlighted
5. Verify icon (HelpCircle) displays correctly
6. Test with collapsed sidebar
7. Verify icon visible, text hidden

## Performance Tests

#### Test 33: Pagination Performance
1. Create 50+ tickets
2. Navigate to ticket list
3. Verify loads quickly (<2 seconds)
4. Click through pages
5. Verify smooth navigation
6. Check network tab for query performance

#### Test 34: Filter Performance
1. Create tickets with various statuses
2. Apply status filter
3. Verify filtered results load quickly
4. Change filter multiple times
5. Verify no lag or freeze

#### Test 35: Large Content Handling
1. Create ticket with 4999-character content
2. Verify saves successfully
3. Open ticket detail
4. Verify content displays correctly
5. Verify no layout breaks
6. Verify scrollable if needed

## Integration Tests

#### Test 36: End-to-End Workflow
1. User creates ticket (pending)
2. Admin views ticket in list
3. Admin opens ticket
4. Admin replies to ticket (status → processing)
5. Admin marks ticket resolved
6. User views resolved ticket and reply
7. User or admin closes ticket
8. Verify full flow works smoothly

#### Test 37: Multi-User Concurrency
1. Log in as User A and User B simultaneously
2. Both create tickets at same time
3. Verify both tickets created
4. Verify each user sees only own tickets
5. Admin sees both tickets
6. No data corruption or confusion

#### Test 38: Redis Failure Handling
1. Stop Redis service
2. Try to create ticket
3. Verify ticket still created (graceful degradation)
4. Verify no crash or error
5. Restart Redis
6. Verify rate limiting works again

## Regression Tests

#### Test 39: Soft Delete Preservation
1. Create a ticket
2. Note the ticket ID
3. Delete the ticket
4. Query database: `SELECT * FROM tickets WHERE id = ? AND deleted_at IS NOT NULL`
5. Verify ticket still exists with deleted_at timestamp
6. Verify ticket not visible in UI

#### Test 40: Timestamp Accuracy
1. Create a ticket
2. Note creation time
3. Wait 1 minute
4. Admin replies to ticket
5. Verify created_at unchanged
6. Verify updated_at changed
7. Verify replied_at is recent timestamp

## Load Tests (Optional)

#### Test 41: High Volume Creation
1. Use script to create 100 tickets rapidly
2. Verify rate limiting works correctly
3. Verify no database deadlocks
4. Verify all valid tickets created
5. Verify data integrity maintained

#### Test 42: Concurrent Admin Replies
1. Multiple admins reply to different tickets simultaneously
2. Verify all replies saved correctly
3. Verify no race conditions
4. Verify replied_by correctly attributed

## Documentation Tests

#### Test 43: API Documentation Match
1. Review TICKET_SYSTEM_IMPLEMENTATION.md
2. Test each API endpoint listed
3. Verify request/response formats match
4. Verify status codes correct
5. Update documentation if discrepancies found

#### Test 44: Error Message Clarity
1. Trigger various errors
2. Verify error messages are:
   - Clear and understandable
   - Actionable (tell user what to fix)
   - Not exposing sensitive system info
   - Properly translated in both languages

## Success Criteria

All tests should pass with:
- ✅ No unhandled exceptions
- ✅ No security vulnerabilities
- ✅ No data corruption
- ✅ Clear error messages
- ✅ Smooth user experience
- ✅ Proper permission enforcement
- ✅ Working rate limiting
- ✅ Bilingual support functioning
- ✅ Responsive design working

## Test Report Template

```markdown
# Ticket System Test Report

**Date:** YYYY-MM-DD
**Tester:** [Name]
**Build:** [Version/Commit]

## Test Summary
- Total Tests: 44
- Passed: X
- Failed: Y
- Skipped: Z

## Failed Tests
### Test [Number]: [Test Name]
- **Expected:** [What should happen]
- **Actual:** [What happened]
- **Severity:** Critical/High/Medium/Low
- **Steps to Reproduce:**
  1. Step 1
  2. Step 2
- **Screenshots:** [If applicable]

## Notes
[Any additional observations or recommendations]

## Sign-off
Tested by: [Name]
Approved by: [Name]
Date: YYYY-MM-DD
```

## Automated Test Script (Bash)

```bash
#!/bin/bash
# Quick automated test for ticket system

BASE_URL="http://localhost:3000"
USER_TOKEN="your_user_session_token"
ADMIN_TOKEN="your_admin_session_token"

echo "Testing Ticket System..."

# Test 1: Create ticket
echo "Test 1: Create Ticket"
curl -s -X POST "$BASE_URL/api/ticket/create" \
  -H "Cookie: session=$USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Test Ticket",
    "content": "Test content",
    "priority": "medium",
    "category": "technical"
  }' | jq .

# Test 2: List tickets
echo "Test 2: List Tickets"
curl -s "$BASE_URL/api/ticket/list" \
  -H "Cookie: session=$USER_TOKEN" | jq .

# Test 3: SQL Injection attempt
echo "Test 3: SQL Injection Test"
curl -s -X POST "$BASE_URL/api/ticket/create" \
  -H "Cookie: session=$USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Test'"'"' OR '"'"'1'"'"'='"'"'1",
    "content": "Test content",
    "priority": "medium",
    "category": "technical"
  }' | jq .

# Test 4: XSS attempt
echo "Test 4: XSS Test"
curl -s -X POST "$BASE_URL/api/ticket/create" \
  -H "Cookie: session=$USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "<script>alert(\"xss\")</script>",
    "content": "<img src=x onerror=alert(1)>",
    "priority": "medium",
    "category": "technical"
  }' | jq .

echo "Tests completed!"
```

## Troubleshooting

### Issue: Rate limiting not working
- Check Redis connection
- Verify `REDIS_CONN_STRING` environment variable set
- Check Redis logs for errors
- Verify key exists: `redis-cli KEYS "ticket:daily:*"`

### Issue: XSS not prevented
- Check `SanitizeInput()` function called
- Verify regex patterns in `ticket_validation.go`
- Test with various XSS payloads

### Issue: Permission checks failing
- Verify user roles correctly set in database
- Check session middleware working
- Verify `isAdmin()` helper function
- Check AdminAuth middleware applied to admin routes

### Issue: Translations missing
- Verify i18n files updated (zh.json, en.json)
- Clear browser cache
- Check console for missing translation warnings
- Verify translation keys match exactly

### Issue: Database migration failed
- Check database connection
- Verify SQL_DSN environment variable
- Check database user has CREATE TABLE permissions
- Run migration manually: `DB.AutoMigrate(&Ticket{})`
- Check for existing tables with conflicts
