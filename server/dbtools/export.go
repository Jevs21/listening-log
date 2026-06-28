package dbtools

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type manifest struct {
	ExportedAt string         `json:"exported_at"`
	Tables     map[string]int `json:"tables"`
}

func Export(db *sql.DB, archivePath string) error {
	f, err := os.Create(archivePath)
	if err != nil {
		return fmt.Errorf("create archive: %w", err)
	}
	defer f.Close()

	gw := gzip.NewWriter(f)
	tw := tar.NewWriter(gw)

	m := manifest{
		ExportedAt: time.Now().UTC().Format(time.RFC3339),
		Tables:     make(map[string]int),
	}

	fmt.Println("Exporting tables...")
	for _, tbl := range tables {
		count, err := exportTable(db, tw, tbl)
		if err != nil {
			tw.Close()
			gw.Close()
			return fmt.Errorf("export %s: %w", tbl.name, err)
		}
		m.Tables[tbl.name] = count
		fmt.Printf("  %s: %d rows\n", tbl.name, count)
	}

	manifestJSON, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		tw.Close()
		gw.Close()
		return fmt.Errorf("marshal manifest: %w", err)
	}
	if err := writeToTar(tw, "manifest.json", manifestJSON); err != nil {
		tw.Close()
		gw.Close()
		return fmt.Errorf("write manifest: %w", err)
	}

	if err := tw.Close(); err != nil {
		gw.Close()
		return fmt.Errorf("close tar: %w", err)
	}
	if err := gw.Close(); err != nil {
		return fmt.Errorf("close gzip: %w", err)
	}

	fmt.Printf("Archive written to %s\n", archivePath)
	return nil
}

func exportTable(db *sql.DB, tw *tar.Writer, tbl tableDef) (int, error) {
	rows, err := db.Query(tbl.query)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	if err := w.Write(tbl.columns); err != nil {
		return 0, err
	}

	count := 0
	for rows.Next() {
		dests := makeScanDests(tbl.types)
		if err := rows.Scan(dests...); err != nil {
			return 0, fmt.Errorf("scan row %d: %w", count, err)
		}
		if err := w.Write(formatRow(dests, tbl.types)); err != nil {
			return 0, err
		}
		count++
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return 0, err
	}

	return count, writeToTar(tw, tbl.name+".csv", buf.Bytes())
}

func writeToTar(tw *tar.Writer, name string, data []byte) error {
	header := &tar.Header{
		Name: name,
		Size: int64(len(data)),
		Mode: 0644,
	}
	if err := tw.WriteHeader(header); err != nil {
		return err
	}
	_, err := tw.Write(data)
	return err
}
