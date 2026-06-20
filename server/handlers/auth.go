package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"

	"listening-log/server/config"
	"listening-log/server/db"
	"listening-log/server/spotify"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	DB  *db.DB
	Cfg config.Config
}

func (h *AuthHandler) Login(c *gin.Context) {
	state := randomState()
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("oauth_state", state, 300, "/", "", false, true)
	url := spotify.AuthorizeURL(h.Cfg.ClientID, h.Cfg.SpotifyRedirectURI, state)
	c.Redirect(http.StatusFound, url)
}

func (h *AuthHandler) Callback(c *gin.Context) {
	// Verify state
	cookieState, err := c.Cookie("oauth_state")
	if err != nil || cookieState != c.Query("state") {
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

	token, err := spotify.ExchangeCode(h.Cfg.ClientID, h.Cfg.ClientSecret, code, h.Cfg.SpotifyRedirectURI)
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
		c.Redirect(http.StatusFound, redirectURL(h.Cfg.ClientBaseURL))
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

	c.Redirect(http.StatusFound, redirectURL(h.Cfg.ClientBaseURL))
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

func redirectURL(base string) string {
	if base == "" {
		return "/"
	}
	return base
}

func randomState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
