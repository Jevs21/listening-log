export async function submitSuggestion(link: string, message: string, source: "home" | "gate"): Promise<Response> {
  return fetch("/api/suggestions", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ link, message, source }),
  });
}

export async function checkSuggestion(): Promise<{ has_suggested: boolean }> {
  const res = await fetch("/api/suggestions/check");
  return res.json();
}
