package controller

import (
    "bytes"
    "encoding/json"
    "io"
    "net/http"
    "strconv"
    "strings"
    "time"

    "github.com/QuantumNous/new-api/common"
    "github.com/QuantumNous/new-api/model"

    "github.com/gin-gonic/gin"
)

func GetAllTokens(c *gin.Context) {
    userId := c.GetInt("id")
    pageInfo := common.GetPageQuery(c)
    tokens, err := model.GetAllUserTokens(userId, pageInfo.GetStartIdx(), pageInfo.GetPageSize())
    if err != nil {
        common.ApiError(c, err)
        return
    }
    total, _ := model.CountUserTokens(userId)
    pageInfo.SetTotal(int(total))
    pageInfo.SetItems(tokens)
    common.ApiSuccess(c, pageInfo)
    return
}

func SearchTokens(c *gin.Context) {
    userId := c.GetInt("id")
    keyword := c.Query("keyword")
    token := c.Query("token")
    tokens, err := model.SearchUserTokens(userId, keyword, token)
    if err != nil {
        common.ApiError(c, err)
        return
    }
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "message": "",
        "data":    tokens,
    })
    return
}

func GetToken(c *gin.Context) {
    id, err := strconv.Atoi(c.Param("id"))
    userId := c.GetInt("id")
    if err != nil {
        common.ApiError(c, err)
        return
    }
    token, err := model.GetTokenByIds(id, userId)
    if err != nil {
        common.ApiError(c, err)
        return
    }
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "message": "",
        "data":    token,
    })
    return
}

func GetTokenStatus(c *gin.Context) {
    tokenId := c.GetInt("token_id")
    userId := c.GetInt("id")
    token, err := model.GetTokenByIds(tokenId, userId)
    if err != nil {
        common.ApiError(c, err)
        return
    }
    expiredAt := token.ExpiredTime
    if expiredAt == -1 {
        expiredAt = 0
    }
    c.JSON(http.StatusOK, gin.H{
        "object":          "credit_summary",
        "total_granted":   token.RemainQuota,
        "total_used":      0, // not supported currently
        "total_available": token.RemainQuota,
        "expires_at":      expiredAt * 1000,
    })
}

func GetTokenUsage(c *gin.Context) {
    authHeader := c.GetHeader("Authorization")
    if authHeader == "" {
        c.JSON(http.StatusUnauthorized, gin.H{
            "success": false,
            "message": "No Authorization header",
        })
        return
    }

    parts := strings.Split(authHeader, " ")
    if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
        c.JSON(http.StatusUnauthorized, gin.H{
            "success": false,
            "message": "Invalid Bearer token",
        })
        return
    }
    tokenKey := parts[1]

    token, err := model.GetTokenByKey(strings.TrimPrefix(tokenKey, "sk-"), false)
    if err != nil {
        c.JSON(http.StatusOK, gin.H{
            "success": false,
            "message": err.Error(),
        })
        return
    }

    expiredAt := token.ExpiredTime
    if expiredAt == -1 {
        expiredAt = 0
    }

    c.JSON(http.StatusOK, gin.H{
        "code":    true,
        "message": "ok",
        "data": gin.H{
            "object":               "token_usage",
            "name":                 token.Name,
            "total_granted":        token.RemainQuota + token.UsedQuota,
            "total_used":           token.UsedQuota,
            "total_available":      token.RemainQuota,
            "unlimited_quota":      token.UnlimitedQuota,
            "model_limits":         token.GetModelLimitsMap(),
            "model_limits_enabled": token.ModelLimitsEnabled,
            "expires_at":           expiredAt,
        },
    })
}

func AddToken(c *gin.Context) {
    token := model.Token{}
    err := c.ShouldBindJSON(&token)
    if err != nil {
        common.ApiError(c, err)
        return
    }
    if len(token.Name) > 30 {
        c.JSON(http.StatusOK, gin.H{
            "success": false,
            "message": "令牌名称过长",
        })
        return
    }
    key, err := common.GenerateKey()
    if err != nil {
        c.JSON(http.StatusOK, gin.H{
            "success": false,
            "message": "生成令牌失败",
        })
        common.SysLog("failed to generate token key: " + err.Error())
        return
    }
    cleanToken := model.Token{
        UserId:             c.GetInt("id"),
        Name:               token.Name,
        Key:                key,
        CreatedTime:        common.GetTimestamp(),
        AccessedTime:       common.GetTimestamp(),
        ExpiredTime:        token.ExpiredTime,
        RemainQuota:        token.RemainQuota,
        UnlimitedQuota:     token.UnlimitedQuota,
        ModelLimitsEnabled: token.ModelLimitsEnabled,
        ModelLimits:        token.ModelLimits,
        AllowIps:           token.AllowIps,
        Group:              token.Group,
    }
    err = cleanToken.Insert()
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

func DeleteToken(c *gin.Context) {
    id, _ := strconv.Atoi(c.Param("id"))
    userId := c.GetInt("id")
    err := model.DeleteTokenById(id, userId)
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

func UpdateToken(c *gin.Context) {
    userId := c.GetInt("id")
    statusOnly := c.Query("status_only")
    
    bodyBytes, err := io.ReadAll(c.Request.Body)
    if err != nil {
        common.ApiError(c, err)
        return
    }
    c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
    
    token := model.Token{}
    err = c.ShouldBindJSON(&token)
    if err != nil {
        common.ApiError(c, err)
        return
    }
    
    c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
    var rawData map[string]interface{}
    _ = json.Unmarshal(bodyBytes, &rawData)
    if len(token.Name) > 30 {
        c.JSON(http.StatusOK, gin.H{
            "success": false,
            "message": "令牌名称过长",
        })
        return
    }
    cleanToken, err := model.GetTokenByIds(token.Id, userId)
    if err != nil {
        common.ApiError(c, err)
        return
    }
    if token.Status == common.TokenStatusEnabled {
        if cleanToken.Status == common.TokenStatusExpired && cleanToken.ExpiredTime <= common.GetTimestamp() && cleanToken.ExpiredTime != -1 {
            c.JSON(http.StatusOK, gin.H{
                "success": false,
                "message": "令牌已过期，无法启用，请先修改令牌过期时间，或者设置为永不过期",
            })
            return
        }
        if cleanToken.Status == common.TokenStatusExhausted && cleanToken.RemainQuota <= 0 && !cleanToken.UnlimitedQuota {
            c.JSON(http.StatusOK, gin.H{
                "success": false,
                "message": "令牌可用额度已用尽，无法启用，请先修改令牌剩余额度，或者设置为无限额度",
            })
            return
        }
    }

    requestedMode := cleanToken.GetBillingMode()
    modeProvided := false
    if rawMode, ok := rawData["billing_mode"]; ok {
        modeProvided = true
        modeStr, ok := rawMode.(string)
        if !ok {
            c.JSON(http.StatusBadRequest, gin.H{
                "success": false,
                "message": "billing_mode must be a string",
            })
            return
        }
        modeStr = strings.ToLower(strings.TrimSpace(modeStr))
        if modeStr == "" {
            modeStr = common.BillingDefaultMode
        }
        switch modeStr {
        case common.BillingModeBalance, common.BillingModePlan, common.BillingModeAuto:
            requestedMode = modeStr
        default:
            c.JSON(http.StatusBadRequest, gin.H{
                "success": false,
                "message": "invalid billing_mode",
            })
            return
        }
    }

    _, planAssignmentProvided := rawData["plan_assignment_id"]
    var targetAssignmentID int
    hasAssignment := false
    if planAssignmentProvided {
        if token.PlanAssignmentId != nil {
            targetAssignmentID = *token.PlanAssignmentId
            hasAssignment = true
        }
    } else if cleanToken.PlanAssignmentId != nil {
        targetAssignmentID = *cleanToken.PlanAssignmentId
        hasAssignment = true
    }

    if requestedMode == common.BillingModePlan {
        if !hasAssignment || targetAssignmentID <= 0 {
            c.JSON(http.StatusBadRequest, gin.H{
                "success": false,
                "message": "plan mode requires an active plan assignment",
            })
            return
        }
        assignment, err := model.GetPlanAssignmentById(targetAssignmentID)
        if err != nil {
            common.ApiError(c, err)
            return
        }
        now := time.Now().UTC()
        if assignment.ActivatedAt.After(now) || (assignment.DeactivatedAt != nil && !assignment.DeactivatedAt.After(now)) {
            c.JSON(http.StatusBadRequest, gin.H{
                "success": false,
                "message": "plan assignment is not active",
            })
            return
        }
        if assignment.SubjectType == common.AssignmentSubjectTypeToken {
            if assignment.SubjectId != cleanToken.Id {
                c.JSON(http.StatusBadRequest, gin.H{
                    "success": false,
                    "message": "plan assignment does not belong to this token",
                })
                return
            }
        } else if assignment.SubjectType == common.AssignmentSubjectTypeUser {
            if assignment.SubjectId != cleanToken.UserId {
                c.JSON(http.StatusBadRequest, gin.H{
                    "success": false,
                    "message": "plan assignment does not belong to this user",
                })
                return
            }
        } else {
            c.JSON(http.StatusBadRequest, gin.H{
                "success": false,
                "message": "unsupported assignment subject type",
            })
            return
        }
    }

    if statusOnly != "" {
        cleanToken.Status = token.Status
    } else {
        // If you add more fields, please also update token.Update()
        cleanToken.Name = token.Name
        cleanToken.ExpiredTime = token.ExpiredTime
        cleanToken.RemainQuota = token.RemainQuota
        cleanToken.UnlimitedQuota = token.UnlimitedQuota
        cleanToken.ModelLimitsEnabled = token.ModelLimitsEnabled
        cleanToken.ModelLimits = token.ModelLimits
        cleanToken.AllowIps = token.AllowIps
        cleanToken.Group = token.Group
        if modeProvided || requestedMode != cleanToken.BillingMode {
            cleanToken.BillingMode = requestedMode
        }
        if requestedMode == common.BillingModePlan && hasAssignment {
            assignmentID := targetAssignmentID
            cleanToken.PlanAssignmentId = &assignmentID
        } else if planAssignmentProvided || (modeProvided && requestedMode != common.BillingModePlan) {
            cleanToken.PlanAssignmentId = nil
        }
    }
    err = cleanToken.Update()
    if err != nil {
        common.ApiError(c, err)
        return
    }
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "message": "",
        "data":    cleanToken,
    })
    return
}

type TokenBatch struct {
    Ids []int `json:"ids"`
}

func DeleteTokenBatch(c *gin.Context) {
    tokenBatch := TokenBatch{}
    if err := c.ShouldBindJSON(&tokenBatch); err != nil || len(tokenBatch.Ids) == 0 {
        c.JSON(http.StatusOK, gin.H{
            "success": false,
            "message": "参数错误",
        })
        return
    }
    userId := c.GetInt("id")
    count, err := model.BatchDeleteTokens(tokenBatch.Ids, userId)
    if err != nil {
        common.ApiError(c, err)
        return
    }
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "message": "",
        "data":    count,
    })
}
