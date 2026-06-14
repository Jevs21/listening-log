import { useNowPlaying } from "../hooks/useNowPlaying";
import { timeAgo } from "../utils/timeAgo";

const STALE_MS = 60_000;

export function NowPlaying() {
  const { data, isLoading, isError } = useNowPlaying();

  if (isLoading) return <p>Loading...</p>;
  if (isError) return <p>Failed to load now playing.</p>;
  if (!data?.track) return <p>No track data yet.</p>;

  const { name, artist_name, album_name, updated_at } = data.track;
  const isStale = Date.now() - new Date(updated_at).getTime() > STALE_MS;

  return (
    <p>
      {isStale ? "Last played" : "Now playing"}: {name} by {artist_name} (
      {album_name}){isStale && ` (${timeAgo(new Date(updated_at))})`}
    </p>
  );
}
