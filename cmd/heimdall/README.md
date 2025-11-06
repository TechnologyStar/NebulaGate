# Heimdall Gateway Quick Start

Heimdall is a secure Go-based API gateway that provides TLS termination and request forwarding for the New API service.

## Quick Start

### 1. Build the Gateway

```bash
# Build Heimdall binary
make build-heimdall

# Or build directly
go build -o bin/heimdall ./cmd/heimdall
```

### 2. Generate Development Certificates

```bash
# Generate self-signed certificates for testing
./scripts/generate-certs.sh
```

### 3. Configure Environment

```bash
# Copy example configuration
cp .env.heimdall.example .env

# Edit configuration
nano .env
```

### 4. Start the Services

```bash
# Start the main API
make start-backend

# Start Heimdall gateway
make start-heimdall

# Or use docker-compose
docker-compose up -d
```

### 5. Test the Gateway

```bash
# Test health endpoint
curl -k https://localhost:8443/health

# Test API proxy
curl -k -X POST https://localhost:8443/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model": "gpt-3.5-turbo", "messages": [{"role": "user", "content": "Hello"}]}'
```

## Configuration Options

### Required Variables

- `HEIMDALL_TLS_ENABLED=true` - Enable TLS (required for production)
- `HEIMDALL_TLS_CERT` - Path to TLS certificate file
- `HEIMDALL_TLS_KEY` - Path to TLS private key file
- `HEIMDALL_BACKEND_URL` - Backend API URL to proxy to

### Optional Variables

- `HEIMDALL_LISTEN_ADDR=:8443` - Server listen address
- `HEIMDALL_CORS_ORIGINS=*` - Allowed CORS origins
- `HEIMDALL_RATE_LIMIT_ENABLED=false` - Enable rate limiting
- `HEIMDALL_LOG_LEVEL=info` - Logging level

### ACME (Let's Encrypt) Configuration

For automatic certificate management:

```bash
HEIMDALL_ACME_ENABLED=true
HEIMDALL_ACME_DOMAIN=api.example.com
HEIMDALL_ACME_EMAIL=admin@example.com
```

## Docker Deployment

### Using Docker Compose

```bash
# Start all services
docker-compose up -d

# Check logs
docker-compose logs -f heimdall

# Stop services
docker-compose down
```

### Manual Docker Build

```bash
# Build image
docker build -t heimdall .

# Run container
docker run -d \
  --name heimdall \
  -p 8443:8443 \
  -v $(pwd)/certs:/certs:ro \
  -e HEIMDALL_TLS_CERT=/certs/heimdall.crt \
  -e HEIMDALL_TLS_KEY=/certs/heimdall.key \
  -e HEIMDALL_BACKEND_URL=http://backend:3000 \
  heimdall
```

## API Endpoints

Heimdall proxies all OpenAI-compatible API endpoints:

- `/v1/chat/completions` - Chat completions
- `/v1/embeddings` - Text embeddings
- `/v1/models` - Model listing
- `/v1/audio/*` - Audio processing
- `/v1/images/*` - Image generation
- `/v1/files/*` - File management
- `/health` - Health check
- `/metrics` - Service metrics

## Security Features

- **TLS Enforcement**: Mandatory HTTPS for all connections
- **Request Tracing**: Unique request IDs for debugging
- **CORS Protection**: Configurable origin restrictions
- **Rate Limiting**: Built-in rate limiting capabilities
- **Header Sanitization**: Removes sensitive headers
- **Health Monitoring**: Backend health checks

## Troubleshooting

### Certificate Issues

```bash
# Verify certificate
openssl x509 -in certs/heimdall.crt -text -noout

# Test TLS connection
openssl s_client -connect localhost:8443 -servername localhost
```

### Connection Issues

```bash
# Check if backend is accessible
curl http://localhost:3000/api/status

# Check Heimdall logs
docker-compose logs heimdall

# Test with verbose curl
curl -v -k https://localhost:8443/health
```

### Common Problems

1. **Certificate not found**: Ensure cert/key files exist and are readable
2. **Port already in use**: Change `HEIMDALL_LISTEN_ADDR` to different port
3. **Backend connection failed**: Check `HEIMDALL_BACKEND_URL` and network connectivity
4. **ACME domain error**: Verify domain ownership and DNS configuration

## Development

### Running Without TLS (Development Only)

```bash
export HEIMDALL_TLS_ENABLED=false
export HEIMDALL_LISTEN_ADDR=:8080
./bin/heimdall
```

### Debug Mode

```bash
export HEIMDALL_LOG_LEVEL=debug
./bin/heimdall
```

### Testing

```bash
# Run unit tests
go test ./cmd/heimdall/...

# Run integration tests (requires service running)
go test -v ./cmd/heimdall/... -tags=integration
```

## Production Deployment

### Security Checklist

- [ ] Use valid TLS certificates from trusted CA
- [ ] Enable rate limiting
- [ ] Configure appropriate CORS origins
- [ ] Set up monitoring and alerting
- [ ] Use reverse proxy (nginx/traefik) for additional security
- [ ] Regularly update dependencies
- [ ] Enable audit logging

### Performance Tuning

- Adjust connection pool settings
- Configure appropriate timeouts
- Monitor memory usage
- Set up horizontal scaling if needed

### Monitoring

Monitor these metrics:
- Request rate and response times
- Error rates and types
- Backend connection health
- TLS handshake success rate
- Memory and CPU usage

## Migration from FastAPI

### Key Changes

1. **TLS Required**: All requests must use HTTPS
2. **New Port**: Default port changed to 8443
3. **Enhanced Security**: Additional security headers and validations
4. **Structured Logging**: JSON-formatted logs with request tracing

### Client Updates

Update your client applications to use:

```javascript
// Old FastAPI endpoint
const response = await fetch('http://localhost:8000/v1/chat/completions', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify(payload)
});

// New Heimdall endpoint
const response = await fetch('https://localhost:8443/v1/chat/completions', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify(payload)
});
```

## Support

For issues and questions:

1. Check the [documentation](../docs/heimdall.md)
2. Review [troubleshooting guide](#troubleshooting)
3. Search existing GitHub issues
4. Create a new issue with detailed information

## License

Heimdall is part of the New API project and follows the same license terms.