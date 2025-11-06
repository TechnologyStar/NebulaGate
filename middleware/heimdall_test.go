package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHeimdallMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userId         int
		setupDirective func(int)
		expectedStatus int
		expectBlock    bool
	}{
		{
			name:   "No block directive - user passes through",
			userId: 1,
			setupDirective: func(userId int) {
			},
			expectedStatus: http.StatusOK,
			expectBlock:    false,
		},
		{
			name:   "Block directive exists - user blocked",
			userId: 2,
			setupDirective: func(userId int) {
				if common.RedisEnabled {
					directive := map[string]interface{}{
						"user_id": userId,
						"action":  "block",
					}
					data, _ := json.Marshal(directive)
					common.RedisSet("heimdall:directive:2", string(data), 3600)
				}
			},
			expectedStatus: http.StatusForbidden,
			expectBlock:    true,
		},
		{
			name:   "Ban directive exists - user blocked",
			userId: 3,
			setupDirective: func(userId int) {
				if common.RedisEnabled {
					directive := map[string]interface{}{
						"user_id": userId,
						"action":  "ban",
					}
					data, _ := json.Marshal(directive)
					common.RedisSet("heimdall:directive:3", string(data), 3600)
				}
			},
			expectedStatus: http.StatusForbidden,
			expectBlock:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(func(c *gin.Context) {
				common.SetContextKey(c, constant.ContextKeyUserId, tt.userId)
				c.Next()
			})
			router.Use(Heimdall())
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			tt.setupDirective(tt.userId)

			req, _ := http.NewRequest("GET", "/test", nil)
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			assert.Equal(t, tt.expectedStatus, resp.Code)

			if tt.expectBlock {
				var response map[string]interface{}
				err := json.Unmarshal(resp.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.NotNil(t, response["error"])
			}
		})
	}
}

func TestCheckHeimdallBlock(t *testing.T) {
	if !common.RedisEnabled {
		t.Skip("Redis not enabled, skipping test")
	}

	userId := 100
	directive := map[string]interface{}{
		"user_id": userId,
		"action":  "block",
	}
	data, _ := json.Marshal(directive)
	common.RedisSet("heimdall:directive:100", string(data), 3600)

	blocked, err := checkHeimdallBlock(userId)
	assert.NoError(t, err)
	assert.True(t, blocked)

	common.RedisDel("heimdall:directive:100")

	blocked, err = checkHeimdallBlock(userId)
	assert.NoError(t, err)
	assert.False(t, blocked)
}
