package handlers

import (
	"database/sql"
	"log"
	"net/http"

	"listening-log/server/db"

	"github.com/gin-gonic/gin"
)

func NowPlaying(database *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		track, err := db.GetNowPlaying(database)
		if err != nil {
			log.Printf("now-playing query error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query now playing"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"track": track})
	}
}
