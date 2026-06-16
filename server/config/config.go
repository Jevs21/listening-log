package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ClientID           string
	ClientSecret       string
	SpotifyRedirectURI string
	ClientBaseURL      string
	Port               string
	DatabasePath         string
	SpotifyAllowedUserID string
}

func Load() Config {
	// Load .env from project root (one level up from server/)
	godotenv.Load("../.env")

	c := Config{
		ClientID:           os.Getenv("CLIENT_ID"),
		ClientSecret:       os.Getenv("CLIENT_SECRET"),
		SpotifyRedirectURI: os.Getenv("SPOTIFY_REDIRECT_URI"),
		ClientBaseURL:      os.Getenv("CLIENT_BASE_URL"),
		Port:               os.Getenv("PORT"),
		DatabasePath:         os.Getenv("DATABASE_PATH"),
		SpotifyAllowedUserID: os.Getenv("SPOTIFY_ALLOWED_USER_ID"),
	}

	if c.Port == "" {
		c.Port = "8080"
	}
	if c.DatabasePath == "" {
		c.DatabasePath = "../data/database.sqlite"
	}
	if c.SpotifyRedirectURI == "" {
		c.SpotifyRedirectURI = "http://127.0.0.1:8080/api/auth/callback"
	}
	if c.ClientBaseURL == "" {
		c.ClientBaseURL = "http://127.0.0.1:5173"
	}

	return c
}
