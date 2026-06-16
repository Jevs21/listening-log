const SERVER_URL = import.meta.env.DEV ? "http://127.0.0.1:8080" : "";

const dotStyle = {
  display: "inline-block",
  width: 12,
  height: 12,
  borderRadius: "50%",
};

export function StatusDot({ connected }: { connected: boolean }) {
  return (
    <div style={{ textAlign: "center", marginTop: "2rem" }}>
      {connected ? (
        <span style={{ ...dotStyle, backgroundColor: "#22c55e" }} />
      ) : (
        <a href={`${SERVER_URL}/api/auth/login`}>
          <span style={{ ...dotStyle, backgroundColor: "#ef4444", cursor: "pointer" }} />
        </a>
      )}
    </div>
  );
}
