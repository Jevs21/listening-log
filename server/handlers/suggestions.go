package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"strings"

	"listening-log/server/db"

	"github.com/gin-gonic/gin"
)

type suggestionRequest struct {
	Link    string `json:"link"`
	Message string `json:"message"`
}

func SubmitSuggestion(database *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req suggestionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		req.Link = strings.TrimSpace(req.Link)
		req.Message = strings.TrimSpace(req.Message)

		if req.Link == "" && req.Message == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "link or message is required"})
			return
		}
		if len(req.Link) > 2048 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "link is too long"})
			return
		}
		if len(req.Message) > 200 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "message is too long"})
			return
		}

		ip := c.ClientIP()

		count, err := db.CountRecentSuggestions(database, ip)
		if err != nil {
			log.Printf("suggestion rate-limit check error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
			return
		}
		if count >= 3 {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}

		if err := db.InsertSuggestion(database, req.Link, req.Message, ip); err != nil {
			log.Printf("suggestion insert error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"ok": true})
	}
}
