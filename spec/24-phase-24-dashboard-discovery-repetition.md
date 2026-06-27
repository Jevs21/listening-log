# Phase 24 — Dashboard: Discovery & Repetition

Builds on phase 23 (overview & listening habits).

## Goal

Add discovery and repetition metrics to the Listening Overview dashboard — how much new music is being explored and how heavily tracks/artists/albums are replayed.

## Scope

### In scope

- Heading: "Discovery"
- Scalars: new artists (30d), new albums (30d), discovery score, avg music age, one-listen artists, one-listen albums, % new music (30d)
- Heading: "Repetition"
- Scalars: repeat ratio, comfort-zone score
- Tables: most replayed song (all time), most replayed song (this month), most replayed album, most replayed artist
- Table: play count milestones (songs at 10+, 25+, 50+, 100+ plays)

### Out of scope

- Genre-based discovery metrics
- Configurable time windows (hardcode 30 days for discovery metrics)

## Dashboard questions

All added to `metabase/dashboards/listening-overview.json`.

### Discovery scalars

```sql
-- ref: new-artists-30d (scalar)
-- Artists whose first listen was in the last 30 days
SELECT COUNT(DISTINCT t.artist_id) AS new_artists
FROM listen l JOIN track t ON t.spotify_id = l.track_id
WHERE t.artist_id NOT IN (
  SELECT DISTINCT t2.artist_id FROM listen l2
  JOIN track t2 ON t2.spotify_id = l2.track_id
  WHERE l2.started_at < CURRENT_DATE - INTERVAL '30 days'
)
AND l.started_at >= CURRENT_DATE - INTERVAL '30 days'

-- ref: new-albums-30d (scalar)
SELECT COUNT(DISTINCT t.album_id) AS new_albums
FROM listen l JOIN track t ON t.spotify_id = l.track_id
WHERE t.album_id NOT IN (
  SELECT DISTINCT t2.album_id FROM listen l2
  JOIN track t2 ON t2.spotify_id = l2.track_id
  WHERE l2.started_at < CURRENT_DATE - INTERVAL '30 days'
)
AND l.started_at >= CURRENT_DATE - INTERVAL '30 days'

-- ref: discovery-score (scalar)
-- New artists in last 30 days / total artists in last 30 days
WITH period AS (
  SELECT DISTINCT t.artist_id,
    MIN(l_all.started_at) AS first_listen
  FROM listen l
  JOIN track t ON t.spotify_id = l.track_id
  JOIN listen l_all ON l_all.track_id IN (
    SELECT spotify_id FROM track WHERE artist_id = t.artist_id
  )
  WHERE l.started_at >= CURRENT_DATE - INTERVAL '30 days'
  GROUP BY t.artist_id
)
SELECT ROUND(
  COUNT(*) FILTER (WHERE first_listen >= CURRENT_DATE - INTERVAL '30 days')::numeric
  / NULLIF(COUNT(*), 0) * 100, 1
) AS discovery_pct FROM period

-- ref: avg-music-age (scalar)
-- Average years since album release for tracks listened to
SELECT ROUND(AVG(
  EXTRACT(YEAR FROM AGE(CURRENT_DATE, al.release_date::date))
), 1) AS avg_years
FROM listen l
JOIN track t ON t.spotify_id = l.track_id
JOIN album al ON al.spotify_id = t.album_id
WHERE al.release_date IS NOT NULL AND al.release_date != ''

-- ref: pct-new-music-30d (scalar)
-- % of listens in last 30d on tracks first heard in last 30d
WITH first_listens AS (
  SELECT track_id, MIN(started_at) AS first_at FROM listen GROUP BY track_id
)
SELECT ROUND(
  COUNT(*) FILTER (WHERE fl.first_at >= CURRENT_DATE - INTERVAL '30 days')::numeric
  / NULLIF(COUNT(*), 0) * 100, 1
) AS pct
FROM listen l
JOIN first_listens fl ON fl.track_id = l.track_id
WHERE l.started_at >= CURRENT_DATE - INTERVAL '30 days'

-- ref: one-listen-artists (scalar)
WITH artist_plays AS (
  SELECT t.artist_id, COUNT(*) AS plays
  FROM listen l JOIN track t ON t.spotify_id = l.track_id
  GROUP BY t.artist_id
)
SELECT COUNT(*) AS artists FROM artist_plays WHERE plays = 1

-- ref: one-listen-albums (scalar)
WITH album_plays AS (
  SELECT t.album_id, COUNT(*) AS plays
  FROM listen l JOIN track t ON t.spotify_id = l.track_id
  GROUP BY t.album_id
)
SELECT COUNT(*) AS albums FROM album_plays WHERE plays = 1
```

### Repetition scalars

```sql
-- ref: repeat-ratio (scalar)
-- Total plays / unique songs
SELECT ROUND(COUNT(*)::numeric / NULLIF(COUNT(DISTINCT track_id), 0), 2) AS ratio
FROM listen

-- ref: comfort-zone-score (scalar)
-- % of total listens spent on top 20 artists
WITH artist_plays AS (
  SELECT t.artist_id, COUNT(*) AS plays
  FROM listen l JOIN track t ON t.spotify_id = l.track_id
  GROUP BY t.artist_id
  ORDER BY plays DESC LIMIT 20
)
SELECT ROUND(SUM(plays)::numeric / NULLIF((SELECT COUNT(*) FROM listen), 0) * 100, 1) AS pct
FROM artist_plays
```

### Repetition tables

```sql
-- ref: most-replayed-song-ever (table)
SELECT t.name AS song, a.name AS artist, COUNT(*) AS plays,
  ROUND(SUM(l.duration_ms) / 60000.0) AS minutes
FROM listen l
JOIN track t ON t.spotify_id = l.track_id
JOIN artist a ON a.spotify_id = t.artist_id
GROUP BY t.name, a.name ORDER BY plays DESC LIMIT 15

-- ref: most-replayed-song-month (table)
SELECT t.name AS song, a.name AS artist, COUNT(*) AS plays
FROM listen l
JOIN track t ON t.spotify_id = l.track_id
JOIN artist a ON a.spotify_id = t.artist_id
WHERE l.started_at >= DATE_TRUNC('month', CURRENT_DATE)
GROUP BY t.name, a.name ORDER BY plays DESC LIMIT 15

-- ref: most-replayed-album (table)
SELECT al.name AS album, a.name AS artist, COUNT(*) AS plays,
  ROUND(SUM(l.duration_ms) / 60000.0) AS minutes
FROM listen l
JOIN track t ON t.spotify_id = l.track_id
JOIN album al ON al.spotify_id = t.album_id
JOIN artist a ON a.spotify_id = t.artist_id
GROUP BY al.name, a.name ORDER BY plays DESC LIMIT 15

-- ref: most-replayed-artist (table)
SELECT a.name AS artist, COUNT(*) AS plays,
  COUNT(DISTINCT l.track_id) AS unique_songs,
  ROUND(SUM(l.duration_ms) / 60000.0) AS minutes
FROM listen l
JOIN track t ON t.spotify_id = l.track_id
JOIN artist a ON a.spotify_id = t.artist_id
GROUP BY a.name ORDER BY plays DESC LIMIT 15

-- ref: play-count-milestones (table)
WITH track_plays AS (
  SELECT track_id, COUNT(*) AS plays FROM listen GROUP BY track_id
)
SELECT
  COUNT(*) FILTER (WHERE plays >= 10) AS "10+ plays",
  COUNT(*) FILTER (WHERE plays >= 25) AS "25+ plays",
  COUNT(*) FILTER (WHERE plays >= 50) AS "50+ plays",
  COUNT(*) FILTER (WHERE plays >= 100) AS "100+ plays"
FROM track_plays
```

## Dashboard card layout

Starting at row 44 (after phase 23 ends at row 43).

Note: Metabase uses a 24-column grid.

| Row | Col | Size | Card |
|-----|-----|------|------|
| 44 | 0 | 24×1 | Heading: "Discovery" |
| 45 | 0 | 6×3 | new-artists-30d |
| 45 | 6 | 6×3 | new-albums-30d |
| 45 | 12 | 6×3 | discovery-score |
| 45 | 18 | 6×3 | avg-music-age |
| 48 | 0 | 8×3 | pct-new-music-30d |
| 48 | 8 | 8×3 | one-listen-artists |
| 48 | 16 | 8×3 | one-listen-albums |
| 51 | 0 | 24×1 | Heading: "Repetition" |
| 52 | 0 | 12×3 | repeat-ratio |
| 52 | 12 | 12×3 | comfort-zone-score |
| 55 | 0 | 12×8 | most-replayed-song-ever |
| 55 | 12 | 12×8 | most-replayed-song-month |
| 63 | 0 | 12×8 | most-replayed-album |
| 63 | 12 | 12×8 | most-replayed-artist |
| 71 | 0 | 24×4 | play-count-milestones |

## Definition of done

- [ ] Dashboard shows "Discovery" heading with 7 scalar cards
- [ ] Discovery score displays as a percentage
- [ ] Average music age displays in years
- [ ] Dashboard shows "Repetition" heading with repeat ratio and comfort-zone score scalars
- [ ] Four "most replayed" tables display correctly with song/artist names and play counts
- [ ] Play count milestones table shows counts for 10+, 25+, 50+, 100+ plays
