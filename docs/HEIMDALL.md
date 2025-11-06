# Heimdall Authentication Gateway

Heimdall is a comprehensive authentication and security middleware for the New API gateway, providing strong authentication, request validation, rate limiting, and audit logging capabilities.

## Features

### 🔐 Authentication Methods
- **API Key Validation**: Validates tokens against the existing database
- **JWT Token Support**: Validates signed JWT tokens with configurable signing methods
- **Mutual TLS (mTLS)**: Client certificate-based authentication
- **Multi-Method Support**: Supports multiple authentication methods simultaneously

### 🛡️ Request Validation
- **JSON Schema Validation**: Validates request payloads against schema requirements
- **Replay Attack Protection**: Prevents duplicate requests using request IDs and Redis
- **Content-Type Validation**: Ensures proper request formatting
- **Payload Size Limits**: Configurable maximum payload sizes

### ⚡ Rate Limiting
- **Per-Token Rate Limiting**: Limits requests per API key/token
- **Per-IP Rate Limiting**: Limits requests per client IP address
- **Sliding Window Algorithm**: Uses Redis for distributed rate limiting
- **Configurable Windows**: Flexible time window configurations
- **Token Bucket Implementation**: Efficient rate limiting algorithm

### 📊 Audit Logging
- **Structured JSON Logging**: Comprehensive audit trail in JSON format
- **Request/Response Tracking**: Logs all requests with detailed metadata
- **Security Event Logging**: Tracks authentication failures and security events
- **Payload Truncation**: Configurable payload logging for security
- **Redis Integration**: Stores audit logs for querying and analysis

## Installation

### Prerequisites
- Redis (recommended for distributed deployments)
- Existing New API installation
- Go 1.25+ (for development)

### Configuration

1. **Environment Variables**
   Copy the example configuration:
   ```bash
   cp .env.heimdall.example .env
   ```

2. **Basic Configuration**
   ```bash
   # Enable Heimdall
   HEIMDALL_AUTH_ENABLED=true
   HEIMDALL_API_KEY_VALIDATION=true
   HEIMDALL_RATE_LIMIT_ENABLED=true
   HEIMDALL_AUDIT_LOGGING_ENABLED=true
   ```

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `HEIMDALL_AUTH_ENABLED` | `true` | Enable/disable Heimdall authentication |
| `HEIMDALL_API_KEY_VALIDATION` | `true` | Enable API key validation |
| `HEIMDALL_JWT_VALIDATION` | `false` | Enable JWT token validation |
| `HEIMDALL_MUTUAL_TLS_VALIDATION` | `false` | Enable mutual TLS validation |
| `HEIMDALL_JWT_SECRET` | - | JWT secret key (required for JWT validation) |
| `HEIMDALL_JWT_SIGNING_METHOD` | `HS256` | JWT signing method |
| `HEIMDALL_SCHEMA_VALIDATION` | `true` | Enable JSON schema validation |
| `HEIMDALL_REPLAY_PROTECTION` | `true` | Enable replay attack protection |
| `HEIMDALL_REPLAY_WINDOW` | `5m` | Replay protection time window |
| `HEIMDALL_RATE_LIMIT_ENABLED` | `true` | Enable rate limiting |
| `HEIMDALL_PER_KEY_RATE_LIMIT` | `100` | Requests per token per window |
| `HEIMDALL_PER_IP_RATE_LIMIT` | `200` | Requests per IP per window |
| `HEIMDALL_RATE_LIMIT_WINDOW` | `1m` | Rate limiting time window |
| `HEIMDALL_AUDIT_LOGGING_ENABLED` | `true` | Enable audit logging |
| `HEIMDALL_LOG_PAYLOAD_TRUNCATE` | `true` | Truncate payloads in logs |
| `HEIMDALL_MAX_PAYLOAD_SIZE` | `1024` | Max payload size to log (bytes) |

## Usage Examples

### API Key Authentication
```bash
curl -H "Authorization: Bearer sk-your-api-key-here" \
     https://your-api.com/v1/chat/completions
```

### JWT Authentication
```bash
curl -H "Authorization: Bearer your-jwt-token-here" \
     https://your-api.com/v1/chat/completions
```

### Mutual TLS
```bash
curl --cert client.crt \
     --key client.key \
     https://your-api.com/v1/chat/completions
```

## Configuration Examples

### Development Environment
```bash
HEIMDALL_AUTH_ENABLED=true
HEIMDALL_API_KEY_VALIDATION=true
HEIMDALL_JWT_VALIDATION=false
HEIMDALL_MUTUAL_TLS_VALIDATION=false
HEIMDALL_SCHEMA_VALIDATION=true
HEIMDALL_REPLAY_PROTECTION=false
HEIMDALL_RATE_LIMIT_ENABLED=false
HEIMDALL_AUDIT_LOGGING_ENABLED=true
```

### Production High Security
```bash
HEIMDALL_AUTH_ENABLED=true
HEIMDALL_API_KEY_VALIDATION=true
HEIMDALL_JWT_VALIDATION=true
HEIMDALL_MUTUAL_TLS_VALIDATION=true
HEIMDALL_JWT_SECRET=your-super-secure-secret-key
HEIMDALL_SCHEMA_VALIDATION=true
HEIMDALL_REPLAY_PROTECTION=true
HEIMDALL_REPLAY_WINDOW=5m
HEIMDALL_RATE_LIMIT_ENABLED=true
HEIMDALL_PER_KEY_RATE_LIMIT=50
HEIMDALL_PER_IP_RATE_LIMIT=100
HEIMDALL_RATE_LIMIT_WINDOW=1m
HEIMDALL_AUDIT_LOGGING_ENABLED=true
HEIMDALL_LOG_PAYLOAD_TRUNCATE=true
HEIMDALL_MAX_PAYLOAD_SIZE=512
```

### High Throughput Configuration
```bash
HEIMDALL_AUTH_ENABLED=true
HEIMDALL_API_KEY_VALIDATION=true
HEIMDALL_JWT_VALIDATION=false
HEIMDALL_MUTUAL_TLS_VALIDATION=false
HEIMDALL_SCHEMA_VALIDATION=false
HEIMDALL_REPLAY_PROTECTION=false
HEIMDALL_RATE_LIMIT_ENABLED=true
HEIMDALL_PER_KEY_RATE_LIMIT=1000
HEIMDALL_PER_IP_RATE_LIMIT=2000
HEIMDALL_RATE_LIMIT_WINDOW=1m
HEIMDALL_AUDIT_LOGGING_ENABLED=false
```

## Response Format

### Success Response
```json
{
  "message": "success",
  "request_id": "2024-01-01-12-00-00-abc12345",
  "user_id": 123,
  "token_id": 456,
  "auth_method": "api_key"
}
```

### Authentication Error
```json
{
  "error": "authentication_failed",
  "message": "invalid API key: invalid token",
  "request_id": "2024-01-01-12-00-00-abc12345"
}
```

### Rate Limit Error
```json
{
  "error": "rate_limit_exceeded",
  "message": "rate limiting failed: IP rate limit: rate limit exceeded for ip: 101/100",
  "request_id": "2024-01-01-12-00-00-abc12345",
  "rate_limit_status": {
    "limit_type": "ip",
    "current_count": 101,
    "limit": 100,
    "window_start": "2024-01-01T12:00:00Z",
    "reset_time": "2024-01-01T12:01:00Z"
  }
}
```

### Validation Error
```json
{
  "error": "validation_failed",
  "message": "request validation failed: replay protection: duplicate request ID detected: 2024-01-01-12-00-00-abc12345",
  "request_id": "2024-01-01-12-00-00-abc12345"
}
```

## Audit Log Format

```json
{
  "request_id": "2024-01-01-12-00-00-abc12345",
  "timestamp": "2024-01-01T12:00:00Z",
  "method": "POST",
  "path": "/v1/chat/completions",
  "client_ip": "192.168.1.100",
  "user_agent": "OpenAI/Python v1.0.0",
  "auth_method": "api_key",
  "user_id": 123,
  "token_id": 456,
  "status_code": 200,
  "response_time": "150ms",
  "request_size": 1024,
  "response_size": 2048,
  "truncated_payload": "{\"model\": \"gpt-3.5-turbo\", \"messages\": [{\"role\": \"user\", \"content\": \"Hello\"}]}...",
  "validation_errors": [],
  "rate_limit_status": {
    "limit_type": "token",
    "current_count": 1,
    "limit": 100,
    "window_start": "2024-01-01T12:00:00Z",
    "reset_time": "2024-01-01T12:01:00Z"
  },
  "tls_info": {
    "version": 771,
    "cipher_suite": 49195,
    "server_name": "api.example.com",
    "peer_certificates": ["CN=client.example.com"]
  }
}
```

## Performance Considerations

### Redis Usage
- **Required for Production**: Redis is required for distributed rate limiting and replay protection
- **Memory Usage**: Each request creates temporary Redis keys with TTL
- **Network Latency**: Consider Redis network latency for high-throughput scenarios

### Rate Limiting Performance
- **Sliding Window**: More accurate but requires more Redis operations
- **Token Bucket**: Efficient and fair resource allocation
- **Fallback**: In-memory rate limiting when Redis is unavailable

### Audit Logging Performance
- **Async Operations**: Audit logging is performed asynchronously to minimize impact
- **Payload Truncation**: Configure truncation to reduce log size
- **TTL Management**: Audit logs automatically expire after 30 days

## Security Considerations

### Sensitive Data
- **No Plaintext Secrets**: API keys and secrets are never logged in plaintext
- **Payload Truncation**: Request payloads are truncated in logs
- **JWT Secret**: Store JWT secret securely (environment variables, secret management)

### Replay Protection
- **Request ID Uniqueness**: Ensure request IDs are sufficiently unique
- **Time Window**: Configure appropriate replay protection windows
- **Redis Persistence**: Ensure Redis persistence for replay protection

### Rate Limiting Bypass
- **IP Spoofing**: Rate limiting by IP can be bypassed with IP spoofing
- **Token Sharing**: Per-token rate limits can be bypassed by sharing tokens
- **Distributed Attacks**: Consider additional protection for distributed attacks

## Monitoring and Debugging

### Health Checks
Monitor the following metrics:
- Authentication success/failure rates
- Rate limit hit rates
- Request validation error rates
- Audit log volume

### Debug Logging
Enable debug mode for detailed logging:
```bash
GIN_MODE=debug
HEIMDALL_AUDIT_LOGGING_ENABLED=true
```

### Redis Monitoring
Monitor Redis keys:
```bash
# Replay protection keys
redis-cli KEYS "heimdall:replay:*"

# Rate limiting keys
redis-cli KEYS "heimdall:rate:*"

# Audit log keys
redis-cli KEYS "heimdall:audit:*"
```

## Testing

### Unit Tests
```bash
go test ./middleware/heimdall_test.go
```

### Integration Tests
```bash
go test ./middleware/heimdall_integration_test.go
```

### Benchmarks
```bash
go test -bench=. ./middleware/
```

## Migration from Standard Authentication

1. **Gradual Migration**: Enable Heimdall alongside existing authentication
2. **Configuration**: Start with basic features, enable advanced features gradually
3. **Monitoring**: Monitor performance and error rates during migration
4. **Rollback**: Keep standard authentication as fallback option

## Troubleshooting

### Common Issues

#### Authentication Failures
- Check API key validity in database
- Verify JWT secret and signing method
- Ensure mutual TLS certificates are valid

#### Rate Limiting Issues
- Verify Redis connectivity
- Check rate limit configuration
- Monitor Redis key expiration

#### Audit Logging Issues
- Check Redis storage capacity
- Verify log rotation settings
- Monitor disk space for system logs

#### Performance Issues
- Monitor Redis latency
- Check rate limiting algorithm efficiency
- Optimize audit log configuration

### Debug Commands
```bash
# Check Heimdall configuration
curl -H "Authorization: Bearer test-key" https://your-api.com/api/test

# Check Redis keys
redis-cli INFO memory
redis-cli INFO stats

# Monitor audit logs
tail -f /var/log/new-api.log | grep heimdall
```

## Contributing

1. **Code Style**: Follow Go conventions and existing code style
2. **Testing**: Add unit tests for new features
3. **Documentation**: Update documentation for configuration changes
4. **Security**: Consider security implications for all changes

## License

This project is licensed under the same terms as the New API project.