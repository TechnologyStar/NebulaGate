package model

import (
	"time"

	"gorm.io/gorm"
)

// IPList represents an IP in blacklist or whitelist
type IPList struct {
	ID        int        `json:"id" gorm:"primaryKey"`
	IP        string     `json:"ip" gorm:"size:64;not null;index"`
	ListType  string     `json:"list_type" gorm:"size:10;not null;index"` // "blacklist" or "whitelist"
	Reason    string     `json:"reason" gorm:"size:255"`
	Scope     string     `json:"scope" gorm:"size:20;default:'global'"` // "global", "user", "key"
	ScopeID   int        `json:"scope_id" gorm:"default:0"`
	ExpiresAt *time.Time `json:"expires_at"`
	CreatedBy int        `json:"created_by" gorm:"index"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func (i *IPList) TableName() string {
	return "ip_lists"
}

// IPRateLimit represents rate limiting rules for IPs
type IPRateLimit struct {
	ID           int       `json:"id" gorm:"primaryKey"`
	Name         string    `json:"name" gorm:"size:100;not null"`
	IP           string    `json:"ip" gorm:"size:64;not null;index"`
	MaxRequests  int       `json:"max_requests" gorm:"not null"`
	TimeWindow   int       `json:"time_window" gorm:"not null"` // in seconds
	Action       string    `json:"action" gorm:"size:20;not null;default:'reject'"` // "reject", "warn", "ban"
	Enabled      bool      `json:"enabled" gorm:"not null;default:true;index"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (i *IPRateLimit) TableName() string {
	return "ip_rate_limits"
}

// IPBan represents a banned IP record
type IPBan struct {
	ID        int        `json:"id" gorm:"primaryKey"`
	IP        string     `json:"ip" gorm:"size:64;not null;index"`
	BanReason string     `json:"ban_reason" gorm:"size:255"`
	BanType   string     `json:"ban_type" gorm:"size:20;not null;default:'temporary'"` // "temporary", "permanent"
	BannedAt  time.Time  `json:"banned_at" gorm:"not null"`
	ExpiresAt *time.Time `json:"expires_at"`
	BannedBy  int        `json:"banned_by" gorm:"index"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func (i *IPBan) TableName() string {
	return "ip_bans"
}

// AddToIPList adds an IP to blacklist or whitelist
func AddToIPList(ip *IPList) error {
	return DB.Create(ip).Error
}

// RemoveFromIPList removes an IP from list
func RemoveFromIPList(id int) error {
	return DB.Where("id = ?", id).Delete(&IPList{}).Error
}

// GetIPLists gets IP lists with filters
func GetIPLists(listType string, page int, pageSize int) ([]IPList, int64, error) {
	var lists []IPList
	var total int64

	query := DB.Model(&IPList{})
	if listType != "" {
		query = query.Where("list_type = ?", listType)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err = query.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&lists).Error
	return lists, total, err
}

// IsIPInList checks if an IP is in a specific list (blacklist/whitelist)
func IsIPInList(ip string, listType string) (bool, *IPList, error) {
	var list IPList
	now := time.Now()
	
	query := DB.Where("ip = ? AND list_type = ?", ip, listType)
	// Check expiration
	query = query.Where("(expires_at IS NULL OR expires_at > ?)", now)
	
	err := query.First(&list).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil, nil
		}
		return false, nil, err
	}
	return true, &list, nil
}

// CreateIPRateLimit creates a rate limit rule
func CreateIPRateLimit(limit *IPRateLimit) error {
	return DB.Create(limit).Error
}

// UpdateIPRateLimit updates a rate limit rule
func UpdateIPRateLimit(limit *IPRateLimit) error {
	return DB.Model(&IPRateLimit{}).Where("id = ?", limit.ID).Updates(limit).Error
}

// DeleteIPRateLimit deletes a rate limit rule
func DeleteIPRateLimit(id int) error {
	return DB.Where("id = ?", id).Delete(&IPRateLimit{}).Error
}

// GetIPRateLimits gets all rate limit rules
func GetIPRateLimits(page int, pageSize int) ([]IPRateLimit, int64, error) {
	var limits []IPRateLimit
	var total int64

	query := DB.Model(&IPRateLimit{})
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err = query.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&limits).Error
	return limits, total, err
}

// GetIPRateLimitForIP gets rate limit rules for a specific IP
func GetIPRateLimitForIP(ip string) (*IPRateLimit, error) {
	var limit IPRateLimit
	err := DB.Where("ip = ? AND enabled = ?", ip, true).First(&limit).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &limit, nil
}

// BanIP bans an IP address
func BanIP(ban *IPBan) error {
	return DB.Create(ban).Error
}

// UnbanIP removes a ban
func UnbanIP(id int) error {
	return DB.Where("id = ?", id).Delete(&IPBan{}).Error
}

// UnbanIPByAddress removes a ban by IP address
func UnbanIPByAddress(ip string) error {
	return DB.Where("ip = ?", ip).Delete(&IPBan{}).Error
}

// GetIPBans gets all banned IPs
func GetIPBans(page int, pageSize int, includeExpired bool) ([]IPBan, int64, error) {
	var bans []IPBan
	var total int64

	query := DB.Model(&IPBan{})
	
	if !includeExpired {
		now := time.Now()
		query = query.Where("ban_type = ? OR (ban_type = ? AND expires_at > ?)", 
			"permanent", "temporary", now)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err = query.Order("banned_at DESC").Limit(pageSize).Offset(offset).Find(&bans).Error
	return bans, total, err
}

// IsIPBanned checks if an IP is currently banned
func IsIPBanned(ip string) (bool, *IPBan, error) {
	var ban IPBan
	now := time.Now()
	
	query := DB.Where("ip = ?", ip)
	query = query.Where("ban_type = ? OR (ban_type = ? AND expires_at > ?)", 
		"permanent", "temporary", now)
	
	err := query.First(&ban).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil, nil
		}
		return false, nil, err
	}
	return true, &ban, nil
}

// CleanupExpiredBans removes expired temporary bans
func CleanupExpiredBans() error {
	now := time.Now()
	return DB.Where("ban_type = ? AND expires_at <= ?", "temporary", now).
		Delete(&IPBan{}).Error
}

// CleanupExpiredIPLists removes expired IP list entries
func CleanupExpiredIPLists() error {
	now := time.Now()
	return DB.Where("expires_at IS NOT NULL AND expires_at <= ?", now).
		Delete(&IPList{}).Error
}
