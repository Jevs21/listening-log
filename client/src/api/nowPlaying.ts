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
