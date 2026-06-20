package db

import (
	"context"

	"listening-log/server/spotify"
)

func (d *DB) InsertMetadata(track spotify.Track) error {
	return insertMetadata(d, track)
}

func InsertMetadataTx(tx Executor, track spotify.Track) error {
	return insertMetadata(tx, track)
}

func insertMetadata(ex Executor, track spotify.Track) error {
	primaryArtist := track.Artists[0]

	// 1. Upsert artist (refresh updated_at on conflict)
	if _, err := ex.ExecContext(context.Background(), `
		INSERT INTO artist (spotify_id, name)
		VALUES ($1, $2)
		ON CONFLICT(spotify_id) DO UPDATE SET updated_at = CURRENT_TIMESTAMP`,
		primaryArtist.ID, primaryArtist.Name,
	); err != nil {
		return err
	}

	// 2. Upsert album (refresh updated_at on conflict)
	if _, err := ex.ExecContext(context.Background(), `
		INSERT INTO album (spotify_id, name, album_type, total_tracks, release_date)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT(spotify_id) DO UPDATE SET updated_at = CURRENT_TIMESTAMP`,
		track.Album.ID, track.Album.Name, track.Album.AlbumType,
		track.Album.TotalTracks, track.Album.ReleaseDate,
	); err != nil {
		return err
	}

	// 3. Insert album images (ignore if exists via unique index)
	for _, img := range track.Album.Images {
		if _, err := ex.ExecContext(context.Background(), `
			INSERT INTO album_image (album_id, url, width, height)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (album_id, url) DO NOTHING`,
			track.Album.ID, img.URL, img.Width, img.Height,
		); err != nil {
			return err
		}
	}

	// 4. Upsert track (refresh updated_at on conflict)
	if _, err := ex.ExecContext(context.Background(), `
		INSERT INTO track (spotify_id, name, album_id, artist_id, duration_ms, track_number, disc_number, explicit, isrc)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT(spotify_id) DO UPDATE SET updated_at = CURRENT_TIMESTAMP`,
		track.ID, track.Name, track.Album.ID, primaryArtist.ID,
		track.DurationMs, track.TrackNumber, track.DiscNumber,
		track.Explicit, track.ExternalIDs.ISRC,
	); err != nil {
		return err
	}

	return nil
}
