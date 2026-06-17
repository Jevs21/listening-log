import { useState, useRef, useEffect } from "react";
import { submitSuggestion } from "../api/suggestions";
import "./SuggestionModal.css";

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
      className="modal-overlay"
      onClick={(e) => e.target === overlayRef.current && onClose()}
    >
      <div className="modal-panel">
        <button onClick={onClose} className="modal-close" aria-label="Close">
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
              className="modal-input"
              value={link}
              onChange={(e) => setLink(e.target.value)}
              placeholder="spotify, apple music, or youtube link"
              maxLength={2048}
              style={{ marginBottom: "0.75rem" }}
            />
            <textarea
              className="modal-input"
              value={message}
              onChange={(e) => setMessage(e.target.value)}
              placeholder="or leave a message"
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
            <button type="submit" disabled={!canSubmit} className="modal-submit">
              {submitting ? "submitting..." : "submit"}
            </button>
          </form>
        )}
      </div>
    </div>
  );
}
