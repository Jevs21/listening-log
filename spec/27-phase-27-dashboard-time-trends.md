# Phase 27 — Dashboard: Time-Based Trends

Builds on phase 26 (extended rankings).

## Goal

Add time-series and calendar-oriented metrics to the Listening Overview dashboard — monthly listening trends, seasonal patterns, and historical comparisons.

## Scope

### In scope

- Heading: "Time-Based Trends"
- Line charts: minutes per month, unique artists per month
- Bar chart: month-over-month listening growth (% change)
- Table: seasonal favorites (top artist per quarter)
- Table: "this day in history" (what was playing on today's date in prior years)
- Line chart: daily top artist calendar (top artist per day, last 30 days)

### Out of scope

- Artist/album obsession graphs (would require parameterized queries — better as standalone Metabase questions, not dashboard cards)
- Genre evolution timeline (no genre data)

## Dashboard questions

All added to `metabase/dashboards/listening-overview.json`.

```sql
-- ref: minutes-per-month (line)
SELECT TO_CHAR(DATE_TRUNC('month', started_at), 'YYYY-MM') AS month,
  ROUND(SUM(duration_ms) / 60000.0) AS minutes
FROM listen
GROUP BY month ORDER BY month

-- ref: unique-artists-per-month (line)
SELECT TO_CHAR(DATE_TRUNC('month', l.started_at), 'YYYY-MM') AS month,
  COUNT(DISTINCT t.artist_id) AS artists
FROM listen l JOIN track t ON t.spotify_id = l.track_id
GROUP BY month ORDER BY month

-- ref: listening-growth-mom (bar)
-- Month-over-month % change in total listens
WITH monthly AS (
  SELECT DATE_TRUNC('month', started_at) AS month, COUNT(*) AS listens
  FROM listen GROUP BY month
)
SELECT TO_CHAR(month, 'YYYY-MM') AS month,
  listens,
  ROUND((listens - LAG(listens) OVER (ORDER BY month))::numeric
    / NULLIF(LAG(listens) OVER (ORDER BY month), 0) * 100, 1) AS growth_pct
FROM monthly ORDER BY month

-- ref: seasonal-favorites (table)
-- Top artist per quarter
WITH quarterly AS (
  SELECT EXTRACT(YEAR FROM l.started_at)::int AS year,
    EXTRACT(QUARTER FROM l.started_at)::int AS quarter,
    a.name AS artist,
    COUNT(*) AS plays,
    ROW_NUMBER() OVER (
      PARTITION BY EXTRACT(YEAR FROM l.started_at), EXTRACT(QUARTER FROM l.started_at)
      ORDER BY COUNT(*) DESC
    ) AS rn
  FROM listen l
  JOIN track t ON t.spotify_id = l.track_id
  JOIN artist a ON a.spotify_id = t.artist_id
  GROUP BY year, quarter, a.name
)
SELECT year, 'Q' || quarter AS quarter, artist, plays
FROM quarterly WHERE rn = 1
ORDER BY year DESC, quarter DESC

-- ref: this-day-in-history (table)
-- What was playing on this calendar date in prior years
SELECT EXTRACT(YEAR FROM l.started_at)::int AS year,
  t.name AS song, a.name AS artist, COUNT(*) AS plays
FROM listen l
JOIN track t ON t.spotify_id = l.track_id
JOIN artist a ON a.spotify_id = t.artist_id
WHERE EXTRACT(MONTH FROM l.started_at) = EXTRACT(MONTH FROM CURRENT_DATE)
  AND EXTRACT(DAY FROM l.started_at) = EXTRACT(DAY FROM CURRENT_DATE)
  AND DATE(l.started_at) != CURRENT_DATE
GROUP BY year, t.name, a.name
ORDER BY year DESC, plays DESC

-- ref: daily-top-artist (table)
-- Top artist per day for the last 30 days
WITH daily AS (
  SELECT DATE(l.started_at) AS day, a.name AS artist, COUNT(*) AS plays,
    ROW_NUMBER() OVER (PARTITION BY DATE(l.started_at) ORDER BY COUNT(*) DESC) AS rn
  FROM listen l
  JOIN track t ON t.spotify_id = l.track_id
  JOIN artist a ON a.spotify_id = t.artist_id
  WHERE l.started_at >= CURRENT_DATE - INTERVAL '30 days'
  GROUP BY day, a.name
)
SELECT day, artist, plays FROM daily WHERE rn = 1 ORDER BY day DESC
```

## Dashboard card layout

Starting at row 166 (after phase 26 ends at row 165).

Note: Metabase uses a 24-column grid.

| Row | Col | Size | Card |
|-----|-----|------|------|
| 166 | 0 | 24×1 | Heading: "Time-Based Trends" |
| 167 | 0 | 24×6 | minutes-per-month |
| 173 | 0 | 24×6 | unique-artists-per-month |
| 179 | 0 | 24×6 | listening-growth-mom |
| 185 | 0 | 12×8 | seasonal-favorites |
| 185 | 12 | 12×8 | this-day-in-history |
| 193 | 0 | 24×8 | daily-top-artist |

## Definition of done

- [ ] Dashboard shows "Time-Based Trends" heading
- [ ] Minutes per month and unique artists per month render as line charts
- [ ] Month-over-month growth shows % change as a bar chart
- [ ] Seasonal favorites shows the top artist for each quarter
- [ ] "This day in history" shows songs played on the same calendar date in prior years (excludes today)
- [ ] Daily top artist table shows the most-played artist per day for the last 30 days
