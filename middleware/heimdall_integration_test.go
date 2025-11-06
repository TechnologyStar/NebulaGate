package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

// setupIntegrationTestRouter creates a router with Heimdall middleware for integration testing
func setupIntegrationTestRouter(config HeimdallConfig) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// Add request ID middleware
	router.Use(func(c *gin.Context) {
		requestID := common.GetTimeString() + common.GetRandomString(8)
		c.Set(common.RequestIdKey, requestID)
		c.Next()
	})
	
	// Add Heimdall middleware
	router.Use(HeimdallAuth(config))
	
	// Add test endpoints
	router.GET("/api/test", func(c *gin.Context) {
		authCtx, exists := c.Get("heimdall_auth_context")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "no auth context"})
			return
		}
		
		ctx := authCtx.(*HeimdallAuthContext)
		c.JSON(http.StatusOK, gin.H{
			"message":    "success",
			"request_id": ctx.RequestID,
			"user_id":    ctx.UserID,
			"token_id":   ctx.TokenID,
			"auth_method": ctx.AuthMethod,
		})
	})
	
	router.POST("/api/test", func(c *gin.Context) {
		var requestBody map[string]interface{}
		if err := c.ShouldBindJSON(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{
			"message": "success",
			"received": requestBody,
		})
	})
	
	return router
}

// createTestToken creates a test token in the database
func createTestToken(t *testing.T) *model.Token {
	token := &model.Token{
		UserId:         1,
		Key:            "sk-test123456789012345678901234567890123456",
		Status:         common.TokenStatusEnabled,
		Name:           "Test Token",
		CreatedTime:    common.GetTimestamp(),
		AccessedTime:   common.GetTimestamp(),
		ExpiredTime:    -1, // Never expires
		RemainQuota:    10000,
		UnlimitedQuota: false,
	}
	
	err := token.Insert()
	assert.NoError(t, err, "Failed to create test token")
	
	return token
}

// cleanupTestToken removes the test token from the database
func cleanupTestToken(t *testing.T, token *model.Token) {
	if token != nil {
		err := token.Delete()
		assert.NoError(t, err, "Failed to cleanup test token")
	}
}

func TestIntegration_Heimdall_AuthFlow(t *testing.T) {
	// Skip if database is not available
	if !common.RedisEnabled && !common.MemoryCacheEnabled {
		t.Skip("Database not available for integration test")
	}
	
	config := DefaultHeimdallConfig()
	config.RateLimitEnabled = false
	config.ReplayProtection = false
	config.AuditLoggingEnabled = true
	
	router := setupIntegrationTestRouter(config)
	
	// Test unauthorized request
	req := httptest.NewRequest("GET", "/api/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "authentication_failed", response["error"])
}

func TestIntegration_Heimdall_ValidAPIKey(t *testing.T) {
	// Skip if database is not available
	if !common.RedisEnabled && !common.MemoryCacheEnabled {
		t.Skip("Database not available for integration test")
	}
	
	config := DefaultHeimdallConfig()
	config.RateLimitEnabled = false
	config.ReplayProtection = false
	config.AuditLoggingEnabled = true
	
	router := setupIntegrationTestRouter(config)
	
	// Create a test token
	token := createTestToken(t)
	defer cleanupTestToken(t, token)
	
	// Test valid API key
	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Authorization", "Bearer "+token.Key)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["message"])
	assert.Equal(t, float64(token.UserId), response["user_id"])
	assert.Equal(t, float64(token.Id), response["token_id"])
	assert.Equal(t, "api_key", response["auth_method"])
}

func TestIntegration_Heimdall_InvalidAPIKey(t *testing.T) {
	config := DefaultHeimdallConfig()
	config.RateLimitEnabled = false
	config.ReplayProtection = false
	
	router := setupIntegrationTestRouter(config)
	
	// Test invalid API key
	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Authorization", "Bearer sk-invalid123456789012345678901234567890")
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "authentication_failed", response["error"])
}

func TestIntegration_Heimdall_JWTAuthentication(t *testing.T) {
	config := DefaultHeimdallConfig()
	config.AuthEnabled = true
	config.APIKeyValidation = false
	config.JWTValidation = true
	config.JWTSecret = "integration-test-secret"
	config.JWTSigningMethod = "HS256"
	config.RateLimitEnabled = false
	config.ReplayProtection = false
	
	router := setupIntegrationTestRouter(config)
	
	// Create a valid JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": 12345,
		"exp":     time.Now().Add(time.Hour).Unix(),
		"iat":     time.Now().Unix(),
		"sub":     "test-user",
	})
	
	tokenString, err := token.SignedString([]byte(config.JWTSecret))
	assert.NoError(t, err)
	
	// Test valid JWT
	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["message"])
	assert.Equal(t, float64(12345), response["user_id"])
	assert.Equal(t, "jwt", response["auth_method"])
}

func TestIntegration_Heimdall_SchemaValidation(t *testing.T) {
	config := DefaultHeimdallConfig()
	config.AuthEnabled = false // Disable auth to focus on validation
	config.SchemaValidation = true
	config.RateLimitEnabled = false
	config.ReplayProtection = false
	
	router := setupIntegrationTestRouter(config)
	
	// Test valid JSON
	validJSON := `{"message": "hello", "data": {"value": 123}}`
	req := httptest.NewRequest("POST", "/api/test", bytes.NewBufferString(validJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	// Test invalid JSON
	invalidJSON := `{"invalid": json, "missing": quote}`
	req = httptest.NewRequest("POST", "/api/test", bytes.NewBufferString(invalidJSON))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "validation_failed", response["error"])
}

func TestIntegration_Heimdall_ReplayProtection(t *testing.T) {
	if !common.RedisEnabled {
		t.Skip("Redis not available for replay protection test")
	}
	
	config := DefaultHeimdallConfig()
	config.AuthEnabled = false
	config.SchemaValidation = false
	config.ReplayProtection = true
	config.ReplayWindow = 1 * time.Minute
	config.RateLimitEnabled = false
	config.AuditLoggingEnabled = true
	
	router := setupIntegrationTestRouter(config)
	
	requestID := "replay-test-request-id-123"
	
	// First request should succeed
	req1 := httptest.NewRequest("GET", "/api/test", nil)
	req1.Header.Set("X-Oneapi-Request-Id", requestID)
	w1 := httptest.NewRecorder()
	
	router.ServeHTTP(w1, req1)
	
	assert.Equal(t, http.StatusOK, w1.Code)
	
	// Second request with same ID should fail
	req2 := httptest.NewRequest("GET", "/api/test", nil)
	req2.Header.Set("X-Oneapi-Request-Id", requestID)
	w2 := httptest.NewRecorder()
	
	router.ServeHTTP(w2, req2)
	
	assert.Equal(t, http.StatusBadRequest, w2.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w2.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "validation_failed", response["error"])
	assert.Contains(t, response["message"], "duplicate request ID")
}

func TestIntegration_Heimdall_RateLimiting(t *testing.T) {
	if !common.RedisEnabled {
		t.Skip("Redis not available for rate limiting test")
	}
	
	config := DefaultHeimdallConfig()
	config.AuthEnabled = false
	config.SchemaValidation = false
	config.ReplayProtection = false
	config.RateLimitEnabled = true
	config.PerKeyRateLimit = 2 // Very low limit for testing
	config.PerIPRateLimit = 5
	config.RateLimitWindow = 1 * time.Second
	
	router := setupIntegrationTestRouter(config)
	
	clientIP := "192.168.1.100"
	
	// Make multiple requests to test rate limiting
	successCount := 0
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest("GET", "/api/test", nil)
		req.RemoteAddr = clientIP
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		if w.Code == http.StatusOK {
			successCount++
		} else if w.Code == http.StatusTooManyRequests {
			// Rate limit hit
			break
		}
	}
	
	// Should allow some requests but then hit rate limit
	assert.True(t, successCount >= 1, "Should allow at least some requests")
	assert.True(t, successCount <= config.PerIPRateLimit, "Should not exceed IP rate limit")
}

func TestIntegration_Heimdall_AuditLogging(t *testing.T) {
	config := DefaultHeimdallConfig()
	config.AuthEnabled = false
	config.SchemaValidation = false
	config.ReplayProtection = false
	config.RateLimitEnabled = false
	config.AuditLoggingEnabled = true
	config.LogPayloadTruncate = true
	config.MaxPayloadSize = 100
	
	router := setupIntegrationTestRouter(config)
	
	// Make a request that should be logged
	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("User-Agent", "test-integration-agent")
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	// The audit log should be created and stored
	// We can't easily test the exact log content without setting up a test Redis instance
	// but we can verify the request was processed successfully
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestIntegration_Heimdall_MultipleAuthMethods(t *testing.T) {
	config := DefaultHeimdallConfig()
	config.AuthEnabled = true
	config.APIKeyValidation = true
	config.JWTValidation = true
	config.JWTSecret = "multi-auth-test-secret"
	config.RateLimitEnabled = false
	config.ReplayProtection = false
	
	router := setupIntegrationTestRouter(config)
	
	// Test API key authentication
	token := createTestToken(t)
	defer cleanupTestToken(t, token)
	
	req1 := httptest.NewRequest("GET", "/api/test", nil)
	req1.Header.Set("Authorization", "Bearer "+token.Key)
	w1 := httptest.NewRecorder()
	
	router.ServeHTTP(w1, req1)
	
	assert.Equal(t, http.StatusOK, w1.Code)
	
	var response1 map[string]interface{}
	err := json.Unmarshal(w1.Body.Bytes(), &response1)
	assert.NoError(t, err)
	assert.Equal(t, "api_key", response1["auth_method"])
	
	// Test JWT authentication
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": 54321,
		"exp":     time.Now().Add(time.Hour).Unix(),
	})
	
	jwtString, err := jwtToken.SignedString([]byte(config.JWTSecret))
	assert.NoError(t, err)
	
	req2 := httptest.NewRequest("GET", "/api/test", nil)
	req2.Header.Set("Authorization", "Bearer "+jwtString)
	w2 := httptest.NewRecorder()
	
	router.ServeHTTP(w2, req2)
	
	assert.Equal(t, http.StatusOK, w2.Code)
	
	var response2 map[string]interface{}
	err = json.Unmarshal(w2.Body.Bytes(), &response2)
	assert.NoError(t, err)
	assert.Equal(t, "jwt", response2["auth_method"])
	assert.Equal(t, float64(54321), response2["user_id"])
}

func TestIntegration_Heimdall_ErrorResponses(t *testing.T) {
	config := DefaultHeimdallConfig()
	config.SchemaValidation = true
	config.AuthEnabled = true
	config.RateLimitEnabled = false
	config.ReplayProtection = false
	
	router := setupIntegrationTestRouter(config)
	
	// Test authentication error response format
	req1 := httptest.NewRequest("GET", "/api/test", nil)
	w1 := httptest.NewRecorder()
	
	router.ServeHTTP(w1, req1)
	
	assert.Equal(t, http.StatusUnauthorized, w1.Code)
	
	var authResponse map[string]interface{}
	err := json.Unmarshal(w1.Body.Bytes(), &authResponse)
	assert.NoError(t, err)
	assert.Equal(t, "authentication_failed", authResponse["error"])
	assert.Contains(t, authResponse, "message")
	assert.Contains(t, authResponse, "request_id")
	
	// Test validation error response format
	req2 := httptest.NewRequest("POST", "/api/test", bytes.NewBufferString(`{invalid json}`))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	
	router.ServeHTTP(w2, req2)
	
	assert.Equal(t, http.StatusBadRequest, w2.Code)
	
	var validationResponse map[string]interface{}
	err = json.Unmarshal(w2.Body.Bytes(), &validationResponse)
	assert.NoError(t, err)
	assert.Equal(t, "validation_failed", validationResponse["error"])
	assert.Contains(t, validationResponse, "message")
	assert.Contains(t, validationResponse, "request_id")
}

// Benchmark_HeimdallAuth benchmarks the Heimdall middleware performance
func Benchmark_HeimdallAuth(b *testing.B) {
	config := DefaultHeimdallConfig()
	config.AuthEnabled = false // Disable auth for pure performance test
	config.SchemaValidation = false
	config.ReplayProtection = false
	config.RateLimitEnabled = false
	config.AuditLoggingEnabled = false // Disable logging for performance
	
	router := setupIntegrationTestRouter(config)
	
	req := httptest.NewRequest("GET", "/api/test", nil)
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
		}
	})
}

// Benchmark_HeimdallAuth_WithAuth benchmarks Heimdall with authentication enabled
func Benchmark_HeimdallAuth_WithAuth(b *testing.B) {
	config := DefaultHeimdallConfig()
	config.RateLimitEnabled = false
	config.ReplayProtection = false
	config.AuditLoggingEnabled = false
	
	router := setupIntegrationTestRouter(config)
	
	// Create a test token for benchmarking
	token := createTestToken(&testing.T{})
	defer cleanupTestToken(&testing.T{}, token)
	
	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Authorization", "Bearer "+token.Key)
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
		}
	})
}