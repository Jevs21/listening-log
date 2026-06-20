package db

import "database/sql"

func CountRecentSuggestions(database *sql.DB, ip string) (int, error) {
	var count int
	err := database.QueryRow(
		`SELECT COUNT(*) FROM song_suggestion WHERE ip_address = $1 AND created_at > NOW() - INTERVAL '1 hour'`,
		ip,
	).Scan(&count)
	return count, err
}

func HasSuggested(database *sql.DB, ip string) (bool, error) {
	var count int
	err := database.QueryRow(
		`SELECT COUNT(*) FROM song_suggestion WHERE ip_address = $1`,
		ip,
	).Scan(&count)
	return count > 0, err
}

func InsertSuggestion(database *sql.DB, link, message, source, ip string) error {
	_, err := database.Exec(
		`INSERT INTO song_suggestion (link, message, source, ip_address) VALUES ($1, $2, $3, $4)`,
		link, message, source, ip,
	)
	return err
}
