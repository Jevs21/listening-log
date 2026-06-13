export async function fetchStatus(): Promise<{ connected: boolean }> {
  const res = await fetch("/api/status");
  return res.json();
}
