package config

import "os"

type Config struct {
	ClientID             string
	ClientSecret         string
	Port                 string
	DatabaseURL          string
	SpotifyAllowedUserID string
}

func Load() Config {
	c := Config{
		ClientID:             os.Getenv("CLIENT_ID"),
		ClientSecret:         os.Getenv("CLIENT_SECRET"),
		Port:                 os.Getenv("PORT"),
		DatabaseURL:          os.Getenv("DATABASE_URL"),
		SpotifyAllowedUserID: os.Getenv("SPOTIFY_ALLOWED_USER_ID"),
	}

	if c.Port == "" {
		c.Port = "8080"
	}
	if c.DatabaseURL == "" {
		c.DatabaseURL = "postgres://listening_log:listening_log@localhost:5432/listening_log?sslmode=disable"
	}

	return c
}
