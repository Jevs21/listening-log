package analysis

import (
	"database/sql"
	"log"
	"time"

	"listening-log/server/db"
)

const (
	gapThreshold = 120 * time.Second
	batchLimit   = 10000
)

func ProcessNewPolls(database *db.DB) error {
	cursor, err := database.GetAnalysisCursor("listen")
	if err != nil {
		return err
	}

	polls, err := database.GetUnprocessedPolls(cursor, batchLimit)
	if err != nil {
		return err
	}

	if len(polls) == 0 {
		return nil
	}

	listens, newCursor := resolveListens(polls)

	if len(listens) == 0 {
		return nil
	}

	return database.WithTx(func(tx *sql.Tx) error {
		for _, l := range listens {
			if err := db.InsertListen(tx, l); err != nil {
				return err
			}
		}
		return db.SetAnalysisCursor(tx, "listen", newCursor)
	})
}

type listenAccumulator struct {
	trackID      string
	startedAt    time.Time
	endedAt      time.Time
	maxProgress  int
	durationMs   int
	pollCount    int
	contextURI   *string
	deviceName   string
	lastPolledAt time.Time
	lastPollID   int64
}

func (a *listenAccumulator) toListen() db.Listen {
	wallClock := int(a.endedAt.Sub(a.startedAt).Milliseconds())
	skipped := float64(a.maxProgress) < 0.10*float64(a.durationMs)
	return db.Listen{
		TrackID:         a.trackID,
		StartedAt:       a.startedAt,
		EndedAt:         a.endedAt,
		DurationMs:      wallClock,
		ProgressMs:      a.maxProgress,
		DurationTrackMs: a.durationMs,
		PollCount:       a.pollCount,
		Skipped:         skipped,
		ContextURI:      a.contextURI,
		DeviceName:      a.deviceName,
	}
}

func resolveListens(polls []db.PlaybackLogRow) ([]db.Listen, int64) {
	var listens []db.Listen
	var current *listenAccumulator
	var lastCompletedCursor int64

	for _, p := range polls {
		isNew := current == nil ||
			p.TrackID != current.trackID ||
			p.PolledAt.Sub(current.lastPolledAt) > gapThreshold

		if isNew {
			if current != nil {
				listens = append(listens, current.toListen())
				lastCompletedCursor = current.lastPollID
			}
			current = &listenAccumulator{
				trackID:      p.TrackID,
				startedAt:    p.PolledAt,
				endedAt:      p.PolledAt,
				maxProgress:  p.ProgressMs,
				durationMs:   p.DurationMs,
				pollCount:    1,
				contextURI:   p.ContextURI,
				deviceName:   p.DeviceName,
				lastPolledAt: p.PolledAt,
				lastPollID:   p.ID,
			}
		} else {
			current.endedAt = p.PolledAt
			current.pollCount++
			if p.ProgressMs > current.maxProgress {
				current.maxProgress = p.ProgressMs
			}
			current.lastPolledAt = p.PolledAt
			current.lastPollID = p.ID
		}
	}

	// Do NOT close the last in-progress listen — it may continue in the next batch.
	// The cursor only advances past fully completed listens.

	if len(listens) == 0 {
		log.Printf("listen analysis: %d polls examined, 0 listens resolved (1 still in progress)", len(polls))
		return nil, lastCompletedCursor
	}

	log.Printf("listen analysis: %d polls examined, %d listens resolved", len(polls), len(listens))
	return listens, lastCompletedCursor
}
