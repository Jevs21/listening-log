package scraper

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"listening-log/server/config"
	"listening-log/server/db"
	"listening-log/server/spotify"
)

func Poll(database *sql.DB, cfg config.Config) {
	auth, err := db.ReadAuth(database)
	if err != nil {
		log.Printf("scraper: error reading auth: %v", err)
		return
	}

	if auth.RefreshToken == "" {
		return
	}

	// Refresh token if expired or within 60s of expiry
	if time.Now().Unix() >= auth.Expiry-60 {
		token, err := spotify.RefreshToken(cfg.ClientID, cfg.ClientSecret, auth.RefreshToken)
		if err != nil {
			log.Printf("scraper: error refreshing token: %v", err)
			return
		}
		expiry := spotify.ExpiryFromNow(token.ExpiresIn)
		if err := db.UpsertAuth(database, db.SpotifyAuth{
			AccessToken:  token.AccessToken,
			RefreshToken: token.RefreshToken,
			Expiry:       expiry,
			Scope:        auth.Scope,
		}); err != nil {
			log.Printf("scraper: error updating tokens: %v", err)
			return
		}
		auth.AccessToken = token.AccessToken
	}

	cp, err := spotify.GetCurrentlyPlaying(auth.AccessToken)
	if err != nil {
		log.Printf("scraper: error fetching currently playing: %v", err)
		return
	}

	now := time.Now().Format("2006-01-02 15:04:05")

	if cp == nil || cp.Item == nil || !cp.IsPlaying {
		fmt.Printf("%s  ⏸ Nothing playing\n", now)
		return
	}

	artists := make([]string, len(cp.Item.Artists))
	for i, a := range cp.Item.Artists {
		artists[i] = a.Name
	}

	fmt.Printf("%s  ♫ \"%s\" — %s (%s)\n",
		now,
		cp.Item.Name,
		strings.Join(artists, ", "),
		cp.Item.Album.Name,
	)
}
