import { useState } from "react";
import { useImageGrid } from "../hooks/useImageGrid";

type Mode = "tracks" | "albums";

export function ImageGrid() {
  const [mode, setMode] = useState<Mode>("tracks");
  const { data } = useImageGrid(mode);
  const images = data?.images ?? [];

  return (
    <div>
      <select
        value={mode}
        onChange={(e) => setMode(e.target.value as Mode)}
      >
        <option value="tracks">Recent Tracks</option>
        <option value="albums">Recent Albums</option>
      </select>
      <div
        style={{
          display: "grid",
          gridTemplateColumns: "repeat(4, 64px)",
          gap: "4px",
          justifyContent: "center",
          marginTop: "1rem",
        }}
      >
        {images.map((img, i) => (
          <img
            key={`${img.url}-${i}`}
            src={img.url}
            alt={img.album_name}
            width={64}
            height={64}
          />
        ))}
      </div>
    </div>
  );
}
