package db

import (
	"database/sql"
	"time"
)

type NowPlayingTrack struct {
	SpotifyID     string    `json:"spotify_id"`
	Name          string    `json:"name"`
	ArtistName    string    `json:"artist_name"`
	AlbumName     string    `json:"album_name"`
	DurationMs    int       `json:"duration_ms"`
	IsExplicit    bool      `json:"is_explicit"`
	UpdatedAt     time.Time `json:"updated_at"`
	AlbumImageURL *string   `json:"album_image_url"`
}

func GetNowPlaying(database *sql.DB) (*NowPlayingTrack, error) {
	row := database.QueryRow(`
		SELECT
			t.spotify_id,
			t.name,
			a.name  AS artist_name,
			al.name AS album_name,
			t.duration_ms,
			t.explicit,
			t.updated_at,
			(SELECT ai.url FROM album_image ai
			 WHERE ai.album_id = t.album_id
			 ORDER BY ABS(ai.width - 300) ASC
			 LIMIT 1) AS album_image_url
		FROM track t
		JOIN artist a  ON t.artist_id = a.spotify_id
		JOIN album  al ON t.album_id  = al.spotify_id
		ORDER BY t.updated_at DESC
		LIMIT 1
	`)

	var track NowPlayingTrack
	err := row.Scan(
		&track.SpotifyID,
		&track.Name,
		&track.ArtistName,
		&track.AlbumName,
		&track.DurationMs,
		&track.IsExplicit,
		&track.UpdatedAt,
		&track.AlbumImageURL,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &track, nil
}
