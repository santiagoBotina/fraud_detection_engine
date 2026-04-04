import React, { useState } from "react";
import { useParams, Link } from "react-router-dom";
import { useTransaction } from "../hooks/useTransaction";
import { useEvaluations } from "../hooks/useEvaluations";
import { useFraudScore } from "../hooks/useFraudScore";
import PageWrapper from "../components/ui/PageWrapper";
import ErrorBanner from "../components/ErrorBanner";
import TransactionFields from "../components/TransactionFields";
import RuleEvaluationsTable from "../components/RuleEvaluationsTable";
import FraudScoreSection from "../components/FraudScoreSection";
import Button from "../components/ui/Button";

const backLinkStyle: React.CSSProperties = {
  display: "inline-flex",
  alignItems: "center",
  gap: "4px",
  fontSize: "0.875rem",
  fontWeight: 500,
  marginBottom: "20px",
  color: "var(--color-text-muted)",
};

const TransactionDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const [refreshingAll, setRefreshingAll] = useState(false);

  // async-parallel: all three hooks fire in parallel on mount
  const { transaction, loading: txnLoading, error: txnError, retry: retryTxn } = useTransaction(id);
  const { evaluations, loading: evalLoading, error: evalError, retry: retryEval } = useEvaluations(id);
  const { score, notFound: scoreNotFound, loading: scoreLoading, error: scoreError, retry: retryScore } = useFraudScore(id);

  const handleRefresh = async () => {
    setRefreshingAll(true);
    await Promise.all([retryTxn(), retryEval(), retryScore()]);
    setRefreshingAll(false);
  };

  if (txnLoading) {
    return <PageWrapper><div data-testid="loading">Loading transaction…</div></PageWrapper>;
  }

  if (txnError) {
    return <PageWrapper><ErrorBanner message={txnError} onRetry={retryTxn} /></PageWrapper>;
  }

  if (!transaction) {
    return <PageWrapper>Transaction not found</PageWrapper>;
  }

  return (
    <PageWrapper>
      <Link to="/" style={backLinkStyle}>← Back to Transactions</Link>
      <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", marginBottom: "16px" }}>
        <h2 style={{ margin: 0 }}>Transaction {transaction.id}</h2>
        <Button loading={refreshingAll} onClick={handleRefresh} aria-label="Refresh transaction">
          {refreshingAll ? "Refreshing…" : "↻ Refresh"}
        </Button>
      </div>

      <TransactionFields transaction={transaction} />

      <RuleEvaluationsTable
        evaluations={evaluations}
        loading={evalLoading}
        error={evalError}
        onRetry={retryEval}
      />

      <FraudScoreSection
        score={score}
        notFound={scoreNotFound}
        loading={scoreLoading}
        error={scoreError}
        onRetry={retryScore}
      />
    </PageWrapper>
  );
};

export default TransactionDetail;
