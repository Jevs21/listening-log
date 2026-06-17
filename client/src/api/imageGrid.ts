export const IMAGE_GRID_MAX = 52;
export const IMAGE_GRID_POLL_MS = 30_000;

export interface ImageGridItem {
  url: string;
  album_name: string;
  track_name?: string;
  artist_name: string;
  updated_at: string;
}

export interface ImageGridResponse {
  images: ImageGridItem[];
}

export async function fetchImageGrid(
  mode: "tracks" | "albums"
): Promise<ImageGridResponse> {
  const res = await fetch(`/api/image-grid?mode=${mode}`);
  return res.json();
}
