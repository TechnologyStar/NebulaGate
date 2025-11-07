package service

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
)

const (
	AnomalyTypeHighFrequency  = "high_frequency"
	AnomalyTypeDataSpike      = "data_spike"
	AnomalyTypeSuspiciousPattern = "suspicious_pattern"
	AnomalyTypeNoAPIActivity  = "no_api_activity"
)

type AnomalyDetectorService struct {
	// Configurable thresholds
	HighAccessRatioThreshold float64
	MinAccessCountThreshold  int
	HighRiskScoreThreshold   float64
	TimeWindowSeconds        int64
}

func NewAnomalyDetectorService() *AnomalyDetectorService {
	return &AnomalyDetectorService{
		HighAccessRatioThreshold: 50.0,  // Access/Login ratio > 50
		MinAccessCountThreshold:  100,   // Minimum access count to consider
		HighRiskScoreThreshold:   70.0,  // Risk score threshold
		TimeWindowSeconds:        3600,  // 1 hour time window
	}
}

// AnalyzeUserBehavior analyzes a user's behavior patterns and detects anomalies
func (ads *AnomalyDetectorService) AnalyzeUserBehavior(userId int) error {
	endTime := time.Now().Unix()
	startTime := endTime - ads.TimeWindowSeconds

	// Get Heimdall logs for the user in the time window
	logs, _, err := model.GetHeimdallLogsByUserId(userId, 0, 10000)
	if err != nil {
		return err
	}

	if len(logs) == 0 {
		return nil // No logs to analyze
	}

	// Aggregate data by device and IP
	deviceStats := make(map[string]*BehaviorStats)
	ipStats := make(map[string]*BehaviorStats)

	for _, log := range logs {
		// Skip logs outside time window
		if log.Timestamp < startTime || log.Timestamp > endTime {
			continue
		}

		// Device aggregation
		if log.DeviceFingerprint != "" {
			if _, exists := deviceStats[log.DeviceFingerprint]; !exists {
				deviceStats[log.DeviceFingerprint] = &BehaviorStats{
					AccessCount: 0,
					Timestamps:  []int64{},
				}
			}
			deviceStats[log.DeviceFingerprint].AccessCount++
			deviceStats[log.DeviceFingerprint].Timestamps = append(deviceStats[log.DeviceFingerprint].Timestamps, log.Timestamp)
		}

		// IP aggregation
		if log.RealIP != "" {
			if _, exists := ipStats[log.RealIP]; !exists {
				ipStats[log.RealIP] = &BehaviorStats{
					AccessCount: 0,
					Timestamps:  []int64{},
				}
			}
			ipStats[log.RealIP].AccessCount++
			ipStats[log.RealIP].Timestamps = append(ipStats[log.RealIP].Timestamps, log.Timestamp)
		}
	}

	// Calculate request frequency
	reqFreq, err := model.GetRequestFrequencyByUser(userId, startTime, endTime)
	if err != nil {
		return err
	}

	// Detect anomalies for each device
	for deviceFp, stats := range deviceStats {
		anomaly := ads.detectDeviceAnomaly(userId, deviceFp, "", stats, int(reqFreq), startTime, endTime)
		if anomaly != nil {
			model.CreateAnomalyDetection(anomaly)
		}
	}

	// Detect anomalies for each IP
	for ip, stats := range ipStats {
		anomaly := ads.detectIPAnomaly(userId, "", ip, stats, int(reqFreq), startTime, endTime)
		if anomaly != nil {
			model.CreateAnomalyDetection(anomaly)
		}
	}

	return nil
}

type BehaviorStats struct {
	AccessCount int
	Timestamps  []int64
}

func (ads *AnomalyDetectorService) detectDeviceAnomaly(
	userId int,
	deviceFingerprint string,
	ip string,
	stats *BehaviorStats,
	apiRequestCount int,
	startTime int64,
	endTime int64,
) *model.AnomalyDetection {
	// Calculate metrics
	loginCount := 1 // Simplified: assume at least 1 login session
	accessRatio := float64(stats.AccessCount) / float64(loginCount)
	avgInterval := ads.calculateAverageInterval(stats.Timestamps)

	// Risk scoring
	riskScore := ads.calculateRiskScore(
		stats.AccessCount,
		loginCount,
		accessRatio,
		apiRequestCount,
		avgInterval,
	)

	// Only create anomaly if risk score exceeds threshold
	if riskScore < ads.HighRiskScoreThreshold {
		return nil
	}

	// Determine anomaly type
	anomalyType := AnomalyTypeSuspiciousPattern
	description := fmt.Sprintf("Suspicious behavior pattern detected for device %s", deviceFingerprint)

	if stats.AccessCount > ads.MinAccessCountThreshold && apiRequestCount == 0 {
		anomalyType = AnomalyTypeNoAPIActivity
		description = fmt.Sprintf("High access count (%d) with no API requests", stats.AccessCount)
	} else if accessRatio > ads.HighAccessRatioThreshold {
		anomalyType = AnomalyTypeDataSpike
		description = fmt.Sprintf("Abnormal access ratio: %.2f (access: %d, login: %d)", accessRatio, stats.AccessCount, loginCount)
	} else if avgInterval < 1.0 {
		anomalyType = AnomalyTypeHighFrequency
		description = fmt.Sprintf("High frequency requests: avg interval %.2f seconds", avgInterval)
	}

	// Determine action based on risk score
	action := "alert"
	if riskScore >= 90.0 {
		action = "block"
	} else if riskScore >= 80.0 {
		action = "rate_limit"
	}

	metadata := map[string]interface{}{
		"timestamps":      stats.Timestamps,
		"access_pattern":  "device_based",
	}
	metadataJSON, _ := json.Marshal(metadata)

	return &model.AnomalyDetection{
		UserId:            userId,
		DeviceFingerprint: deviceFingerprint,
		IPAddress:         ip,
		AnomalyType:       anomalyType,
		RiskScore:         riskScore,
		LoginCount:        loginCount,
		TotalAccessCount:  stats.AccessCount,
		AccessRatio:       accessRatio,
		APIRequestCount:   apiRequestCount,
		TimeWindowStart:   startTime,
		TimeWindowEnd:     endTime,
		AverageInterval:   avgInterval,
		Status:            "detected",
		Action:            action,
		Description:       description,
		Metadata:          string(metadataJSON),
	}
}

func (ads *AnomalyDetectorService) detectIPAnomaly(
	userId int,
	deviceFingerprint string,
	ip string,
	stats *BehaviorStats,
	apiRequestCount int,
	startTime int64,
	endTime int64,
) *model.AnomalyDetection {
	// Similar to device anomaly detection but for IP
	loginCount := 1
	accessRatio := float64(stats.AccessCount) / float64(loginCount)
	avgInterval := ads.calculateAverageInterval(stats.Timestamps)

	riskScore := ads.calculateRiskScore(
		stats.AccessCount,
		loginCount,
		accessRatio,
		apiRequestCount,
		avgInterval,
	)

	// Adjust risk score for NAT considerations
	// IPs with many users are less suspicious
	riskScore *= 0.8 // Reduce risk score by 20% for IP-based detection

	if riskScore < ads.HighRiskScoreThreshold {
		return nil
	}

	anomalyType := AnomalyTypeSuspiciousPattern
	description := fmt.Sprintf("Suspicious behavior pattern detected for IP %s", ip)

	if stats.AccessCount > ads.MinAccessCountThreshold && apiRequestCount == 0 {
		anomalyType = AnomalyTypeNoAPIActivity
		description = fmt.Sprintf("High access count (%d) with no API requests from IP %s", stats.AccessCount, ip)
	} else if accessRatio > ads.HighAccessRatioThreshold {
		anomalyType = AnomalyTypeDataSpike
		description = fmt.Sprintf("Abnormal access ratio from IP %s: %.2f", ip, accessRatio)
	}

	action := "alert"
	if riskScore >= 90.0 {
		action = "block"
	} else if riskScore >= 80.0 {
		action = "rate_limit"
	}

	metadata := map[string]interface{}{
		"timestamps":     stats.Timestamps,
		"access_pattern": "ip_based",
	}
	metadataJSON, _ := json.Marshal(metadata)

	return &model.AnomalyDetection{
		UserId:           userId,
		DeviceFingerprint: deviceFingerprint,
		IPAddress:        ip,
		AnomalyType:      anomalyType,
		RiskScore:        riskScore,
		LoginCount:       loginCount,
		TotalAccessCount: stats.AccessCount,
		AccessRatio:      accessRatio,
		APIRequestCount:  apiRequestCount,
		TimeWindowStart:  startTime,
		TimeWindowEnd:    endTime,
		AverageInterval:  avgInterval,
		Status:           "detected",
		Action:           action,
		Description:      description,
		Metadata:         string(metadataJSON),
	}
}

func (ads *AnomalyDetectorService) calculateAverageInterval(timestamps []int64) float64 {
	if len(timestamps) <= 1 {
		return 0.0
	}

	var totalInterval int64
	for i := 1; i < len(timestamps); i++ {
		totalInterval += timestamps[i] - timestamps[i-1]
	}

	return float64(totalInterval) / float64(len(timestamps)-1)
}

func (ads *AnomalyDetectorService) calculateRiskScore(
	accessCount int,
	loginCount int,
	accessRatio float64,
	apiRequestCount int,
	avgInterval float64,
) float64 {
	riskScore := 0.0

	// Factor 1: Access ratio (40% weight)
	if accessRatio > ads.HighAccessRatioThreshold {
		riskScore += 40.0 * (accessRatio / ads.HighAccessRatioThreshold)
		if riskScore > 40.0 {
			riskScore = 40.0
		}
	}

	// Factor 2: No API activity despite high access (30% weight)
	if accessCount > ads.MinAccessCountThreshold && apiRequestCount == 0 {
		riskScore += 30.0
	}

	// Factor 3: High frequency (20% weight)
	if avgInterval < 1.0 && avgInterval > 0 {
		riskScore += 20.0 * (1.0 - avgInterval)
	}

	// Factor 4: Absolute access count (10% weight)
	if accessCount > ads.MinAccessCountThreshold {
		ratio := float64(accessCount) / float64(ads.MinAccessCountThreshold)
		riskScore += 10.0 * (ratio - 1.0)
		if riskScore > 100.0 {
			riskScore = 100.0
		}
	}

	// Cap at 100
	if riskScore > 100.0 {
		riskScore = 100.0
	}

	return riskScore
}

// RunPeriodicAnalysis runs anomaly detection for all active users
func (ads *AnomalyDetectorService) RunPeriodicAnalysis() {
	common.SysLog("Starting periodic anomaly detection analysis")

	// Get all active users (simplified: get recent users from logs)
	endTime := time.Now().Unix()
	startTime := endTime - ads.TimeWindowSeconds

	// This is a simplified approach - in production, you'd want to optimize this
	// by maintaining a list of active users or using a more efficient query
	pageSize := 100
	offset := 0

	for {
		users, total, err := model.GetAllUsers(&common.PageInfo{
			Page:     offset/pageSize + 1,
			PageSize: pageSize,
		})
		if err != nil {
			common.SysLog(fmt.Sprintf("Error fetching users for anomaly detection: %v", err))
			break
		}

		for _, user := range users {
			if user.Status == common.UserStatusEnabled {
				err := ads.AnalyzeUserBehavior(user.Id)
				if err != nil {
					common.SysLog(fmt.Sprintf("Error analyzing user %d: %v", user.Id, err))
				}
			}
		}

		offset += pageSize
		if int64(offset) >= total {
			break
		}
	}

	common.SysLog("Completed periodic anomaly detection analysis")
}
