package config

import "os"

type Config struct {
	ClientID             string
	ClientSecret         string
	SpotifyRedirectURI   string
	ClientBaseURL        string
	Port                 string
	DatabaseURL          string
	SpotifyAllowedUserID string
	MetabaseURL          string
}

func Load() Config {
	c := Config{
		ClientID:             os.Getenv("CLIENT_ID"),
		ClientSecret:         os.Getenv("CLIENT_SECRET"),
		SpotifyRedirectURI:   os.Getenv("SPOTIFY_REDIRECT_URI"),
		ClientBaseURL:        os.Getenv("CLIENT_BASE_URL"),
		Port:                 os.Getenv("PORT"),
		DatabaseURL:          os.Getenv("DATABASE_URL"),
		SpotifyAllowedUserID: os.Getenv("SPOTIFY_ALLOWED_USER_ID"),
		MetabaseURL:          os.Getenv("METABASE_URL"),
	}

	if c.Port == "" {
		c.Port = "8080"
	}
	if c.DatabaseURL == "" {
		c.DatabaseURL = "postgres://listening_log:listening_log@localhost:5432/listening_log?sslmode=disable"
	}
	if c.SpotifyRedirectURI == "" {
		c.SpotifyRedirectURI = "http://127.0.0.1:8080/api/auth/callback"
	}
	if c.MetabaseURL == "" {
		c.MetabaseURL = "http://localhost:3000"
	}

	return c
}
