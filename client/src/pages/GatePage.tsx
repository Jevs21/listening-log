import { useNavigate } from "react-router-dom";
import { SuggestionForm } from "../components/SuggestionForm";

export function GatePage() {
  const navigate = useNavigate();

  return (
    <div className="app-container">
      <h1 className="app-title">listening-log</h1>
      <p className="app-description">
        woah hold on a second. gonna need that song suggestion first. you thought
        i wasn't gonna check but this is really a ploy to get song suggestions.
        jokes on you.
      </p>
      <SuggestionForm onSuccess={() => navigate("/stats")} />
    </div>
  );
}
