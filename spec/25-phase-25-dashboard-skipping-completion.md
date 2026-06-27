# Phase 25 — Dashboard: Skipping & Completion

Builds on phase 24 (discovery & repetition).

## Goal

Add skip rate analysis and album/track completion metrics to the Listening Overview dashboard. These use the `listen.skipped`, `listen.progress_ms`, and `listen.duration_track_ms` fields from the analysis pipeline.

## Scope

### In scope

- Heading: "Skipping & Completion"
- Scalars: overall skip rate, % of tracks listened >90%
- Tables: most skipped songs, least skipped songs, artists with highest skip rate, songs frequently abandoned early
- Table: album completion rate

### Out of scope

- Modifying skip detection threshold (currently <10% progress in analysis worker)

## Dashboard questions

All added to `metabase/dashboards/listening-overview.json`.

```sql
-- ref: overall-skip-rate (scalar)
SELECT ROUND(COUNT(*) FILTER (WHERE skipped)::numeric / NULLIF(COUNT(*), 0) * 100, 1) AS skip_pct
FROM listen

-- ref: pct-tracks-over-90 (scalar)
-- % of listens where progress reached >90% of track duration
SELECT ROUND(
  COUNT(*) FILTER (WHERE progress_ms::numeric / NULLIF(duration_track_ms, 0) > 0.9)::numeric
  / NULLIF(COUNT(*), 0) * 100, 1
) AS pct
FROM listen

-- ref: most-skipped-songs (table)
SELECT t.name AS song, a.name AS artist,
  COUNT(*) AS total_plays,
  COUNT(*) FILTER (WHERE l.skipped) AS skips,
  ROUND(COUNT(*) FILTER (WHERE l.skipped)::numeric / NULLIF(COUNT(*), 0) * 100, 1) AS skip_pct
FROM listen l
JOIN track t ON t.spotify_id = l.track_id
JOIN artist a ON a.spotify_id = t.artist_id
GROUP BY t.name, a.name
HAVING COUNT(*) >= 3
ORDER BY skip_pct DESC LIMIT 15

-- ref: least-skipped-songs (table)
SELECT t.name AS song, a.name AS artist,
  COUNT(*) AS total_plays,
  COUNT(*) FILTER (WHERE l.skipped) AS skips,
  ROUND(COUNT(*) FILTER (WHERE l.skipped)::numeric / NULLIF(COUNT(*), 0) * 100, 1) AS skip_pct
FROM listen l
JOIN track t ON t.spotify_id = l.track_id
JOIN artist a ON a.spotify_id = t.artist_id
GROUP BY t.name, a.name
HAVING COUNT(*) >= 5
ORDER BY skip_pct ASC, total_plays DESC LIMIT 15

-- ref: artists-highest-skip-rate (table)
SELECT a.name AS artist,
  COUNT(*) AS total_plays,
  COUNT(*) FILTER (WHERE l.skipped) AS skips,
  ROUND(COUNT(*) FILTER (WHERE l.skipped)::numeric / NULLIF(COUNT(*), 0) * 100, 1) AS skip_pct
FROM listen l
JOIN track t ON t.spotify_id = l.track_id
JOIN artist a ON a.spotify_id = t.artist_id
GROUP BY a.name
HAVING COUNT(*) >= 5
ORDER BY skip_pct DESC LIMIT 15

-- ref: frequently-abandoned (table)
-- Songs where avg completion is low but not counted as "skipped" (>10% but <50%)
SELECT t.name AS song, a.name AS artist,
  COUNT(*) AS plays,
  ROUND(AVG(l.progress_ms::numeric / NULLIF(l.duration_track_ms, 0) * 100), 1) AS avg_completion_pct
FROM listen l
JOIN track t ON t.spotify_id = l.track_id
JOIN artist a ON a.spotify_id = t.artist_id
GROUP BY t.name, a.name
HAVING COUNT(*) >= 3
  AND AVG(l.progress_ms::numeric / NULLIF(l.duration_track_ms, 0)) < 0.5
ORDER BY avg_completion_pct ASC LIMIT 15

-- ref: album-completion-rate (table)
-- For each album, what % of its tracks have been listened to
SELECT al.name AS album, a.name AS artist,
  al.total_tracks,
  COUNT(DISTINCT l.track_id) AS tracks_heard,
  ROUND(COUNT(DISTINCT l.track_id)::numeric / NULLIF(al.total_tracks, 0) * 100, 1) AS completion_pct
FROM listen l
JOIN track t ON t.spotify_id = l.track_id
JOIN album al ON al.spotify_id = t.album_id
JOIN artist a ON a.spotify_id = t.artist_id
WHERE al.total_tracks > 1
GROUP BY al.name, a.name, al.total_tracks
ORDER BY completion_pct DESC, tracks_heard DESC LIMIT 20
```

## Dashboard card layout

Starting at row 75 (after phase 24 ends at row 74).

| Row | Col | Size | Card |
|-----|-----|------|------|
| 75 | 0 | 12×1 | Heading: "Skipping & Completion" |
| 76 | 0 | 6×3 | overall-skip-rate |
| 76 | 6 | 6×3 | pct-tracks-over-90 |
| 79 | 0 | 6×8 | most-skipped-songs |
| 79 | 6 | 6×8 | least-skipped-songs |
| 87 | 0 | 6×8 | artists-highest-skip-rate |
| 87 | 6 | 6×8 | frequently-abandoned |
| 95 | 0 | 12×8 | album-completion-rate |

## Definition of done

- [ ] Dashboard shows "Skipping & Completion" heading
- [ ] Skip rate and >90% completion scalars display as percentages
- [ ] Most/least skipped tables show skip percentages with minimum play thresholds
- [ ] Album completion table shows total tracks, tracks heard, and completion percentage
- [ ] Frequently abandoned table shows songs with low average completion
