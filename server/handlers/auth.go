package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"strings"

	"listening-log/server/config"
	"listening-log/server/db"
	"listening-log/server/spotify"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	DB  *db.DB
	Cfg config.Config
}

func buildBaseURL(c *gin.Context) string {
	scheme := c.GetHeader("X-Forwarded-Proto")
	if scheme == "" {
		scheme = c.GetHeader("X-Forwarded-Scheme")
	}
	if scheme == "" {
		if c.Request.TLS != nil {
			scheme = "https"
		} else {
			scheme = "http"
		}
	}

	host := c.GetHeader("X-Forwarded-Host")
	if host == "" {
		host = c.Request.Host
	}

	return scheme + "://" + host
}

func (h *AuthHandler) Login(c *gin.Context) {
	state := randomState()
	baseURL := buildBaseURL(c)
	cookieValue := state + "|" + baseURL
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("oauth_state", cookieValue, 300, "/", "", false, true)

	redirectURI := baseURL + "/api/auth/callback"
	url := spotify.AuthorizeURL(h.Cfg.ClientID, redirectURI, state)
	c.Redirect(http.StatusFound, url)
}

func (h *AuthHandler) Callback(c *gin.Context) {
	// Recover state and base URL from cookie
	cookieValue, err := c.Cookie("oauth_state")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing oauth state"})
		return
	}

	parts := strings.SplitN(cookieValue, "|", 2)
	if len(parts) != 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid oauth state"})
		return
	}
	savedState := parts[0]
	baseURL := parts[1]

	// Verify state
	if savedState != c.Query("state") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "state mismatch"})
		return
	}

	// Clear the state cookie
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("oauth_state", "", -1, "/", "", false, true)

	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing code"})
		return
	}

	redirectURI := baseURL + "/api/auth/callback"
	token, err := spotify.ExchangeCode(h.Cfg.ClientID, h.Cfg.ClientSecret, code, redirectURI)
	if err != nil {
		log.Printf("token exchange error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token exchange failed"})
		return
	}

	profile, err := spotify.GetCurrentUser(token.AccessToken)
	if err != nil {
		log.Printf("get current user error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user profile"})
		return
	}
	log.Printf("spotify auth callback: user_id=%s", profile.ID)

	if h.Cfg.SpotifyAllowedUserID != "" && profile.ID != h.Cfg.SpotifyAllowedUserID {
		log.Printf("spotify auth rejected: user_id=%s not allowed", profile.ID)
		c.Redirect(http.StatusFound, baseURL+"/")
		return
	}

	err = h.DB.UpsertAuth(db.SpotifyAuth{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		Scope:        token.Scope,
		Expiry:       spotify.ExpiryFromNow(token.ExpiresIn),
	})
	if err != nil {
		log.Printf("db upsert error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save tokens"})
		return
	}

	c.Redirect(http.StatusFound, baseURL+"/")
}

func (h *AuthHandler) Status(c *gin.Context) {
	connected, err := h.DB.IsConnected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"connected": connected})
}

func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func randomState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
