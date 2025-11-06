package middleware

import (
    "context"
    "crypto/tls"
    "crypto/x509"
    "encoding/json"
    "fmt"
    "net/http"
    "strings"
    "time"

    "github.com/QuantumNous/new-api/common"
    "github.com/QuantumNous/new-api/model"
    "github.com/gin-gonic/gin"
    "github.com/go-redis/redis/v8"
    "github.com/golang-jwt/jwt/v5"
)

// HeimdallConfig holds configuration for Heimdall authentication
type HeimdallConfig struct {
    // Authentication settings
    AuthEnabled          bool          `json:"auth_enabled"`
    APIKeyValidation     bool          `json:"api_key_validation"`
    JWTValidation        bool          `json:"jwt_validation"`
    MutualTLSValidation  bool          `json:"mutual_tls_validation"`
    JWTSecret           string        `json:"jwt_secret"`
    JWTSigningMethod    string        `json:"jwt_signing_method"`
    
    // Request validation settings
    SchemaValidation     bool          `json:"schema_validation"`
    ReplayProtection     bool          `json:"replay_protection"`
    ReplayWindow        time.Duration `json:"replay_window"`
    
    // Rate limiting settings
    RateLimitEnabled     bool          `json:"rate_limit_enabled"`
    PerKeyRateLimit      int           `json:"per_key_rate_limit"`
    PerIPRateLimit       int           `json:"per_ip_rate_limit"`
    RateLimitWindow      time.Duration `json:"rate_limit_window"`
    
    // Audit logging settings
    AuditLoggingEnabled  bool          `json:"audit_logging_enabled"`
    LogPayloadTruncate   bool          `json:"log_payload_truncate"`
    MaxPayloadSize       int           `json:"max_payload_size"`
}

// DefaultHeimdallConfig returns default configuration for Heimdall
func DefaultHeimdallConfig() HeimdallConfig {
    return HeimdallConfig{
        AuthEnabled:          true,
        APIKeyValidation:     true,
        JWTValidation:        false,
        MutualTLSValidation:  false,
        JWTSecret:           "",
        JWTSigningMethod:    "HS256",
        
        SchemaValidation:     true,
        ReplayProtection:     true,
        ReplayWindow:        5 * time.Minute,
        
        RateLimitEnabled:     true,
        PerKeyRateLimit:      100,
        PerIPRateLimit:       200,
        RateLimitWindow:      time.Minute,
        
        AuditLoggingEnabled:  true,
        LogPayloadTruncate:   true,
        MaxPayloadSize:       1024, // 1KB
    }
}

// HeimdallAuthContext holds authentication context information
type HeimdallAuthContext struct {
    RequestID      string                 `json:"request_id"`
    UserID         int                    `json:"user_id,omitempty"`
    TokenID        int                    `json:"token_id,omitempty"`
    ClientIP       string                 `json:"client_ip"`
    AuthMethod     string                 `json:"auth_method"`
    TLSInfo        *TLSInfo               `json:"tls_info,omitempty"`
    ValidationErrors []string             `json:"validation_errors,omitempty"`
    RateLimitStatus *RateLimitStatus      `json:"rate_limit_status,omitempty"`
    Timestamp      time.Time              `json:"timestamp"`
}

// TLSInfo holds TLS connection information
type TLSInfo struct {
    Version           uint16   `json:"version"`
    CipherSuite       uint16   `json:"cipher_suite"`
    ServerName        string   `json:"server_name"`
    PeerCertificates  []string `json:"peer_certificates,omitempty"`
}

// RateLimitStatus holds rate limiting information
type RateLimitStatus struct {
    LimitType    string    `json:"limit_type"`
    CurrentCount int       `json:"current_count"`
    Limit        int       `json:"limit"`
    WindowStart  time.Time `json:"window_start"`
    ResetTime    time.Time `json:"reset_time"`
}

// AuditLogEntry represents an audit log entry
type AuditLogEntry struct {
    RequestID       string                 `json:"request_id"`
    Timestamp       time.Time              `json:"timestamp"`
    Method          string                 `json:"method"`
    Path            string                 `json:"path"`
    ClientIP        string                 `json:"client_ip"`
    UserAgent       string                 `json:"user_agent"`
    AuthMethod      string                 `json:"auth_method"`
    UserID          *int                   `json:"user_id,omitempty"`
    TokenID         *int                   `json:"token_id,omitempty"`
    StatusCode      int                    `json:"status_code"`
    ResponseTime    time.Duration          `json:"response_time"`
    RequestSize     int64                  `json:"request_size"`
    ResponseSize    int64                  `json:"response_size"`
    TruncatedPayload *string               `json:"truncated_payload,omitempty"`
    ValidationErrors []string              `json:"validation_errors,omitempty"`
    RateLimitStatus *RateLimitStatus       `json:"rate_limit_status,omitempty"`
    TLSInfo         *TLSInfo               `json:"tls_info,omitempty"`
    Additional      map[string]interface{} `json:"additional,omitempty"`
}

// HeimdallAuth creates a new Heimdall authentication middleware
func HeimdallAuth(config HeimdallConfig) gin.HandlerFunc {
    return func(c *gin.Context) {
        startTime := time.Now()
        requestID := c.GetString(common.RequestIdKey)
        if requestID == "" {
            requestID = common.GetTimeString() + common.GetRandomString(8)
            c.Set(common.RequestIdKey, requestID)
        }

        authCtx := &HeimdallAuthContext{
            RequestID:  requestID,
            ClientIP:   c.ClientIP(),
            Timestamp:  startTime,
        }

        // Extract TLS information if available
        if c.Request.TLS != nil {
            authCtx.TLSInfo = extractTLSInfo(c.Request.TLS)
        }

        // Step 1: Authentication
        if config.AuthEnabled {
            if err := authenticateRequest(c, authCtx, config); err != nil {
                logAuditEntry(c, authCtx, config, startTime, http.StatusUnauthorized, err)
                c.JSON(http.StatusUnauthorized, gin.H{
                    "error": "authentication_failed",
                    "message": err.Error(),
                    "request_id": requestID,
                })
                c.Abort()
                return
            }
        }

        // Step 2: Request Validation
        if config.SchemaValidation || config.ReplayProtection {
            if err := validateRequest(c, authCtx, config); err != nil {
                logAuditEntry(c, authCtx, config, startTime, http.StatusBadRequest, err)
                c.JSON(http.StatusBadRequest, gin.H{
                    "error": "validation_failed",
                    "message": err.Error(),
                    "request_id": requestID,
                })
                c.Abort()
                return
            }
        }

        // Step 3: Rate Limiting
        if config.RateLimitEnabled {
            if err := enforceRateLimit(c, authCtx, config); err != nil {
                logAuditEntry(c, authCtx, config, startTime, http.StatusTooManyRequests, err)
                c.JSON(http.StatusTooManyRequests, gin.H{
                    "error": "rate_limit_exceeded",
                    "message": err.Error(),
                    "request_id": requestID,
                    "rate_limit_status": authCtx.RateLimitStatus,
                })
                c.Abort()
                return
            }
        }

        // Store auth context in gin context for later use
        c.Set("heimdall_auth_context", authCtx)

        // Continue with request processing
        c.Next()

        // Log audit entry after request completes
        logAuditEntry(c, authCtx, config, startTime, c.Writer.Status(), nil)
    }
}

// authenticateRequest handles the authentication logic
func authenticateRequest(c *gin.Context, authCtx *HeimdallAuthContext, config HeimdallConfig) error {
    authHeader := c.GetHeader("Authorization")
    
    // Try API key validation first
    if config.APIKeyValidation && strings.HasPrefix(authHeader, "Bearer ") {
        return authenticateWithAPIKey(c, authCtx, authHeader)
    }
    
    // Try JWT validation
    if config.JWTValidation && strings.HasPrefix(authHeader, "Bearer ") {
        return authenticateWithJWT(c, authCtx, authHeader, config)
    }
    
    // Try mutual TLS validation
    if config.MutualTLSValidation && c.Request.TLS != nil {
        return authenticateWithMTLS(c, authCtx, c.Request.TLS)
    }
    
    return fmt.Errorf("no valid authentication method provided")
}

// authenticateWithAPIKey validates API key against the database
func authenticateWithAPIKey(c *gin.Context, authCtx *HeimdallAuthContext, authHeader string) error {
    key := strings.TrimPrefix(authHeader, "Bearer ")
    key = strings.TrimPrefix(key, "sk-")
    
    token, err := model.ValidateUserToken(key)
    if err != nil {
        authCtx.AuthMethod = "api_key_failed"
        return fmt.Errorf("invalid API key: %w", err)
    }
    
    // Store token information in context
    authCtx.AuthMethod = "api_key"
    authCtx.UserID = token.UserId
    authCtx.TokenID = token.Id
    
    // Set gin context variables for compatibility with existing middleware
    c.Set("id", token.UserId)
    c.Set("token_id", token.Id)
    c.Set("token_key", token.Key)
    c.Set("token_name", token.Name)
    c.Set("token_unlimited_quota", token.UnlimitedQuota)
    if !token.UnlimitedQuota {
        c.Set("token_quota", token.RemainQuota)
    }
    
    return nil
}

// authenticateWithJWT validates JWT token
func authenticateWithJWT(c *gin.Context, authCtx *HeimdallAuthContext, authHeader string, config HeimdallConfig) error {
    tokenString := strings.TrimPrefix(authHeader, "Bearer ")
    
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return []byte(config.JWTSecret), nil
    })
    
    if err != nil {
        authCtx.AuthMethod = "jwt_failed"
        return fmt.Errorf("invalid JWT: %w", err)
    }
    
    if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
        authCtx.AuthMethod = "jwt"
        if userID, ok := claims["user_id"].(float64); ok {
            authCtx.UserID = int(userID)
            c.Set("id", int(userID))
        }
    } else {
        return fmt.Errorf("invalid JWT claims")
    }
    
    return nil
}

// authenticateWithMTLS validates mutual TLS certificates
func authenticateWithMTLS(c *gin.Context, authCtx *HeimdallAuthContext, tlsState *tls.ConnectionState) error {
    if len(tlsState.PeerCertificates) == 0 {
        return fmt.Errorf("no client certificate provided")
    }
    
    // Validate client certificate
    cert := tlsState.PeerCertificates[0]
    
    // Check certificate validity
    if time.Now().Before(cert.NotBefore) || time.Now().After(cert.NotAfter) {
        return fmt.Errorf("client certificate is not valid at this time")
    }
    
    // Extract user information from certificate (implementation depends on your cert structure)
    // This is a placeholder - you should implement based on your certificate format
    if subject := cert.Subject.CommonName; subject != "" {
        authCtx.AuthMethod = "mtls"
        // You might want to map CN to user ID or look up in database
    }
    
    return nil
}

// extractTLSInfo extracts TLS connection information
func extractTLSInfo(tlsState *tls.ConnectionState) *TLSInfo {
    info := &TLSInfo{
        Version:     tlsState.Version,
        CipherSuite: tlsState.CipherSuite,
        ServerName:  tlsState.ServerName,
    }
    
    if len(tlsState.PeerCertificates) > 0 {
        info.PeerCertificates = make([]string, len(tlsState.PeerCertificates))
        for i, cert := range tlsState.PeerCertificates {
            info.PeerCertificates[i] = cert.Subject.String()
        }
    }
    
    return info
}

// validateRequest handles request validation including schema validation and replay protection
func validateRequest(c *gin.Context, authCtx *HeimdallAuthContext, config HeimdallConfig) error {
    var errors []string
    
    // Replay protection
    if config.ReplayProtection {
        if err := checkReplayAttack(c, authCtx, config); err != nil {
            errors = append(errors, fmt.Sprintf("replay protection: %v", err))
        }
    }
    
    // Schema validation
    if config.SchemaValidation {
        if err := validateRequestSchema(c); err != nil {
            errors = append(errors, fmt.Sprintf("schema validation: %v", err))
        }
    }
    
    if len(errors) > 0 {
        authCtx.ValidationErrors = errors
        return fmt.Errorf("request validation failed: %s", strings.Join(errors, "; "))
    }
    
    return nil
}

// checkReplayAttack implements replay protection using Redis
func checkReplayAttack(c *gin.Context, authCtx *HeimdallAuthContext, config HeimdallConfig) error {
    if !common.RedisEnabled {
        // If Redis is not available, skip replay protection
        return nil
    }
    
    requestID := authCtx.RequestID
    key := fmt.Sprintf("heimdall:replay:%s", requestID)
    
    ctx := context.Background()
    
    // Check if request ID has been seen before
    exists, err := common.RDB.Exists(ctx, key).Result()
    if err != nil {
        return fmt.Errorf("failed to check replay protection: %w", err)
    }
    
    if exists > 0 {
        return fmt.Errorf("duplicate request ID detected: %s", requestID)
    }
    
    // Store request ID with TTL
    err = common.RDB.Set(ctx, key, time.Now().Unix(), config.ReplayWindow).Err()
    if err != nil {
        return fmt.Errorf("failed to store request ID for replay protection: %w", err)
    }
    
    return nil
}

// validateRequestSchema validates the request body against basic schema requirements
func validateRequestSchema(c *gin.Context) error {
    if c.Request.Method == "GET" || c.Request.Method == "DELETE" {
        // No body validation needed for these methods
        return nil
    }
    
    contentType := c.GetHeader("Content-Type")
    if !strings.Contains(contentType, "application/json") {
        // Only validate JSON requests
        return nil
    }
    
    var body map[string]interface{}
    if err := c.ShouldBindJSON(&body); err != nil {
        return fmt.Errorf("invalid JSON format: %w", err)
    }
    
    // Basic validation - ensure it's not empty
    if len(body) == 0 {
        return fmt.Errorf("request body cannot be empty")
    }
    
    // You can add more specific schema validation here based on your API requirements
    
    return nil
}

// enforceRateLimit implements rate limiting using token bucket algorithm
func enforceRateLimit(c *gin.Context, authCtx *HeimdallAuthContext, config HeimdallConfig) error {
    var errors []string
    
    // Per-key rate limiting
    if authCtx.TokenID > 0 {
        if err := enforceTokenBucketRateLimit(c, authCtx, config, fmt.Sprintf("heimdall:rate:token:%d", authCtx.TokenID), config.PerKeyRateLimit); err != nil {
            errors = append(errors, fmt.Sprintf("token rate limit: %v", err))
        }
    }
    
    // Per-IP rate limiting
    if err := enforceTokenBucketRateLimit(c, authCtx, config, fmt.Sprintf("heimdall:rate:ip:%s", authCtx.ClientIP), config.PerIPRateLimit); err != nil {
        errors = append(errors, fmt.Sprintf("IP rate limit: %v", err))
    }
    
    if len(errors) > 0 {
        return fmt.Errorf("rate limiting failed: %s", strings.Join(errors, "; "))
    }
    
    return nil
}

// enforceTokenBucketRateLimit implements token bucket rate limiting
func enforceTokenBucketRateLimit(c *gin.Context, authCtx *HeimdallAuthContext, config HeimdallConfig, key string, limit int) error {
    if !common.RedisEnabled {
        // If Redis is not available, use in-memory rate limiting
        return enforceInMemoryRateLimit(key, limit, config.RateLimitWindow)
    }
    
    ctx := context.Background()
    now := time.Now()
    windowStart := now.Truncate(config.RateLimitWindow)
    
    // Use a sorted set for sliding window rate limiting
    redisKey := fmt.Sprintf("%s:%d", key, windowStart.Unix())
    
    // Remove old entries
    pipe := common.RDB.Pipeline()
    pipe.ZRemRangeByScore(ctx, redisKey, "-inf", fmt.Sprintf("%d", now.Add(-config.RateLimitWindow).Unix()))
    
    // Count current requests
    currentCountCmd := pipe.ZCard(ctx, redisKey)
    
    // Add current request
    pipe.ZAdd(ctx, redisKey, redis.Z{
        Score:  float64(now.UnixNano()),
        Member: authCtx.RequestID,
    })
    
    // Set expiration
    pipe.Expire(ctx, redisKey, config.RateLimitWindow)
    
    _, err := pipe.Exec(ctx)
    if err != nil {
        return fmt.Errorf("rate limiting pipeline failed: %w", err)
    }
    
    currentCount, err := currentCountCmd.Result()
    if err != nil {
        return fmt.Errorf("failed to get current count: %w", err)
    }
    
    // Update rate limit status
    authCtx.RateLimitStatus = &RateLimitStatus{
        LimitType:    strings.Split(key, ":")[2], // token or ip
        CurrentCount: int(currentCount),
        Limit:        limit,
        WindowStart:  windowStart,
        ResetTime:    windowStart.Add(config.RateLimitWindow),
    }
    
    if currentCount >= int64(limit) {
        return fmt.Errorf("rate limit exceeded for %s: %d/%d", strings.Split(key, ":")[2], currentCount, limit)
    }
    
    return nil
}

// enforceInMemoryRateLimit provides fallback rate limiting when Redis is not available
func enforceInMemoryRateLimit(key string, limit int, window time.Duration) error {
    // This is a simple in-memory rate limiter
    // In production, you might want to use a more sophisticated solution
    if !common.MemoryCacheEnabled {
        return nil // Skip rate limiting if no caching is available
    }
    
    // For now, we'll just return nil to allow requests
    // You can implement a proper in-memory rate limiter here if needed
    return nil
}

// logAuditEntry creates and logs audit entries
func logAuditEntry(c *gin.Context, authCtx *HeimdallAuthContext, config HeimdallConfig, startTime time.Time, statusCode int, processingError error) {
    if !config.AuditLoggingEnabled {
        return
    }
    
    responseTime := time.Since(startTime)
    
    // Create audit log entry
    entry := AuditLogEntry{
        RequestID:    authCtx.RequestID,
        Timestamp:    time.Now(),
        Method:       c.Request.Method,
        Path:         c.Request.URL.Path,
        ClientIP:     authCtx.ClientIP,
        UserAgent:    c.GetHeader("User-Agent"),
        AuthMethod:   authCtx.AuthMethod,
        StatusCode:   statusCode,
        ResponseTime: responseTime,
        RequestSize:  c.Request.ContentLength,
        ResponseSize: int64(c.Writer.Size()),
        TLSInfo:      authCtx.TLSInfo,
    }
    
    // Add user and token information if available
    if authCtx.UserID > 0 {
        entry.UserID = &authCtx.UserID
    }
    if authCtx.TokenID > 0 {
        entry.TokenID = &authCtx.TokenID
    }
    
    // Add validation errors if any
    if len(authCtx.ValidationErrors) > 0 {
        entry.ValidationErrors = authCtx.ValidationErrors
    }
    
    // Add rate limit status if available
    if authCtx.RateLimitStatus != nil {
        entry.RateLimitStatus = authCtx.RateLimitStatus
    }
    
    // Add truncated payload if enabled
    if config.LogPayloadTruncate && statusCode >= 400 {
        payload := extractAndTruncatePayload(c, config.MaxPayloadSize)
        if payload != "" {
            entry.TruncatedPayload = &payload
        }
    }
    
    // Add processing error if any
    if processingError != nil {
        if entry.Additional == nil {
            entry.Additional = make(map[string]interface{})
        }
        entry.Additional["error"] = processingError.Error()
    }
    
    // Log the audit entry
    logAuditEntryToStorage(entry)
}

// extractAndTruncatePayload extracts and truncates request payload for logging
func extractAndTruncatePayload(c *gin.Context, maxSize int) string {
    if c.Request.Body == nil {
        return ""
    }
    
    // Read body (note: this will consume the body, so it should only be used for logging)
    bodyBytes, err := c.GetRawData()
    if err != nil {
        return fmt.Sprintf("Error reading body: %s", err.Error())
    }
    
    bodyStr := string(bodyBytes)
    if len(bodyStr) > maxSize {
        bodyStr = bodyStr[:maxSize] + "...[truncated]"
    }
    
    return bodyStr
}

// logAuditEntryToStorage logs the audit entry to appropriate storage
func logAuditEntryToStorage(entry AuditLogEntry) {
    // Convert to JSON
    jsonData, err := json.Marshal(entry)
    if err != nil {
        common.SysLog(fmt.Sprintf("Failed to marshal audit log entry: %v", err))
        return
    }
    
    // Log to system log
    common.SysLog(string(jsonData))
    
    // If Redis is available, also store there for querying
    if common.RedisEnabled {
        ctx := context.Background()
        key := fmt.Sprintf("heimdall:audit:%s", entry.RequestID)
        
        // Store with TTL (e.g., 30 days)
        ttl := 30 * 24 * time.Hour
        err := common.RDB.Set(ctx, key, string(jsonData), ttl).Err()
        if err != nil {
            common.SysLog(fmt.Sprintf("Failed to store audit log in Redis: %v", err))
        }
        
        // Also add to a time-series index for querying
        tsKey := fmt.Sprintf("heimdall:audit:ts:%d", entry.Timestamp.Unix())
        err = common.RDB.SAdd(ctx, tsKey, entry.RequestID).Err()
        if err != nil {
            common.SysLog(fmt.Sprintf("Failed to add to audit time-series: %v", err))
        }
        // Set TTL for time-series key
        common.RDB.Expire(ctx, tsKey, ttl)
    }
}