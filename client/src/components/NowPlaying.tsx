import { useState, useEffect } from "react";
import { useNowPlaying } from "../hooks/useNowPlaying";
import { timeAgo } from "../utils/timeAgo";
import "./NowPlaying.css";

const STALE_MS = 20_000;

export function NowPlaying() {
  const { data, isLoading, isError } = useNowPlaying();
  const [now, setNow] = useState(Date.now());

  useEffect(() => {
    const id = setInterval(() => setNow(Date.now()), 5_000);
    return () => clearInterval(id);
  }, []);

  if (isLoading) return <p>Loading...</p>;
  if (isError) return <p>Failed to load now playing.</p>;
  if (!data?.track) return <p>No track data yet.</p>;

  const { name, artist_name, album_name, album_image_url, updated_at } =
    data.track;
  const isStale = now - new Date(updated_at).getTime() > STALE_MS;

  return (
    <div className="now-playing">
      <div className="record-container">
        <div className={`record ${isStale ? "" : "spinning"}`}>
          <div className="grooves" />
          <div className="label">
            {album_image_url && (
              <img src={album_image_url} alt={`${album_name} cover`} />
            )}
          </div>
        </div>
      </div>
      <div className="track-info">
        <p className="track-name">{name}</p>
        <p className="track-artist">
          by {artist_name} &mdash; {album_name}
        </p>
        <p className="track-status">
          {isStale
            ? `Last played (${timeAgo(new Date(updated_at))})`
            : "Now playing"}
        </p>
      </div>
    </div>
  );
}
