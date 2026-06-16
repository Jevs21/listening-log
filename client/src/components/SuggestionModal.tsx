import { useState, useRef, useEffect } from "react";
import { submitSuggestion } from "../api/suggestions";

interface Props {
  open: boolean;
  onClose: () => void;
}

export function SuggestionModal({ open, onClose }: Props) {
  const [link, setLink] = useState("");
  const [message, setMessage] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState(false);
  const overlayRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!open) {
      setLink("");
      setMessage("");
      setError("");
      setSuccess(false);
      setSubmitting(false);
    }
  }, [open]);

  useEffect(() => {
    if (success) {
      const timer = setTimeout(onClose, 2000);
      return () => clearTimeout(timer);
    }
  }, [success, onClose]);

  if (!open) return null;

  const canSubmit = (link.trim() !== "" || message.trim() !== "") && !submitting;

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    setSubmitting(true);

    try {
      const res = await submitSuggestion(link.trim(), message.trim());
      if (res.ok) {
        setSuccess(true);
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
    <div
      ref={overlayRef}
      onClick={(e) => e.target === overlayRef.current && onClose()}
      style={{
        position: "fixed",
        inset: 0,
        backgroundColor: "rgba(0,0,0,0.7)",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        zIndex: 1000,
      }}
    >
      <div
        style={{
          background: "#1a1a1a",
          border: "1px solid #333",
          borderRadius: 8,
          padding: "1.5rem",
          width: "min(90vw, 340px)",
          position: "relative",
          color: "#e0e0e0",
        }}
      >
        <button
          onClick={onClose}
          style={{
            position: "absolute",
            top: 8,
            right: 12,
            background: "none",
            border: "none",
            color: "#888",
            fontSize: "1.2rem",
            cursor: "pointer",
          }}
          aria-label="Close"
        >
          ✕
        </button>

        {success ? (
          <p style={{ textAlign: "center", fontSize: "1.2rem", margin: "2rem 0" }}>
            thx 🫶
          </p>
        ) : (
          <form onSubmit={handleSubmit}>
            <input
              type="text"
              value={link}
              onChange={(e) => setLink(e.target.value)}
              placeholder="spotify, apple music, or youtube link"
              maxLength={2048}
              style={{
                width: "100%",
                padding: "0.5rem",
                marginBottom: "0.75rem",
                background: "#111",
                border: "1px solid #333",
                borderRadius: 4,
                color: "#e0e0e0",
                fontSize: "0.9rem",
                boxSizing: "border-box",
              }}
            />
            <textarea
              value={message}
              onChange={(e) => setMessage(e.target.value)}
              placeholder="or leave a message"
              maxLength={200}
              rows={3}
              style={{
                width: "100%",
                padding: "0.5rem",
                background: "#111",
                border: "1px solid #333",
                borderRadius: 4,
                color: "#e0e0e0",
                fontSize: "0.9rem",
                resize: "none",
                boxSizing: "border-box",
              }}
            />
            <div
              style={{
                textAlign: "right",
                fontSize: "0.75rem",
                color: "#666",
                marginBottom: "0.75rem",
              }}
            >
              {message.length}/200
            </div>
            {error && (
              <p style={{ color: "#ef4444", fontSize: "0.85rem", marginBottom: "0.5rem" }}>
                {error}
              </p>
            )}
            <button
              type="submit"
              disabled={!canSubmit}
              style={{
                width: "100%",
                padding: "0.5rem",
                background: canSubmit ? "#e0e0e0" : "#444",
                color: canSubmit ? "#111" : "#888",
                border: "none",
                borderRadius: 4,
                fontSize: "0.9rem",
                cursor: canSubmit ? "pointer" : "default",
              }}
            >
              {submitting ? "submitting..." : "submit"}
            </button>
          </form>
        )}
      </div>
    </div>
  );
}
