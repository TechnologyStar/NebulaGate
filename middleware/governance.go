//go:build governance
package middleware

import (
    "fmt"
    "net/http"
    "strings"
    "time"

    "github.com/QuantumNous/new-api/common"
    "github.com/QuantumNous/new-api/constant"
    "github.com/QuantumNous/new-api/logger"
    "github.com/QuantumNous/new-api/model"
    governanceSvc "github.com/QuantumNous/new-api/service/governance"
    cfg "github.com/QuantumNous/new-api/setting/config"
    "github.com/QuantumNous/new-api/setting/ratio_setting"

    "github.com/gin-gonic/gin"
)

type governanceDecision struct {
    triggered  bool
    severity   string
    detectors  map[string]map[string]string
    reasons    []string
    reasonSet  map[string]struct{}
    flagReason string
}

func newGovernanceDecision() *governanceDecision {
    return &governanceDecision{
        detectors: make(map[string]map[string]string),
        reasonSet: make(map[string]struct{}),
    }
}

func (d *governanceDecision) add(detector string, res governanceSvc.DetectorResult) {
    if !res.Triggered {
        return
    }
    d.triggered = true
    if res.Metadata != nil {
        // Copy metadata to avoid later mutation side-effects
        copied := make(map[string]string, len(res.Metadata))
        for k, v := range res.Metadata {
            copied[k] = v
        }
        d.detectors[detector] = copied
    } else {
        d.detectors[detector] = map[string]string{}
    }
    if rank(res.Severity) > rank(d.severity) {
        d.severity = res.Severity
    } else if d.severity == "" {
        d.severity = res.Severity
    }
    for _, reason := range res.Reasons {
        if _, exists := d.reasonSet[reason]; !exists {
            d.reasonSet[reason] = struct{}{}
            d.reasons = append(d.reasons, reason)
        }
    }
    if d.flagReason == "" {
        if len(res.Reasons) > 0 {
            d.flagReason = res.Reasons[0]
        } else {
            d.flagReason = res.Severity
        }
    }
}

func rank(severity string) int {
    switch severity {
    case governanceSvc.SeverityMalicious:
        return 2
    case governanceSvc.SeverityViolation:
        return 1
    default:
        return 0
    }
}

// Governance injects governance detection into the relay pipeline.
func Governance() gin.HandlerFunc {
    return func(c *gin.Context) {
        if alreadyChecked := c.GetBool("governance_checked"); alreadyChecked {
            c.Next()
            return
        }
        c.Set("governance_checked", true)

        config := cfg.GetGovernanceConfig()
        if config == nil || !config.Enabled {
            c.Next()
            return
        }

        if c.Request == nil || (c.Request.Method != http.MethodPost && c.Request.Method != http.MethodPatch && c.Request.Method != http.MethodPut) {
            c.Next()
            return
        }

        subjectType, subjectID := deriveSubject(c)
        subjectKey := subjectType
        if subjectID > 0 {
            subjectKey = fmt.Sprintf("%s:%d", subjectType, subjectID)
        }

        prompt := extractPromptText(c)
        decision := evaluateGovernance(subjectKey, prompt)
        if decision == nil || !decision.triggered {
            c.Next()
            return
        }

        requestedModel := common.GetContextKeyString(c, constant.ContextKeyRequestedModel)
        if requestedModel == "" {
            requestedModel = common.GetContextKeyString(c, constant.ContextKeyOriginalModel)
        }

        meta := buildGovernanceMetadata(c, decision, subjectType, subjectID, subjectKey, requestedModel)
        alias, applied := applyGovernanceFallback(c, config, requestedModel, decision, meta)
        common.SetContextKey(c, constant.ContextKeyGovernanceMetadata, meta)
        common.SetContextKey(c, constant.ContextKeySkipLeaderboard, true)

        persistRequestFlag(c, config, subjectType, subjectID, alias, decision)

        logger.LogWarn(c, fmt.Sprintf("governance flag triggered severity=%s reasons=%v", decision.severity, decision.reasons))

        c.Next()
    }
}

func evaluateGovernance(subjectKey, prompt string) *governanceDecision {
    decision := newGovernanceDecision()
    now := time.Now()

    decision.add("high_rpm", governanceSvc.DetectHighRPM(subjectKey, now))

    trimmed := strings.TrimSpace(prompt)
    if trimmed != "" {
        decision.add("prompt_sanity", governanceSvc.DetectPromptSanity(trimmed))
        decision.add("keyword_policy", governanceSvc.DetectKeywordPolicy(trimmed))
    }

    if !decision.triggered {
        return nil
    }
    if decision.flagReason == "" {
        if decision.severity == governanceSvc.SeverityMalicious {
            decision.flagReason = common.FlagReasonAbuse
        } else {
            decision.flagReason = common.FlagReasonViolation
        }
    }
    return decision
}

func deriveSubject(c *gin.Context) (string, int) {
    userID := common.GetContextKeyInt(c, constant.ContextKeyUserId)
    if userID > 0 {
        return common.AssignmentSubjectTypeUser, userID
    }
    tokenID := common.GetContextKeyInt(c, constant.ContextKeyTokenId)
    if tokenID > 0 {
        return common.AssignmentSubjectTypeToken, tokenID
    }
    return common.AssignmentSubjectTypeUser, 0
}

func extractPromptText(c *gin.Context) string {
    body, err := common.GetRequestBody(c)
    if err != nil || len(body) == 0 {
        return ""
    }
    contentType := c.Request.Header.Get("Content-Type")
    if strings.HasPrefix(contentType, "application/json") || contentType == "" {
        var payload any
        if err := common.Unmarshal(body, &payload); err == nil {
            builder := &strings.Builder{}
            collectPromptText(payload, builder, 0)
            text := builder.String()
            if len(text) > 8192 {
                return text[:8192]
            }
            return text
        }
    }
    text := string(body)
    if len(text) > 8192 {
        return text[:8192]
    }
    return text
}

func collectPromptText(node any, builder *strings.Builder, depth int) {
    if depth > 8 || node == nil {
        return
    }
    switch val := node.(type) {
    case string:
        if val == "" {
            return
        }
        if builder.Len() > 0 {
            builder.WriteString("\n")
        }
        builder.WriteString(val)
    case []any:
        for _, item := range val {
            collectPromptText(item, builder, depth+1)
        }
    case []interface{}:
        for _, item := range val {
            collectPromptText(item, builder, depth+1)
        }
    case map[string]any:
        collectPromptText(val["prompt"], builder, depth+1)
        collectPromptText(val["input"], builder, depth+1)
        collectPromptText(val["inputs"], builder, depth+1)
        collectPromptText(val["messages"], builder, depth+1)
        collectPromptText(val["content"], builder, depth+1)
        collectPromptText(val["data"], builder, depth+1)
        for _, v := range val {
            collectPromptText(v, builder, depth+1)
        }
    case map[string]interface{}:
        collectPromptText(val["prompt"], builder, depth+1)
        collectPromptText(val["input"], builder, depth+1)
        collectPromptText(val["inputs"], builder, depth+1)
        collectPromptText(val["messages"], builder, depth+1)
        collectPromptText(val["content"], builder, depth+1)
        collectPromptText(val["data"], builder, depth+1)
        for _, v := range val {
            collectPromptText(v, builder, depth+1)
        }
    }
}

func buildGovernanceMetadata(c *gin.Context, decision *governanceDecision, subjectType string, subjectID int, subjectKey, requestedModel string) map[string]interface{} {
    metadata := make(map[string]interface{})
    metadata["severity"] = decision.severity
    metadata["reasons"] = append([]string(nil), decision.reasons...)
    metadata["flag_reason"] = decision.flagReason
    detectors := make(map[string]map[string]string, len(decision.detectors))
    for name, data := range decision.detectors {
        copied := make(map[string]string, len(data))
        for k, v := range data {
            copied[k] = v
        }
        detectors[name] = copied
    }
    metadata["detectors"] = detectors
    metadata["subject_type"] = subjectType
    metadata["subject_id"] = subjectID
    metadata["subject_key"] = subjectKey
    metadata["requested_model"] = requestedModel
    if rid := c.GetString(common.RequestIdKey); rid != "" {
        metadata["request_id"] = rid
    }
    return metadata
}

func applyGovernanceFallback(c *gin.Context, config *cfg.GovernanceConfig, requestedModel string, decision *governanceDecision, metadata map[string]interface{}) (string, bool) {
    alias := selectFallbackAlias(config, decision)
    if alias == "" {
        metadata["fallback_applied"] = false
        return "", false
    }
    metadata["fallback_alias"] = alias

    normalizedRequested := ratio_setting.FormatMatchingModelName(requestedModel)
    normalizedAlias := ratio_setting.FormatMatchingModelName(alias)
    if normalizedAlias == normalizedRequested {
        metadata["fallback_applied"] = false
        metadata["fallback_error"] = "alias_matches_requested"
        return alias, false
    }

    if current := common.GetContextKeyString(c, constant.ContextKeyGovernanceUpstreamModel); current != "" {
        if ratio_setting.FormatMatchingModelName(current) == normalizedAlias {
            metadata["fallback_applied"] = false
            metadata["fallback_error"] = "alias_already_applied"
            return alias, false
        }
    }

    if _, ok, _ := ratio_setting.GetModelRatio(alias); !ok {
        logger.LogError(c, fmt.Sprintf("governance fallback alias %s not registered", alias))
        metadata["fallback_applied"] = false
        metadata["fallback_error"] = "alias_not_registered"
        return alias, false
    }

    common.SetContextKey(c, constant.ContextKeyGovernanceUpstreamModel, alias)
    metadata["fallback_applied"] = true
    metadata["effective_model"] = alias
    logger.LogWarn(c, fmt.Sprintf("governance reroute applied to %s", alias))
    return alias, true
}

func selectFallbackAlias(config *cfg.GovernanceConfig, decision *governanceDecision) string {
    if decision.severity == governanceSvc.SeverityMalicious {
        if config.MaliciousFallbackAlias != "" {
            return config.MaliciousFallbackAlias
        }
        if common.GovernanceMaliciousFallbackAlias != "" {
            return common.GovernanceMaliciousFallbackAlias
        }
    }
    if config.ViolationFallbackAlias != "" {
        return config.ViolationFallbackAlias
    }
    if config.RerouteModelAlias != "" {
        return config.RerouteModelAlias
    }
    if common.GovernanceViolationFallbackAlias != "" {
        return common.GovernanceViolationFallbackAlias
    }
    return common.GovernanceRerouteModelAlias
}

func persistRequestFlag(c *gin.Context, config *cfg.GovernanceConfig, subjectType string, subjectID int, alias string, decision *governanceDecision) {
    requestID := c.GetString(common.RequestIdKey)
    userID := common.GetContextKeyInt(c, constant.ContextKeyUserId)
    tokenID := common.GetContextKeyInt(c, constant.ContextKeyTokenId)

    flag := &model.RequestFlag{
        RequestId:              requestID,
        SubjectType:            subjectType,
        SubjectId:              subjectID,
        ReroutedModelAlias:     alias,
        Reason:                 decision.flagReason,
        ExcludedFromLeaderboard: true,
    }
    if userID > 0 {
        flag.UserId = &userID
    }
    if tokenID > 0 {
        flag.TokenId = &tokenID
    }
    if config.FlagTTLHours > 0 {
        ttl := time.Now().UTC().Add(time.Duration(config.FlagTTLHours) * time.Hour)
        flag.TtlAt = &ttl
    }
    if flag.Reason == "" {
        if decision.severity == governanceSvc.SeverityMalicious {
            flag.Reason = common.FlagReasonAbuse
        } else {
            flag.Reason = common.FlagReasonViolation
        }
    }
    if err := model.CreateRequestFlag(flag); err != nil {
        logger.LogError(c, fmt.Sprintf("failed to persist governance flag: %v", err))
    }

    // Record security violation
    recordSecurityViolation(c, userID, tokenID, decision, alias)
}

func recordSecurityViolation(c *gin.Context, userID, tokenID int, decision *governanceDecision, alias string) {
    if userID == 0 {
        return
    }

    // Extract content snippet from context
    prompt := extractPromptText(c)
    
    // Get matched keywords from decision
    var keywords []string
    if keywordDetector, ok := decision.detectors["keyword_policy"]; ok {
        if kw, exists := keywordDetector["matched_keywords"]; exists {
            keywords = append(keywords, kw)
        }
    }

    // Get request model
    requestedModel := common.GetContextKeyString(c, constant.ContextKeyRequestedModel)
    if requestedModel == "" {
        requestedModel = common.GetContextKeyString(c, constant.ContextKeyOriginalModel)
    }

    // Get IP address
    ipAddress := common.GetClientIP(c)

    // Determine action taken
    action := "log"
    if alias != "" {
        action = "redirect"
    }

    requestID := c.GetString(common.RequestIdKey)

    // Import service package at the top of file and use it
    var tokenIDPtr *int
    if tokenID > 0 {
        tokenIDPtr = &tokenID
    }

    // Call service to record violation (async to avoid blocking request)
    go func() {
        // Use a new service import to avoid circular dependency
        if err := recordViolationAsync(userID, tokenIDPtr, prompt, keywords, requestedModel, ipAddress, requestID, decision.severity, action); err != nil {
            // Log error but don't fail the request
            common.SysLog(fmt.Sprintf("failed to record security violation: %v", err))
        }
    }()
}

func recordViolationAsync(userID int, tokenID *int, content string, keywords []string, model, ipAddress, requestID, severity, action string) error {
    // Import the service package at function level to avoid init-time circular dependency
    // We'll call the model layer directly here
    snippet := content
    if len(snippet) > 500 {
        snippet = snippet[:500] + "..."
    }

    violation := &model.SecurityViolation{
        UserId:          userID,
        TokenId:         tokenID,
        ViolatedAt:      time.Now(),
        ContentSnippet:  snippet,
        MatchedKeywords: joinStrings(keywords, ", "),
        Model:           model,
        IpAddress:       ipAddress,
        RequestId:       requestID,
        Severity:        severity,
        ActionTaken:     action,
    }

    if err := model.CreateSecurityViolation(violation); err != nil {
        return err
    }

    // Increment user violation count
    if err := model.IncrementViolationCount(userID); err != nil {
        return err
    }

    return nil
}

func joinStrings(strs []string, sep string) string {
    if len(strs) == 0 {
        return ""
    }
    result := strs[0]
    for i := 1; i < len(strs); i++ {
        result += sep + strs[i]
    }
    return result
}
