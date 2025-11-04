package controller

import (
    "fmt"
    "net/http"
    "time"

    "github.com/QuantumNous/new-api/model"
    "github.com/gin-gonic/gin"
)

type PlaygroundIPStat struct {
    IP            string   `json:"ip"`
    UserCount     int      `json:"user_count"`
    Usernames     []string `json:"usernames"`
    LastActiveAt  string   `json:"last_active_at"`
    TotalRequests int64    `json:"total_requests"`
}

// GetPlaygroundIPStats gets IP statistics for playground (admin only)
func GetPlaygroundIPStats(c *gin.Context) {
    // Parse time range parameter
    hoursStr := c.DefaultQuery("hours", "24")
    var hours int
    if _, err := fmt.Sscanf(hoursStr, "%d", &hours); err != nil {
        hours = 24
    }
    if hours < 1 {
        hours = 24
    }
    if hours > 720 { // Max 30 days
        hours = 720
    }

    since := time.Now().Add(-time.Duration(hours) * time.Hour)

    // Query all user IP usage since the given time
    var userIPUsages []model.UserIPUsage
    err := model.LOG_DB.Where("last_seen_at >= ?", since).
        Order("last_seen_at desc").
        Find(&userIPUsages).Error

    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "success": false,
            "message": "Failed to get IP statistics: " + err.Error(),
        })
        return
    }

    // Group by IP
    ipMap := make(map[string]*PlaygroundIPStat)
    userIDMap := make(map[string]map[int]bool) // ip -> user_id -> true

    for _, usage := range userIPUsages {
        if _, exists := ipMap[usage.IP]; !exists {
            ipMap[usage.IP] = &PlaygroundIPStat{
                IP:            usage.IP,
                Usernames:     []string{},
                LastActiveAt:  usage.LastSeenAt.Format(time.RFC3339),
                TotalRequests: 0,
            }
            userIDMap[usage.IP] = make(map[int]bool)
        }

        // Track unique users per IP
        if !userIDMap[usage.IP][usage.UserId] {
            userIDMap[usage.IP][usage.UserId] = true

            // Get username
            var user model.User
            if err := model.DB.Where("id = ?", usage.UserId).First(&user).Error; err == nil {
                ipMap[usage.IP].Usernames = append(ipMap[usage.IP].Usernames, user.Username)
            }
        }

        ipMap[usage.IP].TotalRequests += usage.RequestCount

        // Update last active time if newer
        if usage.LastSeenAt.Format(time.RFC3339) > ipMap[usage.IP].LastActiveAt {
            ipMap[usage.IP].LastActiveAt = usage.LastSeenAt.Format(time.RFC3339)
        }
    }

    // Convert map to slice and calculate user counts
    var stats []PlaygroundIPStat
    for _, stat := range ipMap {
        stat.UserCount = len(stat.Usernames)
        stats = append(stats, *stat)
    }

    // Sort by user count (descending)
    for i := 0; i < len(stats); i++ {
        for j := i + 1; j < len(stats); j++ {
            if stats[j].UserCount > stats[i].UserCount {
                stats[i], stats[j] = stats[j], stats[i]
            }
        }
    }

    // Calculate summary
    totalIPs := len(stats)
    var suspiciousIPs int
    for _, stat := range stats {
        if stat.UserCount > 5 {
            suspiciousIPs++
        }
    }

    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "message": "",
        "data": gin.H{
            "total_ips":      totalIPs,
            "suspicious_ips": suspiciousIPs,
            "time_range":     hours,
            "ip_stats":       stats,
        },
    })
}
