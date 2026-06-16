export async function submitSuggestion(link: string, message: string): Promise<Response> {
  return fetch("/api/suggestions", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ link, message }),
  });
}
