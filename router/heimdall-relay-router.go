package router

import (
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/controller"
	"github.com/QuantumNous/new-api/middleware"
	"github.com/QuantumNous/new-api/relay"
	"github.com/QuantumNous/new-api/types"

	"github.com/gin-gonic/gin"
)

// SetHeimdallRelayRouter sets up the relay router with Heimdall authentication
func SetHeimdallRelayRouter(router *gin.Engine) {
	// Initialize Heimdall configuration if not already done
	middleware.InitHeimdallConfig()
	
	// Apply global middleware
	router.Use(middleware.CORS())
	router.Use(middleware.DecompressRequestMiddleware())
	router.Use(middleware.StatsMiddleware())
	
	// Apply Heimdall authentication middleware if enabled
	if middleware.IsHeimdallEnabled() {
		router.Use(middleware.HeimdallAuth(middleware.GetHeimdallConfig()))
	} else {
		// Fall back to original TokenAuth for compatibility
		router.Use(middleware.TokenAuth())
	}
	
	// Models endpoints
	setupModelsRouter(router)
	
	// Playground endpoints
	setupPlaygroundRouter(router)
	
	// Main relay endpoints
	setupRelayV1Router(router)
	
	// Midjourney endpoints
	setupMidjourneyRouter(router)
	
	// Suno endpoints
	setupSunoRouter(router)
	
	// Gemini endpoints
	setupGeminiRouter(router)
}

// setupModelsRouter configures the models endpoints
func setupModelsRouter(router *gin.Engine) {
	// https://platform.openai.com/docs/api-reference/introduction
	modelsRouter := router.Group("/v1/models")
	{
		modelsRouter.GET("", func(c *gin.Context) {
			switch {
			case c.GetHeader("x-api-key") != "" && c.GetHeader("anthropic-version") != "":
				controller.ListModels(c, constant.ChannelTypeAnthropic)
			case c.GetHeader("x-goog-api-key") != "" || c.Query("key") != "":
				controller.RetrieveModel(c, constant.ChannelTypeGemini)
			default:
				controller.ListModels(c, constant.ChannelTypeOpenAI)
			}
		})

		modelsRouter.GET("/:model", func(c *gin.Context) {
			switch {
			case c.GetHeader("x-api-key") != "" && c.GetHeader("anthropic-version") != "":
				controller.RetrieveModel(c, constant.ChannelTypeAnthropic)
			default:
				controller.RetrieveModel(c, constant.ChannelTypeOpenAI)
			}
		})
	}

	geminiRouter := router.Group("/v1beta/models")
	{
		geminiRouter.GET("", func(c *gin.Context) {
			controller.ListModels(c, constant.ChannelTypeGemini)
		})
	}

	geminiCompatibleRouter := router.Group("/v1beta/openai/models")
	{
		geminiCompatibleRouter.GET("", func(c *gin.Context) {
			controller.ListModels(c, constant.ChannelTypeOpenAI)
		})
	}
}

// setupPlaygroundRouter configures the playground endpoints
func setupPlaygroundRouter(router *gin.Engine) {
	playgroundRouter := router.Group("/pg")
	playgroundRouter.Use(middleware.UserAuth(), middleware.Distribute(), middleware.Governance())
	{
		playgroundRouter.POST("/chat/completions", controller.Playground)
	}
}

// setupRelayV1Router configures the main v1 relay endpoints
func setupRelayV1Router(router *gin.Engine) {
	relayV1Router := router.Group("/v1")
	
	// Apply model rate limiting if Heimdall is not handling it
	if !middleware.IsHeimdallEnabled() {
		relayV1Router.Use(middleware.ModelRequestRateLimit())
	}
	
	{
		// WebSocket routes
		wsRouter := relayV1Router.Group("")
		wsRouter.Use(middleware.Distribute(), middleware.Governance())
		wsRouter.GET("/realtime", func(c *gin.Context) {
			controller.Relay(c, types.RelayFormatOpenAIRealtime)
		})
		
		// HTTP routes
		httpRouter := relayV1Router.Group("")
		httpRouter.Use(middleware.Distribute(), middleware.Governance())

		// Claude related routes
		httpRouter.POST("/messages", func(c *gin.Context) {
			controller.Relay(c, types.RelayFormatClaude)
		})

		// Chat related routes
		httpRouter.POST("/completions", func(c *gin.Context) {
			controller.Relay(c, types.RelayFormatOpenAI)
		})
		httpRouter.POST("/chat/completions", func(c *gin.Context) {
			controller.Relay(c, types.RelayFormatOpenAI)
		})

		// Response related routes
		httpRouter.POST("/responses", func(c *gin.Context) {
			controller.Relay(c, types.RelayFormatOpenAIResponses)
		})

		// Image related routes
		httpRouter.POST("/edits", func(c *gin.Context) {
			controller.Relay(c, types.RelayFormatOpenAIImage)
		})
		httpRouter.POST("/images/generations", func(c *gin.Context) {
			controller.Relay(c, types.RelayFormatOpenAIImage)
		})
		httpRouter.POST("/images/edits", func(c *gin.Context) {
			controller.Relay(c, types.RelayFormatOpenAIImage)
		})

		// Embedding related routes
		httpRouter.POST("/embeddings", func(c *gin.Context) {
			controller.Relay(c, types.RelayFormatEmbedding)
		})

		// Audio related routes
		httpRouter.POST("/audio/transcriptions", func(c *gin.Context) {
			controller.Relay(c, types.RelayFormatOpenAIAudio)
		})
		httpRouter.POST("/audio/translations", func(c *gin.Context) {
			controller.Relay(c, types.RelayFormatOpenAIAudio)
		})
		httpRouter.POST("/audio/speech", func(c *gin.Context) {
			controller.Relay(c, types.RelayFormatOpenAIAudio)
		})

		// Rerank related routes
		httpRouter.POST("/rerank", func(c *gin.Context) {
			controller.Relay(c, types.RelayFormatRerank)
		})

		// Gemini relay routes
		httpRouter.POST("/engines/:model/embeddings", func(c *gin.Context) {
			controller.Relay(c, types.RelayFormatGemini)
		})
		httpRouter.POST("/models/*path", func(c *gin.Context) {
			controller.Relay(c, types.RelayFormatGemini)
		})

		// Other relay routes
		httpRouter.POST("/moderations", func(c *gin.Context) {
			controller.Relay(c, types.RelayFormatOpenAI)
		})

		// Not implemented endpoints
		httpRouter.POST("/images/variations", controller.RelayNotImplemented)
		httpRouter.GET("/files", controller.RelayNotImplemented)
		httpRouter.POST("/files", controller.RelayNotImplemented)
		httpRouter.DELETE("/files/:id", controller.RelayNotImplemented)
		httpRouter.GET("/files/:id", controller.RelayNotImplemented)
		httpRouter.GET("/files/:id/content", controller.RelayNotImplemented)
		httpRouter.POST("/fine-tunes", controller.RelayNotImplemented)
		httpRouter.GET("/fine-tunes", controller.RelayNotImplemented)
		httpRouter.GET("/fine-tunes/:id", controller.RelayNotImplemented)
		httpRouter.POST("/fine-tunes/:id/cancel", controller.RelayNotImplemented)
		httpRouter.GET("/fine-tunes/:id/events", controller.RelayNotImplemented)
		httpRouter.DELETE("/models/:model", controller.RelayNotImplemented)
	}
}

// setupMidjourneyRouter configures the Midjourney endpoints
func setupMidjourneyRouter(router *gin.Engine) {
	relayMjRouter := router.Group("/mj")
	registerMjRouterGroup(relayMjRouter)

	relayMjModeRouter := router.Group("/:mode/mj")
	registerMjRouterGroup(relayMjModeRouter)
}

// setupSunoRouter configures the Suno endpoints
func setupSunoRouter(router *gin.Engine) {
	relaySunoRouter := router.Group("/suno")
	
	// Apply authentication if Heimdall is not handling it
	if !middleware.IsHeimdallEnabled() {
		relaySunoRouter.Use(middleware.TokenAuth())
	}
	
	relaySunoRouter.Use(middleware.Distribute(), middleware.Governance())
	{
		relaySunoRouter.POST("/submit/:action", controller.RelayTask)
		relaySunoRouter.POST("/fetch", controller.RelayTask)
		relaySunoRouter.GET("/fetch/:id", controller.RelayTask)
	}
}

// setupGeminiRouter configures the Gemini endpoints
func setupGeminiRouter(router *gin.Engine) {
	relayGeminiRouter := router.Group("/v1beta")
	
	// Apply authentication and rate limiting if Heimdall is not handling it
	if !middleware.IsHeimdallEnabled() {
		relayGeminiRouter.Use(middleware.TokenAuth())
		relayGeminiRouter.Use(middleware.ModelRequestRateLimit())
	}
	
	relayGeminiRouter.Use(middleware.Distribute(), middleware.Governance())
	{
		// Gemini API path format: /v1beta/models/{model_name}:{action}
		relayGeminiRouter.POST("/models/*path", func(c *gin.Context) {
			controller.Relay(c, types.RelayFormatGemini)
		})
	}
}

// registerMjRouterGroup registers Midjourney router group with authentication
func registerMjRouterGroup(relayMjRouter *gin.RouterGroup) {
	relayMjRouter.GET("/image/:id", relay.RelayMidjourneyImage)
	
	// Apply authentication if Heimdall is not handling it
	if !middleware.IsHeimdallEnabled() {
		relayMjRouter.Use(middleware.TokenAuth())
	}
	
	relayMjRouter.Use(middleware.Distribute(), middleware.Governance())
	{
		relayMjRouter.POST("/submit/action", controller.RelayMidjourney)
		relayMjRouter.POST("/submit/shorten", controller.RelayMidjourney)
		relayMjRouter.POST("/submit/modal", controller.RelayMidjourney)
		relayMjRouter.POST("/submit/imagine", controller.RelayMidjourney)
		relayMjRouter.POST("/submit/change", controller.RelayMidjourney)
		relayMjRouter.POST("/submit/simple-change", controller.RelayMidjourney)
		relayMjRouter.POST("/submit/describe", controller.RelayMidjourney)
		relayMjRouter.POST("/submit/blend", controller.RelayMidjourney)
		relayMjRouter.POST("/submit/edits", controller.RelayMidjourney)
		relayMjRouter.POST("/submit/video", controller.RelayMidjourney)
		relayMjRouter.POST("/notify", controller.RelayMidjourney)
		relayMjRouter.GET("/task/:id/fetch", relay.RelayMidjourney)
		relayMjRouter.GET("/task/:id/image-seed", relay.RelayMidjourney)
		relayMjRouter.POST("/task/list-by-condition", relay.RelayMidjourney)
		relayMjRouter.POST("/insight-face/swap", relay.RelayMidjourney)
		relayMjRouter.POST("/submit/upload-discord-images", relay.RelayMidjourney)
	}
}