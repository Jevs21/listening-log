package analysis

import (
	"log"
	"time"

	"listening-log/server/db"
)

func StartWorker(database *db.DB, interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			if err := ProcessNewPolls(database); err != nil {
				log.Printf("listen analysis error: %v", err)
			}
		}
	}()
}
