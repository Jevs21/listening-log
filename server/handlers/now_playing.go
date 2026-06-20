package handlers

import (
	"log"
	"net/http"

	"listening-log/server/db"

	"github.com/gin-gonic/gin"
)

func NowPlaying(database *db.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		track, err := database.GetNowPlaying()
		if err != nil {
			log.Printf("now-playing query error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query now playing"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"track": track})
	}
}
