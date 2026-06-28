package dbtools

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

var importOrder = []string{"artist", "album", "album_image", "track", "playback_log", "song_suggestion"}

func Import(db *sql.DB, archivePath string, mode string) error {
	files, err := extractTarGz(archivePath)
	if err != nil {
		return fmt.Errorf("extract archive: %w", err)
	}

	manifestData, ok := files["manifest.json"]
	if !ok {
		return fmt.Errorf("manifest.json not found in archive")
	}
	var m manifest
	if err := json.Unmarshal(manifestData, &m); err != nil {
		return fmt.Errorf("parse manifest: %w", err)
	}

	csvData := make(map[string][][]string)
	for _, tbl := range tables {
		data, ok := files[tbl.name+".csv"]
		if !ok {
			return fmt.Errorf("%s.csv not found in archive", tbl.name)
		}
		_, rows, err := readCSV(bytes.NewReader(data))
		if err != nil {
			return fmt.Errorf("read %s.csv: %w", tbl.name, err)
		}
		csvData[tbl.name] = rows
	}

	switch mode {
	case "overwrite":
		return importOverwrite(db, csvData)
	case "merge":
		return importMerge(db, csvData)
	default:
		return fmt.Errorf("unknown mode: %s", mode)
	}
}

func importOverwrite(db *sql.DB, csvData map[string][][]string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	fmt.Println("Truncating tables...")
	_, err = tx.Exec("TRUNCATE TABLE album_image, track, album, artist, playback_log, song_suggestion, listen, analysis_cursor")
	if err != nil {
		return fmt.Errorf("truncate: %w", err)
	}

	fmt.Println("Loading data...")
	for _, name := range importOrder {
		tbl := tableByName(name)
		rows := csvData[name]
		if err := batchInsert(tx, name, tbl.columns, tbl.types, rows); err != nil {
			return fmt.Errorf("insert %s: %w", name, err)
		}
		fmt.Printf("  %s: %d rows\n", name, len(rows))
	}

	fmt.Println("Resetting sequences...")
	for _, entry := range []struct{ table, col string }{
		{"playback_log", "id"},
		{"album_image", "id"},
		{"song_suggestion", "id"},
	} {
		_, err := tx.Exec(fmt.Sprintf(`
			SELECT setval(
				pg_get_serial_sequence('%s', '%s'),
				COALESCE((SELECT MAX(%s) FROM %s), 1),
				(SELECT MAX(%s) FROM %s) IS NOT NULL
			)`, entry.table, entry.col, entry.col, entry.table, entry.col, entry.table))
		if err != nil {
			return fmt.Errorf("reset sequence %s.%s: %w", entry.table, entry.col, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	fmt.Println("Import (overwrite) complete.")
	return nil
}

func importMerge(db *sql.DB, csvData map[string][][]string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	fmt.Println("Merging data...")
	for _, name := range importOrder {
		tbl := tableByName(name)
		rows := csvData[name]
		if len(rows) == 0 {
			fmt.Printf("  %s: 0 inserted, 0 updated, 0 skipped\n", name)
			continue
		}

		if _, err := tx.Exec(fmt.Sprintf("CREATE TEMP TABLE temp_%s (LIKE %s) ON COMMIT DROP", name, name)); err != nil {
			return fmt.Errorf("create temp_%s: %w", name, err)
		}

		if err := batchInsert(tx, "temp_"+name, tbl.columns, tbl.types, rows); err != nil {
			return fmt.Errorf("load temp_%s: %w", name, err)
		}

		inserted, updated, err := mergeTable(tx, name)
		if err != nil {
			return fmt.Errorf("merge %s: %w", name, err)
		}
		skipped := int64(len(rows)) - inserted - updated
		fmt.Printf("  %s: %d inserted, %d updated, %d skipped\n", name, inserted, updated, skipped)
	}

	fmt.Println("Resetting analysis state...")
	if err := resetAnalysis(tx); err != nil {
		return fmt.Errorf("reset analysis: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	fmt.Println("Import (merge) complete.")
	return nil
}

func mergeTable(tx *sql.Tx, name string) (inserted, updated int64, err error) {
	switch name {
	case "artist":
		return mergeWithUpdate(tx, name, "spotify_id",
			[]string{"name", "updated_at"},
			[]string{"spotify_id", "name", "created_at", "updated_at"})
	case "album":
		return mergeWithUpdate(tx, name, "spotify_id",
			[]string{"name", "album_type", "total_tracks", "release_date", "updated_at"},
			[]string{"spotify_id", "name", "album_type", "total_tracks", "release_date", "created_at", "updated_at"})
	case "track":
		return mergeWithUpdate(tx, name, "spotify_id",
			[]string{"name", "album_id", "artist_id", "duration_ms", "track_number", "disc_number", "explicit", "isrc", "updated_at"},
			[]string{"spotify_id", "name", "album_id", "artist_id", "duration_ms", "track_number", "disc_number", "explicit", "isrc", "created_at", "updated_at"})
	case "album_image":
		return mergeSkipDuplicates(tx, name,
			"m.album_id = t.album_id AND m.url = t.url",
			[]string{"album_id", "url", "width", "height"})
	case "playback_log":
		return mergeSkipDuplicates(tx, name,
			"m.polled_at = t.polled_at",
			[]string{"polled_at", "track_id", "progress_ms", "duration_ms", "is_playing", "popularity", "device_name", "device_type", "shuffle_state", "repeat_state", "context_uri"})
	case "song_suggestion":
		return mergeSkipDuplicates(tx, name,
			"m.ip_address = t.ip_address AND m.created_at = t.created_at",
			[]string{"link", "message", "source", "ip_address", "created_at"})
	default:
		return 0, 0, fmt.Errorf("unknown table: %s", name)
	}
}

func mergeWithUpdate(tx *sql.Tx, name, keyCol string, updateCols, insertCols []string) (int64, int64, error) {
	selectCols := prefixColumns("t", insertCols)
	insertSQL := fmt.Sprintf(
		"INSERT INTO %s (%s) SELECT %s FROM temp_%s t WHERE NOT EXISTS (SELECT 1 FROM %s WHERE %s.%s = t.%s)",
		name, strings.Join(insertCols, ", "), strings.Join(selectCols, ", "),
		name, name, name, keyCol, keyCol)

	res, err := tx.Exec(insertSQL)
	if err != nil {
		return 0, 0, fmt.Errorf("insert new: %w", err)
	}
	inserted, _ := res.RowsAffected()

	setClauses := make([]string, len(updateCols))
	for i, col := range updateCols {
		setClauses[i] = fmt.Sprintf("%s = t.%s", col, col)
	}
	updateSQL := fmt.Sprintf(
		"UPDATE %s SET %s FROM temp_%s t WHERE %s.%s = t.%s AND t.updated_at > %s.updated_at",
		name, strings.Join(setClauses, ", "), name, name, keyCol, keyCol, name)

	res, err = tx.Exec(updateSQL)
	if err != nil {
		return 0, 0, fmt.Errorf("update existing: %w", err)
	}
	updated, _ := res.RowsAffected()

	return inserted, updated, nil
}

func mergeSkipDuplicates(tx *sql.Tx, name, dedupCondition string, insertCols []string) (int64, int64, error) {
	selectCols := prefixColumns("t", insertCols)
	insertSQL := fmt.Sprintf(
		"INSERT INTO %s (%s) SELECT %s FROM temp_%s t WHERE NOT EXISTS (SELECT 1 FROM %s m WHERE %s)",
		name, strings.Join(insertCols, ", "), strings.Join(selectCols, ", "),
		name, name, dedupCondition)

	res, err := tx.Exec(insertSQL)
	if err != nil {
		return 0, 0, fmt.Errorf("insert: %w", err)
	}
	inserted, _ := res.RowsAffected()
	return inserted, 0, nil
}

func prefixColumns(prefix string, cols []string) []string {
	result := make([]string, len(cols))
	for i, col := range cols {
		result[i] = prefix + "." + col
	}
	return result
}

func extractTarGz(archivePath string) (map[string][]byte, error) {
	f, err := os.Open(archivePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	gr, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	defer gr.Close()

	tr := tar.NewReader(gr)
	files := make(map[string][]byte)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		data, err := io.ReadAll(tr)
		if err != nil {
			return nil, err
		}
		files[header.Name] = data
	}
	return files, nil
}
