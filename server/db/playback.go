package db

import "database/sql"

type PlaybackLog struct {
	TrackID      string
	ProgressMs   int
	DurationMs   int
	IsPlaying    bool
	Popularity   int
	DeviceName   string
	DeviceType   string
	ShuffleState bool
	RepeatState  string
	ContextURI   *string
}

func InsertPlaybackLog(db *sql.DB, log PlaybackLog) error {
	_, err := db.Exec(`
		INSERT INTO playback_log (track_id, progress_ms, duration_ms, is_playing, popularity, device_name, device_type, shuffle_state, repeat_state, context_uri)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		log.TrackID, log.ProgressMs, log.DurationMs, log.IsPlaying, log.Popularity,
		log.DeviceName, log.DeviceType, log.ShuffleState, log.RepeatState, log.ContextURI,
	)
	return err
}
