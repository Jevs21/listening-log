import { useRef, useState, useCallback, type ReactNode } from "react";

function timeAgo(dateString: string): string {
  const now = Date.now();
  const then = new Date(dateString).getTime();
  const seconds = Math.floor((now - then) / 1000);

  if (seconds < 60) return "just now";
  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) return `${minutes}m ago`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}h ago`;
  const days = Math.floor(hours / 24);
  if (days < 7) return `${days}d ago`;
  const weeks = Math.floor(days / 7);
  return `${weeks}w ago`;
}

interface TooltipProps {
  title: string;
  subtitle: string;
  updatedAt: string;
  children: ReactNode;
}

export function Tooltip({ title, subtitle, updatedAt, children }: TooltipProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const [visible, setVisible] = useState(false);
  const [above, setAbove] = useState(true);

  const show = useCallback(() => {
    if (containerRef.current) {
      const rect = containerRef.current.getBoundingClientRect();
      setAbove(rect.top > 60);
    }
    setVisible(true);
  }, []);

  const hide = useCallback(() => setVisible(false), []);

  return (
    <div
      ref={containerRef}
      onMouseEnter={show}
      onMouseLeave={hide}
      style={{ position: "relative" }}
    >
      {children}
      {visible && (
        <div
          style={{
            position: "absolute",
            left: "50%",
            transform: "translateX(-50%)",
            ...(above
              ? { bottom: "calc(100% + 6px)" }
              : { top: "calc(100% + 6px)" }),
            background: "var(--surface)",
            border: "1px solid var(--border)",
            color: "var(--text)",
            fontSize: "0.75rem",
            padding: "0.4rem 0.6rem",
            borderRadius: "4px",
            pointerEvents: "none",
            zIndex: 10,
            whiteSpace: "nowrap",
          }}
        >
          <div style={{ fontFamily: "var(--font-heading)" }}>{title}</div>
          <div style={{ color: "var(--text-secondary)", fontSize: "0.7rem" }}>
            {subtitle} · {timeAgo(updatedAt)}
          </div>
        </div>
      )}
    </div>
  );
}
