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
          gridTemplateColumns: "repeat(4, 1fr)",
          gap: "4px",
          marginTop: "1rem",
        }}
      >
        {images.map((img, i) => (
          <img
            key={`${img.url}-${i}`}
            src={img.url}
            alt={img.album_name}
            style={{
              width: "100%",
              aspectRatio: "1",
              objectFit: "cover",
              borderRadius: "4px",
            }}
          />
        ))}
      </div>
    </div>
  );
}
