# Plan & Voucher API Documentation

This document describes the Plan Management and Voucher APIs introduced in Milestone M5.

## Plan Management APIs

### Admin Endpoints

All admin endpoints require `AdminAuth` middleware.

#### 1. List All Plans

**Endpoint:** `GET /api/plan`

**Description:** Retrieves all active billing plans.

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "code": "plan_abc123",
      "name": "Pro Plan",
      "description": "Professional tier with advanced features",
      "cycle_type": "monthly",
      "cycle_duration_days": 30,
      "quota_metric": "requests",
      "quota_amount": 10000,
      "allow_carry_over": false,
      "carry_limit_percent": 0,
      "is_active": true,
      "is_public": false,
      "is_system": false,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

#### 2. Get Plan by ID

**Endpoint:** `GET /api/plan/:id`

**Description:** Retrieves a single plan by ID.

**Response:** Same as list item above.

#### 3. Create Plan

**Endpoint:** `POST /api/plan`

**Request Body:**
```json
{
  "name": "Pro Plan",
  "description": "Professional tier with advanced features",
  "cycle": "monthly",
  "cycle_length_days": 30,
  "quota": 10000,
  "quota_metric": "requests",
  "rollover_policy": "none"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "code": "plan_xyz",
    "name": "Pro Plan",
    ...
  }
}
```

#### 4. Update Plan

**Endpoint:** `PUT /api/plan/:id`

**Request Body:**
```json
{
  "name": "Updated Pro Plan",
  "quota": 15000
}
```

**Response:** Updated plan object.

#### 5. Delete Plan

**Endpoint:** `DELETE /api/plan/:id`

**Description:** Soft deletes a billing plan.

**Response:**
```json
{
  "success": true,
  "message": "Plan deleted successfully"
}
```

#### 6. Assign Plan

**Endpoint:** `POST /api/plan/assign`

**Description:** Assigns a plan to one or more subjects (users or tokens).

**Request Body:**
```json
{
  "plan_id": 1,
  "targets": [
    {
      "subject_type": "user",
      "subject_id": 123
    },
    {
      "subject_type": "token",
      "subject_id": 456
    }
  ],
  "starts_at": 1640000000
}
```

**Response:**
```json
{
  "success": true,
  "message": "Plan assigned successfully",
  "data": {
    "assigned_count": 2,
    "assignments": [...]
  }
}
```

#### 7. Detach Plan Assignment

**Endpoint:** `DELETE /api/plan/assignment/:id`

**Description:** Deactivates a plan assignment.

**Response:**
```json
{
  "success": true,
  "message": "Plan assignment deactivated successfully"
}
```

#### 8. Get Plan Usage

**Endpoint:** `GET /api/plan/assignment/:assignment_id/usage`

**Description:** Retrieves usage counters for a plan assignment.

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "plan_assignment_id": 1,
      "metric": "requests",
      "cycle_start": "2024-01-01T00:00:00Z",
      "cycle_end": "2024-02-01T00:00:00Z",
      "consumed_amount": 5000,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-15T12:00:00Z"
    }
  ]
}
```

### User Endpoints

#### 9. Get User Plans

**Endpoint:** `GET /api/plan/self`

**Description:** Retrieves active plan assignments for the authenticated user.

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "assignment": {
        "id": 1,
        "subject_type": "user",
        "subject_id": 123,
        "plan_id": 1,
        "billing_mode": "plan",
        "activated_at": "2024-01-01T00:00:00Z",
        "deactivated_at": null,
        ...
      },
      "plan": {
        "id": 1,
        "name": "Pro Plan",
        ...
      }
    }
  ]
}
```

## Voucher APIs

### Admin Endpoints

#### 10. Generate Voucher Batch

**Endpoint:** `POST /api/voucher/batch`

**Description:** Creates a batch of voucher codes.

**Request Body (Credit Vouchers):**
```json
{
  "count": 100,
  "prefix": "PROMO",
  "grant_type": "credit",
  "credit_amount": 10000,
  "expire_days": 30,
  "note": "Promotional vouchers for Q1 2024"
}
```

**Request Body (Plan Vouchers):**
```json
{
  "count": 50,
  "prefix": "PLAN",
  "grant_type": "plan",
  "plan_id": 1,
  "expire_days": 90,
  "note": "Trial plan vouchers"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Vouchers generated successfully",
  "data": {
    "codes": [
      "PROMO-abc123",
      "PROMO-def456",
      ...
    ],
    "count": 100
  }
}
```

#### 11. List Voucher Batches

**Endpoint:** `GET /api/voucher/batch`

**Description:** Retrieves all voucher batches.

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "code_prefix": "PROMO",
      "batch_label": "PROMO_20240101_100",
      "grant_type": "credit",
      "credit_amount": 10000,
      "plan_grant_id": null,
      "is_stackable": false,
      "max_redemptions": 100,
      "max_per_subject": 1,
      "valid_from": "2024-01-01T00:00:00Z",
      "valid_until": "2024-01-31T00:00:00Z",
      "created_by": "admin",
      "notes": "Promotional vouchers",
      ...
    }
  ]
}
```

#### 12. Get Voucher Redemptions

**Endpoint:** `GET /api/voucher/batch/:batch_id/redemptions`

**Description:** Retrieves all redemptions for a specific voucher batch.

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "voucher_batch_id": 1,
      "code": "PROMO-abc123",
      "subject_type": "user",
      "subject_id": 123,
      "plan_assignment_id": null,
      "redeemed_at": "2024-01-15T10:30:00Z",
      "redeemed_by": "user123",
      "credit_amount": 10000,
      "plan_granted_id": null,
      ...
    }
  ]
}
```

### User Endpoints

#### 13. Redeem Voucher

**Endpoint:** `POST /api/voucher/redeem`

**Description:** Redeems a voucher code for the authenticated user.

**Request Body:**
```json
{
  "code": "PROMO-abc123"
}
```

**Response (Credit Voucher):**
```json
{
  "success": true,
  "message": "Successfully redeemed 10000 credits",
  "data": {
    "success": true,
    "message": "Successfully redeemed 10000 credits",
    "credit_amount": 10000,
    "plan_id": 0
  }
}
```

**Response (Plan Voucher):**
```json
{
  "success": true,
  "message": "Successfully redeemed plan: Pro Plan",
  "data": {
    "success": true,
    "message": "Successfully redeemed plan: Pro Plan",
    "credit_amount": 0,
    "plan_id": 1
  }
}
```

**Error Response:**
```json
{
  "success": false,
  "message": "voucher code already redeemed"
}
```

## Token Billing Mode APIs

### Update Token with Billing Mode

**Endpoint:** `PUT /api/token`

**Description:** Updates a token including billing mode and plan assignment.

**Request Body:**
```json
{
  "id": 1,
  "name": "My Token",
  "billing_mode": "plan",
  "plan_assignment_id": 5
}
```

**Billing Modes:**
- `balance` - Uses traditional balance-based billing
- `plan` - Uses plan-based billing (requires active plan assignment)
- `auto` - Automatically determines billing mode

**Notes:**
- When setting `billing_mode` to `plan`, you must provide a valid `plan_assignment_id`
- The plan assignment must be active and belong to either the token or its owner user
- Setting `billing_mode` to `balance` or `auto` will clear the `plan_assignment_id`

**Response:**
```json
{
  "success": true,
  "message": "",
  "data": {
    "id": 1,
    "name": "My Token",
    "billing_mode": "plan",
    "plan_assignment_id": 5,
    ...
  }
}
```

## Authentication

- **Admin endpoints**: Require admin privileges (`middleware.AdminAuth()`)
- **User endpoints**: Require authenticated user (`middleware.UserAuth()`)

## Error Responses

All endpoints return consistent error responses:

```json
{
  "success": false,
  "message": "Error description here"
}
```

Common HTTP status codes:
- `200 OK` - Success
- `400 Bad Request` - Invalid request parameters
- `401 Unauthorized` - Authentication required
- `403 Forbidden` - Insufficient permissions
- `404 Not Found` - Resource not found
- `500 Internal Server Error` - Server error
