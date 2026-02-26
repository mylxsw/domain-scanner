package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/mylxsw/namecheap-domain-probe/backend/internal/api"
	"github.com/mylxsw/namecheap-domain-probe/backend/internal/service"
)

func main() {
	var (
		host   = flag.String("host", "127.0.0.1", "监听地址")
		port   = flag.Int("port", 8080, "监听端口")
		outdir = flag.String("outdir", "./out", "输出目录")
		webDir = flag.String("web-dir", "", "前端静态文件目录（包含 index.html）")
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
	probeService, err := service.NewProbeService(*outdir)
	if err != nil {
		log.Fatalf("Failed to initialize probe service: %v", err)
	}
	namecheapClient := service.NewNamecheapClient(cfg, 45.0, 30.0)
	tldService := service.NewTLDService(namecheapClient)

	// Create handler
	handler := api.NewHandler(probeService, tldService, cfg)

	// Setup router
	resolvedWebDir := resolveWebDir(*webDir)
	router := api.SetupRouter(handler, resolvedWebDir)

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
	if resolvedWebDir != "" {
		fmt.Printf("Frontend static hosting enabled: %s\n", resolvedWebDir)
	}

	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func resolveWebDir(custom string) string {
	candidates := []string{}
	if custom != "" {
		candidates = append(candidates, custom)
	}
	candidates = append(candidates, "./web", "./frontend/dist", "../frontend/dist")

	for _, d := range candidates {
		if d == "" {
			continue
		}
		abs, err := filepath.Abs(d)
		if err != nil {
			continue
		}
		indexFile := filepath.Join(abs, "index.html")
		if stat, err := os.Stat(indexFile); err == nil && !stat.IsDir() {
			return abs
		}
	}
	return ""
}
