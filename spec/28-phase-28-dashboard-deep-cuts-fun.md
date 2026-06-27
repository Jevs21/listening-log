# Phase 28 — Dashboard: Deep Cuts & Fun Stats

Builds on phase 27 (time-based trends).

## Goal

Add the quirkier, personality-driven metrics to the Listening Overview dashboard — deep dives, fun stats, and the "wow, that's me" metrics from the original list.

## Scope

### In scope

- Heading: "Deep Cuts"
- Scalars: avg song length, total days of music, estimated TB streamed
- Tables: deepest artist dive, deepest album dive, longest/shortest songs listened, oldest/newest releases
- Heading: "Fun Stats"
- Tables: one-hit wonders, most loyal artists, most binged album, forgotten favorites, soundtrack of your life, obsession detector

### Out of scope

- Artist network graph (no related-artist data)
- GitHub-style heatmap (Metabase doesn't support heatmap grid natively — would need custom visualization)
- Musical drift score (complex to implement well; save for a standalone feature)
- Artist Retirement Fund (novelty — can add later)

## Dashboard questions

All added to `metabase/dashboards/listening-overview.json`.

### Deep Cuts

```sql
-- ref: avg-song-length (scalar, minutes)
SELECT ROUND(AVG(duration_ms) / 60000.0, 1) AS avg_minutes
FROM (SELECT DISTINCT track_id, duration_track_ms AS duration_ms FROM listen) s

-- ref: total-days-of-music (scalar)
SELECT ROUND(SUM(duration_ms) / 86400000.0, 2) AS days FROM listen

-- ref: est-tb-streamed (scalar)
-- Estimate at 160kbps average (mix of quality levels)
SELECT ROUND(SUM(duration_ms) / 1000.0 * 160 / 8 / 1024 / 1024 / 1024 / 1024, 3) AS tb
FROM listen

-- ref: deepest-artist-dive (table)
-- Artists with the most unique songs listened to
SELECT a.name AS artist,
  COUNT(DISTINCT l.track_id) AS unique_songs,
  COUNT(*) AS total_plays,
  ROUND(SUM(l.duration_ms) / 60000.0) AS minutes
FROM listen l
JOIN track t ON t.spotify_id = l.track_id
JOIN artist a ON a.spotify_id = t.artist_id
GROUP BY a.name ORDER BY unique_songs DESC LIMIT 15

-- ref: deepest-album-dive (table)
-- Albums with the highest track completion (most tracks heard / total tracks)
SELECT al.name AS album, a.name AS artist,
  al.total_tracks,
  COUNT(DISTINCT l.track_id) AS tracks_heard,
  COUNT(*) AS total_plays
FROM listen l
JOIN track t ON t.spotify_id = l.track_id
JOIN album al ON al.spotify_id = t.album_id
JOIN artist a ON a.spotify_id = t.artist_id
WHERE al.total_tracks > 3
GROUP BY al.name, a.name, al.total_tracks
ORDER BY tracks_heard DESC, total_plays DESC LIMIT 15

-- ref: longest-songs-listened (table)
SELECT DISTINCT ON (t.spotify_id)
  t.name AS song, a.name AS artist,
  ROUND(t.duration_ms / 60000.0, 1) AS length_min
FROM listen l
JOIN track t ON t.spotify_id = l.track_id
JOIN artist a ON a.spotify_id = t.artist_id
ORDER BY t.spotify_id, t.duration_ms DESC
LIMIT 10

-- ref: shortest-songs-listened (table)
SELECT DISTINCT ON (t.spotify_id)
  t.name AS song, a.name AS artist,
  ROUND(t.duration_ms / 1000.0) AS length_sec
FROM listen l
JOIN track t ON t.spotify_id = l.track_id
JOIN artist a ON a.spotify_id = t.artist_id
ORDER BY t.spotify_id, t.duration_ms ASC
LIMIT 10

-- ref: oldest-releases (table)
SELECT al.name AS album, a.name AS artist, al.release_date
FROM listen l
JOIN track t ON t.spotify_id = l.track_id
JOIN album al ON al.spotify_id = t.album_id
JOIN artist a ON a.spotify_id = t.artist_id
WHERE al.release_date IS NOT NULL AND al.release_date != ''
GROUP BY al.name, a.name, al.release_date
ORDER BY al.release_date ASC LIMIT 10

-- ref: newest-releases (table)
SELECT al.name AS album, a.name AS artist, al.release_date
FROM listen l
JOIN track t ON t.spotify_id = l.track_id
JOIN album al ON al.spotify_id = t.album_id
JOIN artist a ON a.spotify_id = t.artist_id
WHERE al.release_date IS NOT NULL AND al.release_date != ''
GROUP BY al.name, a.name, al.release_date
ORDER BY al.release_date DESC LIMIT 10
```

### Fun Stats

```sql
-- ref: one-hit-wonders (table)
-- Artists where 80%+ of plays are a single song
WITH artist_songs AS (
  SELECT t.artist_id, a.name AS artist,
    l.track_id, t.name AS song, COUNT(*) AS plays
  FROM listen l
  JOIN track t ON t.spotify_id = l.track_id
  JOIN artist a ON a.spotify_id = t.artist_id
  GROUP BY t.artist_id, a.name, l.track_id, t.name
),
artist_totals AS (
  SELECT artist_id, artist, SUM(plays) AS total_plays,
    MAX(plays) AS top_song_plays,
    (ARRAY_AGG(song ORDER BY plays DESC))[1] AS top_song
  FROM artist_songs GROUP BY artist_id, artist
)
SELECT artist, top_song, top_song_plays, total_plays,
  ROUND(top_song_plays::numeric / total_plays * 100, 1) AS concentration_pct
FROM artist_totals
WHERE total_plays >= 5 AND top_song_plays::numeric / total_plays >= 0.8
ORDER BY total_plays DESC LIMIT 15

-- ref: most-loyal-artists (table)
-- Artists listened to in the most distinct months
SELECT a.name AS artist,
  COUNT(DISTINCT TO_CHAR(l.started_at, 'YYYY-MM')) AS months_listened,
  COUNT(*) AS total_plays
FROM listen l
JOIN track t ON t.spotify_id = l.track_id
JOIN artist a ON a.spotify_id = t.artist_id
GROUP BY a.name
ORDER BY months_listened DESC, total_plays DESC LIMIT 15

-- ref: most-binged-album (table)
-- Albums with the highest plays concentrated in a 7-day window
WITH album_daily AS (
  SELECT t.album_id, al.name AS album, a.name AS artist,
    DATE(l.started_at) AS day, COUNT(*) AS plays
  FROM listen l
  JOIN track t ON t.spotify_id = l.track_id
  JOIN album al ON al.spotify_id = t.album_id
  JOIN artist a ON a.spotify_id = t.artist_id
  GROUP BY t.album_id, al.name, a.name, day
),
rolling AS (
  SELECT album_id, album, artist, day,
    SUM(plays) OVER (PARTITION BY album_id ORDER BY day
      RANGE BETWEEN INTERVAL '6 days' PRECEDING AND CURRENT ROW) AS week_plays
  FROM album_daily
)
SELECT album, artist, MAX(week_plays) AS peak_week_plays
FROM rolling
GROUP BY album_id, album, artist
ORDER BY peak_week_plays DESC LIMIT 15

-- ref: forgotten-favorites (table)
-- Songs heavily played (10+) but not heard in the last 60 days
SELECT t.name AS song, a.name AS artist,
  COUNT(*) AS total_plays,
  MAX(l.started_at) AS last_played
FROM listen l
JOIN track t ON t.spotify_id = l.track_id
JOIN artist a ON a.spotify_id = t.artist_id
GROUP BY t.name, a.name
HAVING COUNT(*) >= 10
  AND MAX(l.started_at) < CURRENT_DATE - INTERVAL '60 days'
ORDER BY total_plays DESC LIMIT 15

-- ref: soundtrack-of-your-life (table)
-- Top song per month since tracking began
WITH monthly AS (
  SELECT TO_CHAR(l.started_at, 'YYYY-MM') AS month,
    t.name AS song, a.name AS artist, COUNT(*) AS plays,
    ROW_NUMBER() OVER (
      PARTITION BY TO_CHAR(l.started_at, 'YYYY-MM') ORDER BY COUNT(*) DESC
    ) AS rn
  FROM listen l
  JOIN track t ON t.spotify_id = l.track_id
  JOIN artist a ON a.spotify_id = t.artist_id
  GROUP BY month, t.name, a.name
)
SELECT month, song, artist, plays
FROM monthly WHERE rn = 1 ORDER BY month DESC

-- ref: obsession-detector (table)
-- Biggest week-over-week spikes for any artist
WITH weekly AS (
  SELECT t.artist_id, a.name AS artist,
    DATE_TRUNC('week', l.started_at)::date AS week, COUNT(*) AS plays
  FROM listen l
  JOIN track t ON t.spotify_id = l.track_id
  JOIN artist a ON a.spotify_id = t.artist_id
  GROUP BY t.artist_id, a.name, week
),
with_prev AS (
  SELECT *, LAG(plays) OVER (PARTITION BY artist_id ORDER BY week) AS prev_plays
  FROM weekly
)
SELECT artist, week, plays, prev_plays,
  plays - COALESCE(prev_plays, 0) AS spike
FROM with_prev
WHERE plays >= 10
ORDER BY spike DESC LIMIT 15
```

## Dashboard card layout

Starting at row 201 (after phase 27 ends at row 200).

| Row | Col | Size | Card |
|-----|-----|------|------|
| 201 | 0 | 12×1 | Heading: "Deep Cuts" |
| 202 | 0 | 4×3 | avg-song-length |
| 202 | 4 | 4×3 | total-days-of-music |
| 202 | 8 | 4×3 | est-tb-streamed |
| 205 | 0 | 6×8 | deepest-artist-dive |
| 205 | 6 | 6×8 | deepest-album-dive |
| 213 | 0 | 6×6 | longest-songs-listened |
| 213 | 6 | 6×6 | shortest-songs-listened |
| 219 | 0 | 6×6 | oldest-releases |
| 219 | 6 | 6×6 | newest-releases |
| 225 | 0 | 12×1 | Heading: "Fun Stats" |
| 226 | 0 | 6×8 | one-hit-wonders |
| 226 | 6 | 6×8 | most-loyal-artists |
| 234 | 0 | 6×8 | most-binged-album |
| 234 | 6 | 6×8 | forgotten-favorites |
| 242 | 0 | 12×10 | soundtrack-of-your-life |
| 252 | 0 | 12×8 | obsession-detector |

## Definition of done

- [ ] Dashboard shows "Deep Cuts" heading with 3 scalars and 6 tables
- [ ] Estimated TB streamed uses 160kbps average bitrate assumption
- [ ] Deepest album dive only includes albums with >3 tracks
- [ ] Dashboard shows "Fun Stats" heading with 6 tables
- [ ] One-hit wonders shows artists where 80%+ of plays are one song (min 5 plays)
- [ ] Forgotten favorites shows songs with 10+ plays not heard in the last 60 days
- [ ] Soundtrack of your life shows the top song for every month, ordered newest first
- [ ] Obsession detector shows the biggest week-over-week play spikes (min 10 plays/week)
