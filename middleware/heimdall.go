package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
)

func Heimdall() gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := common.GetContextKeyInt(c, constant.ContextKeyUserId)
		if userId == 0 {
			c.Next()
			return
		}

		blocked, err := checkHeimdallBlock(userId)
		if err != nil {
			common.SysLog(fmt.Sprintf("heimdall check error: %v", err))
			c.Next()
			return
		}

		if blocked {
			c.JSON(http.StatusForbidden, gin.H{
				"error": gin.H{
					"message": "Access blocked due to security policy",
					"type":    "security_violation",
					"code":    "heimdall_block",
				},
			})
			c.Abort()
			return
		}

		banned, err := service.CheckUserBanned(userId)
		if err == nil && banned {
			c.JSON(http.StatusForbidden, gin.H{
				"error": gin.H{
					"message": "User account has been banned",
					"type":    "security_violation",
					"code":    "user_banned",
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func checkHeimdallBlock(userId int) (bool, error) {
	if !common.RedisEnabled {
		return false, nil
	}

	key := fmt.Sprintf("heimdall:directive:%d", userId)
	data, err := common.RedisGet(key)
	if err != nil || data == "" {
		return false, nil
	}

	var directive map[string]interface{}
	if err := json.Unmarshal([]byte(data), &directive); err != nil {
		return false, err
	}

	action, ok := directive["action"].(string)
	if !ok {
		return false, nil
	}

	return action == "block" || action == "ban", nil
}
