import { useEffect, useState } from "react";
import { fetchStatus } from "./api/client";
import { NowPlaying } from "./components/NowPlaying";
import { ImageGrid } from "./components/ImageGrid";
import { StatusDot } from "./components/StatusDot";
import { SuggestionModal } from "./components/SuggestionModal";

export default function App() {
  const [connected, setConnected] = useState<boolean | null>(null);
  const [modalOpen, setModalOpen] = useState(false);

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
      <p>
        i built this so id never forget a song i listen to. id love if you left me a song{" "}
        <span
          onClick={() => setModalOpen(true)}
          style={{ textDecoration: "underline", cursor: "pointer" }}
        >
          suggestion
        </span>
        .
      </p>
      <NowPlaying />
      <ImageGrid />
      <StatusDot connected={connected} />
      <SuggestionModal open={modalOpen} onClose={() => setModalOpen(false)} />
    </div>
  );
}
