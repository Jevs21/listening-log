# Phase 26 — Dashboard: Extended Rankings

Builds on phase 25 (skipping & completion).

## Goal

Add extended ranking tables to the Listening Overview dashboard — top artists, albums, and songs ranked by multiple dimensions beyond just play count.

## Scope

### In scope

- Heading: "Artist Rankings"
- Tables: top artists by plays, by minutes, by unique songs, by listening days
- Heading: "Album Rankings"
- Tables: top albums by plays, by completion rate, by minutes
- Heading: "Song Rankings"
- Tables: top songs by plays, by minutes, by avg completion %

### Out of scope

- Genre-based rankings
- Configurable time windows (all-time rankings)

## Dashboard questions

All added to `metabase/dashboards/listening-overview.json`.

```sql
-- ref: top-artists-by-plays (table)
SELECT a.name AS artist, COUNT(*) AS plays
FROM listen l
JOIN track t ON t.spotify_id = l.track_id
JOIN artist a ON a.spotify_id = t.artist_id
GROUP BY a.name ORDER BY plays DESC LIMIT 25

-- ref: top-artists-by-minutes (table)
SELECT a.name AS artist, ROUND(SUM(l.duration_ms) / 60000.0) AS minutes
FROM listen l
JOIN track t ON t.spotify_id = l.track_id
JOIN artist a ON a.spotify_id = t.artist_id
GROUP BY a.name ORDER BY minutes DESC LIMIT 25

-- ref: top-artists-by-unique-songs (table)
SELECT a.name AS artist, COUNT(DISTINCT l.track_id) AS unique_songs
FROM listen l
JOIN track t ON t.spotify_id = l.track_id
JOIN artist a ON a.spotify_id = t.artist_id
GROUP BY a.name ORDER BY unique_songs DESC LIMIT 25

-- ref: top-artists-by-listening-days (table)
SELECT a.name AS artist, COUNT(DISTINCT DATE(l.started_at)) AS days
FROM listen l
JOIN track t ON t.spotify_id = l.track_id
JOIN artist a ON a.spotify_id = t.artist_id
GROUP BY a.name ORDER BY days DESC LIMIT 25

-- ref: top-albums-by-plays (table)
SELECT al.name AS album, a.name AS artist, COUNT(*) AS plays
FROM listen l
JOIN track t ON t.spotify_id = l.track_id
JOIN album al ON al.spotify_id = t.album_id
JOIN artist a ON a.spotify_id = t.artist_id
GROUP BY al.name, a.name ORDER BY plays DESC LIMIT 25

-- ref: top-albums-by-completion (table)
SELECT al.name AS album, a.name AS artist, al.total_tracks,
  COUNT(DISTINCT l.track_id) AS tracks_heard,
  ROUND(COUNT(DISTINCT l.track_id)::numeric / NULLIF(al.total_tracks, 0) * 100, 1) AS completion_pct
FROM listen l
JOIN track t ON t.spotify_id = l.track_id
JOIN album al ON al.spotify_id = t.album_id
JOIN artist a ON a.spotify_id = t.artist_id
WHERE al.total_tracks > 1
GROUP BY al.name, a.name, al.total_tracks
ORDER BY completion_pct DESC, tracks_heard DESC LIMIT 25

-- ref: top-albums-by-minutes (table)
SELECT al.name AS album, a.name AS artist,
  ROUND(SUM(l.duration_ms) / 60000.0) AS minutes
FROM listen l
JOIN track t ON t.spotify_id = l.track_id
JOIN album al ON al.spotify_id = t.album_id
JOIN artist a ON a.spotify_id = t.artist_id
GROUP BY al.name, a.name ORDER BY minutes DESC LIMIT 25

-- ref: top-songs-by-plays (table)
SELECT t.name AS song, a.name AS artist, COUNT(*) AS plays
FROM listen l
JOIN track t ON t.spotify_id = l.track_id
JOIN artist a ON a.spotify_id = t.artist_id
GROUP BY t.name, a.name ORDER BY plays DESC LIMIT 25

-- ref: top-songs-by-minutes (table)
SELECT t.name AS song, a.name AS artist,
  ROUND(SUM(l.duration_ms) / 60000.0) AS minutes
FROM listen l
JOIN track t ON t.spotify_id = l.track_id
JOIN artist a ON a.spotify_id = t.artist_id
GROUP BY t.name, a.name ORDER BY minutes DESC LIMIT 25

-- ref: top-songs-by-completion (table)
SELECT t.name AS song, a.name AS artist,
  COUNT(*) AS plays,
  ROUND(AVG(l.progress_ms::numeric / NULLIF(l.duration_track_ms, 0) * 100), 1) AS avg_completion_pct
FROM listen l
JOIN track t ON t.spotify_id = l.track_id
JOIN artist a ON a.spotify_id = t.artist_id
GROUP BY t.name, a.name
HAVING COUNT(*) >= 3
ORDER BY avg_completion_pct DESC LIMIT 25
```

## Dashboard card layout

Starting at row 103 (after phase 25 ends at row 102).

Note: Metabase uses a 24-column grid.

| Row | Col | Size | Card |
|-----|-----|------|------|
| 103 | 0 | 24×1 | Heading: "Artist Rankings" |
| 104 | 0 | 12×10 | top-artists-by-plays |
| 104 | 12 | 12×10 | top-artists-by-minutes |
| 114 | 0 | 12×10 | top-artists-by-unique-songs |
| 114 | 12 | 12×10 | top-artists-by-listening-days |
| 124 | 0 | 24×1 | Heading: "Album Rankings" |
| 125 | 0 | 12×10 | top-albums-by-plays |
| 125 | 12 | 12×10 | top-albums-by-minutes |
| 135 | 0 | 24×10 | top-albums-by-completion |
| 145 | 0 | 24×1 | Heading: "Song Rankings" |
| 146 | 0 | 12×10 | top-songs-by-plays |
| 146 | 12 | 12×10 | top-songs-by-minutes |
| 156 | 0 | 24×10 | top-songs-by-completion |

## Definition of done

- [ ] Dashboard shows "Artist Rankings", "Album Rankings", and "Song Rankings" headings
- [ ] Each ranking table displays 25 entries
- [ ] Artist rankings show plays, minutes, unique songs, and listening days in separate tables
- [ ] Album rankings include completion percentage with total and heard track counts
- [ ] Song completion ranking requires minimum 3 plays to appear
