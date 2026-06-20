package db

import (
	"context"
	"time"
)

type PlaybackLogRow struct {
	ID           int64
	PolledAt     time.Time
	TrackID      string
	ProgressMs   int
	DurationMs   int
	IsPlaying    bool
	DeviceName   string
	ContextURI   *string
}

type Listen struct {
	TrackID        string
	StartedAt      time.Time
	EndedAt        time.Time
	DurationMs     int
	ProgressMs     int
	DurationTrackMs int
	PollCount      int
	Skipped        bool
	ContextURI     *string
	DeviceName     string
}

func (d *DB) GetAnalysisCursor(jobName string) (int64, error) {
	var lastID int64
	err := d.QueryRow(
		"SELECT last_id FROM analysis_cursor WHERE job_name = $1", jobName,
	).Scan(&lastID)
	if err != nil {
		// No row means cursor is 0
		return 0, nil
	}
	return lastID, nil
}

func SetAnalysisCursor(ex Executor, jobName string, lastID int64) error {
	_, err := ex.ExecContext(context.Background(), `
		INSERT INTO analysis_cursor (job_name, last_id, updated_at)
		VALUES ($1, $2, CURRENT_TIMESTAMP)
		ON CONFLICT (job_name) DO UPDATE SET last_id = $2, updated_at = CURRENT_TIMESTAMP`,
		jobName, lastID,
	)
	return err
}

func InsertListen(ex Executor, l Listen) error {
	_, err := ex.ExecContext(context.Background(), `
		INSERT INTO listen (track_id, started_at, ended_at, duration_ms, progress_ms, duration_track_ms, poll_count, skipped, context_uri, device_name)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		l.TrackID, l.StartedAt, l.EndedAt, l.DurationMs, l.ProgressMs,
		l.DurationTrackMs, l.PollCount, l.Skipped, l.ContextURI, l.DeviceName,
	)
	return err
}

func (d *DB) GetUnprocessedPolls(afterID int64, limit int) ([]PlaybackLogRow, error) {
	rows, err := d.Query(`
		SELECT id, polled_at, track_id, progress_ms, duration_ms, is_playing, device_name, context_uri
		FROM playback_log
		WHERE id > $1
		ORDER BY id ASC
		LIMIT $2`,
		afterID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var polls []PlaybackLogRow
	for rows.Next() {
		var p PlaybackLogRow
		if err := rows.Scan(&p.ID, &p.PolledAt, &p.TrackID, &p.ProgressMs, &p.DurationMs, &p.IsPlaying, &p.DeviceName, &p.ContextURI); err != nil {
			return nil, err
		}
		polls = append(polls, p)
	}
	return polls, rows.Err()
}
