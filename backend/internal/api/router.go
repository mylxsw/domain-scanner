package api

import (
	"github.com/gin-gonic/gin"
)

// SetupRouter sets up the API routes
func SetupRouter(handler *Handler) *gin.Engine {
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
	router.GET("/api/probe/:id", handler.GetProbeStatus)
	router.GET("/api/probe/:id/results", handler.GetProbeResults)
	router.GET("/api/probe/:id/stream", handler.StreamProbe)

	return router
}
