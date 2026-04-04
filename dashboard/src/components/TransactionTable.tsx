import React from "react";
import { Link } from "react-router-dom";
import type { Transaction } from "../types";
import LatencyBadge from "./LatencyBadge";
import StatusBadge from "./StatusBadge";
import Skeleton from "./ui/Skeleton";
import { formatCurrency, formatDate } from "../utils/formatters";

interface TransactionTableProps {
  transactions: Transaction[];
  loading: boolean;
}

const cardStyle: React.CSSProperties = {
  backgroundColor: "var(--color-surface)",
  border: "1px solid var(--color-border)",
  borderRadius: "12px",
  boxShadow: "var(--shadow-sm)",
  overflow: "hidden",
};

const tableStyle: React.CSSProperties = {
  width: "100%",
  borderCollapse: "collapse",
  fontFamily: "'Inter', sans-serif",
  fontSize: "0.875rem",
};

const thStyle: React.CSSProperties = {
  textAlign: "left",
  padding: "12px 16px",
  borderBottom: "1px solid var(--color-border)",
  fontWeight: 600,
  fontSize: "0.75rem",
  textTransform: "uppercase",
  letterSpacing: "0.04em",
  color: "var(--color-text-muted)",
};

const tdStyle: React.CSSProperties = {
  padding: "12px 16px",
  borderBottom: "1px solid var(--color-border-muted)",
  color: "var(--color-text)",
};

const COLUMNS = [
  "Transaction ID",
  "Amount",
  "Currency",
  "Payment Method",
  "Customer Name",
  "Status",
  "Latency",
  "Created At",
] as const;

const SKELETON_ROW_COUNT = 5;

function SkeletonRows() {
  return (
    <>
      {Array.from({ length: SKELETON_ROW_COUNT }, (_, i) => (
        <tr key={`skeleton-${i}`} data-testid="skeleton-row">
          {COLUMNS.map((_, j) => (
            <td key={j} style={tdStyle}>
              <Skeleton />
            </td>
          ))}
        </tr>
      ))}
    </>
  );
}

function TransactionRow({ txn }: { txn: Transaction }) {
  return (
    <tr style={{ transition: "background-color 100ms ease" }}>
      <td style={tdStyle}>
        <Link to={`/transactions/${txn.id}`} style={{ fontWeight: 500 }}>{txn.id}</Link>
      </td>
      <td style={{ ...tdStyle, fontVariantNumeric: "tabular-nums" }}>
        {formatCurrency(txn.amount_in_cents, txn.currency)}
      </td>
      <td style={tdStyle}>{txn.currency}</td>
      <td style={tdStyle}>{txn.payment_method}</td>
      <td style={tdStyle}>{txn.customer_name}</td>
      <td style={tdStyle}><StatusBadge status={txn.status} /></td>
      <td style={tdStyle}><LatencyBadge latencyMs={txn.finalization_latency_ms} /></td>
      <td style={{ ...tdStyle, color: "var(--color-text-muted)" }}>{formatDate(txn.created_at)}</td>
    </tr>
  );
}

const TransactionTable: React.FC<TransactionTableProps> = ({ transactions, loading }) => (
  <div style={cardStyle}>
    <table style={tableStyle}>
      <thead>
        <tr>
          {COLUMNS.map((col) => (
            <th key={col} style={thStyle}>{col}</th>
          ))}
        </tr>
      </thead>
      <tbody>
        {loading ? (
          <SkeletonRows />
        ) : (
          transactions.map((txn) => <TransactionRow key={txn.id} txn={txn} />)
        )}
      </tbody>
    </table>
  </div>
);

export default React.memo(TransactionTable);
