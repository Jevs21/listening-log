package db

import "fmt"

const ImageGridMaxResults = 52

type ImageGridItem struct {
	URL        string `json:"url"`
	AlbumName  string `json:"album_name"`
	TrackName  string `json:"track_name,omitempty"`
	ArtistName string `json:"artist_name"`
	UpdatedAt  string `json:"updated_at"`
}

func (d *DB) GetImageGrid(mode string, limit int) ([]ImageGridItem, error) {
	var query string

	switch mode {
	case "albums":
		query = `
			SELECT ai.url, al.name, '' AS track_name,
				COALESCE(
					(SELECT a.name FROM track t2 JOIN artist a ON a.spotify_id = t2.artist_id
					 WHERE t2.album_id = al.spotify_id ORDER BY t2.updated_at DESC LIMIT 1),
					''
				) AS artist_name,
				al.updated_at
			FROM album al
			JOIN album_image ai ON ai.album_id = al.spotify_id
			WHERE ai.id = (
				SELECT ai2.id FROM album_image ai2
				WHERE ai2.album_id = al.spotify_id
				ORDER BY (ai2.width = 64) DESC, ai2.width ASC
				LIMIT 1
			)
			ORDER BY al.updated_at DESC
			LIMIT $1
		`
	case "tracks":
		query = `
			SELECT ai.url, al.name, t.name AS track_name, a.name AS artist_name, t.updated_at
			FROM track t
			JOIN album al ON al.spotify_id = t.album_id
			JOIN artist a ON a.spotify_id = t.artist_id
			JOIN album_image ai ON ai.album_id = al.spotify_id
			WHERE ai.id = (
				SELECT ai2.id FROM album_image ai2
				WHERE ai2.album_id = al.spotify_id
				ORDER BY (ai2.width = 64) DESC, ai2.width ASC
				LIMIT 1
			)
			ORDER BY t.updated_at DESC
			LIMIT $1
		`
	default:
		return nil, fmt.Errorf("invalid mode: %s", mode)
	}

	rows, err := d.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []ImageGridItem
	for rows.Next() {
		var item ImageGridItem
		if err := rows.Scan(&item.URL, &item.AlbumName, &item.TrackName, &item.ArtistName, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
