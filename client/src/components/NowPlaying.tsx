import { useNowPlaying } from "../hooks/useNowPlaying";

export function NowPlaying() {
  const { data, isLoading, isError } = useNowPlaying();

  if (isLoading) return <p>Loading...</p>;
  if (isError) return <p>Failed to load now playing.</p>;
  if (!data?.track) return <p>No track data yet.</p>;

  const { name, artist_name, album_name } = data.track;
  return (
    <p>
      Now playing: {name} by {artist_name} ({album_name})
    </p>
  );
}
