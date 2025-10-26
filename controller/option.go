package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting"
	"github.com/QuantumNous/new-api/setting/config"
	"github.com/QuantumNous/new-api/setting/console_setting"
	"github.com/QuantumNous/new-api/setting/ratio_setting"
	"github.com/QuantumNous/new-api/setting/system_setting"

	"github.com/gin-gonic/gin"
)

var featureSectionAliases = map[string]string{
	"billing":      "billing",
	"finance":      "billing",
	"fea_mance":    "billing",
	"governance":   "governance",
	"public_logs":  "public_logs",
	"publiclog":    "public_logs",
	"publiclogs":   "public_logs",
	"public_log":   "public_logs",
	"public-log":   "public_logs",
	"public log":   "public_logs",
	"public\tlogs": "public_logs",
}

func canonicalFeatureSection(raw string) (string, bool) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", false
	}

	normalised := strings.ToLower(trimmed)
	normalised = strings.ReplaceAll(normalised, "-", "_")
	normalised = strings.ReplaceAll(normalised, " ", "_")
	normalised = strings.ReplaceAll(normalised, "\t", "_")
	normalised = strings.Trim(normalised, "_")
	if normalised == "" {
		return "", false
	}

	if canonical, ok := featureSectionAliases[normalised]; ok {
		return canonical, true
	}

	return normalised, true
}

func addFeatureSection(sectionSet map[string]struct{}, raw string) {
	if canonical, ok := canonicalFeatureSection(raw); ok {
		sectionSet[canonical] = struct{}{}
	}
}

func parseFeatureSections(param string) map[string]struct{} {
	sections := make(map[string]struct{})
	if param == "" {
		return sections
	}

	for _, section := range strings.Split(param, ",") {
		addFeatureSection(sections, section)
	}

	return sections
}

func parseFeatureSectionsFromQuery(c *gin.Context) map[string]struct{} {
	sections := make(map[string]struct{})

	values := c.QueryArray("sections")
	for _, value := range values {
		addFeatureSection(sections, value)
	}

	if len(sections) == 0 {
		legacy := c.Query("sections")
		for _, value := range strings.Split(legacy, ",") {
			addFeatureSection(sections, value)
		}
	}

	return sections
}

func shouldIncludeSection(sectionSet map[string]struct{}, name string) bool {
	if len(sectionSet) == 0 {
		return true
	}
	_, ok := sectionSet[name]
	return ok
}

func buildFeatureConfigResponse(sectionSet map[string]struct{}) dto.FeatureConfigResponse {
	response := dto.FeatureConfigResponse{}

	if shouldIncludeSection(sectionSet, "billing") {
		cfg := config.GetBillingConfig()
		response.Billing = &dto.BillingFeatureConfig{
			Enabled:       cfg.Enabled,
			DefaultMode:   cfg.DefaultMode,
			AutoFallback:  cfg.AutoFallback,
			ResetHourUTC:  cfg.ResetHourUTC,
			ResetTimezone: cfg.ResetTimezone,
			NotifyOnReset: cfg.NotifyOnReset,
		}
	}

	if shouldIncludeSection(sectionSet, "governance") {
		cfg := config.GetGovernanceConfig()
		response.Governance = &dto.GovernanceFeatureConfig{
			Enabled:           cfg.Enabled,
			AbuseRPMThreshold: cfg.AbuseRPMThreshold,
			RerouteModelAlias: cfg.RerouteModelAlias,
			FlagTTLHours:      cfg.FlagTTLHours,
		}
	}

	if shouldIncludeSection(sectionSet, "public_logs") {
		cfg := config.GetPublicLogsConfig()
		response.PublicLogs = &dto.PublicLogsFeatureConfig{
			Enabled:          cfg.Enabled,
			PublicVisibility: cfg.PublicVisibility,
			RetentionDays:    cfg.RetentionDays,
		}
	}

	return response
}

func persistFeatureConfig(section string, cfg interface{}) error {
	configMap, err := config.ConfigToMap(cfg)
	if err != nil {
		return err
	}

	for key, value := range configMap {
		if err := model.UpdateOption(section+"."+key, value); err != nil {
			return err
		}
	}

	return nil
}

// GetFeatureOptions returns non-sensitive feature flag configuration for billing/governance/public logs.
func GetFeatureOptions(c *gin.Context) {
	sectionSet := parseFeatureSectionsFromQuery(c)
	response := buildFeatureConfigResponse(sectionSet)
	common.ApiSuccess(c, response)
}

// UpdateFeatureOptions updates persisted feature flag configuration (root only).
func UpdateFeatureOptions(c *gin.Context) {
	var req dto.FeatureConfigUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request payload",
		})
		return
	}

	if req.Billing == nil && req.Governance == nil && req.PublicLogs == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "No feature sections provided",
		})
		return
	}

	updatedSections := make(map[string]struct{})

	if req.Billing != nil {
		cfg := config.GetBillingConfig()
		if req.Billing.Enabled != nil {
			cfg.Enabled = *req.Billing.Enabled
		}
		if req.Billing.DefaultMode != nil {
			cfg.DefaultMode = strings.TrimSpace(*req.Billing.DefaultMode)
		}
		if req.Billing.AutoFallback != nil {
			cfg.AutoFallback = *req.Billing.AutoFallback
		}
		if req.Billing.ResetHourUTC != nil {
			cfg.ResetHourUTC = *req.Billing.ResetHourUTC
		}
		if req.Billing.ResetTimezone != nil {
			cfg.ResetTimezone = strings.TrimSpace(*req.Billing.ResetTimezone)
		}
		if req.Billing.NotifyOnReset != nil {
			cfg.NotifyOnReset = *req.Billing.NotifyOnReset
		}

		if err := persistFeatureConfig("billing", cfg); err != nil {
			common.ApiError(c, err)
			return
		}
		updatedSections["billing"] = struct{}{}
	}

	if req.Governance != nil {
		cfg := config.GetGovernanceConfig()
		if req.Governance.Enabled != nil {
			cfg.Enabled = *req.Governance.Enabled
		}
		if req.Governance.AbuseRPMThreshold != nil {
			cfg.AbuseRPMThreshold = *req.Governance.AbuseRPMThreshold
		}
		if req.Governance.RerouteModelAlias != nil {
			cfg.RerouteModelAlias = strings.TrimSpace(*req.Governance.RerouteModelAlias)
		}
		if req.Governance.FlagTTLHours != nil {
			cfg.FlagTTLHours = *req.Governance.FlagTTLHours
		}

		if err := persistFeatureConfig("governance", cfg); err != nil {
			common.ApiError(c, err)
			return
		}
		updatedSections["governance"] = struct{}{}
	}

	if req.PublicLogs != nil {
		if req.PublicLogs.RetentionDays != nil && *req.PublicLogs.RetentionDays <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Retention days must be greater than zero",
			})
			return
		}

		cfg := config.GetPublicLogsConfig()
		if req.PublicLogs.Enabled != nil {
			cfg.Enabled = *req.PublicLogs.Enabled
		}
		if req.PublicLogs.PublicVisibility != nil {
			cfg.PublicVisibility = strings.TrimSpace(*req.PublicLogs.PublicVisibility)
		}
		if req.PublicLogs.RetentionDays != nil {
			cfg.RetentionDays = *req.PublicLogs.RetentionDays
		}

		if err := persistFeatureConfig("public_logs", cfg); err != nil {
			common.ApiError(c, err)
			return
		}
		updatedSections["public_logs"] = struct{}{}
	}

	response := buildFeatureConfigResponse(updatedSections)
	common.ApiSuccess(c, response)
}

func GetOptions(c *gin.Context) {
	var options []*model.Option
	common.OptionMapRWMutex.Lock()
	for k, v := range common.OptionMap {
		if strings.HasSuffix(k, "Token") || strings.HasSuffix(k, "Secret") || strings.HasSuffix(k, "Key") {
			continue
		}
		options = append(options, &model.Option{
			Key:   k,
			Value: common.Interface2String(v),
		})
	}
	common.OptionMapRWMutex.Unlock()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    options,
	})
	return
}

type OptionUpdateRequest struct {
	Key   string `json:"key"`
	Value any    `json:"value"`
}

func UpdateOption(c *gin.Context) {
	var option OptionUpdateRequest
	err := json.NewDecoder(c.Request.Body).Decode(&option)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的参数",
		})
		return
	}
	switch option.Value.(type) {
	case bool:
		option.Value = common.Interface2String(option.Value.(bool))
	case float64:
		option.Value = common.Interface2String(option.Value.(float64))
	case int:
		option.Value = common.Interface2String(option.Value.(int))
	default:
		option.Value = fmt.Sprintf("%v", option.Value)
	}
	switch option.Key {
	case "GitHubOAuthEnabled":
		if option.Value == "true" && common.GitHubClientId == "" {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "无法启用 GitHub OAuth，请先填入 GitHub Client Id 以及 GitHub Client Secret！",
			})
			return
		}
	case "oidc.enabled":
		if option.Value == "true" && system_setting.GetOIDCSettings().ClientId == "" {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "无法启用 OIDC 登录，请先填入 OIDC Client Id 以及 OIDC Client Secret！",
			})
			return
		}
	case "LinuxDOOAuthEnabled":
		if option.Value == "true" && common.LinuxDOClientId == "" {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "无法启用 LinuxDO OAuth，请先填入 LinuxDO Client Id 以及 LinuxDO Client Secret！",
			})
			return
		}
	case "EmailDomainRestrictionEnabled":
		if option.Value == "true" && len(common.EmailDomainWhitelist) == 0 {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "无法启用邮箱域名限制，请先填入限制的邮箱域名！",
			})
			return
		}
	case "WeChatAuthEnabled":
		if option.Value == "true" && common.WeChatServerAddress == "" {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "无法启用微信登录，请先填入微信登录相关配置信息！",
			})
			return
		}
	case "TurnstileCheckEnabled":
		if option.Value == "true" && common.TurnstileSiteKey == "" {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "无法启用 Turnstile 校验，请先填入 Turnstile 校验相关配置信息！",
			})

			return
		}
	case "TelegramOAuthEnabled":
		if option.Value == "true" && common.TelegramBotToken == "" {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "无法启用 Telegram OAuth，请先填入 Telegram Bot Token！",
			})
			return
		}
	case "GroupRatio":
		err = ratio_setting.CheckGroupRatio(option.Value.(string))
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}
	case "ImageRatio":
		err = ratio_setting.UpdateImageRatioByJSONString(option.Value.(string))
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "图片倍率设置失败: " + err.Error(),
			})
			return
		}
	case "AudioRatio":
		err = ratio_setting.UpdateAudioRatioByJSONString(option.Value.(string))
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "音频倍率设置失败: " + err.Error(),
			})
			return
		}
	case "AudioCompletionRatio":
		err = ratio_setting.UpdateAudioCompletionRatioByJSONString(option.Value.(string))
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "音频补全倍率设置失败: " + err.Error(),
			})
			return
		}
	case "ModelRequestRateLimitGroup":
		err = setting.CheckModelRequestRateLimitGroup(option.Value.(string))
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}
	case "console_setting.api_info":
		err = console_setting.ValidateConsoleSettings(option.Value.(string), "ApiInfo")
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}
	case "console_setting.announcements":
		err = console_setting.ValidateConsoleSettings(option.Value.(string), "Announcements")
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}
	case "console_setting.faq":
		err = console_setting.ValidateConsoleSettings(option.Value.(string), "FAQ")
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}
	case "console_setting.uptime_kuma_groups":
		err = console_setting.ValidateConsoleSettings(option.Value.(string), "UptimeKumaGroups")
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}
	}
	err = model.UpdateOption(option.Key, option.Value.(string))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
	return
}
