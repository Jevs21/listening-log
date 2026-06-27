# Phase 23 — Dashboard: Overview & Listening Habits

Builds on phase 22 (dashboard dark theme) and phase 20 (dashboards as code).

## Goal

Add summary scalar cards and listening habit visualizations to the Listening Overview dashboard. Also extend the dashboard loader to support heading cards and to sync cards on every boot (not just on dashboard creation), so that subsequent phases can add cards incrementally.

## Scope

### In scope

- Loader changes: heading card support + always-sync cards
- Dashboard heading: "Overview"
- Scalar cards: total hours, unique songs/albums/artists, avg minutes/day, current streak, longest streak
- Dashboard heading: "Listening Habits"
- Scalar cards: total sessions, avg session length, longest session
- Table: On Repeat (last 7 days)
- Bar charts: listens by hour of day, listens by day of week
- Bar chart: weekday vs weekend minutes

### Out of scope

- Metrics requiring genre data
- Session detection tuning (use 30-minute gap threshold)

## Loader changes

### `metabase/load-dashboards.sh`

Two changes:

**1. Support heading cards in the JSON format.** Cards with `"type": "heading"` should be included in the dashboard PUT as virtual cards (no saved question). In the cards payload building step, handle both types:

```bash
# Build cards payload — handles both question cards and heading cards
CARDS_PAYLOAD=$(jq -c --argjson ref_map "$REF_MAP" \
  '[.cards | to_entries[] | .value as $card | .key as $idx |
    if $card.type == "heading" then
      {
        id: (-$idx - 1),
        card_id: null,
        row: $card.row,
        col: $card.col,
        size_x: $card.size_x,
        size_y: $card.size_y,
        visualization_settings: {
          virtual_card: {
            name: null,
            display: "heading",
            visualization_settings: {},
            dataset_query: {},
            archived: false
          },
          text: $card.text
        }
      }
    else
      {
        id: (-$idx - 1),
        card_id: ($ref_map[$card.ref] // "" | tonumber),
        row: $card.row,
        col: $card.col,
        size_x: $card.size_x,
        size_y: $card.size_y
      }
    end
  ]' "$file")
```

**2. Always sync cards, not just on creation.** Move the cards PUT outside the `if/else` block so it runs for both new and existing dashboards. Use `FINAL_DASH_ID` instead of `DASH_ID`:

```bash
# After dashboard create-or-find block, always sync cards:
echo "  Syncing cards to dashboard..."
curl -s -X PUT "$MB_URL/api/dashboard/$FINAL_DASH_ID" \
  -H "Content-Type: application/json" \
  -H "$AUTH_HEADER" \
  -d "$(jq -n --argjson dashcards "$CARDS_PAYLOAD" '{ dashcards: $dashcards }')" > /dev/null
echo "  Synced cards."
```

### Dashboard JSON format addition

Cards array now supports two types:

```json
{ "ref": "some-question", "row": 0, "col": 0, "size_x": 12, "size_y": 6 }
{ "type": "heading", "text": "Section Title", "row": 0, "col": 0, "size_x": 12, "size_y": 1 }
```

Question cards (no `type` field or `type: "question"`) work as before.

## Dashboard questions

All questions added to `metabase/dashboards/listening-overview.json`.

### Scalars

```sql
-- ref: total-hours
SELECT ROUND(SUM(duration_ms) / 3600000.0, 1) AS hours FROM listen

-- ref: total-unique-songs
SELECT COUNT(DISTINCT track_id) AS songs FROM listen

-- ref: total-unique-albums
SELECT COUNT(DISTINCT t.album_id) AS albums
FROM listen l JOIN track t ON t.spotify_id = l.track_id

-- ref: total-unique-artists
SELECT COUNT(DISTINCT t.artist_id) AS artists
FROM listen l JOIN track t ON t.spotify_id = l.track_id

-- ref: avg-minutes-per-day
SELECT ROUND(AVG(daily_min)) AS avg_minutes
FROM (SELECT DATE(started_at) AS day, SUM(duration_ms) / 60000.0 AS daily_min
      FROM listen GROUP BY day) s

-- ref: current-listening-streak
WITH listen_days AS (
  SELECT DISTINCT DATE(started_at) AS day FROM listen
),
islands AS (
  SELECT day, day - (ROW_NUMBER() OVER (ORDER BY day))::int AS grp
  FROM listen_days
),
streaks AS (
  SELECT COUNT(*) AS len, MAX(day) AS last_day FROM islands GROUP BY grp
)
SELECT COALESCE(MAX(len), 0) AS days
FROM streaks WHERE last_day >= CURRENT_DATE - 1

-- ref: longest-listening-streak
WITH listen_days AS (
  SELECT DISTINCT DATE(started_at) AS day FROM listen
),
islands AS (
  SELECT day, day - (ROW_NUMBER() OVER (ORDER BY day))::int AS grp
  FROM listen_days
)
SELECT COALESCE(MAX(cnt), 0) AS days
FROM (SELECT COUNT(*) AS cnt FROM islands GROUP BY grp) s
```

### Session-based scalars

Define a session as consecutive listens where the gap between `ended_at` and the next `started_at` exceeds 30 minutes.

```sql
-- ref: total-sessions (scalar)
WITH ordered AS (
  SELECT started_at, ended_at,
    CASE WHEN started_at - LAG(ended_at) OVER (ORDER BY started_at) > INTERVAL '30 minutes'
         OR LAG(ended_at) OVER (ORDER BY started_at) IS NULL
    THEN 1 ELSE 0 END AS new_session
  FROM listen
)
SELECT SUM(new_session) AS sessions FROM ordered

-- ref: avg-session-minutes (scalar)
WITH ordered AS (
  SELECT started_at, ended_at,
    CASE WHEN started_at - LAG(ended_at) OVER (ORDER BY started_at) > INTERVAL '30 minutes'
         OR LAG(ended_at) OVER (ORDER BY started_at) IS NULL
    THEN 1 ELSE 0 END AS new_session
  FROM listen
),
sessioned AS (
  SELECT *, SUM(new_session) OVER (ORDER BY started_at) AS sid FROM ordered
)
SELECT ROUND(AVG(EXTRACT(EPOCH FROM (session_end - session_start)) / 60)) AS avg_minutes
FROM (SELECT sid, MIN(started_at) AS session_start, MAX(ended_at) AS session_end
      FROM sessioned GROUP BY sid) s

-- ref: longest-session-minutes (scalar)
WITH ordered AS (
  SELECT started_at, ended_at,
    CASE WHEN started_at - LAG(ended_at) OVER (ORDER BY started_at) > INTERVAL '30 minutes'
         OR LAG(ended_at) OVER (ORDER BY started_at) IS NULL
    THEN 1 ELSE 0 END AS new_session
  FROM listen
),
sessioned AS (
  SELECT *, SUM(new_session) OVER (ORDER BY started_at) AS sid FROM ordered
)
SELECT ROUND(MAX(EXTRACT(EPOCH FROM (session_end - session_start)) / 60)) AS minutes
FROM (SELECT sid, MIN(started_at) AS session_start, MAX(ended_at) AS session_end
      FROM sessioned GROUP BY sid) s
```

### Charts & tables

```sql
-- ref: on-repeat-7d (table)
SELECT t.name AS song, a.name AS artist, COUNT(*) AS plays,
  ROUND(SUM(l.duration_ms) / 60000.0) AS minutes
FROM listen l
JOIN track t ON t.spotify_id = l.track_id
JOIN artist a ON a.spotify_id = t.artist_id
WHERE l.started_at >= CURRENT_DATE - INTERVAL '7 days'
GROUP BY t.name, a.name ORDER BY plays DESC LIMIT 10

-- ref: listens-by-hour (bar)
SELECT EXTRACT(HOUR FROM started_at)::int AS hour, COUNT(*) AS listens
FROM listen GROUP BY hour ORDER BY hour

-- ref: listens-by-day-of-week (bar)
SELECT TO_CHAR(started_at, 'Dy') AS day,
  EXTRACT(ISODOW FROM started_at)::int AS day_num, COUNT(*) AS listens
FROM listen GROUP BY day, day_num ORDER BY day_num

-- ref: weekday-vs-weekend (bar)
SELECT CASE WHEN EXTRACT(ISODOW FROM started_at) <= 5 THEN 'Weekday' ELSE 'Weekend' END AS period,
  ROUND(SUM(duration_ms) / 60000.0) AS minutes
FROM listen GROUP BY period
```

## Dashboard card layout

Starting at row 14 (after existing cards which occupy rows 0–13).

Note: Metabase uses a 24-column grid.

| Row | Col | Size | Card |
|-----|-----|------|------|
| 14 | 0 | 24×1 | Heading: "Overview" |
| 15 | 0 | 6×3 | total-hours |
| 15 | 6 | 6×3 | total-unique-songs |
| 15 | 12 | 6×3 | total-unique-albums |
| 15 | 18 | 6×3 | total-unique-artists |
| 18 | 0 | 8×3 | avg-minutes-per-day |
| 18 | 8 | 8×3 | current-listening-streak |
| 18 | 16 | 8×3 | longest-listening-streak |
| 21 | 0 | 24×1 | Heading: "Listening Habits" |
| 22 | 0 | 8×3 | total-sessions |
| 22 | 8 | 8×3 | avg-session-minutes |
| 22 | 16 | 8×3 | longest-session-minutes |
| 25 | 0 | 24×8 | on-repeat-7d |
| 33 | 0 | 12×6 | listens-by-hour |
| 33 | 12 | 12×6 | listens-by-day-of-week |
| 39 | 0 | 12×5 | weekday-vs-weekend |

## Definition of done

- [ ] Loader supports `"type": "heading"` cards in dashboard JSON
- [ ] Loader syncs cards on every boot, not just when creating a new dashboard
- [ ] Dashboard shows "Overview" heading with 7 scalar summary cards
- [ ] Dashboard shows "Listening Habits" heading with 3 session scalars, on-repeat table, hour/day bar charts, and weekday/weekend bar
- [ ] All scalar cards display a single number
- [ ] On Repeat table shows song, artist, plays, and minutes for the last 7 days
