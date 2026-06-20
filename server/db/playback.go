package db

import "context"

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

func (d *DB) InsertPlaybackLog(log PlaybackLog) error {
	return insertPlaybackLog(d, log)
}

func InsertPlaybackLogTx(tx Executor, log PlaybackLog) error {
	return insertPlaybackLog(tx, log)
}

func insertPlaybackLog(ex Executor, log PlaybackLog) error {
	_, err := ex.ExecContext(context.Background(), `
		INSERT INTO playback_log (track_id, progress_ms, duration_ms, is_playing, popularity, device_name, device_type, shuffle_state, repeat_state, context_uri)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		log.TrackID, log.ProgressMs, log.DurationMs, log.IsPlaying, log.Popularity,
		log.DeviceName, log.DeviceType, log.ShuffleState, log.RepeatState, log.ContextURI,
	)
	return err
}
