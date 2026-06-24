# Phase 22 — Dashboard Dark Theme

Builds on phase 21 (embed dashboard).

## Goal

Fix unreadable dark-on-dark text in the embedded Metabase dashboard by switching from `theme=transparent` to `theme=night` with a transparent background, so the dashboard content uses light text that's readable against the app's dark background.

## Scope

### In scope

- Change the embed URL hash parameters in `server/handlers/stats.go`

### Out of scope

- Metabase appearance API settings (requires Pro/Enterprise)
- Custom fonts (requires Pro/Enterprise)
- Custom chart color palettes (requires Pro/Enterprise)
- CSS injection or other iframe styling hacks

## Backend changes

### `server/handlers/stats.go`

Change the URL hash fragment from:

```
#bordered=false&titled=false&theme=transparent
```

to:

```
#theme=night&background=false&bordered=false&titled=false
```

`theme=night` switches all Metabase text/UI to light colors for dark backgrounds. `background=false` removes the dashboard background so the app's own background shows through. Both parameters are available on OSS Metabase.

## Definition of done

- [ ] `GET /api/stats/dashboard` returns a URL with `#theme=night&background=false&bordered=false&titled=false`
- [ ] Embedded dashboard on `/stats` shows light text on the app's dark background
- [ ] Chart labels, axis text, and card titles are all readable
