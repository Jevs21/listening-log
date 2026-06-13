# Phase 9: Now Playing Staleness Detection

> Builds on Phase 6. The frontend detects when the most recent track is no
> longer actively playing and switches from "Now playing" to "Last played"
> with a relative time-ago label.

## Goal

Show "Now playing: ..." only when the track is plausibly still playing.
When the track's `updated_at` plus its duration (plus a buffer) is in the
past, show "Last played: ... (N ago)" instead. All logic lives in the
frontend — no server changes.

## Scope

### In scope
- Staleness check in the `<NowPlaying />` component.
- A `STALENESS_BUFFER_MS` constant (30 000 ms).
- A reusable `timeAgo(date: Date): string` utility.
- Updated display text for the stale state.

### Out of scope
- Server-side changes.
- Any styling or layout changes.
- Hiding the component entirely after a long period.

## Frontend changes

### New utility: `client/src/utils/timeAgo.ts`

```ts
export function timeAgo(date: Date): string {
  const seconds = Math.floor((Date.now() - date.getTime()) / 1000);
  if (seconds < 60) return "just now";
  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) return `${minutes} min ago`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours} hrs ago`;
  const days = Math.floor(hours / 24);
  return `${days} days ago`;
}
```

### Staleness constant: `client/src/components/NowPlaying.tsx`

```ts
const STALENESS_BUFFER_MS = 30_000;
```

### Staleness check logic

A track is stale when:

```ts
Date.now() > new Date(track.updated_at).getTime() + track.duration_ms + STALENESS_BUFFER_MS
```

### Updated `<NowPlaying />` component

When track exists and is **not stale**:
```
Now playing: {name} by {artist_name} ({album_name})
```

When track exists and **is stale**:
```
Last played: {name} by {artist_name} ({album_name}) ({timeAgo})
```

The staleness check should re-evaluate on each render. Since React Query
refetches every 10s, the component will naturally transition from
"Now playing" to "Last played" within ~10s of a track ending.

## Definition of done

1. While a track is plausibly still playing, display shows "Now playing: ...".
2. After `updated_at + duration_ms + 30s` has passed, display switches to
   "Last played: ... (N ago)".
3. `timeAgo` is a standalone utility in `client/src/utils/timeAgo.ts`
   returning "just now", "N min ago", "N hrs ago", or "N days ago".
4. No server changes.
