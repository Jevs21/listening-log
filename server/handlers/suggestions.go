package handlers

import (
	"log"
	"net/http"
	"net/url"
	"strings"

	"listening-log/server/db"

	"github.com/gin-gonic/gin"
)

type suggestionRequest struct {
	Link    string `json:"link"`
	Message string `json:"message"`
	Source  string `json:"source"`
}

func CheckSuggestion(database *db.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		has, err := database.HasSuggested(ip)
		if err != nil {
			log.Printf("suggestion check error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"has_suggested": has})
	}
}

func SubmitSuggestion(database *db.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req suggestionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		req.Link = strings.TrimSpace(req.Link)
		req.Message = strings.TrimSpace(req.Message)

		if req.Link == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "link is required"})
			return
		}
		parsed, err := url.ParseRequestURI(req.Link)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "link must be a valid url"})
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

		count, err := database.CountRecentSuggestions(ip)
		if err != nil {
			log.Printf("suggestion rate-limit check error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
			return
		}
		if count >= 3 {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}

		if req.Source == "" {
			req.Source = "home"
		}
		if req.Source != "home" && req.Source != "gate" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid source"})
			return
		}

		if err := database.InsertSuggestion(req.Link, req.Message, req.Source, ip); err != nil {
			log.Printf("suggestion insert error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"ok": true})
	}
}
