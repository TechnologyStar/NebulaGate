package main

import (
    "context"
    "crypto/tls"
    "fmt"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/QuantumNous/new-api/common"
    "github.com/QuantumNous/new-api/middleware"
    "github.com/QuantumNous/new-api/setting/heimdall"
    "github.com/gin-contrib/cors"
    "github.com/gin-gonic/gin"
    "golang.org/x/crypto/acme/autocert"
)

func main() {
    // Load configuration
    config, err := heimdall.LoadConfig()
    if err != nil {
        log.Fatalf("Failed to load configuration: %v", err)
    }

    // Setup logging based on configuration
    setupLogging(config)

    // Create Gin router
    router := setupRouter(config)

    // Create HTTP server
    server := &http.Server{
        Addr:    config.ListenAddr,
        Handler: router,
    }

    // Setup TLS
    if config.TLSEnabled {
        if config.ACMEEnabled {
            // Setup ACME (Let's Encrypt)
            setupACMEServer(server, config)
        } else {
            // Setup manual TLS
            server.TLSConfig = &tls.Config{
                MinVersion: tls.VersionTLS12,
            }
            server.TLSCertFile = config.TLSCertPath
            server.TLSKeyFile = config.TLSKeyPath
        }
    }

    // Start server in a goroutine
    go func() {
        if config.TLSEnabled {
            if config.ACMEEnabled {
                log.Printf("Starting Heimdall server with ACME TLS on %s", config.ListenAddr)
                if err := server.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
                    log.Fatalf("Failed to start server: %v", err)
                }
            } else {
                log.Printf("Starting Heimdall server with manual TLS on %s", config.ListenAddr)
                if err := server.ListenAndServeTLS(config.TLSCertPath, config.TLSKeyPath); err != nil && err != http.ErrServerClosed {
                    log.Fatalf("Failed to start server: %v", err)
                }
            }
        } else {
            log.Printf("Starting Heimdall server without TLS on %s (NOT RECOMMENDED FOR PRODUCTION)", config.ListenAddr)
            if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
                log.Fatalf("Failed to start server: %v", err)
            }
        }
    }()

    // Wait for interrupt signal to gracefully shutdown the server
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    log.Println("Shutting down server...")

    // The context is used to inform the server it has 5 seconds to finish
    // the request it is currently handling
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    if err := server.Shutdown(ctx); err != nil {
        log.Fatal("Server forced to shutdown:", err)
    }

    log.Println("Server exited")
}

func setupLogging(config *heimdall.Config) {
    // Set Gin mode based on log level
    if config.LogLevel == "debug" {
        gin.SetMode(gin.DebugMode)
    } else {
        gin.SetMode(gin.ReleaseMode)
    }

    // Setup common logger (reuse existing logging infrastructure)
    common.SetupLogger()
}

func setupRouter(config *heimdall.Config) *gin.Engine {
    // Create new Gin engine
    router := gin.New()

    // Add recovery middleware
    router.Use(gin.CustomRecovery(func(c *gin.Context, err any) {
        common.SysLog(fmt.Sprintf("panic detected: %v", err))
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": gin.H{
                "message": fmt.Sprintf("Internal server error: %v", err),
                "type":    "heimdall_panic",
            },
        })
    }))

    // Add request ID middleware (reuse existing)
    router.Use(middleware.RequestId())

    // Add CORS middleware
    corsConfig := cors.DefaultConfig()
    corsConfig.AllowOrigins = config.CORSOrigins
    corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
    corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"}
    corsConfig.AllowCredentials = true
    router.Use(cors.New(corsConfig))

    // Add logging middleware
    router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
        return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
            param.ClientIP,
            param.TimeStamp.Format(time.RFC3339),
            param.Method,
            param.Path,
            param.Request.Proto,
            param.StatusCode,
            param.Latency,
            param.Request.UserAgent(),
            param.ErrorMessage,
        )
    }))

    // Setup API routes
    setupAPIRoutes(router, config)

    // Root endpoint
    router.GET("/", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{
            "service": "Heimdall Gateway",
            "version": "1.0.0",
            "status": "running",
        })
    })

    return router
}

func setupAPIRoutes(router *gin.Engine, config *heimdall.Config) {
    // Create proxy handler
    proxy := NewProxyHandler(config)

    // API v1 group
    v1 := router.Group("/v1")
    {
        // Chat completions
        chat := v1.Group("/chat")
        {
            chat.POST("/completions", proxy.Handle("/v1/chat/completions"))
        }

        // Embeddings
        embeddings := v1.Group("/embeddings")
        {
            embeddings.POST("", proxy.Handle("/v1/embeddings"))
        }

        // Audio
        audio := v1.Group("/audio")
        {
            audio.POST("/transcriptions", proxy.Handle("/v1/audio/transcriptions"))
            audio.POST("/translations", proxy.Handle("/v1/audio/translations"))
            audio.POST("/speech", proxy.Handle("/v1/audio/speech"))
        }

        // Models
        models := v1.Group("/models")
        {
            models.GET("", proxy.Handle("/v1/models"))
        }

        // Moderations
        moderations := v1.Group("/moderations")
        {
            moderations.POST("", proxy.Handle("/v1/moderations"))
        }

        // Images
        images := v1.Group("/images")
        {
            images.POST("/generations", proxy.Handle("/v1/images/generations"))
            images.POST("/edits", proxy.Handle("/v1/images/edits"))
            images.POST("/variations", proxy.Handle("/v1/images/variations"))
        }

        // Files
        files := v1.Group("/files")
        {
            files.POST("", proxy.Handle("/v1/files"))
            files.GET("", proxy.Handle("/v1/files"))
            files.DELETE("/:file_id", proxy.Handle("/v1/files/:file_id"))
        }

        // Fine-tuning
        fineTuning := v1.Group("/fine_tuning")
        {
            fineTuning.POST("/jobs", proxy.Handle("/v1/fine_tuning/jobs"))
            fineTuning.GET("/jobs", proxy.Handle("/v1/fine_tuning/jobs"))
            fineTuning.GET("/jobs/:job_id", proxy.Handle("/v1/fine_tuning/jobs/:job_id"))
            fineTuning.POST("/jobs/:job_id/cancel", proxy.Handle("/v1/fine_tuning/jobs/:job_id/cancel"))
        }

        // Batch
        batch := v1.Group("/batch")
        {
            batch.POST("", proxy.Handle("/v1/batch"))
            batch.GET("/:batch_id", proxy.Handle("/v1/batch/:batch_id"))
            batch.POST("/:batch_id/cancel", proxy.Handle("/v1/batch/:batch_id/cancel"))
        }
    }

    // Enhanced endpoints using proxy handler
    router.GET("/health", proxy.HealthCheckHandler)
    router.GET("/metrics", proxy.MetricsHandler)
}

func setupACMEServer(server *http.Server, config *heimdall.Config) {
    // Setup ACME certificate manager
    certManager := &autocert.Manager{
        Prompt:     autocert.AcceptTOS,
        HostPolicy: autocert.HostWhitelist(config.ACMEDomain),
        Email:      config.ACMEEmail,
        Cache:      autocert.DirCache(config.ACMECacheDir),
    }

    // Configure server to use ACME
    server.TLSConfig = &tls.Config{
        GetCertificate: certManager.GetCertificate,
        MinVersion:     tls.VersionTLS12,
    }

    // Start HTTP server for ACME challenges
    go func() {
        log.Printf("Starting ACME HTTP server on :80 for challenges")
        if err := http.ListenAndServe(":80", certManager.HTTPHandler(nil)); err != nil && err != http.ErrServerClosed {
            log.Printf("ACME HTTP server error: %v", err)
        }
    }()
}