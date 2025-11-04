package controller

import (
	"net/http"
	"strconv"
	"time"

	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
)

// AddIPToBlacklist adds an IP to blacklist
func AddIPToBlacklist(c *gin.Context) {
	var req struct {
		IP        string     `json:"ip" binding:"required"`
		Reason    string     `json:"reason"`
		Scope     string     `json:"scope"`
		ScopeID   int        `json:"scope_id"`
		ExpiresAt *time.Time `json:"expires_at"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}

	if !service.IsValidIP(req.IP) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid IP address or CIDR format",
		})
		return
	}

	userId := c.GetInt("id")
	ipList := &model.IPList{
		IP:        req.IP,
		ListType:  "blacklist",
		Reason:    req.Reason,
		Scope:     req.Scope,
		ScopeID:   req.ScopeID,
		ExpiresAt: req.ExpiresAt,
		CreatedBy: userId,
	}

	if err := model.AddToIPList(ipList); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to add IP to blacklist: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "IP added to blacklist successfully",
		"data":    ipList,
	})
}

// AddIPToWhitelist adds an IP to whitelist
func AddIPToWhitelist(c *gin.Context) {
	var req struct {
		IP        string     `json:"ip" binding:"required"`
		Reason    string     `json:"reason"`
		Scope     string     `json:"scope"`
		ScopeID   int        `json:"scope_id"`
		ExpiresAt *time.Time `json:"expires_at"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}

	if !service.IsValidIP(req.IP) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid IP address or CIDR format",
		})
		return
	}

	userId := c.GetInt("id")
	ipList := &model.IPList{
		IP:        req.IP,
		ListType:  "whitelist",
		Reason:    req.Reason,
		Scope:     req.Scope,
		ScopeID:   req.ScopeID,
		ExpiresAt: req.ExpiresAt,
		CreatedBy: userId,
	}

	if err := model.AddToIPList(ipList); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to add IP to whitelist: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "IP added to whitelist successfully",
		"data":    ipList,
	})
}

// RemoveIPFromList removes an IP from blacklist or whitelist
func RemoveIPFromList(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid list entry ID",
		})
		return
	}

	if err := model.RemoveFromIPList(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to remove IP from list: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "IP removed from list successfully",
	})
}

// GetIPLists gets IP blacklist or whitelist
func GetIPLists(c *gin.Context) {
	listType := c.Query("list_type")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	lists, total, err := model.GetIPLists(listType, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get IP lists: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"items":      lists,
			"total":      total,
			"page":       page,
			"page_size":  pageSize,
			"total_page": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// CreateIPRateLimit creates an IP rate limit rule
func CreateIPRateLimit(c *gin.Context) {
	var limit model.IPRateLimit
	if err := c.ShouldBindJSON(&limit); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}

	if !service.IsValidIP(limit.IP) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid IP address or CIDR format",
		})
		return
	}

	if limit.MaxRequests <= 0 || limit.TimeWindow <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Max requests and time window must be greater than 0",
		})
		return
	}

	if err := model.CreateIPRateLimit(&limit); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to create rate limit: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Rate limit created successfully",
		"data":    limit,
	})
}

// UpdateIPRateLimit updates an IP rate limit rule
func UpdateIPRateLimit(c *gin.Context) {
	var limit model.IPRateLimit
	if err := c.ShouldBindJSON(&limit); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}

	if limit.ID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Rate limit ID is required",
		})
		return
	}

	if err := model.UpdateIPRateLimit(&limit); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to update rate limit: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Rate limit updated successfully",
		"data":    limit,
	})
}

// DeleteIPRateLimit deletes an IP rate limit rule
func DeleteIPRateLimit(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid rate limit ID",
		})
		return
	}

	if err := model.DeleteIPRateLimit(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to delete rate limit: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Rate limit deleted successfully",
	})
}

// GetIPRateLimits gets all IP rate limit rules
func GetIPRateLimits(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	limits, total, err := model.GetIPRateLimits(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get rate limits: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"items":      limits,
			"total":      total,
			"page":       page,
			"page_size":  pageSize,
			"total_page": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// BanIPAddress bans an IP address
func BanIPAddress(c *gin.Context) {
	var req struct {
		IP        string     `json:"ip" binding:"required"`
		Reason    string     `json:"reason"`
		BanType   string     `json:"ban_type"` // "temporary" or "permanent"
		Duration  int        `json:"duration"` // in hours, for temporary bans
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}

	if !service.IsValidIP(req.IP) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid IP address",
		})
		return
	}

	userId := c.GetInt("id")
	ban := &model.IPBan{
		IP:        req.IP,
		BanReason: req.Reason,
		BanType:   req.BanType,
		BannedAt:  time.Now(),
		BannedBy:  userId,
	}

	if req.BanType == "temporary" {
		if req.Duration <= 0 {
			req.Duration = 24 // Default 24 hours
		}
		expiresAt := time.Now().Add(time.Duration(req.Duration) * time.Hour)
		ban.ExpiresAt = &expiresAt
	} else {
		ban.BanType = "permanent"
	}

	if err := model.BanIP(ban); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to ban IP: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "IP banned successfully",
		"data":    ban,
	})
}

// UnbanIPAddress unbans an IP address
func UnbanIPAddress(c *gin.Context) {
	var req struct {
		IP string `json:"ip" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}

	if err := model.UnbanIPByAddress(req.IP); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to unban IP: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "IP unbanned successfully",
	})
}

// GetIPBans gets all banned IPs
func GetIPBans(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	includeExpired := c.Query("include_expired") == "true"

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	bans, total, err := model.GetIPBans(page, pageSize, includeExpired)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get banned IPs: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"items":      bans,
			"total":      total,
			"page":       page,
			"page_size":  pageSize,
			"total_page": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// GetIPProtectionStats gets IP protection statistics
func GetIPProtectionStats(c *gin.Context) {
	stats, err := service.GetIPStatistics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get statistics: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    stats,
	})
}
