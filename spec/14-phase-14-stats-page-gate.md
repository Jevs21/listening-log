# Phase 14 — Stats Page with Suggestion Gate

Builds on Phase 12 (song suggestions).

## Goal

Add a `/stats` route gated by IP-based suggestion check. Visitors without a suggestion are redirected to `/woah-hold-it-right-there-buckaroo` where they see a playful denial message and can submit a suggestion inline (no modal). After submitting, they're redirected to `/stats`. The stats page itself is a placeholder for now.

## Scope

### In scope
- Client-side routing via React Router (`/`, `/stats`, `/woah-hold-it-right-there-buckaroo`)
- "stats" link in intro text (same style as "suggestion")
- `GET /api/suggestions/check` endpoint — returns whether requesting IP has any submissions
- Refactor suggestion form out of modal into reusable `SuggestionForm` component
- Gate page with inline suggestion form (not a modal)
- Auto-redirect to `/stats` after successful suggestion submit from gate page
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

**Query:** `SELECT COUNT(*) FROM song_suggestion WHERE ip_address = ?`

**Files:**
- `CheckSuggestion` handler in `server/handlers/suggestions.go`
- `HasSuggested(db, ip) (bool, error)` in `server/db/suggestions.go`
- Register `GET /api/suggestions/check` route in `server/main.go`

## Frontend changes

### Component refactor

Extract the suggestion form (inputs, submit, success/error states) from `SuggestionModal` into a standalone `SuggestionForm` component:

- `client/src/components/SuggestionForm.tsx` — form with `onSuccess?: () => void` prop. Shows "thx 🫶" on success, calls `onSuccess` after 2s delay.
- `client/src/components/SuggestionForm.css` — input and submit button styles (extracted from `SuggestionModal.css`, renamed from `modal-*` to `suggestion-*`)
- `client/src/components/SuggestionModal.tsx` — simplified to just modal chrome (overlay, panel, close button) wrapping `SuggestionForm`. Passes `onClose` as `onSuccess`.
- `client/src/components/SuggestionModal.css` — only modal chrome styles (overlay, panel, close button)

### Routing — `App.tsx` / `main.tsx`

Add `react-router-dom`. Wrap app in `BrowserRouter` in `main.tsx`. Three routes:

| Path | Component |
|------|-----------|
| `/` | `HomePage` (extracted from current `App`) |
| `/stats` | `StatsPage` |
| `/woah-hold-it-right-there-buckaroo` | `GatePage` |

### Intro text link

Make "stats" a `<Link to="/stats">` styled with `.suggestion-link`.

### `client/src/pages/StatsPage.tsx`

On mount, calls `GET /api/suggestions/check`:
- Loading: render nothing
- `has_suggested: false`: `<Navigate to="/woah-hold-it-right-there-buckaroo" replace />`
- `has_suggested: true`: render placeholder "stats coming soon"

### `client/src/pages/GatePage.tsx`

Renders the denial message as `app-description` text followed by an inline `SuggestionForm`. The form's `onSuccess` calls `navigate("/stats")`.

Message text:
> woah hold on a second. gonna need that song suggestion first. you thought i wasn't gonna check but this is really a ploy to get song suggestions. jokes on you.

### `client/src/api/suggestions.ts`

Add:

```ts
export async function checkSuggestion(): Promise<{ has_suggested: boolean }> {
  const res = await fetch("/api/suggestions/check");
  return res.json();
}
```

### Server-side catch-all

Already handled by existing `spaMiddleware` in `server/main.go` — serves `index.html` for any non-`/api/` route.

## Definition of done

- [ ] Suggestion form extracted into reusable `SuggestionForm` component
- [ ] `SuggestionModal` uses `SuggestionForm` internally (no duplication)
- [ ] App uses React Router with `/`, `/stats`, and `/woah-hold-it-right-there-buckaroo` routes
- [ ] "stats" in intro text is a link to `/stats`, styled like "suggestion"
- [ ] `GET /api/suggestions/check` returns `{ "has_suggested": true/false }` based on IP
- [ ] `/stats` redirects to `/woah-hold-it-right-there-buckaroo` if no prior suggestion
- [ ] Gate page shows denial message with inline suggestion form
- [ ] After submitting on gate page, user is redirected to `/stats`
- [ ] `/stats` with a prior suggestion shows placeholder stats content
- [ ] Direct navigation to any route works (SPA fallback)
