import { useQuery } from "@tanstack/react-query";
import { fetchNowPlaying } from "../api/nowPlaying";

export function useNowPlaying() {
  return useQuery({
    queryKey: ["now-playing"],
    queryFn: fetchNowPlaying,
    refetchInterval: 10_000,
  });
}
