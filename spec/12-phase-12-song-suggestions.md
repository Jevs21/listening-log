# Phase 12 — Song Suggestions

## Goal

Make the word "suggestion" in the intro text clickable to open a modal where visitors can submit a song suggestion (link and/or message). Store submissions in SQLite with basic abuse protection.

## Scope

### In scope
- Clickable "suggestion" trigger in `App.tsx` intro text
- Modal with link input, message textarea, and submit button
- `POST /api/suggestions` endpoint
- `song_suggestion` table in SQLite
- Input validation (length limits, at least one field required)
- IP-based rate limiting (3 submissions per hour)
- Success confirmation: "thx 🫶"

### Out of scope
- Admin UI to view/manage suggestions
- Link preview or metadata fetching
- Email/push notifications on new suggestions

## Data model

```sql
CREATE TABLE IF NOT EXISTS song_suggestion (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    link       TEXT    NOT NULL DEFAULT '',
    message    TEXT    NOT NULL DEFAULT '',
    ip_address TEXT    NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_song_suggestion_ip_created ON song_suggestion (ip_address, created_at);
```

- `link` and `message` are both optional, but the API enforces at least one is non-empty.
- `ip_address` stored for rate limiting lookups.

## API changes

### `POST /api/suggestions`

**Request body:**
```json
{
  "link": "https://open.spotify.com/track/...",
  "message": "you'd love this one"
}
```

**Validation rules:**
| Field   | Max length | Required          |
|---------|-----------|-------------------|
| link    | 2048 chars | No (but one of link/message must be present) |
| message | 200 chars  | No (but one of link/message must be present) |

**Rate limit:** 3 per IP per rolling 60-minute window. Checked via `SELECT COUNT(*) FROM song_suggestion WHERE ip_address = ? AND created_at > datetime('now', '-1 hour')`.

**Responses:**
| Status | Body | When |
|--------|------|------|
| 201    | `{"ok": true}` | Success |
| 400    | `{"error": "..."}` | Validation failure (both empty, too long) |
| 429    | `{"error": "rate limit exceeded"}` | >3 submissions/hour from same IP |

**Files:**
- `server/handlers/suggestions.go` — handler func
- `server/db/suggestions.go` — `InsertSuggestion(db, link, message, ip) error` and `CountRecentSuggestions(db, ip) (int, error)`
- Register route in `server/main.go`

**IP extraction:** Use `c.ClientIP()` (Gin's built-in, handles X-Forwarded-For).

## Frontend changes

### App.tsx

Split the intro text so "suggestion" is wrapped in a `<span>` with `onClick` to open the modal. Style the span with `cursor: pointer` and `text-decoration: underline` to hint it's interactive.

### `client/src/components/SuggestionModal.tsx`

New component. Props: `open: boolean`, `onClose: () => void`.

- Dark overlay (`rgba(0,0,0,0.7)`) covering viewport, click to close
- Centered card matching app's dark minimal aesthetic
- X button top-right to close
- Form fields:
  - Text input for link — placeholder: `"spotify, apple music, or youtube link"`
  - Textarea for message — placeholder: `"or leave a message"`, maxLength 200, show char count
- Submit button, disabled when both fields are empty or while submitting
- On success: replace form content with "thx 🫶", auto-close after ~2 seconds
- On 429: show inline error "too many suggestions, try again later"
- On other errors: show inline error "something went wrong"

### `client/src/api/suggestions.ts`

New file. Single function:

```ts
export async function submitSuggestion(link: string, message: string): Promise<Response> {
  return fetch("/api/suggestions", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ link, message }),
  });
}
```

No React Query needed — this is a fire-and-forget mutation.

## Definition of done

- [ ] Clicking "suggestion" in intro text opens modal
- [ ] Modal closes on overlay click or X button
- [ ] Can submit with only a link, only a message, or both
- [ ] Cannot submit when both fields are empty
- [ ] Message textarea enforces 200 char limit with visible counter
- [ ] Successful submit shows "thx 🫶" then auto-closes
- [ ] Submissions are stored in `song_suggestion` table with IP and timestamp
- [ ] 4th submission from same IP within an hour returns 429
- [ ] Link over 2048 chars or message over 200 chars returns 400
- [ ] Schema migration runs on startup (CREATE TABLE IF NOT EXISTS)
