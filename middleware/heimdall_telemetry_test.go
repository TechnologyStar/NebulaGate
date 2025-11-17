package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/QuantumNous/new-api/model"
)

// MockDB is a mock database for testing
type MockDB struct {
	mock.Mock
}

func (m *MockDB) Create(value interface{}) error {
	args := m.Called(value)
	return args.Error(0)
}

// MockRedis is a mock Redis client for testing
type MockRedis struct {
	mock.Mock
}

func (m *MockRedis) Incr(ctx context.Context, key string) (int64, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRedis) Expire(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	args := m.Called(ctx, key, expiration)
	return args.Bool(0), args.Error(1)
}

func TestDefaultTelemetryConfig(t *testing.T) {
	config := DefaultTelemetryConfig()
	
	assert.True(t, config.Enabled)
	assert.False(t, config.GeolocationEnabled)
	assert.Equal(t, 10000, config.BufferSize)
	assert.Equal(t, 5, config.WorkerCount)
	assert.Equal(t, 3, config.RetryAttempts)
	assert.Equal(t, time.Second, config.RetryDelay)
	assert.True(t, config.DiskQueueEnabled)
	assert.Equal(t, "/tmp/heimdall_queue", config.DiskQueuePath)
	assert.Equal(t, 5*time.Second, config.FlushInterval)
}

func TestHeimdallTelemetryMiddleware_Disabled(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	config := TelemetryConfig{
		Enabled: false,
	}
	
	middleware := HeimdallTelemetryMiddleware(config)
	
	router := gin.New()
	router.Use(middleware)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})
	
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHeimdallTelemetryMiddleware_Enabled(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	config := TelemetryConfig{
		Enabled:     true,
		BufferSize:  10,
		WorkerCount: 1,
	}
	
	middleware := HeimdallTelemetryMiddleware(config)
	
	router := gin.New()
	router.Use(middleware)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})
	
	// Create a request with headers
	req := httptest.NewRequest("GET", "/test?param=value", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.1")
	req.Header.Set("User-Agent", "Test Agent")
	req.Header.Set("X-Device-Id", "device123")
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHeimdallTelemetryMiddleware_WithRequestBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	config := TelemetryConfig{
		Enabled:     true,
		BufferSize:  10,
		WorkerCount: 1,
	}
	
	middleware := HeimdallTelemetryMiddleware(config)
	
	router := gin.New()
	router.Use(middleware)
	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})
	
	// Create a request with body
	requestBody := map[string]interface{}{
		"model":    "gpt-4",
		"messages": []interface{}{map[string]interface{}{"role": "user", "content": "Hello"}},
	}
	bodyBytes, _ := json.Marshal(requestBody)
	
	req := httptest.NewRequest("POST", "/test", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Forwarded-For", "192.168.1.1")
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHeimdallTelemetryMiddleware_ErrorResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	config := TelemetryConfig{
		Enabled:     true,
		BufferSize:  10,
		WorkerCount: 1,
	}
	
	middleware := HeimdallTelemetryMiddleware(config)
	
	router := gin.New()
	router.Use(middleware)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
	})
	
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.1")
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestBuildTelemetryLog(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	// Create a gin context
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewBuffer([]byte(`{"model": "gpt-4", "messages": [{"role": "user", "content": "Hello"}]}`)))
	c.Request.Header.Set("X-Forwarded-For", "192.168.1.1")
	c.Request.Header.Set("User-Agent", "Test Agent")
	c.Request.Header.Set("Authorization", "Bearer sk-test123")
	c.Request.Header.Set("Cookie", "session=abc123; theme=dark")
	
	// Set context values
	c.Set("request_id", "test-request-123")
	c.Set("id", 42)
	c.Set("token_id", 123)
	c.Set("model_name", "gpt-4")
	c.Set("channel_name", "openai")
	
	start := time.Now()
	end := start.Add(100 * time.Millisecond)
	requestBody := []byte(`{"model": "gpt-4", "messages": [{"role": "user", "content": "Hello"}]}`)
	responseBody := []byte(`{"choices": [{"message": {"content": "Hello!"}}]}`)
	
	log := buildTelemetryLog(c, start, end, requestBody, responseBody)
	
	require.NotNil(t, log)
	assert.Equal(t, "test-request-123", log.RequestId)
	assert.Equal(t, "/v1/chat/completions", log.NormalizedURL)
	assert.Equal(t, "POST", log.HTTPMethod)
	assert.Equal(t, int64(100), log.LatencyMs)
	assert.Equal(t, "192.168.1.1", log.ClientIP)
	assert.Equal(t, "Test Agent", log.ClientUserAgent)
	assert.Equal(t, int64(requestBody), log.RequestSizeBytes)
	assert.Equal(t, int64(responseBody), log.ResponseSizeBytes)
	assert.NotEmpty(t, log.ParamDigest)
	assert.Equal(t, "session=***; theme=dark", log.SanitizedCookies)
	assert.Equal(t, "gpt-4", log.ModelName)
	assert.Equal(t, "openai", log.UpstreamProvider)
	
	// Check user and token IDs
	assert.NotNil(t, log.UserId)
	assert.Equal(t, 42, *log.UserId)
	assert.NotNil(t, log.TokenId)
	assert.Equal(t, 123, *log.TokenId)
}

func TestResponseBodyWriter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	// Create a response writer wrapper
	recorder := httptest.NewRecorder()
	writer := &responseBodyWriter{
		ResponseWriter: recorder,
		body:          &bytes.Buffer{},
	}
	
	// Write some data
	data := []byte("test response data")
	n, err := writer.Write(data)
	
	assert.NoError(t, err)
	assert.Equal(t, len(data), n)
	assert.Equal(t, data, writer.body.Bytes())
	assert.Equal(t, data, recorder.Body.Bytes())
}

func TestCategorizeError(t *testing.T) {
	tests := []struct {
		statusCode int
		expected   string
	}{
		{200, ""},
		{201, ""},
		{399, ""},
		{400, "client_error"},
		{401, "client_error"},
		{404, "client_error"},
		{499, "client_error"},
		{500, "server_error"},
		{502, "server_error"},
		{599, "server_error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := categorizeError(tt.statusCode)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTelemetryWorker_StartStop(t *testing.T) {
	config := TelemetryConfig{
		Enabled:         true,
		BufferSize:      10,
		WorkerCount:     2,
		RetryAttempts:   1,
		RetryDelay:      10 * time.Millisecond,
		DiskQueueEnabled: false, // Disable disk queue for testing
		FlushInterval:   100 * time.Millisecond,
	}
	
	worker := NewTelemetryWorker(config)
	require.NotNil(t, worker)
	
	// Test start
	worker.Start()
	
	// Check if worker is running
	stats := worker.GetTelemetryStats()
	assert.True(t, stats["running"].(bool))
	assert.Equal(t, 2, stats["worker_count"].(int))
	assert.Equal(t, 10, stats["buffer_capacity"].(int))
	
	// Test stop
	worker.Stop()
	
	// Check if worker is stopped
	stats = worker.GetTelemetryStats()
	assert.False(t, stats["running"].(bool))
}

func TestTelemetryWorker_ProcessEntry(t *testing.T) {
	// This test would require mocking the database
	// For now, we'll test the basic structure
	config := TelemetryConfig{
		Enabled:         true,
		BufferSize:      10,
		WorkerCount:     1,
		RetryAttempts:   1,
		RetryDelay:      10 * time.Millisecond,
		DiskQueueEnabled: false,
		FlushInterval:   100 * time.Millisecond,
	}
	
	worker := NewTelemetryWorker(config)
	require.NotNil(t, worker)
	
	// Create a test entry
	entry := &TelemetryEntry{
		RequestStart: time.Now(),
		RequestEnd:   time.Now().Add(100 * time.Millisecond),
		Log: &model.HeimdallRequestLog{
			RequestId:     "test-123",
			NormalizedURL: "/test",
			HTTPMethod:    "GET",
			HTTPStatus:    200,
			LatencyMs:     100,
			ClientIP:      "192.168.1.1",
		},
	}
	
	// Process the entry (this will fail to persist to DB since we don't have a real DB)
	// But it should not panic
	worker.processEntry(entry)
	
	worker.Stop()
}

func TestNewTelemetryWorker(t *testing.T) {
	config := TelemetryConfig{
		Enabled:         true,
		BufferSize:      100,
		WorkerCount:     3,
		DiskQueueEnabled: true,
		DiskQueuePath:    "/tmp/test_queue",
	}
	
	worker := NewTelemetryWorker(config)
	
	require.NotNil(t, worker)
	assert.Equal(t, config, worker.config)
	assert.NotNil(t, worker.entryChan)
	assert.NotNil(t, worker.stopChan)
	assert.NotNil(t, worker.diskQueue)
	
	worker.Stop()
}

func TestGetGeolocation(t *testing.T) {
	countryCode, region, city := getGeolocation("192.168.1.1")
	
	// Currently returns empty strings (placeholder implementation)
	assert.Empty(t, countryCode)
	assert.Empty(t, region)
	assert.Empty(t, city)
}

func TestGlobalTelemetryFunctions(t *testing.T) {
	// Test initialization
	InitHeimdallTelemetry()
	
	// Test getting stats
	stats := GetHeimdallTelemetryStats()
	assert.NotNil(t, stats)
	
	// Test stopping
	StopHeimdallTelemetry()
}

// Integration test with actual disk queue
func TestDiskQueueIntegration(t *testing.T) {
	tempDir := t.TempDir()
	config := TelemetryConfig{
		Enabled:         true,
		BufferSize:      10,
		WorkerCount:     1,
		RetryAttempts:   1,
		RetryDelay:      10 * time.Millisecond,
		DiskQueueEnabled: true,
		DiskQueuePath:    tempDir,
		FlushInterval:   50 * time.Millisecond,
	}
	
	worker := NewTelemetryWorker(config)
	require.NotNil(t, worker)
	
	worker.Start()
	
	// Create a test entry
	entry := &TelemetryEntry{
		RequestStart: time.Now(),
		RequestEnd:   time.Now().Add(100 * time.Millisecond),
		Log: &model.HeimdallRequestLog{
			RequestId:     "test-integration-123",
			NormalizedURL: "/test",
			HTTPMethod:    "GET",
			HTTPStatus:    200,
			LatencyMs:     100,
			ClientIP:      "192.168.1.1",
		},
	}
	
	// Send entry to worker
	select {
	case worker.entryChan <- entry:
		// Successfully queued
	default:
		t.Fatal("Failed to queue entry")
	}
	
	// Wait a bit for processing
	time.Sleep(100 * time.Millisecond)
	
	worker.Stop()
	
	// Check disk queue stats
	if worker.diskQueue != nil {
		stats := worker.diskQueue.GetQueueStats()
		assert.NotNil(t, stats)
	}
}

// Benchmark tests
func BenchmarkBuildTelemetryLog(b *testing.B) {
	gin.SetMode(gin.TestMode)
	
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewBuffer([]byte(`{"model": "gpt-4"}`)))
	c.Request.Header.Set("X-Forwarded-For", "192.168.1.1")
	c.Request.Header.Set("User-Agent", "Test Agent")
	c.Set("request_id", "test-123")
	
	start := time.Now()
	end := start.Add(100 * time.Millisecond)
	requestBody := []byte(`{"model": "gpt-4"}`)
	responseBody := []byte(`{"response": "test"}`)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = buildTelemetryLog(c, start, end, requestBody, responseBody)
	}
}

func BenchmarkResponseBodyWriter_Write(b *testing.B) {
	recorder := httptest.NewRecorder()
	writer := &responseBodyWriter{
		ResponseWriter: recorder,
		body:          &bytes.Buffer{},
	}
	
	data := []byte("test response data")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		writer.body.Reset()
		recorder.Body.Reset()
		writer.Write(data)
	}
}
