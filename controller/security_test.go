package controller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}

func TestGetSecurityDashboard(t *testing.T) {
	router := setupTestRouter()
	router.GET("/api/security/dashboard", GetSecurityDashboard)

	req, _ := http.NewRequest("GET", "/api/security/dashboard", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var response map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response["success"].(bool))
	assert.NotNil(t, response["data"])
}

func TestGetDeviceClusters(t *testing.T) {
	router := setupTestRouter()
	router.GET("/api/security/devices", GetDeviceClusters)

	req, _ := http.NewRequest("GET", "/api/security/devices?page=1&page_size=10", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var response map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response["success"].(bool))
	
	data := response["data"].(map[string]interface{})
	assert.NotNil(t, data["devices"])
	assert.NotNil(t, data["total"])
}

func TestGetIPClusters(t *testing.T) {
	router := setupTestRouter()
	router.GET("/api/security/ip-clusters", GetIPClusters)

	req, _ := http.NewRequest("GET", "/api/security/ip-clusters?page=1&page_size=10", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var response map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response["success"].(bool))
	
	data := response["data"].(map[string]interface{})
	assert.NotNil(t, data["clusters"])
	assert.NotNil(t, data["total"])
}

func TestGetSecurityAnomalies(t *testing.T) {
	router := setupTestRouter()
	router.GET("/api/security/anomalies", GetSecurityAnomalies)

	req, _ := http.NewRequest("GET", "/api/security/anomalies?page=1&page_size=10&severity=malicious", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var response map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response["success"].(bool))
	
	data := response["data"].(map[string]interface{})
	assert.NotNil(t, data["anomalies"])
	assert.NotNil(t, data["total"])
}

func TestApproveAnomaly(t *testing.T) {
	router := setupTestRouter()
	router.POST("/api/security/anomalies/:id/approve", func(c *gin.Context) {
		c.Set("id", 1)
		ApproveAnomaly(c)
	})

	body := `{"rationale": "False positive, user verified"}`
	req, _ := http.NewRequest("POST", "/api/security/anomalies/1/approve", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var response map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response["success"].(bool))
}

func TestIgnoreAnomaly(t *testing.T) {
	router := setupTestRouter()
	router.POST("/api/security/anomalies/:id/ignore", func(c *gin.Context) {
		c.Set("id", 1)
		IgnoreAnomaly(c)
	})

	body := `{"rationale": "Testing anomaly, can be ignored"}`
	req, _ := http.NewRequest("POST", "/api/security/anomalies/1/ignore", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var response map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response["success"].(bool))
}

func TestEnforcementWorkflow(t *testing.T) {
	userId := 999

	anomaly, err := service.CreateAnomaly(
		userId,
		nil,
		service.AnomalyTypeHighRPM,
		"malicious",
		"Test anomaly for enforcement",
		map[string]interface{}{
			"rpm": 150,
		},
		"192.168.1.100",
		"test-device-123",
		85,
	)

	assert.NoError(t, err)
	assert.NotNil(t, anomaly)
	assert.Equal(t, userId, anomaly.UserId)
	assert.Equal(t, "malicious", anomaly.Severity)

	time.Sleep(100 * time.Millisecond)

	updatedAnomaly, err := model.GetSecurityAnomaly(anomaly.Id)
	assert.NoError(t, err)
	
	if updatedAnomaly.Status == service.StatusActioned {
		assert.NotEmpty(t, updatedAnomaly.ActionTaken)
		assert.NotNil(t, updatedAnomaly.ActionedAt)
	}
}
