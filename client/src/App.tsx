import { useEffect, useState } from "react";
import { Routes, Route, Link } from "react-router-dom";
import { fetchStatus } from "./api/client";
import { NowPlaying } from "./components/NowPlaying";
import { ImageGrid } from "./components/ImageGrid";
import { StatusDot } from "./components/StatusDot";
import { SuggestionModal } from "./components/SuggestionModal";
import { StatsPage } from "./pages/StatsPage";
import { GatePage } from "./pages/GatePage";

function HomePage() {
  const [connected, setConnected] = useState<boolean | null>(null);
  const [modalOpen, setModalOpen] = useState(false);

  useEffect(() => {
    fetchStatus()
      .then((data) => setConnected(data.connected))
      .catch(() => setConnected(false));
  }, []);

  if (connected === null) {
    return null;
  }

  return (
    <div className="app-container">
      <h1 className="app-title">listening-log</h1>
      <p className="app-description">
        spotify could but they wont so i did.
        <br/>
        <br/>
        leave me a song{" "}
        <span className="suggestion-link" onClick={() => setModalOpen(true)}>
          suggestion
        </span>
         {" "}and i will let you see my{" "}
        <Link to="/stats" className="suggestion-link">
          stats
        </Link>
        .
      </p>
      <NowPlaying />
      <ImageGrid />
      <StatusDot connected={connected} />
      <SuggestionModal open={modalOpen} onClose={() => setModalOpen(false)} />
    </div>
  );
}

export default function App() {
  return (
    <Routes>
      <Route path="/" element={<HomePage />} />
      <Route path="/stats" element={<StatsPage />} />
      <Route path="/woah-hold-it-right-there-buckaroo" element={<GatePage />} />
    </Routes>
  );
}
