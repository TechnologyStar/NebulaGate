package model

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// ConversationLog stores encrypted conversation history for users who enable end-to-end encryption
type ConversationLog struct {
	Id              int            `json:"id" gorm:"primaryKey"`
	UserId          int            `json:"user_id" gorm:"index;not null"`
	TokenId         int            `json:"token_id" gorm:"index;not null"`
	Model           string         `json:"model" gorm:"type:varchar(255);index"`
	EncryptedData   string         `json:"encrypted_data" gorm:"type:text;not null"` // AES-256-GCM encrypted conversation data
	Nonce           string         `json:"nonce" gorm:"type:varchar(64);not null"`   // Encryption nonce/IV
	Timestamp       int64          `json:"timestamp" gorm:"index;not null"`
	RequestId       string         `json:"request_id" gorm:"type:varchar(64);index"`
	MessageCount    int            `json:"message_count" gorm:"default:0"`
	PromptTokens    int            `json:"prompt_tokens" gorm:"default:0"`
	CompletionTokens int           `json:"completion_tokens" gorm:"default:0"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

func (ConversationLog) TableName() string {
	return "conversation_logs"
}

// CreateConversationLog creates a new encrypted conversation log entry
func CreateConversationLog(log *ConversationLog) error {
	if log.UserId == 0 || log.TokenId == 0 {
		return errors.New("user_id and token_id are required")
	}
	if log.EncryptedData == "" || log.Nonce == "" {
		return errors.New("encrypted_data and nonce are required")
	}
	log.Timestamp = time.Now().Unix()
	return DB.Create(log).Error
}

// GetConversationLogsByUserId retrieves conversation logs for a specific user with pagination
func GetConversationLogsByUserId(userId int, startIdx int, num int) ([]*ConversationLog, int64, error) {
	var logs []*ConversationLog
	var total int64

	err := DB.Model(&ConversationLog{}).Where("user_id = ?", userId).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = DB.Where("user_id = ?", userId).
		Order("timestamp desc").
		Limit(num).
		Offset(startIdx).
		Find(&logs).Error

	return logs, total, err
}

// GetConversationLogsByTokenId retrieves conversation logs for a specific token
func GetConversationLogsByTokenId(userId int, tokenId int, startIdx int, num int) ([]*ConversationLog, int64, error) {
	var logs []*ConversationLog
	var total int64

	query := DB.Model(&ConversationLog{}).Where("user_id = ? AND token_id = ?", userId, tokenId)
	
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Order("timestamp desc").
		Limit(num).
		Offset(startIdx).
		Find(&logs).Error

	return logs, total, err
}

// GetConversationLogById retrieves a specific conversation log by ID
func GetConversationLogById(id int, userId int) (*ConversationLog, error) {
	var log ConversationLog
	err := DB.Where("id = ? AND user_id = ?", id, userId).First(&log).Error
	return &log, err
}

// DeleteConversationLog soft deletes a conversation log
func DeleteConversationLog(id int, userId int) error {
	return DB.Where("id = ? AND user_id = ?", id, userId).Delete(&ConversationLog{}).Error
}

// DeleteConversationLogsByTokenId deletes all conversation logs for a specific token
func DeleteConversationLogsByTokenId(userId int, tokenId int) error {
	return DB.Where("user_id = ? AND token_id = ?", userId, tokenId).Delete(&ConversationLog{}).Error
}

// CountConversationLogsByUserId returns the total number of conversation logs for a user
func CountConversationLogsByUserId(userId int) (int64, error) {
	var count int64
	err := DB.Model(&ConversationLog{}).Where("user_id = ?", userId).Count(&count).Error
	return count, err
}

// SearchConversationLogs searches conversation logs by model or request_id
func SearchConversationLogs(userId int, keyword string, startIdx int, num int) ([]*ConversationLog, int64, error) {
	var logs []*ConversationLog
	var total int64

	query := DB.Model(&ConversationLog{}).Where("user_id = ?", userId)
	if keyword != "" {
		query = query.Where("model LIKE ? OR request_id LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Order("timestamp desc").
		Limit(num).
		Offset(startIdx).
		Find(&logs).Error

	return logs, total, err
}
