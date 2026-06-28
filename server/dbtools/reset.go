package dbtools

import "database/sql"

func resetAnalysis(tx *sql.Tx) error {
	if _, err := tx.Exec("TRUNCATE TABLE listen"); err != nil {
		return err
	}
	_, err := tx.Exec(`
		INSERT INTO analysis_cursor (job_name, last_id, updated_at)
		VALUES ('listen', 0, NOW())
		ON CONFLICT (job_name) DO UPDATE SET last_id = 0, updated_at = NOW()`)
	return err
}
