import { useEffect, useState } from "react";
import { Navigate } from "react-router-dom";
import { checkSuggestion } from "../api/suggestions";
import { getDashboardURL } from "../api/stats";
import { Loading } from "../components/Loading";

export function StatsPage() {
  const [hasSuggested, setHasSuggested] = useState<boolean | null>(null);
  const [dashboardURL, setDashboardURL] = useState<string | null>(null);

  useEffect(() => {
    checkSuggestion().then((data) => setHasSuggested(data.has_suggested));
    getDashboardURL().then((data) => setDashboardURL(data.url)).catch(() => {});
  }, []);

  if (hasSuggested === null) return null;
  if (!hasSuggested) return <Navigate to="/woah-hold-it-right-there-buckaroo" replace />;
  if (!dashboardURL) return <Loading />;

  return (
    <div className="app-container" style={{ maxWidth: "90vw" }}>
      <h1 className="app-title" style={{ textAlign: "center" }}>
        listening-log
      </h1>
      <iframe
        src={dashboardURL}
        style={{ width: "100%", height: "calc(100vh - 80px)", border: "none" }}
        title="Listening Overview"
      />
    </div>
  );
}
