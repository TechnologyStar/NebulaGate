package securityanalytics

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
)

// DeviceAggregationResult contains aggregated device activity
type DeviceAggregationResult struct {
	NormalizedDeviceId string
	UserId             int
	RequestCount       int64
	UniqueIPs          []string
	UniqueModels       []string
	LastSeenAt         time.Time
	FirstSeenAt        time.Time
}

// AggregateDeviceActivity groups logs by normalized device_id and user_id
func AggregateDeviceActivity(userId int, startTime, endTime time.Time) ([]*DeviceAggregationResult, error) {
	var logs []*model.Log
	err := model.LOG_DB.
		Where("user_id = ? AND created_at >= ? AND created_at <= ?", userId, startTime.Unix(), endTime.Unix()).
		Order("created_at DESC").
		Find(&logs).Error
	if err != nil {
		return nil, err
	}

	// Group by normalized device_id
	deviceMap := make(map[string]*DeviceAggregationResult)
	ipSet := make(map[string]map[string]bool)     // device -> set of IPs
	modelSet := make(map[string]map[string]bool)  // device -> set of models

	for _, log := range logs {
		deviceId := normalizeDeviceId(log.Other)
		if deviceId == "" {
			deviceId = "unknown"
		}

		if _, exists := deviceMap[deviceId]; !exists {
			deviceMap[deviceId] = &DeviceAggregationResult{
				NormalizedDeviceId: deviceId,
				UserId:             userId,
				UniqueIPs:          []string{},
				UniqueModels:       []string{},
				LastSeenAt:         time.Unix(log.CreatedAt, 0),
				FirstSeenAt:        time.Unix(log.CreatedAt, 0),
			}
			ipSet[deviceId] = make(map[string]bool)
			modelSet[deviceId] = make(map[string]bool)
		}

		result := deviceMap[deviceId]
		result.RequestCount++

		if log.Ip != "" {
			ipSet[deviceId][log.Ip] = true
		}
		if log.ModelName != "" {
			modelSet[deviceId][log.ModelName] = true
		}

		logTime := time.Unix(log.CreatedAt, 0)
		if logTime.After(result.LastSeenAt) {
			result.LastSeenAt = logTime
		}
		if logTime.Before(result.FirstSeenAt) {
			result.FirstSeenAt = logTime
		}
	}

	// Convert sets to slices
	results := make([]*DeviceAggregationResult, 0, len(deviceMap))
	for deviceId, result := range deviceMap {
		result.UniqueIPs = mapKeysToSlice(ipSet[deviceId])
		result.UniqueModels = mapKeysToSlice(modelSet[deviceId])
		results = append(results, result)
	}

	return results, nil
}

// IPAggregationWindow represents aggregated IP activity in a time window
type IPAggregationWindow struct {
	IP                string
	UserId            int
	TokenId           *int
	WindowStart       time.Time
	WindowEnd         time.Time
	RequestCount      int64
	UniqueDevices     int64
	UniqueModels      []string
	ASN               string
	Subnet            string
	LastActivityTime  time.Time
}

// AggregateIPActivity aggregates IP-based activity with sliding time windows
func AggregateIPActivity(userId int, startTime, endTime time.Time) ([]*IPAggregationWindow, error) {
	var logs []*model.Log
	err := model.LOG_DB.
		Where("user_id = ? AND created_at >= ? AND created_at <= ? AND ip != ?", userId, startTime.Unix(), endTime.Unix(), "").
		Order("created_at DESC").
		Find(&logs).Error
	if err != nil {
		return nil, err
	}

	// Group by IP address
	ipMap := make(map[string]*IPAggregationWindow)
	deviceSet := make(map[string]map[string]bool)  // ip -> set of devices
	modelSet := make(map[string]map[string]bool)   // ip -> set of models

	for _, log := range logs {
		ip := log.Ip
		if ip == "" {
			continue
		}

		if _, exists := ipMap[ip]; !exists {
			asn, subnet := extractASNAndSubnet(ip)
			ipMap[ip] = &IPAggregationWindow{
				IP:               ip,
				UserId:           userId,
				TokenId:          func() *int { if log.TokenId > 0 { return &log.TokenId } else { return nil } }(),
				WindowStart:      startTime,
				WindowEnd:        endTime,
				ASN:              asn,
				Subnet:           subnet,
				UniqueModels:     []string{},
				LastActivityTime: time.Unix(log.CreatedAt, 0),
			}
			deviceSet[ip] = make(map[string]bool)
			modelSet[ip] = make(map[string]bool)
		}

		result := ipMap[ip]
		result.RequestCount++

		deviceId := normalizeDeviceId(log.Other)
		if deviceId != "" {
			deviceSet[ip][deviceId] = true
		}
		if log.ModelName != "" {
			modelSet[ip][log.ModelName] = true
		}

		logTime := time.Unix(log.CreatedAt, 0)
		if logTime.After(result.LastActivityTime) {
			result.LastActivityTime = logTime
		}
	}

	// Convert sets to slices and count unique devices
	results := make([]*IPAggregationWindow, 0, len(ipMap))
	for ip, result := range ipMap {
		result.UniqueDevices = int64(len(deviceSet[ip]))
		result.UniqueModels = mapKeysToSlice(modelSet[ip])
		results = append(results, result)
	}

	return results, nil
}

// normalizeDeviceId extracts and normalizes device ID from log Other field
func normalizeDeviceId(otherJSON string) string {
	if otherJSON == "" {
		return ""
	}

	otherMap, err := common.StrToMap(otherJSON)
	if err != nil {
		return ""
	}

	// Try common device field names
	deviceFields := []string{"device_id", "deviceId", "device", "user_agent_hash", "fingerprint"}
	for _, field := range deviceFields {
		if val, exists := otherMap[field]; exists {
			if str, ok := val.(string); ok && str != "" {
				return strings.ToLower(str)
			}
		}
	}

	return ""
}

// extractASNAndSubnet extracts ASN and subnet from IP address
// This is a simplified implementation; in production, use GeoIP/ASN lookup services
func extractASNAndSubnet(ip string) (string, string) {
	// Extract subnet using CIDR notation
	parts := strings.Split(ip, ".")
	if len(parts) == 4 {
		subnet := strings.Join(parts[:3], ".") + ".0/24"
		return "", subnet
	}

	// For IPv6
	if strings.Contains(ip, ":") {
		ipAddr := net.ParseIP(ip)
		if ipAddr != nil && ipAddr.To4() == nil {
			// IPv6
			parts := strings.Split(ip, ":")
			if len(parts) >= 4 {
				subnet := strings.Join(parts[:4], ":") + "::/64"
				return "", subnet
			}
		}
	}

	return "", ""
}

// ConversationLinkage joins conversation sessions with request logs
type ConversationLinkageResult struct {
	ConversationId    string
	RequestIds        []string
	UserId            int
	TokenId           *int
	StartTime         time.Time
	EndTime           time.Time
	RequestCount      int64
	QuotaUsed         int64
	Models            []string
}

// LinkConversationsWithRequests links conversation sessions to request logs
// This is a placeholder; actual implementation depends on conversation log schema
func LinkConversationsWithRequests(userId int, startTime, endTime time.Time) ([]*ConversationLinkageResult, error) {
	// Get request logs for the user
	var logs []*model.Log
	err := model.LOG_DB.
		Where("user_id = ? AND created_at >= ? AND created_at <= ? AND type = ?", userId, startTime.Unix(), endTime.Unix(), model.LogTypeConsume).
		Order("created_at ASC").
		Find(&logs).Error
	if err != nil {
		return nil, err
	}

	// Group consecutive requests by time proximity (within 30 minutes)
	const sessionTimeout = 30 * time.Minute
	results := make([]*ConversationLinkageResult, 0)

	if len(logs) == 0 {
		return results, nil
	}

	currentSession := &ConversationLinkageResult{
		ConversationId: fmt.Sprintf("conv_%d_%d", userId, logs[0].Id),
		RequestIds:     []string{},
		UserId:         userId,
		TokenId:        func() *int { if logs[0].TokenId > 0 { return &logs[0].TokenId } else { return nil } }(),
		StartTime:      time.Unix(logs[0].CreatedAt, 0),
		Models:         []string{},
	}

	for _, log := range logs {
		logTime := time.Unix(log.CreatedAt, 0)

		// Check if we should start a new session
		if logTime.Sub(currentSession.EndTime) > sessionTimeout {
			if currentSession.RequestCount > 0 {
				results = append(results, currentSession)
			}
			currentSession = &ConversationLinkageResult{
				ConversationId: fmt.Sprintf("conv_%d_%d", userId, log.Id),
				RequestIds:     []string{},
				UserId:         userId,
				TokenId:        func() *int { if log.TokenId > 0 { return &log.TokenId } else { return nil } }(),
				StartTime:      logTime,
				Models:         []string{},
			}
		}

		currentSession.RequestIds = append(currentSession.RequestIds, log.Other)
		currentSession.RequestCount++
		currentSession.QuotaUsed += int64(log.Quota)
		currentSession.EndTime = logTime

		// Add unique models
		if log.ModelName != "" && !stringInSlice(log.ModelName, currentSession.Models) {
			currentSession.Models = append(currentSession.Models, log.ModelName)
		}
	}

	// Add the last session
	if currentSession.RequestCount > 0 {
		results = append(results, currentSession)
	}

	return results, nil
}

// Helper functions

func mapKeysToSlice(m map[string]bool) []string {
	result := make([]string, 0, len(m))
	for k := range m {
		result = append(result, k)
	}
	return result
}

func stringInSlice(s string, list []string) bool {
	for _, item := range list {
		if item == s {
			return true
		}
	}
	return false
}
