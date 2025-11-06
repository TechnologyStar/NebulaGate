package service

import (
	"fmt"
	"strconv"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service/securityanalytics"
)

var globalAnomalyEngine *securityanalytics.Engine
var anomalyEngineStopCh chan struct{}

// InitAnomalyEngine initializes and starts the anomaly detection engine
func InitAnomalyEngine() *securityanalytics.Engine {
	engine := securityanalytics.NewEngine()
	globalAnomalyEngine = engine

	// Get configuration from options or environment
	intervalStr := model.GetOptionValue("anomaly_detection_interval_seconds")
	if intervalStr == "" {
		intervalStr = "3600" // default 1 hour
	}

	windowStr := model.GetOptionValue("anomaly_detection_window_hours")
	if windowStr == "" {
		windowStr = "24" // default 24 hours
	}

	interval, err := strconv.Atoi(intervalStr)
	if err != nil {
		interval = 3600
	}

	window, err := strconv.Atoi(windowStr)
	if err != nil {
		window = 24
	}

	// Start background processing
	anomalyEngineStopCh = engine.Start(
		time.Duration(interval)*time.Second,
		time.Duration(window)*time.Hour,
		5, // concurrency
	)

	common.SysLog(fmt.Sprintf("anomaly engine initialized with interval=%ds, window=%dh", interval, window))

	return engine
}

// GetAnomalyEngine returns the global anomaly engine
func GetAnomalyEngine() *securityanalytics.Engine {
	return globalAnomalyEngine
}

// StopAnomalyEngine stops the anomaly detection engine
func StopAnomalyEngine() {
	if globalAnomalyEngine != nil && anomalyEngineStopCh != nil {
		globalAnomalyEngine.Stop(anomalyEngineStopCh)
		common.SysLog("anomaly engine stopped")
	}
}

// ProcessUserAnomalies analyzes a specific user for anomalies
func ProcessUserAnomalies(userId int, windowDuration time.Duration) ([]*model.SecurityAnomaly, error) {
	if globalAnomalyEngine == nil {
		return nil, fmt.Errorf("anomaly engine not initialized")
	}

	return globalAnomalyEngine.ProcessUser(userId, windowDuration)
}

// GetAnomalyStatistics retrieves statistics about detected anomalies
func GetAnomalyStatistics(startTime, endTime time.Time) (map[string]interface{}, error) {
	return model.GetAnomalyStatsByDateRange(startTime, endTime)
}

// UpdateAnomalyBaseline updates the baseline metrics for a user
func UpdateAnomalyBaseline(userId int) error {
	return securityanalytics.UpdateUserBaselines(userId)
}

// ResolveAnomaly marks an anomaly as resolved
func ResolveAnomaly(anomalyId int) error {
	return model.ResolveSecurityAnomaly(anomalyId)
}

// CleanupExpiredAnomalies removes expired anomalies based on TTL
func CleanupExpiredAnomalies() (int64, error) {
	return model.CleanupExpiredAnomalies()
}
