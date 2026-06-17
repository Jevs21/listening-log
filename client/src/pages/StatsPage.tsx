import { useEffect, useState } from "react";
import { Navigate } from "react-router-dom";
import { checkSuggestion } from "../api/suggestions";

export function StatsPage() {
  const [hasSuggested, setHasSuggested] = useState<boolean | null>(null);

  useEffect(() => {
    checkSuggestion().then((data) => setHasSuggested(data.has_suggested));
  }, []);

  if (hasSuggested === null) {
    return null;
  }

  if (!hasSuggested) {
    return <Navigate to="/woah-hold-it-right-there-buckaroo" replace />;
  }

  return (
    <div className="app-container">
      <h1 className="app-title">listening-log</h1>
      <p className="app-description">stats coming soon</p>
    </div>
  );
}
