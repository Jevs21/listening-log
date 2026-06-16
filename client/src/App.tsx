import { useEffect, useState } from "react";
import { fetchStatus } from "./api/client";
import { NowPlaying } from "./components/NowPlaying";
import { ImageGrid } from "./components/ImageGrid";
import { StatusDot } from "./components/StatusDot";

export default function App() {
  const [connected, setConnected] = useState<boolean | null>(null);

  useEffect(() => {
    fetchStatus()
      .then((data) => setConnected(data.connected))
      .catch(() => setConnected(false));
  }, []);

  if (connected === null) {
    return <p>Loading...</p>;
  }

  return (
    <div style={{ maxWidth: "300px", margin: "4rem auto 0" }}>
      <h1>Listening Log</h1>
      <p>i built this so id never forget a song i listen to. id love if you left me a song suggestion.</p>
      <NowPlaying />
      <ImageGrid />
      <StatusDot connected={connected} />
    </div>
  );
}
