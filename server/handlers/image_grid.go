package handlers

import (
	"database/sql"
	"log"
	"net/http"

	"listening-log/server/db"

	"github.com/gin-gonic/gin"
)

func ImageGrid(database *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		mode := c.DefaultQuery("mode", "tracks")
		if mode != "tracks" && mode != "albums" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "mode must be 'tracks' or 'albums'"})
			return
		}

		images, err := db.GetImageGrid(database, mode, db.ImageGridMaxResults)
		if err != nil {
			log.Printf("image-grid query error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query image grid"})
			return
		}

		if images == nil {
			images = []db.ImageGridItem{}
		}

		c.JSON(http.StatusOK, gin.H{"images": images})
	}
}
