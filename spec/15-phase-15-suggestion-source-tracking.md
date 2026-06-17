# Phase 15 — Suggestion Source Tracking

Builds on Phase 12 (song suggestions) and Phase 14 (stats page gate).

## Goal

Track whether each song suggestion was submitted from the home page modal or the forced suggestion gate page.

## Scope

### In scope
- Add `source` column to `song_suggestion` table
- Pass source from frontend to `POST /api/suggestions`
- Store source value on insert

### Out of scope
- Displaying source data on stats page (future spec)

## Data model

Add column to `song_suggestion`:

```sql
CREATE TABLE IF NOT EXISTS song_suggestion (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    link       TEXT    NOT NULL DEFAULT '',
    message    TEXT    NOT NULL DEFAULT '',
    source     TEXT    NOT NULL DEFAULT 'home',
    ip_address TEXT    NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

Valid values for `source`: `"home"`, `"gate"`.

User is wiping the database — no migration needed, just update the CREATE TABLE statement.

## API changes

### `POST /api/suggestions`

Add `source` to request body:

```json
{
  "link": "...",
  "message": "...",
  "source": "home"
}
```

| Field  | Max length | Required | Validation |
|--------|-----------|----------|------------|
| source | —         | No       | Must be `"home"` or `"gate"` if provided. Defaults to `"home"`. |

Rate limiting remains combined across sources (unchanged).

## Backend changes

- `server/db/suggestions.go` — `InsertSuggestion` gains a `source string` parameter. Pass it into the INSERT.
- `server/handlers/suggestions.go` — parse `source` from request body, validate against allowed values, pass to `InsertSuggestion`.

## Frontend changes

- `client/src/api/suggestions.ts` — `submitSuggestion` gains a `source` parameter, includes it in the POST body.
- `client/src/components/SuggestionForm.tsx` — accept `source` prop (`"home" | "gate"`), pass it through to `submitSuggestion`.
- `client/src/components/SuggestionModal.tsx` — pass `source="home"` to `SuggestionForm`.
- `client/src/pages/GatePage.tsx` — pass `source="gate"` to `SuggestionForm`.

## Definition of done

- [ ] `song_suggestion` table includes `source` column
- [ ] Suggestions from home page modal are stored with `source = "home"`
- [ ] Suggestions from gate page are stored with `source = "gate"`
- [ ] API rejects invalid source values with 400
- [ ] API defaults to `"home"` if source is omitted
- [ ] Rate limiting is unchanged (combined across sources)
