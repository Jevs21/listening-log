export async function getDashboardURL(): Promise<{ url: string }> {
  const res = await fetch("/api/stats/dashboard");
  if (!res.ok) throw new Error("dashboard not available");
  return res.json();
}
