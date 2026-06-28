package handlers

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

func DashboardURL() gin.HandlerFunc {
	return func(c *gin.Context) {
		uuid, err := os.ReadFile("/shared/dashboard-uuid")
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "dashboard not available yet"})
			return
		}

		url := fmt.Sprintf("/metabase/public/dashboard/%s#theme=night&background=false&bordered=false&titled=false",
			strings.TrimSpace(string(uuid)),
		)

		c.JSON(http.StatusOK, gin.H{"url": url})
	}
}
