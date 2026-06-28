import { useState } from "react";
import { submitSuggestion } from "../api/suggestions";
import "./SuggestionForm.css";

interface Props {
  source: "home" | "gate";
  onSuccess?: () => void;
}

export function SuggestionForm({ source, onSuccess }: Props) {
  const [link, setLink] = useState("");
  const [message, setMessage] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState(false);

  if (success) {
    return (
      <p style={{ textAlign: "center", fontSize: "1.2rem", margin: "2rem 0" }}>
        thx 🫶
      </p>
    );
  }

  function isValidUrl(str: string): boolean {
    try {
      const url = new URL(str);
      return url.protocol === "http:" || url.protocol === "https:";
    } catch {
      return false;
    }
  }

  const linkTrimmed = link.trim();
  const messageTrimmed = message.trim();
  const linkValid = linkTrimmed === "" || isValidUrl(linkTrimmed);
  const hasContent = linkTrimmed !== "" || messageTrimmed !== "";
  const canSubmit = hasContent && linkValid && !submitting;

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (linkTrimmed !== "" && !isValidUrl(linkTrimmed)) {
      setError("please enter a valid url");
      return;
    }
    setError("");
    setSubmitting(true);

    try {
      const res = await submitSuggestion(link.trim(), message.trim(), source);
      if (res.ok) {
        setSuccess(true);
        if (onSuccess) {
          setTimeout(onSuccess, 2000);
        }
      } else if (res.status === 429) {
        setError("too many suggestions, try again later");
      } else {
        setError("something went wrong");
      }
    } catch {
      setError("something went wrong");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <form onSubmit={handleSubmit}>
      <input
        type="text"
        className="suggestion-input"
        value={link}
        onChange={(e) => setLink(e.target.value)}
        placeholder="spotify, youtube, or any link (optional)"
        maxLength={2048}
        style={{ marginBottom: "0.75rem" }}
      />
      <textarea
        className="suggestion-input"
        value={message}
        onChange={(e) => setMessage(e.target.value)}
        placeholder="song name, artist, or a message"
        maxLength={200}
        rows={3}
        style={{ resize: "none" }}
      />
      <div
        style={{
          textAlign: "right",
          fontSize: "0.75rem",
          color: "var(--text-muted)",
          marginBottom: "0.75rem",
        }}
      >
        {message.length}/200
      </div>
      {error && (
        <p style={{ color: "var(--error)", fontSize: "0.85rem", marginBottom: "0.5rem" }}>
          {error}
        </p>
      )}
      <button type="submit" disabled={!canSubmit} className="suggestion-submit">
        {submitting ? "submitting..." : "submit"}
      </button>
    </form>
  );
}
