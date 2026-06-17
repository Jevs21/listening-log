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
    <div className="app-container">
      <h1 className="app-title">listening log</h1>
      <p className="app-description">
        spotify could but they wont so i did. <br/> pls leave me a song{" "}
        <span className="suggestion-link" onClick={() => setModalOpen(true)}>
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
