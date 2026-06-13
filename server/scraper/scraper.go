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

const (
	PollIntervalActive = 10 * time.Second
	PollIntervalIdle   = 45 * time.Second
)

func Poll(database *sql.DB, cfg config.Config) bool {
	auth, err := db.ReadAuth(database)
	if err != nil {
		log.Printf("scraper: error reading auth: %v", err)
		return false
	}

	if auth.RefreshToken == "" {
		return false
	}

	// Refresh token if expired or within 60s of expiry
	if time.Now().Unix() >= auth.Expiry-60 {
		token, err := spotify.RefreshToken(cfg.ClientID, cfg.ClientSecret, auth.RefreshToken)
		if err != nil {
			log.Printf("scraper: error refreshing token: %v", err)
			return false
		}
		expiry := spotify.ExpiryFromNow(token.ExpiresIn)
		if err := db.UpsertAuth(database, db.SpotifyAuth{
			AccessToken:  token.AccessToken,
			RefreshToken: token.RefreshToken,
			Expiry:       expiry,
			Scope:        auth.Scope,
		}); err != nil {
			log.Printf("scraper: error updating tokens: %v", err)
			return false
		}
		auth.AccessToken = token.AccessToken
	}

	ps, err := spotify.GetPlaybackState(auth.AccessToken)
	if err != nil {
		log.Printf("scraper: error fetching playback state: %v", err)
		return false
	}

	now := time.Now().Format("2006-01-02 15:04:05")

	// Nothing playing
	if ps == nil || ps.Item == nil {
		fmt.Printf("%s  ⏸ Nothing playing\n", now)
		return false
	}

	// Skip non-track types (episodes, ads)
	if ps.CurrentlyPlayingType != "track" {
		return true
	}

	// Skip local files
	if ps.Item.IsLocal {
		return true
	}

	// Insert metadata (artist, album, album images, track)
	if err := db.InsertMetadata(database, *ps.Item); err != nil {
		log.Printf("scraper: error inserting metadata: %v", err)
		return true
	}

	// Log to database
	var contextURI *string
	if ps.Context != nil {
		contextURI = &ps.Context.URI
	}

	if err := db.InsertPlaybackLog(database, db.PlaybackLog{
		TrackID:      ps.Item.ID,
		ProgressMs:   ps.ProgressMs,
		DurationMs:   ps.Item.DurationMs,
		IsPlaying:    ps.IsPlaying,
		Popularity:   ps.Item.Popularity,
		DeviceName:   ps.Device.Name,
		DeviceType:   ps.Device.Type,
		ShuffleState: ps.ShuffleState,
		RepeatState:  ps.RepeatState,
		ContextURI:   contextURI,
	}); err != nil {
		log.Printf("scraper: error inserting playback log: %v", err)
		return true
	}

	// Stdout logging
	artists := make([]string, len(ps.Item.Artists))
	for i, a := range ps.Item.Artists {
		artists[i] = a.Name
	}

	status := "♫"
	if !ps.IsPlaying {
		status = "⏸"
	}

	fmt.Printf("%s  %s \"%s\" — %s (%s)\n",
		now,
		status,
		ps.Item.Name,
		strings.Join(artists, ", "),
		ps.Item.Album.Name,
	)

	return true
}
