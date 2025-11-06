# Heimdall Gateway

Heimdall is a hardened Go-based API gateway that provides secure TLS termination, structured routing, and modular extensions for the New API monorepo. It replaces the previous insecure FastAPI-based implementation with a production-ready service.

## Architecture

### Core Components

- **Service Skeleton**: Built with Gin framework for consistency with the monolith stack
- **Configuration Management**: Environment-based configuration via `setting/heimdall` package
- **TLS Support**: Mandatory TLS with manual certificates or automatic ACME (Let's Encrypt)
- **Structured Routing**: Explicit route groups for different API endpoints
- **Modular Middlewares**: Request tracing, panic recovery, CORS, rate limiting
- **Proxy Handler**: Full HTTP request forwarding to backend services

### Directory Structure

```
cmd/heimdall/
├── main.go          # Main service entry point
└── proxy.go         # HTTP proxy implementation
setting/heimdall/
└── config.go        # Configuration management
```

## Configuration

Heimdall uses environment variables for configuration. Create a `.env` file or set environment variables directly.

### Required Configuration

```bash
# Basic TLS (required for production)
HEIMDALL_TLS_ENABLED=true
HEIMDALL_TLS_CERT=/path/to/cert.pem
HEIMDALL_TLS_KEY=/path/to/key.pem
HEIMDALL_LISTEN_ADDR=:8443

# Backend service
HEIMDALL_BACKEND_URL=http://localhost:3000
```

### ACME (Let's Encrypt) Configuration

```bash
# Use automatic certificate management
HEIMDALL_ACME_ENABLED=true
HEIMDALL_ACME_DOMAIN=api.example.com
HEIMDALL_ACME_EMAIL=admin@example.com
HEIMDALL_ACME_CACHE_DIR=/tmp/heimdall-acme
```

### Optional Configuration

```bash
# CORS settings
HEIMDALL_CORS_ORIGINS=*

# Rate limiting
HEIMDALL_RATE_LIMIT_ENABLED=true
HEIMDALL_RATE_LIMIT_REQUESTS=100
HEIMDALL_RATE_LIMIT_WINDOW_MINUTES=1

# Authentication
HEIMDALL_API_KEY_HEADER=Authorization
HEIMDALL_API_KEY_VALUE=Bearer your-token

# Logging
HEIMDALL_LOG_LEVEL=info
HEIMDALL_LOG_FORMAT=json
```

## API Routes

Heimdall provides explicit routing for OpenAI-compatible API endpoints:

### Core API Endpoints

- `/v1/chat/completions` - Chat completions
- `/v1/embeddings` - Text embeddings
- `/v1/models` - Model listing
- `/v1/moderations` - Content moderation

### Audio Endpoints

- `/v1/audio/transcriptions` - Speech to text
- `/v1/audio/translations` - Audio translation
- `/v1/audio/speech` - Text to speech

### Image Endpoints

- `/v1/images/generations` - Image generation
- `/v1/images/edits` - Image editing
- `/v1/images/variations` - Image variations

### File Management

- `/v1/files` - File upload/listing
- `/v1/files/:file_id` - File deletion

### Advanced Features

- `/v1/fine_tuning/jobs` - Fine-tuning job management
- `/v1/batch` - Batch request processing

### System Endpoints

- `/` - Service information
- `/health` - Health check (includes backend status)
- `/metrics` - Basic service metrics

## Deployment

### Building the Service

```bash
# Build the Heimdall binary
make build-heimdall

# Or build directly
go build -o bin/heimdall ./cmd/heimdall
```

### Running with TLS

#### Manual Certificates

```bash
export HEIMDALL_TLS_ENABLED=true
export HEIMDALL_TLS_CERT=/etc/ssl/certs/heimdall.crt
export HEIMDALL_TLS_KEY=/etc/ssl/private/heimdall.key
export HEIMDALL_LISTEN_ADDR=:8443
export HEIMDALL_BACKEND_URL=http://localhost:3000

./bin/heimdall
```

#### ACME (Let's Encrypt)

```bash
export HEIMDALL_ACME_ENABLED=true
export HEIMDALL_ACME_DOMAIN=api.example.com
export HEIMDALL_ACME_EMAIL=admin@example.com
export HEIMDALL_LISTEN_ADDR=:8443
export HEIMDALL_BACKEND_URL=http://localhost:3000

./bin/heimdall
```

### Docker Deployment

#### Dockerfile Extension

Add to your existing Dockerfile:

```dockerfile
# Build Heimdall
FROM builder AS heimdall-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o heimdall ./cmd/heimdall

# Final image
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=heimdall-builder /app/heimdall .
COPY --from=backend-builder /app/new-api .
EXPOSE 8443
CMD ["./heimdall"]
```

#### Docker Compose

```yaml
version: '3.8'
services:
  heimdall:
    build: .
    ports:
      - "8443:8443"
      - "80:80"  # Required for ACME challenges
    environment:
      - HEIMDALL_TLS_ENABLED=true
      - HEIMDALL_ACME_ENABLED=true
      - HEIMDALL_ACME_DOMAIN=api.example.com
      - HEIMDALL_ACME_EMAIL=admin@example.com
      - HEIMDALL_BACKEND_URL=http://backend:3000
    volumes:
      - /etc/letsencrypt:/etc/letsencrypt
    depends_on:
      - backend

  backend:
    build: .
    ports:
      - "3000:3000"
```

## Migration from FastAPI

### Key Differences

1. **TLS Enforcement**: Heimdall requires TLS by default (configurable)
2. **Explicit Routing**: No wildcard catch-all routes
3. **Structured Configuration**: Environment-based config management
4. **Enhanced Security**: Built-in rate limiting, CORS, and security headers
5. **Production Ready**: Graceful shutdown, health checks, and metrics

### Migration Steps

1. **Update Configuration**: Convert FastAPI config to environment variables
2. **Update Client Code**: Change endpoints to use HTTPS on port 8443
3. **TLS Setup**: Obtain SSL certificates or configure ACME
4. **Update Reverse Proxy**: Point to Heimdall instead of FastAPI
5. **Monitor Logs**: Check Heimdall logs for request forwarding

### Configuration Mapping

| FastAPI Setting | Heimdall Environment Variable |
|------------------|------------------------------|
| `host` | `HEIMDALL_LISTEN_ADDR` |
| `ssl_certfile` | `HEIMDALL_TLS_CERT` |
| `ssl_keyfile` | `HEIMDALL_TLS_KEY` |
| `backend_url` | `HEIMDALL_BACKEND_URL` |
| `cors_origins` | `HEIMDALL_CORS_ORIGINS` |

## Security Features

### TLS Configuration

- **Minimum TLS Version**: TLS 1.2
- **Certificate Validation**: Automatic cert file validation
- **ACME Support**: Let's Encrypt integration with automatic renewal

### Request Security

- **Request ID Tracing**: Unique IDs for all requests
- **CORS Protection**: Configurable origin restrictions
- **Rate Limiting**: Built-in rate limiting capabilities
- **Security Headers**: Common security headers automatically added

### Proxy Security

- **Header Sanitization**: Removes hop-by-hop headers
- **Backend Authentication**: Optional API key forwarding
- **Request Context**: Preserves request context across proxy
- **Error Handling**: Comprehensive error responses

## Monitoring and Observability

### Health Checks

The `/health` endpoint provides:
- Service status
- Backend connectivity status
- Response timestamps

### Metrics

The `/metrics` endpoint provides:
- Service version and uptime
- Request statistics (placeholder for implementation)
- Backend status information

### Logging

Structured logging includes:
- Request timestamps and duration
- Client IP addresses
- Response status codes
- Error details and stack traces

## Development

### Local Development

For development without TLS:

```bash
export HEIMDALL_TLS_ENABLED=false
export HEIMDALL_LISTEN_ADDR=:8080
export HEIMDALL_BACKEND_URL=http://localhost:3000

./bin/heimdall
```

### Testing

Integration tests are included that verify:
- TLS handshake functionality
- Request forwarding
- Response structure validation
- Health check functionality

Run tests with:

```bash
go test ./cmd/heimdall/...
```

## Troubleshooting

### Common Issues

1. **Certificate Not Found**: Ensure cert/key files exist and are readable
2. **ACME Domain Error**: Verify domain ownership and DNS configuration
3. **Backend Connection Failed**: Check backend URL and network connectivity
4. **Port Already in Use**: Change `HEIMDALL_LISTEN_ADDR` to different port

### Debug Mode

Enable debug logging:

```bash
export HEIMDALL_LOG_LEVEL=debug
```

### Log Analysis

Monitor logs for:
- TLS handshake errors
- Backend connection issues
- Request forwarding failures
- Rate limiting activations

## Performance Considerations

### Connection Pooling

Heimdall uses HTTP connection pooling for backend connections:
- Max idle connections: 100
- Idle connection timeout: 90 seconds
- Connection reuse enabled

### Resource Usage

Typical resource consumption:
- Memory: ~50MB base + request buffering
- CPU: Low overhead, mainly proxy forwarding
- Network: Proportional to traffic volume

### Scaling

For high-traffic deployments:
- Consider horizontal scaling with load balancer
- Monitor connection pool metrics
- Implement rate limiting per client
- Use CDN for static content if needed

## Future Enhancements

Planned improvements:
- Advanced metrics collection (Prometheus)
- WebSocket proxy support
- Request/response transformation
- Advanced authentication methods
- Circuit breaker patterns
- Distributed tracing integration