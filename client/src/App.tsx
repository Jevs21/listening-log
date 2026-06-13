import { useEffect, useState } from "react";
import { fetchStatus } from "./api/client";

const SERVER_URL = import.meta.env.DEV ? "http://127.0.0.1:8080" : "";

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

  if (connected) {
    return (
      <div style={{ textAlign: "center", marginTop: "4rem" }}>
        <h1>Listening Log</h1>
        <p>Spotify connected</p>
      </div>
    );
  }

  return (
    <div style={{ textAlign: "center", marginTop: "4rem" }}>
      <h1>Listening Log</h1>
      <a href={`${SERVER_URL}/api/auth/login`}>
        <button style={{ padding: "0.75rem 1.5rem", fontSize: "1rem", cursor: "pointer" }}>
          Connect Spotify
        </button>
      </a>
    </div>
  );
}
