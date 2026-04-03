import React from "react";
import type { FraudScore } from "../types";
import ScoreIndicator from "./ScoreIndicator";
import Card from "./ui/Card";
import EmptyState from "./ui/EmptyState";
import ErrorBanner from "./ErrorBanner";

interface FraudScoreSectionProps {
  score: FraudScore | null;
  notFound: boolean;
  loading: boolean;
  error: string | null;
  onRetry: () => void;
}

const FraudScoreSection: React.FC<FraudScoreSectionProps> = ({
  score,
  notFound,
  loading,
  error,
  onRetry,
}) => (
  <>
    <h3>Fraud Score</h3>
    {error ? <ErrorBanner message={error} onRetry={onRetry} /> : null}
    {loading ? <div data-testid="score-loading">Loading fraud score…</div> : null}
    {!loading && !error && notFound ? (
      <EmptyState message="No fraud score computed" />
    ) : null}
    {!loading && !error && score ? (
      <Card>
        <ScoreIndicator score={score.fraud_score} detailed />
      </Card>
    ) : null}
  </>
);

export default React.memo(FraudScoreSection);
