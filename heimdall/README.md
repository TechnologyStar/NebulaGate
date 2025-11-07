# Heimdall Security Gateway

Heimdall is a comprehensive security gateway for NewAPI that provides:

- **TLS/HTTPS Support**: Encrypted communication between clients and the API
- **Request Logging**: Detailed logging of all API requests with IP tracking, device fingerprinting, and user identification
- **Transparent Proxying**: Routes all requests to NewAPI while maintaining full compatibility with OpenAI API
- **Anomaly Detection Integration**: Feeds data to NewAPI's anomaly detection system

## Features

### 1. TLS/HTTPS Encryption
- Configurable TLS certificate and key
- Secure communication channel
- Supports both HTTP and HTTPS modes

### 2. Comprehensive Request Logging
- Real IP extraction from X-Forwarded-For, X-Real-IP headers
- Device fingerprinting based on browser characteristics
- Request/response timing
- Full header and body capture (configurable limits)
- Cookie tracking for session analysis

### 3. Clear Routing Mechanism
- Explicit routes for OpenAI-compatible endpoints
- Catch-all proxy for any other routes
- No wildcard routing confusion

### 4. Data Collection & Analysis
- API token extraction from Authorization headers
- Request frequency tracking
- Content fingerprinting
- Device and IP aggregation data

## Installation

1. Install Python dependencies:
```bash
cd heimdall
pip install -r requirements.txt
```

2. Configure environment variables:
```bash
export NEWAPI_BASE_URL="http://localhost:3000"  # NewAPI backend URL
export HEIMDALL_PORT="8000"                      # Heimdall port
export TLS_CERT_PATH="/path/to/cert.pem"        # Optional: TLS certificate
export TLS_KEY_PATH="/path/to/key.pem"          # Optional: TLS key
```

3. Run Heimdall:
```bash
python main.py
```

## TLS Setup

### Generate Self-Signed Certificate (for testing)
```bash
openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365 -nodes
```

### Use Let's Encrypt (for production)
```bash
certbot certonly --standalone -d your-domain.com
export TLS_CERT_PATH="/etc/letsencrypt/live/your-domain.com/fullchain.pem"
export TLS_KEY_PATH="/etc/letsencrypt/live/your-domain.com/privkey.pem"
```

## Architecture

```
Client → Heimdall (TLS) → NewAPI → AI Models
           ↓
    Log Collection
           ↓
    Anomaly Detection
```

### Request Flow
1. Client sends request to Heimdall
2. Heimdall extracts metadata (IP, device fingerprint, token, etc.)
3. Request is proxied to NewAPI
4. Response is captured and logged
5. Log data is sent to NewAPI's log storage
6. Anomaly detection analyzes patterns

## API Endpoints

### Health Check
```
GET /health
```

### OpenAI-Compatible Endpoints
```
POST /v1/chat/completions
POST /v1/completions
POST /v1/embeddings
```

### Log Ingestion
```
POST /api/log
```

### Catch-All Proxy
```
ANY /{path}
```

## Security Features

### IP Tracking
- Extracts real IP from X-Forwarded-For (first value to avoid proxy pollution)
- Falls back to X-Real-IP and client IP
- Tracks forwarded chains for NAT analysis

### Device Fingerprinting
- SHA-256 hash of User-Agent, Accept headers, language, and encoding
- Helps identify unique devices across sessions

### Content Fingerprinting
- MD5 hash of request body
- Detects repeated or suspicious payloads

### Token Identification
- Extracts API keys from Authorization header
- Tracks API usage per token
- Enables per-key anomaly detection

## Integration with NewAPI

Heimdall sends log data to NewAPI's `/api/heimdall/log` endpoint. NewAPI stores this data in the `heimdall_logs` table and uses it for:

1. **User Activity Monitoring**: Track user behavior across sessions
2. **Anomaly Detection**: Identify suspicious patterns
3. **Abuse Prevention**: Rate limiting and blocking
4. **Security Analytics**: Dashboard and reporting

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| NEWAPI_BASE_URL | NewAPI backend URL | http://localhost:3000 |
| HEIMDALL_PORT | Port for Heimdall | 8000 |
| HEIMDALL_HOST | Host to bind | 0.0.0.0 |
| TLS_CERT_PATH | TLS certificate path | (empty) |
| TLS_KEY_PATH | TLS key path | (empty) |

## Deployment

### Docker
```dockerfile
FROM python:3.11-slim
WORKDIR /app
COPY requirements.txt .
RUN pip install -r requirements.txt
COPY main.py .
CMD ["python", "main.py"]
```

### Systemd Service
```ini
[Unit]
Description=Heimdall Security Gateway
After=network.target

[Service]
Type=simple
User=heimdall
WorkingDirectory=/opt/heimdall
Environment="NEWAPI_BASE_URL=http://localhost:3000"
Environment="HEIMDALL_PORT=8000"
ExecStart=/usr/bin/python3 /opt/heimdall/main.py
Restart=always

[Install]
WantedBy=multi-user.target
```

### Reverse Proxy (Nginx)
```nginx
upstream heimdall {
    server localhost:8000;
}

server {
    listen 80;
    listen 443 ssl;
    server_name api.yourdomain.com;

    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    location / {
        proxy_pass http://heimdall;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## Monitoring

Heimdall logs to both file (`heimdall.log`) and console. Monitor the logs for:
- Request volume
- Error rates
- Response times
- Failed proxy attempts

## Performance

- Asynchronous request handling with FastAPI
- Connection pooling with httpx
- Minimal overhead (~5-10ms per request)
- Horizontal scaling supported

## Security Considerations

1. **Always use TLS in production**
2. **Rotate TLS certificates regularly**
3. **Limit request body logging size** (default: 10KB)
4. **Monitor log storage size** and implement retention policies
5. **Secure the log database** and restrict access
6. **Use environment variables** for sensitive configuration

## Troubleshooting

### "Connection refused" errors
- Check that NewAPI is running on the configured URL
- Verify network connectivity between Heimdall and NewAPI

### TLS errors
- Verify certificate and key paths
- Check certificate validity: `openssl x509 -in cert.pem -text -noout`
- Ensure certificate matches the domain

### High memory usage
- Reduce request body logging size
- Implement log rotation
- Increase server resources

## License

Part of the NewAPI project.
