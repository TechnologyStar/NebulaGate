package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHeimdallRequestLog_Integration tests the HeimdallRequestLog model integration
func TestHeimdallRequestLog_Integration(t *testing.T) {
	// Skip if LOG_DB is not available
	if LOG_DB == nil {
		t.Skip("LOG_DB not available for integration test")
	}
	
	// Create a test log entry
	log := &HeimdallRequestLog{
		RequestId:            "test-integration-123",
		OccurredAt:           time.Now().UTC(),
		AuthKeyFingerprint:   "fp1234567890abcdef",
		UserId:               intPtr(42),
		TokenId:              intPtr(123),
		NormalizedURL:        "/v1/chat/completions",
		HTTPMethod:           "POST",
		HTTPStatus:           200,
		LatencyMs:            150,
		ClientIP:             "192.168.1.1",
		ClientUserAgent:      "Test Agent",
		ClientDeviceId:       "device123",
		RequestSizeBytes:     1024,
		ResponseSizeBytes:    2048,
		ParamDigest:          "abcd1234efgh5678",
		SanitizedCookies:     "session=***; theme=dark",
		CountryCode:          "US",
		Region:               "California",
		City:                "San Francisco",
		ProcessingTimeMs:     145,
		UpstreamProvider:     "openai",
		ModelName:            "gpt-4",
	}
	
	// Test Create
	err := LOG_DB.Create(log).Error
	assert.NoError(t, err)
	assert.NotZero(t, log.Id)
	
	// Test Read
	var retrievedLog HeimdallRequestLog
	err = LOG_DB.First(&retrievedLog, log.Id).Error
	assert.NoError(t, err)
	
	// Verify fields
	assert.Equal(t, log.RequestId, retrievedLog.RequestId)
	assert.Equal(t, log.AuthKeyFingerprint, retrievedLog.AuthKeyFingerprint)
	assert.Equal(t, log.UserId, retrievedLog.UserId)
	assert.Equal(t, log.TokenId, retrievedLog.TokenId)
	assert.Equal(t, log.NormalizedURL, retrievedLog.NormalizedURL)
	assert.Equal(t, log.HTTPMethod, retrievedLog.HTTPMethod)
	assert.Equal(t, log.HTTPStatus, retrievedLog.HTTPStatus)
	assert.Equal(t, log.LatencyMs, retrievedLog.LatencyMs)
	assert.Equal(t, log.ClientIP, retrievedLog.ClientIP)
	assert.Equal(t, log.ClientUserAgent, retrievedLog.ClientUserAgent)
	assert.Equal(t, log.ClientDeviceId, retrievedLog.ClientDeviceId)
	assert.Equal(t, log.RequestSizeBytes, retrievedLog.RequestSizeBytes)
	assert.Equal(t, log.ResponseSizeBytes, retrievedLog.ResponseSizeBytes)
	assert.Equal(t, log.ParamDigest, retrievedLog.ParamDigest)
	assert.Equal(t, log.SanitizedCookies, retrievedLog.SanitizedCookies)
	assert.Equal(t, log.CountryCode, retrievedLog.CountryCode)
	assert.Equal(t, log.Region, retrievedLog.Region)
	assert.Equal(t, log.City, retrievedLog.City)
	assert.Equal(t, log.ProcessingTimeMs, retrievedLog.ProcessingTimeMs)
	assert.Equal(t, log.UpstreamProvider, retrievedLog.UpstreamProvider)
	assert.Equal(t, log.ModelName, retrievedLog.ModelName)
	
	// Test Update
	retrievedLog.HTTPStatus = 500
	retrievedLog.ErrorMessage = "Internal server error"
	retrievedLog.ErrorType = "server_error"
	
	err = LOG_DB.Save(&retrievedLog).Error
	assert.NoError(t, err)
	
	// Verify update
	var updatedLog HeimdallRequestLog
	err = LOG_DB.First(&updatedLog, log.Id).Error
	assert.NoError(t, err)
	assert.Equal(t, 500, updatedLog.HTTPStatus)
	assert.Equal(t, "Internal server error", updatedLog.ErrorMessage)
	assert.Equal(t, "server_error", updatedLog.ErrorType)
	
	// Test Delete
	err = LOG_DB.Delete(&updatedLog).Error
	assert.NoError(t, err)
	
	// Verify deletion
	var deletedLog HeimdallRequestLog
	err = LOG_DB.First(&deletedLog, log.Id).Error
	assert.Error(t, err) // Should not find the record
}

// TestHeimdallRequestLog_BeforeCreate tests the BeforeCreate hook
func TestHeimdallRequestLog_BeforeCreate(t *testing.T) {
	log := &HeimdallRequestLog{
		RequestId:     "test-hook-123",
		NormalizedURL:  "/test",
		HTTPMethod:     "GET",
		HTTPStatus:     200,
	}
	
	// Call BeforeCreate
	err := log.BeforeCreate(nil)
	assert.NoError(t, err)
	assert.False(t, log.OccurredAt.IsZero())
	assert.True(t, log.OccurredAt.Before(time.Now().UTC().Add(time.Second)))
}

// TestHeimdallRequestLog_TableName tests the TableName method
func TestHeimdallRequestLog_TableName(t *testing.T) {
	log := &HeimdallRequestLog{}
	expected := "heimdall_request_logs"
	assert.Equal(t, expected, log.TableName())
}

// TestHeimdallRequestLog_QueryPerformance tests query performance
func TestHeimdallRequestLog_QueryPerformance(t *testing.T) {
	if LOG_DB == nil {
		t.Skip("LOG_DB not available for integration test")
	}
	
	// Create multiple test entries
	for i := 0; i < 100; i++ {
		log := &HeimdallRequestLog{
			RequestId:     "perf-test-" + string(rune(i)),
			OccurredAt:     time.Now().UTC().Add(-time.Duration(i) * time.Minute),
			NormalizedURL:  "/v1/chat/completions",
			HTTPMethod:     "POST",
			HTTPStatus:     200,
			LatencyMs:      int64(100 + i),
			ClientIP:       "192.168.1.1",
			ParamDigest:    "digest123",
		}
		
		err := LOG_DB.Create(log).Error
		assert.NoError(t, err)
	}
	
	// Test query by URL
	start := time.Now()
	var logs []HeimdallRequestLog
	err := LOG_DB.Where("normalized_url = ?", "/v1/chat/completions").
		Order("occurred_at DESC").
		Limit(50).
		Find(&logs).Error
	queryTime := time.Since(start)
	
	assert.NoError(t, err)
	assert.LessOrEqual(t, len(logs), 50)
	assert.Less(t, queryTime, 100*time.Millisecond, "Query should complete within 100ms")
	
	// Test query by time range
	start = time.Now()
	var timeRangeLogs []HeimdallRequestLog
	cutoff := time.Now().UTC().Add(-30 * time.Minute)
	err = LOG_DB.Where("occurred_at >= ?", cutoff).
		Order("occurred_at DESC").
		Find(&timeRangeLogs).Error
	queryTime = time.Since(start)
	
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(timeRangeLogs), 0)
	assert.Less(t, queryTime, 100*time.Millisecond, "Time range query should complete within 100ms")
	
	// Cleanup
	LOG_DB.Where("request_id LIKE ?", "perf-test-%").Delete(&HeimdallRequestLog{})
}

// TestHeimdallRequestLog_Indexes tests that indexes work correctly
func TestHeimdallRequestLog_Indexes(t *testing.T) {
	if LOG_DB == nil {
		t.Skip("LOG_DB not available for integration test")
	}
	
	// Create test entries with different values
	logs := []*HeimdallRequestLog{
		{
			RequestId:     "index-test-1",
			OccurredAt:     time.Now().UTC(),
			NormalizedURL:  "/v1/chat/completions",
			HTTPMethod:     "POST",
			HTTPStatus:     200,
			LatencyMs:      100,
			ClientIP:       "192.168.1.1",
			ParamDigest:    "digest1",
		},
		{
			RequestId:     "index-test-2",
			OccurredAt:     time.Now().UTC(),
			NormalizedURL:  "/v1/models",
			HTTPMethod:     "GET",
			HTTPStatus:     200,
			LatencyMs:      50,
			ClientIP:       "192.168.1.2",
			ParamDigest:    "digest2",
		},
		{
			RequestId:     "index-test-3",
			OccurredAt:     time.Now().UTC(),
			NormalizedURL:  "/v1/chat/completions",
			HTTPMethod:     "POST",
			HTTPStatus:     500,
			LatencyMs:      1000,
			ClientIP:       "192.168.1.1",
			ParamDigest:    "digest3",
		},
	}
	
	for _, log := range logs {
		err := LOG_DB.Create(log).Error
		assert.NoError(t, err)
	}
	
	// Test index on normalized_url
	start := time.Now()
	var urlLogs []HeimdallRequestLog
	err := LOG_DB.Where("normalized_url = ?", "/v1/chat/completions").Find(&urlLogs).Error
	urlQueryTime := time.Since(start)
	
	assert.NoError(t, err)
	assert.Len(t, urlLogs, 2)
	assert.Less(t, urlQueryTime, 50*time.Millisecond, "URL index query should be fast")
	
	// Test index on client_ip
	start = time.Now()
	var ipLogs []HeimdallRequestLog
	err = LOG_DB.Where("client_ip = ?", "192.168.1.1").Find(&ipLogs).Error
	ipQueryTime := time.Since(start)
	
	assert.NoError(t, err)
	assert.Len(t, ipLogs, 2)
	assert.Less(t, ipQueryTime, 50*time.Millisecond, "IP index query should be fast")
	
	// Test index on request_id (unique)
	start = time.Now()
	var uniqueLog HeimdallRequestLog
	err = LOG_DB.Where("request_id = ?", "index-test-1").First(&uniqueLog).Error
	uniqueQueryTime := time.Since(start)
	
	assert.NoError(t, err)
	assert.Equal(t, "index-test-1", uniqueLog.RequestId)
	assert.Less(t, uniqueQueryTime, 10*time.Millisecond, "Unique request_id query should be very fast")
	
	// Cleanup
	LOG_DB.Where("request_id LIKE ?", "index-test-%").Delete(&HeimdallRequestLog{})
}

// Helper function to create int pointer
func intPtr(i int) *int {
	return &i
}
