package service

import (
    "encoding/hex"
    "fmt"
    "time"

    "github.com/QuantumNous/new-api/common"
    "github.com/QuantumNous/new-api/model"
)

// AuditMetadata represents additional audit info stored in RequestLog.Metadata.
type AuditMetadata struct {
    Mode       string `json:"mode"`
    Amount     int64  `json:"amount"`
    SubjectKey string `json:"subject_key"`
    Engine     string `json:"engine"`
    Idempotent bool   `json:"idempotent"`
}

func anonymizeSubject(subjectType string, subjectId int) string {
    plain := []byte(subjectType + ":" + fmt.Sprintf("%d", subjectId))
    sum := common.HmacSha256Raw(plain, []byte(common.SessionSecret))
    return hex.EncodeToString(sum)
}

// BuildRequestLog builds a RequestLog ready to be created.
func BuildRequestLog(requestId string, subjectType string, subjectId int, modelAlias string, upstream string,
    assignment *model.PlanAssignment, usageMetric string, promptTokens, completionTokens, totalTokens int64,
    latencyMs int64, amount int64, mode string, meta AuditMetadata) *model.RequestLog {
    payload := map[string]any{
        "mode":       mode,
        "amount":     amount,
        "audit_meta": meta,
    }
    b, _ := common.Marshal(payload)
    log := &model.RequestLog{
        RequestId:             requestId,
        OccurredAt:            time.Now().UTC(),
        ModelAlias:            modelAlias,
        UpstreamProvider:      upstream,
        SubjectType:           subjectType,
        AnonymizedSubjectHash: anonymizeSubject(subjectType, subjectId),
        UsageMetric:           usageMetric,
        PromptTokens:          promptTokens,
        CompletionTokens:      completionTokens,
        TotalTokens:           totalTokens,
        LatencyMs:             latencyMs,
        Metadata:              model.JSONValue(b),
    }
    if assignment != nil {
        log.PlanId = &assignment.PlanId
        log.PlanAssignmentId = &assignment.Id
    }
    return log
}
