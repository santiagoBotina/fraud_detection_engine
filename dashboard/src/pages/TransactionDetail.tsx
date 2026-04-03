import React from "react";
import { useParams, Link } from "react-router-dom";
import { useTransaction } from "../hooks/useTransaction";
import { useEvaluations } from "../hooks/useEvaluations";
import { useFraudScore } from "../hooks/useFraudScore";
import PageWrapper from "../components/ui/PageWrapper";
import ErrorBanner from "../components/ErrorBanner";
import TransactionFields from "../components/TransactionFields";
import RuleEvaluationsTable from "../components/RuleEvaluationsTable";
import FraudScoreSection from "../components/FraudScoreSection";

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

  // async-parallel: all three hooks fire in parallel on mount
  const { transaction, loading: txnLoading, error: txnError, retry: retryTxn } = useTransaction(id);
  const { evaluations, loading: evalLoading, error: evalError, retry: retryEval } = useEvaluations(id);
  const { score, notFound: scoreNotFound, loading: scoreLoading, error: scoreError, retry: retryScore } = useFraudScore(id);

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
      <h2>Transaction {transaction.id}</h2>

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
