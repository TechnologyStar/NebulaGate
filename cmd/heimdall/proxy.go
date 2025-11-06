package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/setting/heimdall"
	"github.com/gin-gonic/gin"
)

type ProxyHandler struct {
	config *heimdall.Config
	client *http.Client
}

func NewProxyHandler(config *heimdall.Config) *ProxyHandler {
	return &ProxyHandler{
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				IdleConnTimeout:     90 * time.Second,
				DisableCompression:  false,
			},
		},
	}
}

func (p *ProxyHandler) Handle(targetPath string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Build target URL
		targetURL, err := url.Parse(p.config.BackendURL)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("Invalid backend URL: %v", err),
				"type":  "heimdall_config_error",
			})
			return
		}

		// Append the target path to the backend URL
		targetURL.Path = strings.TrimSuffix(targetURL.Path, "/") + targetPath
		
		// Preserve query parameters
		if c.Request.URL.RawQuery != "" {
			targetURL.RawQuery = c.Request.URL.RawQuery
		}

		// Read request body
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		// Create new request to backend
		req, err := http.NewRequestWithContext(c.Request.Context(), c.Request.Method, targetURL.String(), bytes.NewBuffer(bodyBytes))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("Failed to create proxy request: %v", err),
				"type":  "heimdall_proxy_error",
			})
			return
		}

		// Copy headers, excluding hop-by-hop headers
		p.copyHeaders(c.Request.Header, req.Header, true)

		// Add authentication if configured
		if p.config.APIKeyValue != "" {
			req.Header.Set(p.config.APIKeyHeader, p.config.APIKeyValue)
		}

		// Add custom headers to identify the gateway
		req.Header.Set("X-Forwarded-For", c.ClientIP())
		req.Header.Set("X-Forwarded-Proto", "https")
		req.Header.Set("X-Forwarded-Host", c.Request.Host)
		req.Header.Set("X-Gateway", "heimdall")
		req.Header.Set("X-Gateway-Version", "1.0.0")

		// Make the request
		resp, err := p.client.Do(req)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{
				"error": fmt.Sprintf("Backend request failed: %v", err),
				"type":  "heimdall_backend_error",
			})
			return
		}
		defer resp.Body.Close()

		// Read response body
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("Failed to read response: %v", err),
				"type":  "heimdall_response_error",
			})
			return
		}

		// Copy response headers, excluding hop-by-hop headers
		p.copyHeaders(resp.Header, c.Writer.Header(), false)

		// Set status code
		c.Status(resp.StatusCode)

		// Write response body
		c.Writer.Write(respBody)
	}
}

func (p *ProxyHandler) copyHeaders(src, dst http.Header, isRequest bool) {
	// Hop-by-hop headers that should not be copied
	hopByHopHeaders := map[string]bool{
		"Connection":          true,
		"Keep-Alive":         true,
		"Proxy-Authenticate": true,
		"Proxy-Authorization": true,
		"Te":                 true,
		"Trailers":           true,
		"Transfer-Encoding":   true,
		"Upgrade":           true,
	}

	// Additional headers to handle differently for requests
	requestSpecificHeaders := map[string]bool{
		"Host":              true,
		"Content-Length":    true,
	}

	for key, values := range src {
		// Skip hop-by-hop headers
		if hopByHopHeaders[http.CanonicalHeaderKey(key)] {
			continue
		}

		// Skip request-specific headers when copying request headers
		if isRequest && requestSpecificHeaders[http.CanonicalHeaderKey(key)] {
			continue
		}

		// Copy header values
		for _, value := range values {
			dst.Add(key, value)
		}
	}
}

// HealthCheckHandler provides a more detailed health check
func (p *ProxyHandler) HealthCheckHandler(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check backend connectivity
	healthURL := p.config.BackendURL + "/health"
	req, err := http.NewRequestWithContext(ctx, "GET", healthURL, nil)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unhealthy",
			"service": "heimdall",
			"error": fmt.Sprintf("Failed to create health check request: %v", err),
			"timestamp": time.Now().UTC(),
		})
		return
	}

	// Add authentication if configured
	if p.config.APIKeyValue != "" {
		req.Header.Set(p.config.APIKeyHeader, p.config.APIKeyValue)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unhealthy",
			"service": "heimdall",
			"error": fmt.Sprintf("Backend health check failed: %v", err),
			"timestamp": time.Now().UTC(),
		})
		return
	}
	defer resp.Body.Close()

	// Read backend health response
	backendBody, _ := io.ReadAll(resp.Body)

	c.JSON(resp.StatusCode, gin.H{
		"status": "healthy",
		"service": "heimdall",
		"backend": gin.H{
			"status": resp.Status,
			"response": string(backendBody),
		},
		"timestamp": time.Now().UTC(),
	})
}

// MetricsHandler provides basic metrics
func (p *ProxyHandler) MetricsHandler(c *gin.Context) {
	// This is a placeholder for metrics
	// In a full implementation, you would track:
	// - Request counts by endpoint
	// - Response times
	// - Error rates
	// - Backend health status
	c.JSON(http.StatusOK, gin.H{
		"service": "heimdall",
		"version": "1.0.0",
		"uptime": "0s", // This should be tracked from service start
		"requests": gin.H{
			"total": 0,
			"success": 0,
			"error": 0,
		},
		"backend": gin.H{
			"url": p.config.BackendURL,
			"status": "unknown", // This should be tracked from health checks
		},
		"timestamp": time.Now().UTC(),
	})
}