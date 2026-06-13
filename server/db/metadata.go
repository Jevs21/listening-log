package db

import (
	"database/sql"

	"listening-log/server/spotify"
)

func InsertMetadata(db *sql.DB, track spotify.Track) error {
	primaryArtist := track.Artists[0]

	// 1. Insert artist (ignore if exists)
	if _, err := db.Exec(`
		INSERT OR IGNORE INTO artist (spotify_id, name)
		VALUES (?, ?)`,
		primaryArtist.ID, primaryArtist.Name,
	); err != nil {
		return err
	}

	// 2. Insert album (ignore if exists)
	if _, err := db.Exec(`
		INSERT OR IGNORE INTO album (spotify_id, name, album_type, total_tracks, release_date)
		VALUES (?, ?, ?, ?, ?)`,
		track.Album.ID, track.Album.Name, track.Album.AlbumType,
		track.Album.TotalTracks, track.Album.ReleaseDate,
	); err != nil {
		return err
	}

	// 3. Insert album images (ignore if exists via unique index)
	for _, img := range track.Album.Images {
		if _, err := db.Exec(`
			INSERT OR IGNORE INTO album_image (album_id, url, width, height)
			VALUES (?, ?, ?, ?)`,
			track.Album.ID, img.URL, img.Width, img.Height,
		); err != nil {
			return err
		}
	}

	// 4. Insert track (ignore if exists)
	explicit := 0
	if track.Explicit {
		explicit = 1
	}
	if _, err := db.Exec(`
		INSERT OR IGNORE INTO track (spotify_id, name, album_id, artist_id, duration_ms, track_number, disc_number, explicit, isrc)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		track.ID, track.Name, track.Album.ID, primaryArtist.ID,
		track.DurationMs, track.TrackNumber, track.DiscNumber,
		explicit, track.ExternalIDs.ISRC,
	); err != nil {
		return err
	}

	return nil
}
