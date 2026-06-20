package db

type SpotifyAuth struct {
	AccessToken  string
	RefreshToken string
	Scope        string
	Expiry       int64
}

func (d *DB) ReadAuth() (*SpotifyAuth, error) {
	var auth SpotifyAuth
	err := d.QueryRow(
		"SELECT access_token, refresh_token, scope, expiry FROM spotify_auth WHERE id = 1",
	).Scan(&auth.AccessToken, &auth.RefreshToken, &auth.Scope, &auth.Expiry)
	if err != nil {
		return nil, err
	}
	return &auth, nil
}

func (d *DB) IsConnected() (bool, error) {
	var refreshToken string
	err := d.QueryRow("SELECT refresh_token FROM spotify_auth WHERE id = 1").Scan(&refreshToken)
	if err != nil {
		return false, err
	}
	return refreshToken != "", nil
}

func (d *DB) UpsertAuth(auth SpotifyAuth) error {
	// If no new refresh token provided, keep the existing one
	if auth.RefreshToken == "" {
		var existing string
		d.QueryRow("SELECT refresh_token FROM spotify_auth WHERE id = 1").Scan(&existing)
		auth.RefreshToken = existing
	}

	_, err := d.Exec(`
		UPDATE spotify_auth
		SET access_token = $1, refresh_token = $2, scope = $3, expiry = $4, updated_at = CURRENT_TIMESTAMP
		WHERE id = 1`,
		auth.AccessToken, auth.RefreshToken, auth.Scope, auth.Expiry,
	)
	return err
}
