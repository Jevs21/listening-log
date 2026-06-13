package db

import (
	"database/sql"
	"fmt"
)

const ImageGridMaxResults = 50

type ImageGridItem struct {
	URL       string `json:"url"`
	AlbumName string `json:"album_name"`
}

func GetImageGrid(database *sql.DB, mode string, limit int) ([]ImageGridItem, error) {
	var query string

	switch mode {
	case "albums":
		query = `
			SELECT ai.url, al.name
			FROM album al
			JOIN album_image ai ON ai.album_id = al.spotify_id
			WHERE ai.id = (
				SELECT ai2.id FROM album_image ai2
				WHERE ai2.album_id = al.spotify_id
				ORDER BY (ai2.width = 64) DESC, ai2.width ASC
				LIMIT 1
			)
			ORDER BY al.updated_at DESC
			LIMIT ?
		`
	case "tracks":
		query = `
			SELECT ai.url, al.name
			FROM track t
			JOIN album al ON al.spotify_id = t.album_id
			JOIN album_image ai ON ai.album_id = al.spotify_id
			WHERE ai.id = (
				SELECT ai2.id FROM album_image ai2
				WHERE ai2.album_id = al.spotify_id
				ORDER BY (ai2.width = 64) DESC, ai2.width ASC
				LIMIT 1
			)
			ORDER BY t.updated_at DESC
			LIMIT ?
		`
	default:
		return nil, fmt.Errorf("invalid mode: %s", mode)
	}

	rows, err := database.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []ImageGridItem
	for rows.Next() {
		var item ImageGridItem
		if err := rows.Scan(&item.URL, &item.AlbumName); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
