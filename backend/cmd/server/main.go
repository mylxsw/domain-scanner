package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/mylxsw/namecheap-domain-probe/backend/internal/api"
	"github.com/mylxsw/namecheap-domain-probe/backend/internal/service"
)

func main() {
	var (
		host   = flag.String("host", "127.0.0.1", "监听地址")
		port   = flag.Int("port", 8000, "监听端口")
		outdir = flag.String("outdir", "./out", "输出目录")
	)
	flag.Parse()

	// Set Gin mode
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.DebugMode)
	}

	// Load configuration
	cfg, err := service.FromEnv()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create output directory
	if err := os.MkdirAll(*outdir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Initialize services
	probeService := service.NewProbeService(*outdir)
	namecheapClient := service.NewNamecheapClient(cfg, 45.0, 30.0)
	tldService := service.NewTLDService(namecheapClient)

	// Create handler
	handler := api.NewHandler(probeService, tldService, cfg)

	// Setup router
	router := api.SetupRouter(handler)

	// Start server
	addr := fmt.Sprintf("%s:%d", *host, *port)
	fmt.Printf("Server starting on http://%s\n", addr)
	fmt.Printf("API endpoints:\n")
	fmt.Printf("  - GET  /api/health\n")
	fmt.Printf("  - GET  /api/tlds\n")
	fmt.Printf("  - POST /api/probe\n")
	fmt.Printf("  - GET  /api/probe/:id\n")
	fmt.Printf("  - GET  /api/probe/:id/results\n")
	fmt.Printf("  - GET  /api/probe/:id/stream\n")

	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
