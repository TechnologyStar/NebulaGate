package migrations

import (
	"time"

	"gorm.io/gorm"
)

func init() {
	RegisterSchemaProvider("20250207_e2e_encryption_heimdall", E2EEncryptionHeimdallSchema)
}

func E2EEncryptionHeimdallSchema() []interface{} {
	return []interface{}{
		&UserEncryption{},
		&TokenEncryption{},
		&ConversationLog{},
		&HeimdallLog{},
		&AnomalyDetection{},
	}
}

// UserEncryption adds encryption fields to User table
type UserEncryption struct {
	EncryptionKeyHash string `gorm:"type:varchar(255);column:encryption_key_hash"`
	EncryptionEnabled bool   `gorm:"type:boolean;default:false;column:encryption_enabled"`
}

func (UserEncryption) TableName() string {
	return "users"
}

// TokenEncryption adds conversation logging field to Token table
type TokenEncryption struct {
	ConversationLoggingEnabled bool `gorm:"type:boolean;default:false;column:conversation_logging_enabled"`
}

func (TokenEncryption) TableName() string {
	return "tokens"
}

// ConversationLog stores encrypted conversation history
type ConversationLog struct {
	Id               int            `json:"id" gorm:"primaryKey;autoIncrement"`
	UserId           int            `json:"user_id" gorm:"index;not null"`
	TokenId          int            `json:"token_id" gorm:"index;not null"`
	Model            string         `json:"model" gorm:"type:varchar(255);index"`
	EncryptedData    string         `json:"encrypted_data" gorm:"type:text;not null"`
	Nonce            string         `json:"nonce" gorm:"type:varchar(64);not null"`
	Timestamp        int64          `json:"timestamp" gorm:"index;not null"`
	RequestId        string         `json:"request_id" gorm:"type:varchar(64);index"`
	MessageCount     int            `json:"message_count" gorm:"default:0"`
	PromptTokens     int            `json:"prompt_tokens" gorm:"default:0"`
	CompletionTokens int            `json:"completion_tokens" gorm:"default:0"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

func (ConversationLog) TableName() string {
	return "conversation_logs"
}

// HeimdallLog stores request logs from Heimdall security gateway
type HeimdallLog struct {
	Id                 int            `json:"id" gorm:"primaryKey;autoIncrement"`
	UserId             int            `json:"user_id" gorm:"index"`
	TokenKey           string         `json:"token_key" gorm:"type:varchar(255);index"`
	RequestPath        string         `json:"request_path" gorm:"type:varchar(512)"`
	RequestMethod      string         `json:"request_method" gorm:"type:varchar(16)"`
	RealIP             string         `json:"real_ip" gorm:"type:varchar(64);index"`
	ForwardedFor       string         `json:"forwarded_for" gorm:"type:varchar(255)"`
	UserAgent          string         `json:"user_agent" gorm:"type:text"`
	RequestHeaders     string         `json:"request_headers" gorm:"type:text"`
	RequestBody        string         `json:"request_body" gorm:"type:text"`
	ContentFingerprint string         `json:"content_fingerprint" gorm:"type:varchar(255)"`
	DeviceFingerprint  string         `json:"device_fingerprint" gorm:"type:varchar(255);index"`
	Cookies            string         `json:"cookies" gorm:"type:text"`
	ResponseStatus     int            `json:"response_status" gorm:"index"`
	ResponseTime       int            `json:"response_time"`
	Timestamp          int64          `json:"timestamp" gorm:"index;not null"`
	CreatedAt          time.Time      `json:"created_at"`
	DeletedAt          gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

func (HeimdallLog) TableName() string {
	return "heimdall_logs"
}

// AnomalyDetection stores anomaly detection results
type AnomalyDetection struct {
	Id                int            `json:"id" gorm:"primaryKey;autoIncrement"`
	UserId            int            `json:"user_id" gorm:"index;not null"`
	DeviceFingerprint string         `json:"device_fingerprint" gorm:"type:varchar(255);index"`
	IPAddress         string         `json:"ip_address" gorm:"type:varchar(64);index"`
	AnomalyType       string         `json:"anomaly_type" gorm:"type:varchar(64);index"`
	RiskScore         float64        `json:"risk_score" gorm:"index"`
	LoginCount        int            `json:"login_count"`
	TotalAccessCount  int            `json:"total_access_count"`
	AccessRatio       float64        `json:"access_ratio"`
	APIRequestCount   int            `json:"api_request_count"`
	TimeWindowStart   int64          `json:"time_window_start"`
	TimeWindowEnd     int64          `json:"time_window_end"`
	AverageInterval   float64        `json:"average_interval"`
	Status            string         `json:"status" gorm:"type:varchar(32);index;default:'detected'"`
	Action            string         `json:"action" gorm:"type:varchar(64)"`
	Description       string         `json:"description" gorm:"type:text"`
	Metadata          string         `json:"metadata" gorm:"type:text"`
	DetectedAt        int64          `json:"detected_at" gorm:"index;not null"`
	ReviewedAt        int64          `json:"reviewed_at"`
	ReviewedBy        int            `json:"reviewed_by"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

func (AnomalyDetection) TableName() string {
	return "anomaly_detections"
}
