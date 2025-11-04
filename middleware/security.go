package middleware

import (
	"fmt"
	"net/http"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
)

// SecurityCheck middleware checks user security status (ban, redirect)
func SecurityCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip if not a relay request
		if c.Request == nil || (c.Request.Method != http.MethodPost && c.Request.Method != http.MethodPatch && c.Request.Method != http.MethodPut) {
			c.Next()
			return
		}

		userId := common.GetContextKeyInt(c, constant.ContextKeyUserId)
		if userId == 0 {
			// No user ID, skip security check
			c.Next()
			return
		}

		// Check if user is banned
		banned, err := service.CheckUserBanned(userId)
		if err != nil {
			logger.LogError(c, fmt.Sprintf("failed to check user ban status: %v", err))
			c.Next()
			return
		}

		if banned {
			c.JSON(http.StatusForbidden, gin.H{
				"error": gin.H{
					"message": "Access denied. Your account has been banned due to policy violations.",
					"type":    "forbidden",
					"code":    "user_banned",
				},
			})
			c.Abort()
			return
		}

		// Check for user-level redirect
		redirectModel, err := service.GetUserRedirectModel(userId)
		if err != nil {
			logger.LogError(c, fmt.Sprintf("failed to get user redirect model: %v", err))
		}

		if redirectModel != "" {
			// Apply user-level redirect
			common.SetContextKey(c, constant.ContextKeySecurityRedirectModel, redirectModel)
			logger.LogWarn(c, fmt.Sprintf("user %d redirected to model %s", userId, redirectModel))
		}

		c.Next()
	}
}
