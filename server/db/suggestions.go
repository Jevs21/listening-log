package db

import "database/sql"

func CountRecentSuggestions(database *sql.DB, ip string) (int, error) {
	var count int
	err := database.QueryRow(
		`SELECT COUNT(*) FROM song_suggestion WHERE ip_address = ? AND created_at > datetime('now', '-1 hour')`,
		ip,
	).Scan(&count)
	return count, err
}

func InsertSuggestion(database *sql.DB, link, message, ip string) error {
	_, err := database.Exec(
		`INSERT INTO song_suggestion (link, message, ip_address) VALUES (?, ?, ?)`,
		link, message, ip,
	)
	return err
}
