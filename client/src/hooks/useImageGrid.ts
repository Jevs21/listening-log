import { useQuery } from "@tanstack/react-query";
import { fetchImageGrid, IMAGE_GRID_POLL_MS } from "../api/imageGrid";

export function useImageGrid(mode: "tracks" | "albums") {
  return useQuery({
    queryKey: ["image-grid", mode],
    queryFn: () => fetchImageGrid(mode),
    refetchInterval: IMAGE_GRID_POLL_MS,
  });
}
