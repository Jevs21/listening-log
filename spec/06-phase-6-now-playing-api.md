# Phase 6: Now Playing API

> Builds on Phase 5. The frontend gains its first data-driven feature: a
> "now playing" indicator showing the most recently updated track. A new Go
> API endpoint serves the data; TanStack React Query handles fetching and
> automatic re-fetching on the client.

## Goal

Display the currently (or most recently) playing song in the frontend by
querying the metadata tables. Establish the React Query infrastructure so
future API queries can be added cleanly with minimal boilerplate.

## Scope

### In scope
- New Go endpoint `GET /api/now-playing` that returns the track with the
  most recent `updated_at`, joined with its artist and album names.
- Install and configure TanStack React Query in the client.
- A reusable `<NowPlaying />` component that displays the now-playing data
  as plain text (no styling).
- React Query polls the endpoint every 10 seconds.
- Organized query hook and API client patterns that are easy to extend
  with future endpoints.

### Out of scope
- Routing (no react-router yet).
- Any CSS, Tailwind, or visual styling.
- Playback controls or Spotify interaction from the frontend.
- Historical playback data or analytics queries.
- Error boundary or retry UI beyond React Query defaults.

## API

### `GET /api/now-playing`

Returns the track whose `updated_at` is most recent, joined with artist
and album metadata.

**Response `200 OK`** (when a track exists):

```json
{
  "track": {
    "spotify_id": "4PTG3Z6ehGkBFwjybzWkR8",
    "name": "Never Gonna Give You Up",
    "artist_name": "Rick Astley",
    "album_name": "Whenever You Need Somebody",
    "duration_ms": 213573,
    "is_explicit": false,
    "updated_at": "2026-06-13T12:34:56Z"
  }
}
```

**Response `200 OK`** (when no tracks exist yet):

```json
{
  "track": null
}
```

**SQL query:**

```sql
SELECT
    t.spotify_id,
    t.name,
    a.name  AS artist_name,
    al.name AS album_name,
    t.duration_ms,
    t.explicit,
    t.updated_at
FROM track t
JOIN artist a  ON t.artist_id = a.spotify_id
JOIN album  al ON t.album_id  = al.spotify_id
ORDER BY t.updated_at DESC
LIMIT 1
```

## Backend changes

### New handler: `handlers/now_playing.go`

- Query function in the `db` package that runs the SQL above and returns a
  struct (or nil if no rows).
- Handler registered on the Gin router at `GET /api/now-playing`.
- Returns JSON as described above.

### Router change in `main.go`

Register the new route alongside the existing `/api/*` routes:

```go
api.GET("/now-playing", handlers.NowPlaying(database))
```

## Frontend changes

### Install dependencies

```
pnpm add @tanstack/react-query
```

### React Query setup

Create a `QueryClient` and wrap the app in `<QueryClientProvider>` in
`main.tsx`.

```tsx
// client/src/main.tsx
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";

const queryClient = new QueryClient();

// wrap <App /> with <QueryClientProvider client={queryClient}>
```

### API client

Add a new function in `client/src/api/` for the now-playing endpoint:

```ts
// client/src/api/nowPlaying.ts
export interface NowPlayingTrack {
  spotify_id: string;
  name: string;
  artist_name: string;
  album_name: string;
  duration_ms: number;
  is_explicit: boolean;
  updated_at: string;
}

export interface NowPlayingResponse {
  track: NowPlayingTrack | null;
}

export async function fetchNowPlaying(): Promise<NowPlayingResponse> {
  const res = await fetch("/api/now-playing");
  return res.json();
}
```

### Query hook

```ts
// client/src/hooks/useNowPlaying.ts
import { useQuery } from "@tanstack/react-query";
import { fetchNowPlaying } from "../api/nowPlaying";

export function useNowPlaying() {
  return useQuery({
    queryKey: ["now-playing"],
    queryFn: fetchNowPlaying,
    refetchInterval: 10_000,
  });
}
```

### `<NowPlaying />` component

```tsx
// client/src/components/NowPlaying.tsx
import { useNowPlaying } from "../hooks/useNowPlaying";

export function NowPlaying() {
  const { data, isLoading, isError } = useNowPlaying();

  if (isLoading) return <p>Loading...</p>;
  if (isError)   return <p>Failed to load now playing.</p>;
  if (!data?.track) return <p>No track data yet.</p>;

  const { name, artist_name, album_name } = data.track;
  return <p>Now playing: {name} by {artist_name} ({album_name})</p>;
}
```

This is a standalone reusable component. For now, render it inside
`App.tsx` when `connected` is true. Later it can be placed on any page or
layout.

## File structure

New and changed files:

```
client/
  src/
    main.tsx                    # changed — add QueryClientProvider
    App.tsx                     # changed — render <NowPlaying />
    api/
      nowPlaying.ts             # new — fetch function + types
    hooks/
      useNowPlaying.ts          # new — React Query hook
    components/
      NowPlaying.tsx            # new — display component
server/
  handlers/
    now_playing.go              # new — handler
  db/
    now_playing.go              # new — query function
  main.go                       # changed — register route
```

## Definition of done

1. `GET /api/now-playing` returns the most recently updated track with
   artist and album names, or `null` if no tracks exist.
2. TanStack React Query is installed and the app is wrapped in
   `QueryClientProvider`.
3. `<NowPlaying />` renders the track name, artist, and album as plain
   text when Spotify is connected.
4. The component automatically re-fetches every 10 seconds without user
   interaction.
5. The hook, API client, and component are in separate files, each easy
   to use as a pattern for future queries.
6. No new styling is introduced.
