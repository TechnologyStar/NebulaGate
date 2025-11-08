#!/usr/bin/env python3
"""
Heimdall Security Gateway
A comprehensive security gateway for NewAPI with TLS support, request logging, and anomaly detection.
"""

import os
import sys
import json
import time
import hashlib
import logging
from datetime import datetime
from typing import Optional, Dict, Any

import uvicorn
from fastapi import FastAPI, Request, HTTPException, Header, Response
from fastapi.responses import JSONResponse, StreamingResponse
from fastapi.middleware.cors import CORSMiddleware
import httpx
from pydantic import BaseModel

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
    handlers=[
        logging.FileHandler('heimdall.log'),
        logging.StreamHandler()
    ]
)
logger = logging.getLogger(__name__)

# Configuration
NEWAPI_BASE_URL = os.getenv("NEWAPI_BASE_URL", "http://localhost:3000")
HEIMDALL_PORT = int(os.getenv("HEIMDALL_PORT", "8000"))
HEIMDALL_HOST = os.getenv("HEIMDALL_HOST", "0.0.0.0")
TLS_CERT_PATH = os.getenv("TLS_CERT_PATH", "")
TLS_KEY_PATH = os.getenv("TLS_KEY_PATH", "")
LOG_DB_URL = os.getenv("LOG_DB_URL", "sqlite:///heimdall_logs.db")

# FastAPI app
app = FastAPI(
    title="Heimdall Security Gateway",
    description="Security gateway with TLS, logging, and anomaly detection",
    version="1.0.0"
)

# CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# HTTP client for proxying requests
http_client = httpx.AsyncClient(timeout=300.0)


class LogEntry(BaseModel):
    """Model for log entries"""
    user_id: Optional[int] = None
    token_key: Optional[str] = None
    request_path: str
    request_method: str
    real_ip: str
    forwarded_for: Optional[str] = None
    user_agent: Optional[str] = None
    request_headers: Optional[str] = None
    request_body: Optional[str] = None
    content_fingerprint: Optional[str] = None
    device_fingerprint: Optional[str] = None
    cookies: Optional[str] = None
    response_status: int
    response_time: int
    timestamp: int


def extract_real_ip(request: Request) -> str:
    """
    Extract real IP from request headers
    Priority: X-Forwarded-For (first value) > X-Real-IP > client IP
    """
    # Check X-Forwarded-For header (take first IP)
    forwarded_for = request.headers.get("X-Forwarded-For", "")
    if forwarded_for:
        # Take the first IP in the chain
        ips = [ip.strip() for ip in forwarded_for.split(",")]
        if ips:
            return ips[0]
    
    # Check X-Real-IP header
    real_ip = request.headers.get("X-Real-IP", "")
    if real_ip:
        return real_ip
    
    # Fallback to client host
    if request.client:
        return request.client.host
    
    return "unknown"


def extract_token_from_auth(authorization: Optional[str]) -> Optional[str]:
    """Extract API token from Authorization header"""
    if not authorization:
        return None
    
    # Handle "Bearer sk-xxx" format
    if authorization.startswith("Bearer "):
        return authorization[7:].strip()
    
    # Handle direct token
    return authorization.strip()


def generate_device_fingerprint(request: Request) -> str:
    """
    Generate device fingerprint from request characteristics
    """
    components = []
    
    # User Agent
    user_agent = request.headers.get("User-Agent", "")
    components.append(user_agent)
    
    # Accept headers
    accept = request.headers.get("Accept", "")
    components.append(accept)
    
    # Accept-Language
    accept_lang = request.headers.get("Accept-Language", "")
    components.append(accept_lang)
    
    # Accept-Encoding
    accept_encoding = request.headers.get("Accept-Encoding", "")
    components.append(accept_encoding)
    
    # Create hash
    fingerprint_str = "|".join(components)
    return hashlib.sha256(fingerprint_str.encode()).hexdigest()


def generate_content_fingerprint(body: str) -> str:
    """Generate fingerprint for request content"""
    return hashlib.md5(body.encode()).hexdigest()


async def log_request_to_newapi(log_entry: Dict[str, Any]):
    """Send log entry to NewAPI for storage"""
    try:
        response = await http_client.post(
            f"{NEWAPI_BASE_URL}/api/heimdall/log",
            json=log_entry,
            timeout=5.0
        )
        if response.status_code != 200:
            logger.warning(f"Failed to log to NewAPI: {response.status_code}")
    except Exception as e:
        logger.error(f"Error logging to NewAPI: {e}")


@app.middleware("http")
async def security_logging_middleware(request: Request, call_next):
    """
    Main security middleware that logs all requests
    """
    start_time = time.time()
    
    # Extract request details
    real_ip = extract_real_ip(request)
    forwarded_for = request.headers.get("X-Forwarded-For", "")
    user_agent = request.headers.get("User-Agent", "")
    authorization = request.headers.get("Authorization", "")
    token_key = extract_token_from_auth(authorization)
    device_fingerprint = generate_device_fingerprint(request)
    
    # Read request body (if present)
    body_bytes = await request.body()
    body_str = body_bytes.decode('utf-8') if body_bytes else ""
    content_fingerprint = generate_content_fingerprint(body_str) if body_str else None
    
    # Get cookies
    cookies_str = json.dumps(dict(request.cookies)) if request.cookies else None
    
    # Process request
    response = await call_next(request)
    
    # Calculate response time
    response_time = int((time.time() - start_time) * 1000)  # milliseconds
    
    # Prepare log entry
    log_entry = {
        "token_key": token_key,
        "request_path": str(request.url.path),
        "request_method": request.method,
        "real_ip": real_ip,
        "forwarded_for": forwarded_for,
        "user_agent": user_agent,
        "request_headers": json.dumps(dict(request.headers)),
        "request_body": body_str[:10000] if body_str else None,  # Limit to 10KB
        "content_fingerprint": content_fingerprint,
        "device_fingerprint": device_fingerprint,
        "cookies": cookies_str,
        "response_status": response.status_code,
        "response_time": response_time,
        "timestamp": int(time.time())
    }
    
    # Log to NewAPI asynchronously
    await log_request_to_newapi(log_entry)
    
    return response


@app.get("/health")
async def health_check():
    """Health check endpoint"""
    return {
        "status": "healthy",
        "service": "Heimdall Security Gateway",
        "timestamp": int(time.time())
    }


@app.post("/api/log")
async def receive_log(log_entry: LogEntry):
    """
    Receive log entries from external sources
    """
    try:
        await log_request_to_newapi(log_entry.dict())
        return {"status": "success", "message": "Log entry received"}
    except Exception as e:
        logger.error(f"Error receiving log: {e}")
        raise HTTPException(status_code=500, detail=str(e))


# OpenAI-compatible endpoints
@app.post("/v1/chat/completions")
async def proxy_chat_completions(request: Request):
    """Proxy chat completions to NewAPI"""
    try:
        body = await request.body()
        headers = dict(request.headers)
        
        # Remove host header to avoid conflicts
        headers.pop("host", None)
        
        response = await http_client.post(
            f"{NEWAPI_BASE_URL}/v1/chat/completions",
            content=body,
            headers=headers
        )
        
        return Response(
            content=response.content,
            status_code=response.status_code,
            headers=dict(response.headers)
        )
    except Exception as e:
        logger.error(f"Error proxying request: {e}")
        raise HTTPException(status_code=500, detail=str(e))


@app.post("/v1/completions")
async def proxy_completions(request: Request):
    """Proxy completions to NewAPI"""
    try:
        body = await request.body()
        headers = dict(request.headers)
        headers.pop("host", None)
        
        response = await http_client.post(
            f"{NEWAPI_BASE_URL}/v1/completions",
            content=body,
            headers=headers
        )
        
        return Response(
            content=response.content,
            status_code=response.status_code,
            headers=dict(response.headers)
        )
    except Exception as e:
        logger.error(f"Error proxying request: {e}")
        raise HTTPException(status_code=500, detail=str(e))


@app.post("/v1/embeddings")
async def proxy_embeddings(request: Request):
    """Proxy embeddings to NewAPI"""
    try:
        body = await request.body()
        headers = dict(request.headers)
        headers.pop("host", None)
        
        response = await http_client.post(
            f"{NEWAPI_BASE_URL}/v1/embeddings",
            content=body,
            headers=headers
        )
        
        return Response(
            content=response.content,
            status_code=response.status_code,
            headers=dict(response.headers)
        )
    except Exception as e:
        logger.error(f"Error proxying request: {e}")
        raise HTTPException(status_code=500, detail=str(e))


@app.api_route("/{path:path}", methods=["GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"])
async def proxy_all(request: Request, path: str):
    """
    Catch-all proxy for any other routes
    """
    try:
        body = await request.body()
        headers = dict(request.headers)
        headers.pop("host", None)
        
        # Build target URL
        target_url = f"{NEWAPI_BASE_URL}/{path}"
        if request.url.query:
            target_url += f"?{request.url.query}"
        
        # Proxy request
        response = await http_client.request(
            method=request.method,
            url=target_url,
            content=body,
            headers=headers
        )
        
        return Response(
            content=response.content,
            status_code=response.status_code,
            headers=dict(response.headers)
        )
    except Exception as e:
        logger.error(f"Error proxying request: {e}")
        raise HTTPException(status_code=500, detail=str(e))


if __name__ == "__main__":
    # Check TLS configuration
    use_tls = bool(TLS_CERT_PATH and TLS_KEY_PATH)
    
    if use_tls:
        logger.info(f"Starting Heimdall with TLS on {HEIMDALL_HOST}:{HEIMDALL_PORT}")
        uvicorn.run(
            app,
            host=HEIMDALL_HOST,
            port=HEIMDALL_PORT,
            ssl_certfile=TLS_CERT_PATH,
            ssl_keyfile=TLS_KEY_PATH,
            log_level="info"
        )
    else:
        logger.info(f"Starting Heimdall without TLS on {HEIMDALL_HOST}:{HEIMDALL_PORT}")
        logger.warning("TLS is not enabled. Set TLS_CERT_PATH and TLS_KEY_PATH for encrypted communication.")
        uvicorn.run(
            app,
            host=HEIMDALL_HOST,
            port=HEIMDALL_PORT,
            log_level="info"
        )
