package db

import (
	"encoding/json"
	"fmt"
)

type SpotifyAuth struct {
	AccessToken  string
	RefreshToken string
	Expiry       int64
}

// Temporary stub: Just prints the now playing data
func InsertPlaybackData(data map[string]interface{}) {
	fmt.Println("🎵 Now Playing Data (truncated):")

	pretty, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Println("Error formatting JSON:", err)
		return
	}

	if len(pretty) > 1000 {
		fmt.Println(string(pretty[:1000]) + "...")
	} else {
		fmt.Println(string(pretty))
	}
}

// Temporary stub: Return nil to simulate no credentials stored yet
func GetSpotifyCreds() *SpotifyAuth {
	// TODO: query Postgres and return actual creds
	return nil
}

// Temporary stub: Log credentials instead of storing
func SetSpotifyCreds(access, refresh string, expiry int64) bool {
	// TODO: insert into Postgres
	fmt.Println("💾 Saving Spotify credentials:")
	fmt.Printf("Access: %s\n", access)
	fmt.Printf("Refresh: %s\n", refresh)
	fmt.Printf("Expiry: %d\n", expiry)
	return true
}