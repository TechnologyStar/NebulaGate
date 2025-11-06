# Heimdall Quick Start Guide

## 🚀 Getting Started

### 1. Basic Setup
```bash
# Copy the example configuration
cp .env.heimdall.example .env

# Edit the configuration with your settings
nano .env
```

### 2. Minimum Configuration
Add these to your `.env` file:
```bash
# Enable Heimdall
HEIMDALL_AUTH_ENABLED=true
HEIMDALL_API_KEY_VALIDATION=true

# Basic rate limiting
HEIMDALL_RATE_LIMIT_ENABLED=true
HEIMDALL_PER_KEY_RATE_LIMIT=100
HEIMDALL_PER_IP_RATE_LIMIT=200

# Enable audit logging
HEIMDALL_AUDIT_LOGGING_ENABLED=true
```

### 3. Start the Application
```bash
# The application will automatically use Heimdall if enabled
./new-api
```

## 🧪 Testing the Implementation

### Test Authentication
```bash
# Test with API key (replace with your actual key)
curl -H "Authorization: Bearer sk-your-api-key-here" \
     -X POST https://localhost:3000/v1/chat/completions \
     -H "Content-Type: application/json" \
     -d '{"model": "gpt-3.5-turbo", "messages": [{"role": "user", "content": "Hello"}]}'
```

### Test Rate Limiting
```bash
# Send multiple requests quickly to test rate limiting
for i in {1..110}; do
    curl -s -o /dev/null -w "%{http_code}" \
         -H "Authorization: Bearer sk-your-api-key-here" \
         https://localhost:3000/v1/models
    echo ""
done
```

### Test Replay Protection
```bash
# Send same request ID twice (should fail on second request)
curl -H "Authorization: Bearer sk-your-api-key-here" \
     -H "X-Oneapi-Request-Id: test-replay-123" \
     https://localhost:3000/v1/models

curl -H "Authorization: Bearer sk-your-api-key-here" \
     -H "X-Oneapi-Request-Id: test-replay-123" \
     https://localhost:3000/v1/models  # Should return 400
```

## 📊 Monitoring

### Check Audit Logs
```bash
# View system logs (audit logs are included)
tail -f /var/log/new-api.log | grep heimdall

# Check Redis for audit logs
redis-cli KEYS "heimdall:audit:*" | head -10
```

### Monitor Rate Limiting
```bash
# Check rate limit keys in Redis
redis-cli KEYS "heimdall:rate:*" | head -10

# View specific rate limit status
redis-cli ZRANGE "heimdall:rate:token:123:1640995200" 0 -1 WITHSCORES
```

## 🔧 Advanced Configuration

### JWT Authentication
```bash
# Enable JWT authentication
HEIMDALL_JWT_VALIDATION=true
HEIMDALL_JWT_SECRET=your-super-secret-jwt-key
HEIMDALL_JWT_SIGNING_METHOD=HS256

# Test with JWT token
curl -H "Authorization: Bearer your-jwt-token-here" \
     https://localhost:3000/v1/models
```

### High Security Setup
```bash
# Enable all security features
HEIMDALL_AUTH_ENABLED=true
HEIMDALL_API_KEY_VALIDATION=true
HEIMDALL_JWT_VALIDATION=true
HEIMDALL_MUTUAL_TLS_VALIDATION=true
HEIMDALL_JWT_SECRET=your-production-secret-key
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

### High Throughput Setup
```bash
# Optimize for performance
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

## 🐛 Troubleshooting

### Common Issues

#### Authentication Fails
```bash
# Check if token exists in database
# (Use your database client to verify)

# Check Heimdall configuration
grep HEIMDALL_ .env

# Check logs for errors
tail -f /var/log/new-api.log | grep -i error
```

#### Rate Limiting Not Working
```bash
# Check Redis connection
redis-cli ping

# Verify rate limit configuration
grep HEIMDALL_RATE .env

# Check rate limit keys
redis-cli KEYS "heimdall:rate:*"
```

#### Audit Logs Missing
```bash
# Check if audit logging is enabled
grep HEIMDALL_AUDIT_LOGGING_ENABLED .env

# Check Redis storage
redis-cli KEYS "heimdall:audit:*"

# Verify permissions
ls -la /var/log/new-api.log
```

## 📈 Performance Tuning

### Redis Optimization
```bash
# Redis configuration for high performance
redis-cli CONFIG SET maxmemory 2gb
redis-cli CONFIG SET maxmemory-policy allkeys-lru
redis-cli CONFIG SET save "900 1 300 10 60 10000"
```

### Monitoring Metrics
```bash
# Monitor Redis memory usage
redis-cli INFO memory | grep used_memory_human

# Monitor rate limiting performance
redis-cli INFO stats | grep instantaneous_ops_per_sec
```

## 🔄 Migration from Standard Auth

### Gradual Migration
1. **Phase 1**: Enable Heimdall with API key validation only
2. **Phase 2**: Add rate limiting
3. **Phase 3**: Enable audit logging
4. **Phase 4**: Add advanced authentication methods

### Configuration Migration
```bash
# Start with compatible settings
HEIMDALL_AUTH_ENABLED=true
HEIMDALL_API_KEY_VALIDATION=true
HEIMDALL_RATE_LIMIT_ENABLED=false
HEIMDALL_REPLAY_PROTECTION=false
HEIMDALL_AUDIT_LOGGING_ENABLED=true

# Gradually enable features
HEIMDALL_RATE_LIMIT_ENABLED=true
HEIMDALL_REPLAY_PROTECTION=true
HEIMDALL_JWT_VALIDATION=true
```

## 📚 Additional Resources

- **Full Documentation**: `docs/HEIMDALL.md`
- **Implementation Details**: `HEIMDALL_IMPLEMENTATION.md`
- **Configuration Examples**: `.env.heimdall.example`
- **Test Files**: `middleware/heimdall_test.go`, `middleware/heimdall_integration_test.go`

## 🆘 Getting Help

### Debug Mode
```bash
# Enable debug logging
GIN_MODE=debug ./new-api

# Check configuration
curl -H "Authorization: Bearer test-key" https://localhost:3000/api/test
```

### Health Check
```bash
# Verify Heimdall is working
curl -I https://localhost:3000/v1/models

# Check response headers for Heimdall information
curl -v https://localhost:3000/v1/models
```

### Support
- Check the logs for detailed error messages
- Verify Redis connectivity
- Ensure environment variables are set correctly
- Review the comprehensive documentation in `docs/HEIMDALL.md`

---

**🎉 Congratulations! Heimdall is now securing your API gateway with enterprise-grade authentication, validation, rate limiting, and audit logging.**