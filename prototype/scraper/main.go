package main

import (
	// "fmt"
	"io"
	"log"
	"net/http"
	// "os"
	"time"

	"scraper/internal/spotify"
)

const (
	intervalSeconds = 10 
	endpointURL     = "https://example.com"
)

func main() {
	log.Println("starting scraper service...")
	log.Println("initializing spotify")
	ctrl := spotify.NewSpotifyController()

	for {
		if err := ctrl.GetNowPlaying(); err != nil {
			authUrl := ctrl.GetAuthRedirectURL()
			log.Printf("Error fetching now playing: %v: %v", err, authUrl)
		}
		time.Sleep(10 * time.Second)
	}
}

func checkAndStore() {
	resp, err := http.Get(endpointURL)
	if err != nil {
		log.Printf("Failed to GET %s: %v\n", endpointURL, err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response: %v\n", err)
		return
	}

	// Pseudocode: store in DB
	storeInDatabase(time.Now(), resp.StatusCode, string(body))
}

func storeInDatabase(timestamp time.Time, statusCode int, responseBody string) {
	// Replace this pseudocode with actual DB logic
	log.Printf("Storing to DB: time=%s status=%d body_len=%d\n",
		timestamp.Format(time.RFC3339), statusCode, len(responseBody))
}
