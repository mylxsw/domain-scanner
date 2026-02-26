package api

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mylxsw/namecheap-domain-probe/backend/internal/model"
)

// SetupRouter sets up the API routes
func SetupRouter(handler *Handler, webDir string) *gin.Engine {
	router := gin.New()

	// Use custom middleware
	router.Use(ErrorHandlerMiddleware())
	router.Use(LoggerMiddleware())
	router.Use(CORSMiddleware())

	// Health check
	router.GET("/api/health", handler.HealthCheck)

	// TLD routes
	router.GET("/api/tlds", handler.GetTLDs)

	// Probe routes
	router.POST("/api/probe", handler.CreateProbe)
	router.GET("/api/probe", handler.ListProbes)
	router.GET("/api/probe/:id", handler.GetProbeStatus)
	router.GET("/api/probe/:id/results", handler.GetProbeResults)
	router.GET("/api/probe/:id/stream", handler.StreamProbe)

	// Static frontend hosting (optional)
	if webDir != "" {
		indexFile := filepath.Join(webDir, "index.html")
		if stat, err := os.Stat(indexFile); err == nil && !stat.IsDir() {
			// Serve known static assets without using root wildcard route.
			router.GET("/", func(c *gin.Context) {
				c.File(indexFile)
			})
			router.GET("/favicon.ico", func(c *gin.Context) {
				c.File(filepath.Join(webDir, "favicon.ico"))
			})
			router.GET("/assets/*filepath", func(c *gin.Context) {
				fp := strings.TrimPrefix(c.Param("filepath"), "/")
				c.File(filepath.Join(webDir, "assets", fp))
			})

			router.NoRoute(func(c *gin.Context) {
				// Keep API 404 responses as JSON.
				if strings.HasPrefix(c.Request.URL.Path, "/api/") {
					c.JSON(http.StatusNotFound, model.ErrorResponse{
						Error:   "not_found",
						Message: "API endpoint not found",
					})
					return
				}

				// Try to serve direct static file first (e.g. /manifest.webmanifest).
				reqPath := strings.TrimPrefix(filepath.Clean(c.Request.URL.Path), "/")
				if reqPath != "" && reqPath != "." {
					fullPath := filepath.Join(webDir, reqPath)
					if fi, err := os.Stat(fullPath); err == nil && !fi.IsDir() {
						c.File(fullPath)
						return
					}
				}

				// SPA fallback
				c.File(indexFile)
			})
		}
	}

	return router
}
