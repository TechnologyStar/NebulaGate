package main

import (
    "bytes"
    "context"
    "embed"
    "fmt"
    "log"
    "net/http"
    "os"
    "strconv"
    "strings"
    "time"

    "github.com/QuantumNous/new-api/common"
    "github.com/QuantumNous/new-api/constant"
    "github.com/QuantumNous/new-api/controller"
    "github.com/QuantumNous/new-api/logger"
    "github.com/QuantumNous/new-api/middleware"
    "github.com/QuantumNous/new-api/model"
    "github.com/QuantumNous/new-api/router"
    "github.com/QuantumNous/new-api/service"
    sched "github.com/QuantumNous/new-api/service/scheduler"
    "github.com/QuantumNous/new-api/setting/ratio_setting"

    "github.com/bytedance/gopkg/util/gopool"
    "github.com/gin-contrib/sessions"
    "github.com/gin-contrib/sessions/cookie"
    "github.com/gin-gonic/gin"
    "github.com/joho/godotenv"

    _ "net/http/pprof"
)

//go:embed web/dist
var buildFS embed.FS

//go:embed web/dist/index.html
var indexPage []byte

func main() {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("[FATAL PANIC] Application crashed during startup: %v", r)
            os.Exit(1)
        }
    }()
    
    startTime := time.Now()
    
    common.SysLog("=== Starting NebulaGate initialization ===")

    err := InitResources()
    if err != nil {
        common.FatalLog("failed to initialize resources: " + err.Error())
        return
    }

    common.SysLog("=== NebulaGate " + common.Version + " initialization completed ===")
    if os.Getenv("GIN_MODE") != "debug" {
        gin.SetMode(gin.ReleaseMode)
    }
    if common.DebugEnabled {
        common.SysLog("running in debug mode")
    }

    defer func() {
        err := model.CloseDB()
        if err != nil {
            common.FatalLog("failed to close database: " + err.Error())
        }
    }()

    if common.RedisEnabled {
        // for compatibility with old versions
        common.MemoryCacheEnabled = true
    }
    if common.MemoryCacheEnabled {
        common.SysLog("memory cache enabled")
        common.SysLog(fmt.Sprintf("sync frequency: %d seconds", common.SyncFrequency))

        // Add panic recovery and retry for InitChannelCache
        initSuccess := false
        maxRetries := 2
        
        for attempt := 0; attempt < maxRetries && !initSuccess; attempt++ {
            func() {
                defer func() {
                    if r := recover(); r != nil {
                        common.SysLog(fmt.Sprintf("InitChannelCache panic on attempt %d: %v", attempt+1, r))
                        if attempt == 0 {
                            // First failure: try to fix abilities
                            common.SysLog("attempting to fix abilities...")
                            successCount, failCount, fixErr := model.FixAbility()
                            if fixErr != nil {
                                common.SysLog(fmt.Sprintf("FixAbility failed: %s", fixErr.Error()))
                            } else {
                                common.SysLog(fmt.Sprintf("FixAbility completed: %d success, %d failed", successCount, failCount))
                            }
                        }
                    }
                }()
                model.InitChannelCache()
                initSuccess = true
                common.SysLog("channel cache initialized successfully")
            }()
        }

        if !initSuccess {
            common.SysLog("WARNING: failed to initialize channel cache after multiple attempts, continuing without cache")
            common.MemoryCacheEnabled = false
        } else {
            go model.SyncChannelCache(common.SyncFrequency)
        }
    }

    // 热更新配置
    go model.SyncOptions(common.SyncFrequency)

    // 数据看板
    go model.UpdateQuotaData()

    go func() {
        ticker := time.NewTicker(1 * time.Hour)
        defer ticker.Stop()
        for {
            if !common.IsMasterNode {
                <-ticker.C
                continue
            }
            count, err := model.UpdateExpiredPackages()
            if err != nil {
                common.SysLog(fmt.Sprintf("failed to update expired user packages: %v", err))
            } else if count > 0 {
                common.SysLog(fmt.Sprintf("marked %d user packages as expired", count))
            }
            <-ticker.C
        }
    }()

    if os.Getenv("CHANNEL_UPDATE_FREQUENCY") != "" {
        frequency, err := strconv.Atoi(os.Getenv("CHANNEL_UPDATE_FREQUENCY"))
        if err != nil {
            common.SysLog("WARNING: failed to parse CHANNEL_UPDATE_FREQUENCY: " + err.Error() + ", skipping automatic channel updates")
        } else {
            go controller.AutomaticallyUpdateChannels(frequency)
        }
    }

    go controller.AutomaticallyTestChannels()

    if common.IsMasterNode && constant.UpdateTask {
        gopool.Go(func() {
            controller.UpdateMidjourneyTaskBulk()
        })
        gopool.Go(func() {
            controller.UpdateTaskBulk()
        })
    }
    if os.Getenv("BATCH_UPDATE_ENABLED") == "true" {
        common.BatchUpdateEnabled = true
        common.SysLog("batch update enabled with interval " + strconv.Itoa(common.BatchUpdateInterval) + "s")
        model.InitBatchUpdater()
    }

    if os.Getenv("ENABLE_PPROF") == "true" {
        gopool.Go(func() {
            log.Println(http.ListenAndServe("0.0.0.0:8005", nil))
        })
        go common.Monitor()
        common.SysLog("pprof enabled")
    }

    // Initialize HTTP server
    server := gin.New()
    server.Use(gin.CustomRecovery(func(c *gin.Context, err any) {
        common.SysLog(fmt.Sprintf("panic detected: %v", err))
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": gin.H{
                "message": fmt.Sprintf("Panic detected, error: %v. Please submit a issue here: https://github.com/Calcium-Ion/new-api", err),
                "type":    "new_api_panic",
            },
        })
    }))
    // This will cause SSE not to work!!!
    //server.Use(gzip.Gzip(gzip.DefaultCompression))
    server.Use(middleware.RequestId())
    middleware.SetUpLogger(server)
    // Initialize session store
    store := cookie.NewStore([]byte(common.SessionSecret))
    store.Options(sessions.Options{
        Path:     "/",
        MaxAge:   2592000, // 30 days
        HttpOnly: true,
        Secure:   false,
        SameSite: http.SameSiteStrictMode,
    })
    server.Use(sessions.Sessions("session", store))

    InjectUmamiAnalytics()
    InjectGoogleAnalytics()

    // 设置路由
    router.SetRouter(server, buildFS, indexPage)
    var port = os.Getenv("PORT")
    if port == "" {
        port = strconv.Itoa(*common.Port)
    }

    // Log startup success message
    common.LogStartupSuccess(startTime, port)

    err = server.Run(":" + port)
    if err != nil {
        common.FatalLog("failed to start HTTP server: " + err.Error())
    }
}

func InjectUmamiAnalytics() {
    analyticsInjectBuilder := &strings.Builder{}
    if os.Getenv("UMAMI_WEBSITE_ID") != "" {
        umamiSiteID := os.Getenv("UMAMI_WEBSITE_ID")
        umamiScriptURL := os.Getenv("UMAMI_SCRIPT_URL")
        if umamiScriptURL == "" {
            umamiScriptURL = "https://analytics.umami.is/script.js"
        }
        analyticsInjectBuilder.WriteString("<script defer src=\"")
        analyticsInjectBuilder.WriteString(umamiScriptURL)
        analyticsInjectBuilder.WriteString("\" data-website-id=\"")
        analyticsInjectBuilder.WriteString(umamiSiteID)
        analyticsInjectBuilder.WriteString("\"></script>")
    }
    analyticsInject := analyticsInjectBuilder.String()
    indexPage = bytes.ReplaceAll(indexPage, []byte("<!--umami-->\n"), []byte(analyticsInject))
}

func InjectGoogleAnalytics() {
    analyticsInjectBuilder := &strings.Builder{}
    if os.Getenv("GOOGLE_ANALYTICS_ID") != "" {
        gaID := os.Getenv("GOOGLE_ANALYTICS_ID")
        // Google Analytics 4 (gtag.js)
        analyticsInjectBuilder.WriteString("<script async src=\"https://www.googletagmanager.com/gtag/js?id=")
        analyticsInjectBuilder.WriteString(gaID)
        analyticsInjectBuilder.WriteString("\"></script>")
        analyticsInjectBuilder.WriteString("<script>")
        analyticsInjectBuilder.WriteString("window.dataLayer = window.dataLayer || [];")
        analyticsInjectBuilder.WriteString("function gtag(){dataLayer.push(arguments);}")
        analyticsInjectBuilder.WriteString("gtag('js', new Date());")
        analyticsInjectBuilder.WriteString("gtag('config', '")
        analyticsInjectBuilder.WriteString(gaID)
        analyticsInjectBuilder.WriteString("');")
        analyticsInjectBuilder.WriteString("</script>")
    }
    analyticsInject := analyticsInjectBuilder.String()
    indexPage = bytes.ReplaceAll(indexPage, []byte("<!--Google Analytics-->\n"), []byte(analyticsInject))
}

func InitResources() error {
    common.SysLog("Step 1/10: Loading environment variables...")
    err := godotenv.Load(".env")
    if err != nil {
        if common.DebugEnabled {
            common.SysLog("No .env file found, using default environment variables. If needed, please create a .env file and set the relevant variables.")
        }
    }

    common.SysLog("Step 2/10: Initializing environment...")
    common.InitEnv()

    common.SysLog("Step 3/10: Setting up logger...")
    logger.SetupLogger()

    common.SysLog("Step 4/10: Initializing model settings...")
    ratio_setting.InitRatioSettings()

    common.SysLog("Step 5/10: Initializing HTTP client...")
    service.InitHttpClient()

    common.SysLog("Step 6/10: Initializing token encoders...")
    service.InitTokenEncoders()

    common.SysLog("Step 7/10: Initializing SQL database...")
    err = model.InitDB()
    if err != nil {
        common.SysLog("ERROR: failed to initialize database: " + err.Error())
        return err
    }

    common.SysLog("Step 8/10: Checking setup status...")
    model.CheckSetup()

    common.SysLog("Step 9/10: Initializing option map and pricing...")
    model.InitOptionMap()
    model.GetPricing()

    common.SysLog("Step 10/10: Initializing log database and Redis...")
    err = model.InitLogDB()
    if err != nil {
        common.SysLog("ERROR: failed to initialize log database: " + err.Error())
        return err
    }

    err = common.InitRedisClient()
    if err != nil {
        common.SysLog("ERROR: failed to initialize Redis: " + err.Error())
        return err
    }

    common.SysLog("Bootstrapping background scheduler...")
    _ = bootstrapScheduler()

    common.SysLog("Resource initialization completed successfully")
    return nil
}

var schedulerCancel context.CancelFunc

func bootstrapScheduler() context.CancelFunc {
    // Start background jobs only once
    if schedulerCancel != nil {
        return schedulerCancel
    }
    schedulerCancel = sched.Start()
    return schedulerCancel
}
