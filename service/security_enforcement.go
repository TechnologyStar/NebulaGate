package service

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
)

const (
	AnomalyTypeHighRPM         = "high_rpm"
	AnomalyTypeSuspiciousIP    = "suspicious_ip"
	AnomalyTypeDeviceAnomaly   = "device_anomaly"
	AnomalyTypeContentViolation = "content_violation"

	ActionBlock    = "block"
	ActionRedirect = "redirect"
	ActionBan      = "ban"
	ActionLog      = "log"

	StatusPending  = "pending"
	StatusActioned = "actioned"
	StatusApproved = "approved"
	StatusIgnored  = "ignored"
)

type EnforcementAction struct {
	UserId      int
	Action      string
	Reason      string
	AnomalyId   int
	Severity    string
	Metadata    map[string]interface{}
}

func CreateAnomaly(userId int, tokenId *int, anomalyType, severity, description string, metadata map[string]interface{}, ipAddress, deviceId string, riskScore int) (*model.SecurityAnomaly, error) {
	metadataJSON := ""
	if metadata != nil {
		data, err := json.Marshal(metadata)
		if err == nil {
			metadataJSON = string(data)
		}
	}

	anomaly := &model.SecurityAnomaly{
		UserId:      userId,
		TokenId:     tokenId,
		AnomalyType: anomalyType,
		Severity:    severity,
		Description: description,
		Metadata:    metadataJSON,
		IpAddress:   ipAddress,
		DeviceId:    deviceId,
		RiskScore:   riskScore,
		Status:      StatusPending,
	}

	if err := model.CreateSecurityAnomaly(anomaly); err != nil {
		return nil, err
	}

	if severity == "malicious" && isAutoEnforcementEnabled() {
		go func() {
			if err := ProcessAnomaly(anomaly); err != nil {
				common.SysLog(fmt.Sprintf("failed to process anomaly %d: %v", anomaly.Id, err))
			}
		}()
	}

	return anomaly, nil
}

func ProcessAnomaly(anomaly *model.SecurityAnomaly) error {
	if anomaly.Status != StatusPending {
		return nil
	}

	action := determineAction(anomaly)
	if action == ActionLog {
		return nil
	}

	if err := executeAction(anomaly.UserId, action, anomaly); err != nil {
		return err
	}

	now := time.Now()
	anomaly.ActionTaken = action
	anomaly.ActionedAt = &now
	anomaly.Status = StatusActioned

	return model.UpdateSecurityAnomaly(anomaly)
}

func determineAction(anomaly *model.SecurityAnomaly) string {
	settings := GetSecuritySettings()
	
	if anomaly.Severity == "malicious" {
		if autoBanEnabled, ok := settings["auto_ban_enabled"].(bool); ok && autoBanEnabled {
			return ActionBan
		}
		if blockEnabled, ok := settings["auto_block_enabled"].(bool); ok && blockEnabled {
			return ActionBlock
		}
	}

	if anomaly.RiskScore > 80 {
		return ActionBan
	} else if anomaly.RiskScore > 50 {
		return ActionBlock
	} else if anomaly.RiskScore > 30 {
		return ActionRedirect
	}

	return ActionLog
}

func executeAction(userId int, action string, anomaly *model.SecurityAnomaly) error {
	switch action {
	case ActionBan:
		if err := BanUser(userId); err != nil {
			return fmt.Errorf("failed to ban user: %w", err)
		}
		publishHeimdallDirective(userId, "ban", anomaly)
		sendNotification(userId, "banned", anomaly)

	case ActionBlock:
		publishHeimdallDirective(userId, "block", anomaly)
		sendNotification(userId, "blocked", anomaly)

	case ActionRedirect:
		redirectModel := GetViolationRedirectModel()
		if redirectModel != "" {
			if err := SetUserRedirect(userId, redirectModel); err != nil {
				return fmt.Errorf("failed to set redirect: %w", err)
			}
		}
		publishHeimdallDirective(userId, "redirect", anomaly)
		sendNotification(userId, "redirected", anomaly)
	}

	return nil
}

func publishHeimdallDirective(userId int, action string, anomaly *model.SecurityAnomaly) {
	if !common.RedisEnabled {
		return
	}

	directive := map[string]interface{}{
		"user_id":     userId,
		"action":      action,
		"anomaly_id":  anomaly.Id,
		"severity":    anomaly.Severity,
		"timestamp":   time.Now().Unix(),
		"description": anomaly.Description,
	}

	data, err := json.Marshal(directive)
	if err != nil {
		common.SysLog(fmt.Sprintf("failed to marshal heimdall directive: %v", err))
		return
	}

	key := fmt.Sprintf("heimdall:directive:%d", userId)
	common.RedisSet(key, string(data), 3600)

	channel := "heimdall:directives"
	if err := common.RedisPublish(channel, string(data)); err != nil {
		common.SysLog(fmt.Sprintf("failed to publish heimdall directive: %v", err))
	}
}

func sendNotification(userId int, actionType string, anomaly *model.SecurityAnomaly) {
	user, err := model.GetUserById(userId, false)
	if err != nil {
		return
	}

	message := fmt.Sprintf("Security action taken: %s due to %s (severity: %s)", 
		actionType, anomaly.AnomalyType, anomaly.Severity)

	notificationSettings := GetUserNotificationSettings(userId)
	if notificationSettings != nil && notificationSettings.SecurityAlerts {
		common.SysLog(fmt.Sprintf("notification sent to user %s: %s", user.Username, message))
	}
}

func ApproveAnomaly(anomalyId int, reviewerId int, rationale string) error {
	anomaly, err := model.GetSecurityAnomaly(anomalyId)
	if err != nil {
		return err
	}

	now := time.Now()
	anomaly.Status = StatusApproved
	anomaly.ReviewedBy = &reviewerId
	anomaly.ReviewedAt = &now
	anomaly.ReviewDecision = "approved"
	anomaly.ReviewRationale = rationale

	return model.UpdateSecurityAnomaly(anomaly)
}

func IgnoreAnomaly(anomalyId int, reviewerId int, rationale string) error {
	anomaly, err := model.GetSecurityAnomaly(anomalyId)
	if err != nil {
		return err
	}

	now := time.Now()
	anomaly.Status = StatusIgnored
	anomaly.ReviewedBy = &reviewerId
	anomaly.ReviewedAt = &now
	anomaly.ReviewDecision = "ignored"
	anomaly.ReviewRationale = rationale

	if anomaly.ActionTaken != "" && anomaly.ActionTaken != ActionLog {
		if err := rollbackAction(anomaly); err != nil {
			return fmt.Errorf("failed to rollback action: %w", err)
		}
	}

	return model.UpdateSecurityAnomaly(anomaly)
}

func rollbackAction(anomaly *model.SecurityAnomaly) error {
	switch anomaly.ActionTaken {
	case ActionBan:
		return UnbanUser(anomaly.UserId)
	case ActionRedirect:
		return ClearUserRedirect(anomaly.UserId)
	}
	return nil
}

func isAutoEnforcementEnabled() bool {
	settings := GetSecuritySettings()
	if enabled, ok := settings["auto_enforcement_enabled"].(bool); ok {
		return enabled
	}
	return true
}

type NotificationSettings struct {
	SecurityAlerts bool
}

func GetUserNotificationSettings(userId int) *NotificationSettings {
	return &NotificationSettings{
		SecurityAlerts: true,
	}
}

func TrackDeviceFingerprint(userId int, fingerprint, userAgent, ipAddress string) error {
	device, err := model.GetDeviceFingerprint(fingerprint, userId)
	if err != nil {
		now := time.Now()
		device = &model.DeviceFingerprint{
			UserId:       userId,
			Fingerprint:  fingerprint,
			UserAgent:    userAgent,
			IpAddress:    ipAddress,
			FirstSeenAt:  now,
			LastSeenAt:   now,
			RequestCount: 1,
			RiskScore:    0,
		}
		return model.CreateDeviceFingerprint(device)
	}

	device.LastSeenAt = time.Now()
	device.RequestCount++
	device.IpAddress = ipAddress

	return model.UpdateDeviceFingerprint(device)
}

func TrackIPCluster(ipAddress string, userId int) error {
	cluster, err := model.GetIPCluster(ipAddress)
	if err != nil {
		now := time.Now()
		cluster = &model.IPCluster{
			IpAddress:     ipAddress,
			UniqueUsers:   1,
			TotalRequests: 1,
			RiskScore:     0,
			FirstSeenAt:   now,
			LastSeenAt:    now,
		}
		return model.CreateIPCluster(cluster)
	}

	cluster.LastSeenAt = time.Now()
	cluster.TotalRequests++

	return model.UpdateIPCluster(cluster)
}
