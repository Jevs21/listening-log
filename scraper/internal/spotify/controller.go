package spotify

import (
	// "context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"scraper/internal/db"
)

type SpotifyController struct {
	IsAuthenticated bool
	AuthCode        string
	AccessToken     string
	RefreshToken    string
	Expiry          int64
}

var (
	clientID     = os.Getenv("CLIENT_ID")
	clientSecret = os.Getenv("CLIENT_SECRET")
	redirectURI  = "http://localhost/setup"
	scope        = "user-read-currently-playing"
)

func NewSpotifyController() *SpotifyController {
	authData := db.GetSpotifyCreds()

	ctrl := &SpotifyController{}
	if authData != nil {
		ctrl.IsAuthenticated = true
		ctrl.AccessToken = authData.AccessToken
		ctrl.RefreshToken = authData.RefreshToken
		ctrl.Expiry = authData.Expiry
	}
	return ctrl
}

func (s *SpotifyController) GetAuthRedirectURL() string {
	params := url.Values{}
	params.Set("client_id", clientID)
	params.Set("response_type", "code")
	params.Set("redirect_uri", redirectURI)
	params.Set("scope", scope)
	return fmt.Sprintf("https://accounts.spotify.com/authorize?%s", params.Encode())
}

func (s *SpotifyController) Authenticate(refresh bool, code string) error {
	headers := map[string]string{"Content-Type": "application/x-www-form-urlencoded"}
	var body url.Values = url.Values{}

	if refresh && s.RefreshToken != "" {
		authHeader := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", clientID, clientSecret)))
		headers["Authorization"] = "Basic " + authHeader

		body.Set("grant_type", "refresh_token")
		body.Set("refresh_token", s.RefreshToken)
	} else {
		body.Set("grant_type", "authorization_code")
		body.Set("code", code)
		body.Set("redirect_uri", redirectURI)
		body.Set("client_id", clientID)
		body.Set("client_secret", clientSecret)
	}

	req, _ := http.NewRequest("POST", "https://accounts.spotify.com/api/token", strings.NewReader(body.Encode()))
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("Spotify auth failed: %s", resp.Status)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	s.AccessToken = result["access_token"].(string)
	s.RefreshToken = result["refresh_token"].(string)
	s.Expiry = time.Now().Unix() + int64(result["expires_in"].(float64))
	s.IsAuthenticated = true

	// Pseudocode: save to DB
	db.SetSpotifyCreds(s.AccessToken, s.RefreshToken, s.Expiry)

	return nil
}

func (s *SpotifyController) HasTokenExpired() bool {
	return time.Now().Unix() > s.Expiry
}

func (s *SpotifyController) GetNowPlaying() error {
	if !s.IsAuthenticated {
		return errors.New("Not authenticated — visit auth URL")
	}

	if s.HasTokenExpired() {
		log.Println("Refreshing Spotify token...")
		if err := s.Authenticate(true, ""); err != nil {
			return err
		}
	}

	req, _ := http.NewRequest("GET", "https://api.spotify.com/v1/me/player/currently-playing", nil)
	req.Header.Set("Authorization", "Bearer "+s.AccessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 204 {
		log.Println("No song currently playing")
		return nil
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("Spotify error: %s", resp.Status)
	}

	var data map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&data)

	// Pseudocode: save to DB
	db.InsertPlaybackData(data)

	return nil
}
