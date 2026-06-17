import { useNavigate } from "react-router-dom";
import { SuggestionForm } from "../components/SuggestionForm";

export function GatePage() {
  const navigate = useNavigate();

  return (
    <div className="app-container">
      <div style={{ textAlign: "center" }}>
        <div style={{ fontSize: "6rem", marginBottom: "1rem" }}>🤨</div>
        <p className="app-description">oh you thought i wasn't gonna check?</p>
      </div>
      <SuggestionForm source="gate" onSuccess={() => navigate("/stats")} />
    </div>
  );
}
