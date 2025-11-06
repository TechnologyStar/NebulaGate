package service

import (
	"testing"
	"time"

	"github.com/QuantumNous/new-api/model"
	"github.com/stretchr/testify/assert"
)

func TestCreateAnomaly(t *testing.T) {
	anomaly, err := CreateAnomaly(
		1,
		nil,
		AnomalyTypeHighRPM,
		"malicious",
		"High request rate detected",
		map[string]interface{}{
			"rpm": 120,
		},
		"192.168.1.1",
		"device-fp-123",
		75,
	)

	assert.NoError(t, err)
	assert.NotNil(t, anomaly)
	assert.Equal(t, 1, anomaly.UserId)
	assert.Equal(t, "malicious", anomaly.Severity)
	assert.Equal(t, StatusPending, anomaly.Status)
}

func TestDetermineAction(t *testing.T) {
	tests := []struct {
		name      string
		anomaly   *model.SecurityAnomaly
		expected  string
	}{
		{
			name: "High risk score triggers ban",
			anomaly: &model.SecurityAnomaly{
				Severity:  "malicious",
				RiskScore: 85,
			},
			expected: ActionBan,
		},
		{
			name: "Medium risk score triggers block",
			anomaly: &model.SecurityAnomaly{
				Severity:  "violation",
				RiskScore: 60,
			},
			expected: ActionBlock,
		},
		{
			name: "Low risk score triggers redirect",
			anomaly: &model.SecurityAnomaly{
				Severity:  "violation",
				RiskScore: 40,
			},
			expected: ActionRedirect,
		},
		{
			name: "Very low risk score only logs",
			anomaly: &model.SecurityAnomaly{
				Severity:  "violation",
				RiskScore: 20,
			},
			expected: ActionLog,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action := determineAction(tt.anomaly)
			assert.Equal(t, tt.expected, action)
		})
	}
}

func TestTrackDeviceFingerprint(t *testing.T) {
	err := TrackDeviceFingerprint(
		1,
		"device-fp-test",
		"Mozilla/5.0",
		"192.168.1.1",
	)

	assert.NoError(t, err)

	device, err := model.GetDeviceFingerprint("device-fp-test", 1)
	assert.NoError(t, err)
	assert.Equal(t, "device-fp-test", device.Fingerprint)
	assert.Equal(t, 1, device.UserId)
}

func TestTrackIPCluster(t *testing.T) {
	err := TrackIPCluster("192.168.1.1", 1)
	assert.NoError(t, err)

	cluster, err := model.GetIPCluster("192.168.1.1")
	assert.NoError(t, err)
	assert.Equal(t, "192.168.1.1", cluster.IpAddress)
	assert.GreaterOrEqual(t, cluster.TotalRequests, 1)
}

func TestApproveAndIgnoreAnomaly(t *testing.T) {
	anomaly := &model.SecurityAnomaly{
		UserId:      1,
		AnomalyType: AnomalyTypeHighRPM,
		Severity:    "malicious",
		Description: "Test anomaly",
		Status:      StatusPending,
	}

	err := model.CreateSecurityAnomaly(anomaly)
	assert.NoError(t, err)

	err = ApproveAnomaly(anomaly.Id, 100, "Reviewed and approved")
	assert.NoError(t, err)

	updated, err := model.GetSecurityAnomaly(anomaly.Id)
	assert.NoError(t, err)
	assert.Equal(t, StatusApproved, updated.Status)
	assert.Equal(t, "approved", updated.ReviewDecision)
	assert.NotNil(t, updated.ReviewedAt)
	assert.NotNil(t, updated.ReviewedBy)

	anomaly2 := &model.SecurityAnomaly{
		UserId:      1,
		AnomalyType: AnomalyTypeHighRPM,
		Severity:    "violation",
		Description: "Test anomaly 2",
		Status:      StatusPending,
	}

	err = model.CreateSecurityAnomaly(anomaly2)
	assert.NoError(t, err)

	err = IgnoreAnomaly(anomaly2.Id, 100, "False positive")
	assert.NoError(t, err)

	updated2, err := model.GetSecurityAnomaly(anomaly2.Id)
	assert.NoError(t, err)
	assert.Equal(t, StatusIgnored, updated2.Status)
	assert.Equal(t, "ignored", updated2.ReviewDecision)
	assert.Equal(t, "False positive", updated2.ReviewRationale)
}

func TestProcessAnomalyWithAutoEnforcement(t *testing.T) {
	anomaly := &model.SecurityAnomaly{
		UserId:      1,
		AnomalyType: AnomalyTypeHighRPM,
		Severity:    "malicious",
		Description: "Auto enforcement test",
		Status:      StatusPending,
		RiskScore:   90,
	}

	err := model.CreateSecurityAnomaly(anomaly)
	assert.NoError(t, err)

	err = ProcessAnomaly(anomaly)
	assert.NoError(t, err)

	updated, err := model.GetSecurityAnomaly(anomaly.Id)
	assert.NoError(t, err)

	if updated.Status == StatusActioned {
		assert.NotEmpty(t, updated.ActionTaken)
		assert.NotNil(t, updated.ActionedAt)
	}
}

func TestGetAnomalyTrends(t *testing.T) {
	startTime := time.Now().AddDate(0, 0, -7)
	endTime := time.Now()

	trends, err := model.GetAnomalyTrends(startTime, endTime)
	assert.NoError(t, err)
	assert.NotNil(t, trends)
}

func TestGetAnomalyCountsBySeverity(t *testing.T) {
	startTime := time.Now().AddDate(0, 0, -7)
	endTime := time.Now()

	counts, err := model.GetAnomalyCountsBySeverity(startTime, endTime)
	assert.NoError(t, err)
	assert.NotNil(t, counts)
}
