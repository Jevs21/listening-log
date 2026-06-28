package dbtools

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

type colType int

const (
	tText      colType = iota // NOT NULL text
	tInt                      // NOT NULL integer
	tBool                     // NOT NULL boolean
	tTimestamp                // NOT NULL timestamp
	tNullText                 // nullable text
	tNullInt                  // nullable integer
	tNullBool                 // nullable boolean
)

type tableDef struct {
	name    string
	columns []string
	types   []colType
	query   string
}

var tables = []tableDef{
	{
		name:    "playback_log",
		columns: []string{"id", "polled_at", "track_id", "progress_ms", "duration_ms", "is_playing", "popularity", "device_name", "device_type", "shuffle_state", "repeat_state", "context_uri"},
		types:   []colType{tInt, tTimestamp, tText, tInt, tInt, tBool, tInt, tText, tText, tBool, tText, tNullText},
		query:   `SELECT id, polled_at, track_id, progress_ms, duration_ms, is_playing, popularity, device_name, device_type, shuffle_state, repeat_state, context_uri FROM playback_log ORDER BY id`,
	},
	{
		name:    "artist",
		columns: []string{"spotify_id", "name", "created_at", "updated_at"},
		types:   []colType{tText, tText, tTimestamp, tTimestamp},
		query:   `SELECT spotify_id, name, created_at, updated_at FROM artist ORDER BY spotify_id`,
	},
	{
		name:    "album",
		columns: []string{"spotify_id", "name", "album_type", "total_tracks", "release_date", "created_at", "updated_at"},
		types:   []colType{tText, tText, tNullText, tNullInt, tNullText, tTimestamp, tTimestamp},
		query:   `SELECT spotify_id, name, album_type, total_tracks, release_date, created_at, updated_at FROM album ORDER BY spotify_id`,
	},
	{
		name:    "album_image",
		columns: []string{"id", "album_id", "url", "width", "height"},
		types:   []colType{tInt, tText, tText, tNullInt, tNullInt},
		query:   `SELECT id, album_id, url, width, height FROM album_image ORDER BY id`,
	},
	{
		name:    "track",
		columns: []string{"spotify_id", "name", "album_id", "artist_id", "duration_ms", "track_number", "disc_number", "explicit", "isrc", "created_at", "updated_at"},
		types:   []colType{tText, tText, tText, tText, tInt, tNullInt, tNullInt, tNullBool, tNullText, tTimestamp, tTimestamp},
		query:   `SELECT spotify_id, name, album_id, artist_id, duration_ms, track_number, disc_number, explicit, isrc, created_at, updated_at FROM track ORDER BY spotify_id`,
	},
	{
		name:    "song_suggestion",
		columns: []string{"id", "link", "message", "source", "ip_address", "created_at"},
		types:   []colType{tInt, tText, tText, tText, tText, tTimestamp},
		query:   `SELECT id, link, message, source, ip_address, created_at FROM song_suggestion ORDER BY id`,
	},
}

func tableByName(name string) *tableDef {
	for i := range tables {
		if tables[i].name == name {
			return &tables[i]
		}
	}
	return nil
}

func makeScanDests(types []colType) []any {
	dests := make([]any, len(types))
	for i, t := range types {
		switch t {
		case tText:
			dests[i] = new(string)
		case tInt:
			dests[i] = new(int64)
		case tBool:
			dests[i] = new(bool)
		case tTimestamp:
			dests[i] = new(time.Time)
		case tNullText:
			dests[i] = new(sql.NullString)
		case tNullInt:
			dests[i] = new(sql.NullInt64)
		case tNullBool:
			dests[i] = new(sql.NullBool)
		}
	}
	return dests
}

func formatRow(dests []any, types []colType) []string {
	row := make([]string, len(dests))
	for i, d := range dests {
		switch types[i] {
		case tText:
			row[i] = *d.(*string)
		case tInt:
			row[i] = strconv.FormatInt(*d.(*int64), 10)
		case tBool:
			if *d.(*bool) {
				row[i] = "true"
			} else {
				row[i] = "false"
			}
		case tTimestamp:
			row[i] = d.(*time.Time).UTC().Format(time.RFC3339)
		case tNullText:
			if ns := d.(*sql.NullString); ns.Valid {
				row[i] = ns.String
			}
		case tNullInt:
			if ni := d.(*sql.NullInt64); ni.Valid {
				row[i] = strconv.FormatInt(ni.Int64, 10)
			}
		case tNullBool:
			if nb := d.(*sql.NullBool); nb.Valid {
				if nb.Bool {
					row[i] = "true"
				} else {
					row[i] = "false"
				}
			}
		}
	}
	return row
}

func parseRow(row []string, types []colType) ([]any, error) {
	if len(row) != len(types) {
		return nil, fmt.Errorf("expected %d columns, got %d", len(types), len(row))
	}
	values := make([]any, len(row))
	for i, s := range row {
		switch types[i] {
		case tText:
			values[i] = s
		case tInt:
			n, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("column %d: %w", i, err)
			}
			values[i] = n
		case tBool:
			values[i] = s == "true"
		case tTimestamp:
			t, err := time.Parse(time.RFC3339, s)
			if err != nil {
				return nil, fmt.Errorf("column %d: %w", i, err)
			}
			values[i] = t
		case tNullText:
			if s == "" {
				values[i] = nil
			} else {
				values[i] = s
			}
		case tNullInt:
			if s == "" {
				values[i] = nil
			} else {
				n, err := strconv.ParseInt(s, 10, 64)
				if err != nil {
					return nil, fmt.Errorf("column %d: %w", i, err)
				}
				values[i] = n
			}
		case tNullBool:
			if s == "" {
				values[i] = nil
			} else {
				values[i] = s == "true"
			}
		}
	}
	return values, nil
}

func readCSV(r io.Reader) ([]string, [][]string, error) {
	reader := csv.NewReader(r)
	header, err := reader.Read()
	if err != nil {
		return nil, nil, fmt.Errorf("reading header: %w", err)
	}
	var rows [][]string
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil, err
		}
		rows = append(rows, row)
	}
	return header, rows, nil
}

const batchSize = 500

func batchInsert(tx *sql.Tx, table string, columns []string, types []colType, csvRows [][]string) error {
	for i := 0; i < len(csvRows); i += batchSize {
		end := i + batchSize
		if end > len(csvRows) {
			end = len(csvRows)
		}
		batch := csvRows[i:end]

		query := buildInsertQuery(table, columns, len(batch))
		args := make([]any, 0, len(batch)*len(columns))
		for j, row := range batch {
			parsed, err := parseRow(row, types)
			if err != nil {
				return fmt.Errorf("row %d: %w", i+j, err)
			}
			args = append(args, parsed...)
		}

		if _, err := tx.Exec(query, args...); err != nil {
			return fmt.Errorf("batch starting at row %d: %w", i, err)
		}
	}
	return nil
}

func buildInsertQuery(table string, columns []string, numRows int) string {
	var b strings.Builder
	numCols := len(columns)
	b.WriteString("INSERT INTO ")
	b.WriteString(table)
	b.WriteString(" (")
	b.WriteString(strings.Join(columns, ", "))
	b.WriteString(") VALUES ")
	for i := 0; i < numRows; i++ {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteByte('(')
		for j := 0; j < numCols; j++ {
			if j > 0 {
				b.WriteString(", ")
			}
			b.WriteByte('$')
			b.WriteString(strconv.Itoa(i*numCols + j + 1))
		}
		b.WriteByte(')')
	}
	return b.String()
}
