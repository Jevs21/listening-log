# Phase 14 — Stats Page with Suggestion Gate

Builds on Phase 12 (song suggestions).

## Goal

Add a `/stats` route with a gate: visitors who haven't submitted a song suggestion (by IP) see a playful denial message with a link to the suggestion modal. Once they submit, they're redirected to the real stats page. The stats page itself is a placeholder for now.

## Scope

### In scope
- Client-side routing via React Router
- "stats" link in intro text (same style as "suggestion")
- `GET /api/suggestions/check` endpoint — returns whether requesting IP has any submissions
- Gate component that checks suggestion status and renders accordingly
- Gated message with inline "suggestion" link that opens the modal
- Auto-redirect to stats content after successful suggestion submit from the gate
- Placeholder stats page content

### Out of scope
- Actual stats content (future spec)
- Any auth or accounts

## API changes

### `GET /api/suggestions/check`

Returns whether the requesting IP has submitted at least one suggestion.

**Response:**
```json
{ "has_suggested": true }
```

Uses `c.ClientIP()` like the existing POST endpoint.

**Query:** `SELECT COUNT(*) FROM song_suggestion WHERE ip_address = ? LIMIT 1`

**Files:**
- Add handler in `server/handlers/suggestions.go`
- Add `HasSuggested(db, ip) (bool, error)` in `server/db/suggestions.go`
- Register route in `server/main.go`

## Frontend changes

### Routing — `App.tsx`

Add `react-router-dom`. Wrap app in `BrowserRouter`. Two routes:

| Path | Component |
|------|-----------|
| `/` | Current home page content |
| `/stats` | `StatsPage` |

Extract current home content into a `HomePage` component or keep inline — either way, the suggestion modal state and rendering stays on the home route.

### Intro text link

In the intro text, make "stats" a `<Link to="/stats">` styled identically to the "suggestion" span (`.suggestion-link` class or equivalent).

### `client/src/pages/StatsPage.tsx`

New component. On mount:

1. Call `GET /api/suggestions/check`
2. While loading: show nothing (or the same `Loading...` pattern)
3. If `has_suggested` is false: render the gate message
4. If `has_suggested` is true: render placeholder stats content

**Gate message** — centered `<p>` matching `.app-description` style:

> woah hold on a second. gonna need that song <span>suggestion</span> first. you thought i wasn't gonna check but this is really a ploy to get song suggestions. jokes on you.

Where "suggestion" is a clickable span (same style as home page) that opens the `SuggestionModal`. On successful suggestion submit (modal closes after "thx 🫶"), automatically re-check the endpoint and show the stats content — no manual navigation needed.

**Placeholder stats content** — centered `<p>` in same style:

> stats coming soon

### `client/src/api/suggestions.ts`

Add:

```ts
export async function checkSuggestion(): Promise<{ has_suggested: boolean }> {
  const res = await fetch("/api/suggestions/check");
  return res.json();
}
```

### Server-side catch-all

The Go server needs to serve `index.html` for any non-`/api/` route so that navigating directly to `/stats` works (client-side routing). If this isn't already handled, add a catch-all fallback in `server/main.go` that serves the SPA index for unmatched routes.

## Definition of done

- [ ] App uses React Router with `/` and `/stats` routes
- [ ] "stats" in intro text is a link to `/stats`, styled like "suggestion"
- [ ] `GET /api/suggestions/check` returns `{ "has_suggested": true/false }` based on IP
- [ ] Visiting `/stats` with no prior suggestion shows the gate message
- [ ] "suggestion" in the gate message opens the suggestion modal
- [ ] After submitting a suggestion from the gate, stats content appears automatically (no manual nav)
- [ ] Visiting `/stats` with a prior suggestion shows placeholder stats content
- [ ] Navigating directly to `/stats` (browser URL bar) works (server serves SPA fallback)
