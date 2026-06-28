package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"listening-log/server/analysis"
	"listening-log/server/config"
	"listening-log/server/db"
	"listening-log/server/handlers"
	"listening-log/server/scraper"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	database, err := db.Open(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	// Start scraper with adaptive polling
	go func() {
		timer := time.NewTimer(0) // fire immediately on startup
		defer timer.Stop()
		for {
			<-timer.C
			active := scraper.Poll(database, cfg)
			if active {
				timer.Reset(scraper.PollIntervalActive)
			} else {
				timer.Reset(scraper.PollIntervalIdle)
			}
		}
	}()
	log.Println("scraper started — polling every 15s (active) / 30s (idle)")

	// Start listen analysis worker
	analysis.StartWorker(database, 5*time.Minute)
	log.Println("listen analysis worker started — running every 5m")

	auth := &handlers.AuthHandler{DB: database, Cfg: cfg}

	r := gin.Default()

	// API routes
	r.GET("/api/auth/login", auth.Login)
	r.GET("/api/auth/callback", auth.Callback)
	r.GET("/api/status", auth.Status)
	r.GET("/api/health", handlers.Health)
	r.GET("/api/now-playing", handlers.NowPlaying(database))
	r.GET("/api/image-grid", handlers.ImageGrid(database))
	r.POST("/api/suggestions", handlers.SubmitSuggestion(database))
	r.GET("/api/suggestions/check", handlers.CheckSuggestion(database))
	r.GET("/api/stats/dashboard", handlers.DashboardURL())
	r.Any("/metabase/*path", handlers.MetabaseProxy("http://metabase:3000"))

	// Serve built client in prod (if client/dist exists)
	// Check ../client/dist first (running from server/), then /client/dist (Docker)
	clientDist := filepath.Join("..", "client", "dist")
	if info, err := os.Stat(clientDist); err != nil || !info.IsDir() {
		clientDist = filepath.Join("/", "client", "dist")
	}
	if info, err := os.Stat(clientDist); err == nil && info.IsDir() {
		r.Use(spaMiddleware(clientDist))
	}

	log.Printf("server listening on :%s", cfg.Port)
	r.Run(":" + cfg.Port)
}

func spaMiddleware(staticDir string) gin.HandlerFunc {
	fs := http.Dir(staticDir)
	fileServer := http.FileServer(fs)

	return func(c *gin.Context) {
		// Skip API and Metabase proxy routes
		if strings.HasPrefix(c.Request.URL.Path, "/api") || strings.HasPrefix(c.Request.URL.Path, "/metabase") {
			c.Next()
			return
		}

		// Try to serve the file directly
		path := c.Request.URL.Path
		if f, err := fs.Open(path); err == nil {
			f.Close()
			fileServer.ServeHTTP(c.Writer, c.Request)
			c.Abort()
			return
		}

		// SPA fallback: serve index.html
		c.Request.URL.Path = "/"
		fileServer.ServeHTTP(c.Writer, c.Request)
		c.Abort()
	}
}
