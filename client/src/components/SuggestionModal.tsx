import { useRef } from "react";
import { SuggestionForm } from "./SuggestionForm";
import "./SuggestionModal.css";

interface Props {
  open: boolean;
  onClose: () => void;
}

export function SuggestionModal({ open, onClose }: Props) {
  const overlayRef = useRef<HTMLDivElement>(null);

  if (!open) return null;

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
        <SuggestionForm source="home" onSuccess={onClose} />
      </div>
    </div>
  );
}
