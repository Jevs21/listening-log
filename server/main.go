package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"listening-log/server/config"
	"listening-log/server/db"
	"listening-log/server/handlers"
	"listening-log/server/scraper"

	"github.com/gin-gonic/gin"
	"github.com/go-co-op/gocron/v2"
)

func main() {
	cfg := config.Load()

	database, err := db.Open(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	// Start scraper on a 30s schedule
	s, err := gocron.NewScheduler()
	if err != nil {
		log.Fatalf("failed to create scheduler: %v", err)
	}
	s.NewJob(
		gocron.DurationJob(30*time.Second),
		gocron.NewTask(scraper.Poll, database, cfg),
	)
	s.Start()
	defer s.Shutdown()
	log.Println("scraper started — polling every 30s")

	auth := &handlers.AuthHandler{DB: database, Cfg: cfg}

	r := gin.Default()

	// API routes
	r.GET("/api/auth/login", auth.Login)
	r.GET("/api/auth/callback", auth.Callback)
	r.GET("/api/status", auth.Status)
	r.GET("/api/health", handlers.Health)
	r.GET("/api/now-playing", handlers.NowPlaying(database))

	// Serve built client in prod (if client/dist exists)
	clientDist := filepath.Join("..", "client", "dist")
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
		// Skip API routes
		if len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:4] == "/api" {
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
