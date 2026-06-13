package db

import "database/sql"

type SpotifyAuth struct {
	AccessToken  string
	RefreshToken string
	Scope        string
	Expiry       int64
}

func IsConnected(db *sql.DB) (bool, error) {
	var refreshToken string
	err := db.QueryRow("SELECT refresh_token FROM spotify_auth WHERE id = 1").Scan(&refreshToken)
	if err != nil {
		return false, err
	}
	return refreshToken != "", nil
}

func UpsertAuth(db *sql.DB, auth SpotifyAuth) error {
	// If no new refresh token provided, keep the existing one
	if auth.RefreshToken == "" {
		var existing string
		db.QueryRow("SELECT refresh_token FROM spotify_auth WHERE id = 1").Scan(&existing)
		auth.RefreshToken = existing
	}

	_, err := db.Exec(`
		UPDATE spotify_auth
		SET access_token = ?, refresh_token = ?, scope = ?, expiry = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = 1`,
		auth.AccessToken, auth.RefreshToken, auth.Scope, auth.Expiry,
	)
	return err
}
