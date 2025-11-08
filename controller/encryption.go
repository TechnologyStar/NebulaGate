package controller

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
)

type EncryptionKeyRequest struct {
	EncryptionKey string `json:"encryption_key" binding:"required"`
}

type EncryptionEnableRequest struct {
	EncryptionKey string `json:"encryption_key" binding:"required"`
	Enable        bool   `json:"enable"`
}

type ConversationLogRequest struct {
	TokenId       int    `json:"token_id" binding:"required"`
	Model         string `json:"model" binding:"required"`
	ConversationData string `json:"conversation_data" binding:"required"`
	PromptTokens  int    `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	MessageCount  int    `json:"message_count"`
	RequestId     string `json:"request_id"`
}

type ConversationLogResponse struct {
	Id                int    `json:"id"`
	TokenId           int    `json:"token_id"`
	Model             string `json:"model"`
	ConversationData  string `json:"conversation_data"` // Decrypted data
	Timestamp         int64  `json:"timestamp"`
	RequestId         string `json:"request_id"`
	MessageCount      int    `json:"message_count"`
	PromptTokens      int    `json:"prompt_tokens"`
	CompletionTokens  int    `json:"completion_tokens"`
}

// GenerateEncryptionKey generates a new encryption key for the user
func GenerateEncryptionKey(c *gin.Context) {
	userId := c.GetInt("id")
	
	// Check if user already has encryption enabled
	user, err := model.GetUserById(userId, true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get user information",
		})
		return
	}
	
	if user.EncryptionEnabled {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Encryption is already enabled. Cannot regenerate key.",
		})
		return
	}
	
	// Generate new encryption key
	encryptionKey, err := common.GenerateEncryptionKey()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to generate encryption key",
		})
		return
	}
	
	// Hash the key for storage
	keyHash := common.HashEncryptionKey(encryptionKey, "")
	
	// Update user
	user.EncryptionKeyHash = keyHash
	user.EncryptionEnabled = false // User must explicitly enable it
	
	err = user.Update(false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to save encryption key",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Encryption key generated successfully. Please save it securely - it cannot be recovered!",
		"data": gin.H{
			"encryption_key": encryptionKey,
		},
	})
}

// EnableEncryption enables end-to-end encryption for the user
func EnableEncryption(c *gin.Context) {
	userId := c.GetInt("id")
	
	var req EncryptionEnableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}
	
	// Get user
	user, err := model.GetUserById(userId, true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get user information",
		})
		return
	}
	
	if user.EncryptionKeyHash == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "No encryption key found. Please generate a key first.",
		})
		return
	}
	
	// Verify the provided key
	if !common.VerifyEncryptionKey(req.EncryptionKey, user.EncryptionKeyHash) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Invalid encryption key",
		})
		return
	}
	
	// Update encryption status
	user.EncryptionEnabled = req.Enable
	err = user.Update(false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to update encryption status",
		})
		return
	}
	
	status := "disabled"
	if req.Enable {
		status = "enabled"
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Encryption " + status + " successfully",
	})
}

// GetEncryptionStatus returns the user's encryption status
func GetEncryptionStatus(c *gin.Context) {
	userId := c.GetInt("id")
	
	user, err := model.GetUserById(userId, true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get user information",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"encryption_enabled": user.EncryptionEnabled,
			"has_encryption_key": user.EncryptionKeyHash != "",
		},
	})
}

// CreateConversationLog creates an encrypted conversation log entry
func CreateConversationLog(c *gin.Context) {
	userId := c.GetInt("id")
	
	var req ConversationLogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}
	
	// Get user
	user, err := model.GetUserById(userId, true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get user information",
		})
		return
	}
	
	if !user.EncryptionEnabled {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Encryption is not enabled",
		})
		return
	}
	
	// Verify token belongs to user
	token, err := model.GetTokenByIds(req.TokenId, userId)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Token not found",
		})
		return
	}
	
	if !token.ConversationLoggingEnabled {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Conversation logging is not enabled for this token",
		})
		return
	}
	
	// Get encryption key from request header
	encryptionKey := c.GetHeader("X-Encryption-Key")
	if encryptionKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Encryption key is required",
		})
		return
	}
	
	// Verify encryption key
	if !common.VerifyEncryptionKey(encryptionKey, user.EncryptionKeyHash) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Invalid encryption key",
		})
		return
	}
	
	// Encrypt conversation data
	encryptedData, nonce, err := common.EncryptData(req.ConversationData, encryptionKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to encrypt conversation data: " + err.Error(),
		})
		return
	}
	
	// Create log entry
	log := &model.ConversationLog{
		UserId:           userId,
		TokenId:          req.TokenId,
		Model:            req.Model,
		EncryptedData:    encryptedData,
		Nonce:            nonce,
		RequestId:        req.RequestId,
		MessageCount:     req.MessageCount,
		PromptTokens:     req.PromptTokens,
		CompletionTokens: req.CompletionTokens,
	}
	
	err = model.CreateConversationLog(log)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to save conversation log",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Conversation log saved successfully",
		"data": gin.H{
			"id": log.Id,
		},
	})
}

// GetConversationLogs retrieves and decrypts conversation logs
func GetConversationLogs(c *gin.Context) {
	userId := c.GetInt("id")
	
	// Get pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	tokenIdStr := c.Query("token_id")
	
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	
	startIdx := (page - 1) * pageSize
	
	// Get user
	user, err := model.GetUserById(userId, true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get user information",
		})
		return
	}
	
	if !user.EncryptionEnabled {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Encryption is not enabled",
		})
		return
	}
	
	// Get encryption key from header
	encryptionKey := c.GetHeader("X-Encryption-Key")
	if encryptionKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Encryption key is required",
		})
		return
	}
	
	// Verify encryption key
	if !common.VerifyEncryptionKey(encryptionKey, user.EncryptionKeyHash) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Invalid encryption key",
		})
		return
	}
	
	// Get logs
	var logs []*model.ConversationLog
	var total int64
	
	if tokenIdStr != "" {
		tokenId, err := strconv.Atoi(tokenIdStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid token_id",
			})
			return
		}
		logs, total, err = model.GetConversationLogsByTokenId(userId, tokenId, startIdx, pageSize)
	} else {
		logs, total, err = model.GetConversationLogsByUserId(userId, startIdx, pageSize)
	}
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to retrieve conversation logs",
		})
		return
	}
	
	// Decrypt logs
	decryptedLogs := make([]ConversationLogResponse, 0, len(logs))
	for _, log := range logs {
		decryptedData, err := common.DecryptData(log.EncryptedData, log.Nonce, encryptionKey)
		if err != nil {
			// Log error but continue with other logs
			common.SysLog("Failed to decrypt log " + strconv.Itoa(log.Id) + ": " + err.Error())
			continue
		}
		
		decryptedLogs = append(decryptedLogs, ConversationLogResponse{
			Id:                log.Id,
			TokenId:           log.TokenId,
			Model:             log.Model,
			ConversationData:  decryptedData,
			Timestamp:         log.Timestamp,
			RequestId:         log.RequestId,
			MessageCount:      log.MessageCount,
			PromptTokens:      log.PromptTokens,
			CompletionTokens:  log.CompletionTokens,
		})
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"logs":  decryptedLogs,
			"total": total,
			"page":  page,
			"page_size": pageSize,
		},
	})
}

// DeleteConversationLog deletes a conversation log
func DeleteConversationLog(c *gin.Context) {
	userId := c.GetInt("id")
	logId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid log ID",
		})
		return
	}
	
	err = model.DeleteConversationLog(logId, userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to delete conversation log",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Conversation log deleted successfully",
	})
}

// HeimdallLogReceiver receives logs from Heimdall gateway
func HeimdallLogReceiver(c *gin.Context) {
	var logData map[string]interface{}
	if err := c.ShouldBindJSON(&logData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid log data",
		})
		return
	}
	
	// Extract fields
	log := &model.HeimdallLog{}
	
	if tokenKey, ok := logData["token_key"].(string); ok {
		log.TokenKey = tokenKey
		
		// Try to get user ID from token
		if tokenKey != "" {
			token, err := model.GetTokenByKey(tokenKey, false)
			if err == nil {
				log.UserId = token.UserId
			}
		}
	}
	
	if requestPath, ok := logData["request_path"].(string); ok {
		log.RequestPath = requestPath
	}
	
	if requestMethod, ok := logData["request_method"].(string); ok {
		log.RequestMethod = requestMethod
	}
	
	if realIP, ok := logData["real_ip"].(string); ok {
		log.RealIP = realIP
	}
	
	if forwardedFor, ok := logData["forwarded_for"].(string); ok {
		log.ForwardedFor = forwardedFor
	}
	
	if userAgent, ok := logData["user_agent"].(string); ok {
		log.UserAgent = userAgent
	}
	
	if requestHeaders, ok := logData["request_headers"].(string); ok {
		log.RequestHeaders = requestHeaders
	}
	
	if requestBody, ok := logData["request_body"].(string); ok {
		log.RequestBody = requestBody
	}
	
	if contentFingerprint, ok := logData["content_fingerprint"].(string); ok {
		log.ContentFingerprint = contentFingerprint
	}
	
	if deviceFingerprint, ok := logData["device_fingerprint"].(string); ok {
		log.DeviceFingerprint = deviceFingerprint
	}
	
	if cookies, ok := logData["cookies"].(string); ok {
		log.Cookies = cookies
	}
	
	if responseStatus, ok := logData["response_status"].(float64); ok {
		log.ResponseStatus = int(responseStatus)
	}
	
	if responseTime, ok := logData["response_time"].(float64); ok {
		log.ResponseTime = int(responseTime)
	}
	
	if timestamp, ok := logData["timestamp"].(float64); ok {
		log.Timestamp = int64(timestamp)
	}
	
	// Save log
	err := model.CreateHeimdallLog(log)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to save log",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Log received successfully",
	})
}
