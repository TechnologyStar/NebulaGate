package router

import (
    "embed"
    "fmt"
    "net/http"
    "os"
    "strings"

    "github.com/QuantumNous/new-api/common"
    "github.com/QuantumNous/new-api/middleware"

    "github.com/gin-gonic/gin"
)

func SetRouter(router *gin.Engine, buildFS embed.FS, indexPage []byte) {
    SetApiRouter(router)
    SetDashboardRouter(router)
    
    // Use Heimdall relay router if enabled, otherwise fall back to original
    if IsHeimdallEnabled() {
        SetHeimdallRelayRouter(router)
        common.SysLog("Using Heimdall enhanced relay router")
    } else {
        SetRelayRouter(router)
        common.SysLog("Using standard relay router")
    }
    
    SetVideoRouter(router)
    frontendBaseUrl := os.Getenv("FRONTEND_BASE_URL")
    if common.IsMasterNode && frontendBaseUrl != "" {
        frontendBaseUrl = ""
        common.SysLog("FRONTEND_BASE_URL is ignored on master node")
    }
    if frontendBaseUrl == "" {
        SetWebRouter(router, buildFS, indexPage)
    } else {
        frontendBaseUrl = strings.TrimSuffix(frontendBaseUrl, "/")
        router.NoRoute(func(c *gin.Context) {
            c.Redirect(http.StatusMovedPermanently, fmt.Sprintf("%s%s", frontendBaseUrl, c.Request.RequestURI))
        })
    }
}
