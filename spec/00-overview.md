# Listening Log

The listening log will be a place for a user to see a detailed breakdown of their spotify listening data.

The core of the app will be a scraper hitting the spotify API every ~15s to see what the user is listening to. It will log this to a database. From this core database, insights can be gathered by analyzing the data to see what music is being listened to and for how long - and potentially more later on.

The core database will be analyzed and re-analyzed by other system components to build a second data layer, constructed more for presentation in a web app.

## Phase 1: Web App Scaffolding

The first thing needed will be a frontend to do the Spotify auth flow. This will require a very basic UI and corresponding HTTP server. When the user completes the auth flow, the required tokens (that are safe and necessary to store) will be put into a database so other system components can use them. Be sure to use a .env file for any shared globals. I should be able to pnpm dev in one terminal, and ./server in another and the app work.

Technologies:
- Frontend - Vite + Typescript + React (Bare Minimum MVP)
- Server - Go + Gin for HTTP routes

Structure:
- Frontend in ./client
- Server in ./server
- Database in ./data/database.sqlite


## Phase 2: Scraper MVP

The MVP will consist of just the core scraper inserting data into the core database layer.

Technologies:
- Database - File based SQLite
- Server - Go

## Phase 3: Core Polling Table

The scraper writes each poll result to a `playback_log` table — one row per poll, storing only data that can't be retrieved later by ID. This is the permanent source of truth for all future analysis. Track/album/artist metadata will live in separate lookup tables built in a later phase.

## Phase 4: Metadata Lookup Tables

On each poll, the scraper inserts track, album, and artist metadata into lookup tables (insert-or-ignore by Spotify ID). All data comes from the existing now-playing response — no extra API calls. Adds `artist`, `album`, `album_image`, and `track` tables.

## Phase 5: Metadata Upsert

Add `updated_at` column to `artist`, `album`, and `track` tables. Switch from insert-or-ignore to upsert so `updated_at` is refreshed each time the scraper encounters an existing row. No other columns change on conflict.

## Phase 6: Now Playing API

First data-driven frontend feature. A new `GET /api/now-playing` Go endpoint returns the most recently updated track joined with artist and album names. TanStack React Query is added to the client with a 10-second polling interval. A reusable `<NowPlaying />` component displays the result as plain text. Establishes the query hook + API client pattern for all future frontend data queries.
