import React, { useState, useEffect, useCallback } from "react";
import { useParams, Link } from "react-router-dom";
import { fetchTransaction, Transaction } from "../api/transactions";
import { fetchEvaluations, RuleEvaluationResult } from "../api/evaluations";
import { fetchScore, FraudScore } from "../api/scores";
import ScoreIndicator from "../components/ScoreIndicator";
import StatusBadge from "../components/StatusBadge";
import ErrorBanner from "../components/ErrorBanner";
import { formatCurrency, formatDate, computeEvaluationTime } from "../utils/formatters";

const wrapperStyle: React.CSSProperties = {
  maxWidth: "920px",
  margin: "0 auto",
  fontFamily: "'Inter', sans-serif",
};

const cardStyle: React.CSSProperties = {
  backgroundColor: "var(--color-surface)",
  border: "1px solid var(--color-border)",
  borderRadius: "12px",
  boxShadow: "var(--shadow-sm)",
  padding: "24px",
  marginBottom: "24px",
};

const scrollCardStyle: React.CSSProperties = {
  backgroundColor: "var(--color-surface)",
  border: "1px solid var(--color-border)",
  borderRadius: "12px",
  boxShadow: "var(--shadow-sm)",
  marginBottom: "24px",
  overflowX: "auto",
};

const tableStyle: React.CSSProperties = {
  width: "100%",
  minWidth: "650px",
  borderCollapse: "collapse",
  fontSize: "0.875rem",
  tableLayout: "auto",
};

const thStyle: React.CSSProperties = {
  textAlign: "left",
  padding: "10px 16px",
  borderBottom: "1px solid var(--color-border)",
  fontWeight: 600,
  fontSize: "0.75rem",
  textTransform: "uppercase",
  letterSpacing: "0.04em",
  color: "var(--color-text-muted)",
  whiteSpace: "nowrap",
  resize: "horizontal",
  overflow: "hidden",
};

const tdStyle: React.CSSProperties = {
  padding: "10px 16px",
  borderBottom: "1px solid var(--color-border-muted)",
  whiteSpace: "nowrap",
};

const fieldRowStyle: React.CSSProperties = {
  display: "flex",
  padding: "8px 0",
  borderBottom: "1px solid var(--color-border-muted)",
  fontSize: "0.875rem",
};

const fieldLabelStyle: React.CSSProperties = {
  fontWeight: 600,
  width: "180px",
  flexShrink: 0,
  color: "var(--color-text-muted)",
  fontSize: "0.8rem",
  textTransform: "uppercase",
  letterSpacing: "0.03em",
};

const backLinkStyle: React.CSSProperties = {
  display: "inline-flex",
  alignItems: "center",
  gap: "4px",
  fontSize: "0.875rem",
  fontWeight: 500,
  marginBottom: "20px",
  color: "var(--color-text-muted)",
};

const condFieldStyle: React.CSSProperties = {
  display: "inline-block",
  padding: "2px 6px",
  borderRadius: "4px",
  backgroundColor: "#dbeafe",
  color: "#1e40af",
  fontFamily: "monospace",
  fontSize: "0.78rem",
  fontWeight: 500,
};

const condOperatorStyle: React.CSSProperties = {
  display: "inline-block",
  padding: "2px 6px",
  borderRadius: "4px",
  backgroundColor: "#f3e8ff",
  color: "#6b21a8",
  fontFamily: "monospace",
  fontSize: "0.78rem",
  fontWeight: 600,
};

const condValueStyle: React.CSSProperties = {
  display: "inline-block",
  padding: "2px 6px",
  borderRadius: "4px",
  backgroundColor: "#fef3c7",
  color: "#92400e",
  fontFamily: "monospace",
  fontSize: "0.78rem",
  fontWeight: 500,
};

const priorityBadgeStyle: React.CSSProperties = {
  display: "inline-flex",
  alignItems: "center",
  justifyContent: "center",
  width: "28px",
  height: "28px",
  borderRadius: "50%",
  backgroundColor: "var(--color-border-muted, #ededf0)",
  color: "var(--color-text-muted)",
  fontSize: "0.75rem",
  fontWeight: 700,
  fontFamily: "'Inter', sans-serif",
};

const TransactionDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();

  const [transaction, setTransaction] = useState<Transaction | null>(null);
  const [txnLoading, setTxnLoading] = useState(true);
  const [txnError, setTxnError] = useState<string | null>(null);

  const [evaluations, setEvaluations] = useState<RuleEvaluationResult[]>([]);
  const [evalLoading, setEvalLoading] = useState(true);
  const [evalError, setEvalError] = useState<string | null>(null);

  const [score, setScore] = useState<FraudScore | null>(null);
  const [scoreNotFound, setScoreNotFound] = useState(false);
  const [scoreLoading, setScoreLoading] = useState(true);
  const [scoreError, setScoreError] = useState<string | null>(null);

  const loadTransaction = useCallback(async () => {
    if (!id) return;
    setTxnLoading(true);
    setTxnError(null);
    try {
      const response = await fetchTransaction(id);
      setTransaction(response.data);
    } catch {
      setTxnError("Unable to connect to Transaction service");
    } finally {
      setTxnLoading(false);
    }
  }, [id]);

  const loadEvaluations = useCallback(async () => {
    if (!id) return;
    setEvalLoading(true);
    setEvalError(null);
    try {
      const response = await fetchEvaluations(id);
      setEvaluations(response.data);
    } catch {
      setEvalError("Unable to load rule evaluations");
    } finally {
      setEvalLoading(false);
    }
  }, [id]);

  const loadScore = useCallback(async () => {
    if (!id) return;
    setScoreLoading(true);
    setScoreError(null);
    setScoreNotFound(false);
    try {
      const result = await fetchScore(id);
      setScore(result);
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : "";
      if (message.includes("404")) {
        setScoreNotFound(true);
      } else {
        setScoreError("Unable to load fraud score");
      }
    } finally {
      setScoreLoading(false);
    }
  }, [id]);

  useEffect(() => {
    loadTransaction();
    loadEvaluations();
    loadScore();
  }, [loadTransaction, loadEvaluations, loadScore]);

  if (txnLoading) {
    return <div data-testid="loading" style={wrapperStyle}>Loading transaction…</div>;
  }

  if (txnError) {
    return <div style={wrapperStyle}><ErrorBanner message={txnError} onRetry={loadTransaction} /></div>;
  }

  if (!transaction) {
    return <div style={wrapperStyle}>Transaction not found</div>;
  }

  const evalTime = computeEvaluationTime(transaction.created_at, transaction.updated_at);

  return (
    <div style={wrapperStyle}>
      <Link to="/" style={backLinkStyle}>← Back to Transactions</Link>
      <h2>Transaction {transaction.id}</h2>

      <div style={cardStyle}>
        <h3 style={{ marginBottom: "16px" }}>Details</h3>
        <div style={fieldRowStyle}>
          <span style={fieldLabelStyle}>Status</span>
          <span><StatusBadge status={transaction.status} /></span>
        </div>
        <div style={fieldRowStyle}>
          <span style={fieldLabelStyle}>Amount</span>
          <span>{formatCurrency(transaction.amount_in_cents, transaction.currency)}</span>
        </div>
        <div style={fieldRowStyle}>
          <span style={fieldLabelStyle}>Currency</span>
          <span>{transaction.currency}</span>
        </div>
        <div style={fieldRowStyle}>
          <span style={fieldLabelStyle}>Payment Method</span>
          <span>{transaction.payment_method}</span>
        </div>
        <div style={fieldRowStyle}>
          <span style={fieldLabelStyle}>Customer ID</span>
          <span>{transaction.customer_id}</span>
        </div>
        <div style={fieldRowStyle}>
          <span style={fieldLabelStyle}>Customer Name</span>
          <span>{transaction.customer_name}</span>
        </div>
        <div style={fieldRowStyle}>
          <span style={fieldLabelStyle}>Customer Email</span>
          <span>{transaction.customer_email}</span>
        </div>
        <div style={fieldRowStyle}>
          <span style={fieldLabelStyle}>Customer Phone</span>
          <span>{transaction.customer_phone}</span>
        </div>
        <div style={fieldRowStyle}>
          <span style={fieldLabelStyle}>Customer IP</span>
          <span>{transaction.customer_ip_address}</span>
        </div>
        <div style={fieldRowStyle}>
          <span style={fieldLabelStyle}>Created At</span>
          <span>{formatDate(transaction.created_at)}</span>
        </div>
        <div style={fieldRowStyle}>
          <span style={fieldLabelStyle}>Updated At</span>
          <span>{formatDate(transaction.updated_at)}</span>
        </div>
        <div style={{ ...fieldRowStyle, borderBottom: "none" }}>
          <span style={fieldLabelStyle}>Evaluation Time</span>
          <span>{evalTime} ms</span>
        </div>
      </div>

      <h3>Rule Evaluations</h3>
      {evalError && <ErrorBanner message={evalError} onRetry={loadEvaluations} />}
      {evalLoading && <div data-testid="eval-loading">Loading evaluations…</div>}
      {!evalLoading && !evalError && evaluations.length === 0 && (
        <div style={{ ...cardStyle, color: "var(--color-text-muted)", textAlign: "center" }}>
          No rule evaluations found
        </div>
      )}
      {!evalLoading && !evalError && evaluations.length > 0 && (
        <div style={scrollCardStyle}>
          <table style={tableStyle}>
            <thead>
              <tr>
                <th style={thStyle}>Priority</th>
                <th style={thStyle}>Rule Name</th>
                <th style={thStyle}>Condition</th>
                <th style={thStyle}>Matched</th>
                <th style={thStyle}>Result Status</th>
              </tr>
            </thead>
            <tbody>
              {evaluations.map((ev) => (
                <tr key={ev.rule_id}>
                  <td style={{ ...tdStyle, textAlign: "center" }}>
                    <span style={priorityBadgeStyle}>{ev.priority}</span>
                  </td>
                  <td style={{ ...tdStyle, fontWeight: 500 }}>{ev.rule_name}</td>
                  <td style={tdStyle}>
                    <span style={condFieldStyle}>{ev.condition_field}</span>
                    {" "}
                    <span style={condOperatorStyle}>{ev.condition_operator}</span>
                    {" "}
                    <span style={condValueStyle}>{ev.condition_value}</span>
                  </td>
                  <td style={tdStyle}>
                    <span style={{
                      display: "inline-block",
                      width: "8px",
                      height: "8px",
                      borderRadius: "50%",
                      backgroundColor: ev.matched ? "#a3d9a5" : "#f5a3a3",
                      marginRight: "6px",
                    }} />
                    {ev.matched ? "Yes" : "No"}
                  </td>
                  <td style={tdStyle}>{ev.result_status}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      <h3>Fraud Score</h3>
      {scoreError && <ErrorBanner message={scoreError} onRetry={loadScore} />}
      {scoreLoading && <div data-testid="score-loading">Loading fraud score…</div>}
      {!scoreLoading && !scoreError && scoreNotFound && (
        <div style={{ ...cardStyle, color: "var(--color-text-muted)", textAlign: "center" }}>
          No fraud score computed
        </div>
      )}
      {!scoreLoading && !scoreError && score && (
        <div style={cardStyle}>
          <ScoreIndicator score={score.fraud_score} detailed />
        </div>
      )}
    </div>
  );
};

export default TransactionDetail;
