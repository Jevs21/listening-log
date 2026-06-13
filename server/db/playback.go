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
	isPlaying := 0
	if log.IsPlaying {
		isPlaying = 1
	}
	shuffle := 0
	if log.ShuffleState {
		shuffle = 1
	}

	_, err := db.Exec(`
		INSERT INTO playback_log (track_id, progress_ms, duration_ms, is_playing, popularity, device_name, device_type, shuffle_state, repeat_state, context_uri)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		log.TrackID, log.ProgressMs, log.DurationMs, isPlaying, log.Popularity,
		log.DeviceName, log.DeviceType, shuffle, log.RepeatState, log.ContextURI,
	)
	return err
}
