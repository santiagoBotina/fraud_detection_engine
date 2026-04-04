import React from "react";
import type { Transaction } from "../types";
import StatusBadge from "./StatusBadge";
import LatencyBadge from "./LatencyBadge";
import FieldRow from "./ui/FieldRow";
import Card from "./ui/Card";
import { formatCurrency, formatDate, computeEvaluationTime } from "../utils/formatters";

interface TransactionFieldsProps {
  transaction: Transaction;
}

const awaitingBadgeStyle: React.CSSProperties = {
  display: "inline-block",
  padding: "3px 12px",
  borderRadius: "9999px",
  fontSize: "0.75rem",
  fontWeight: 600,
  fontFamily: "'Inter', sans-serif",
  letterSpacing: "0.02em",
  textTransform: "uppercase",
  backgroundColor: "var(--color-pending, #f5d89a)",
  color: "var(--color-pending-text, #6b4f10)",
};

const AwaitingBadge = () => (
  <span style={awaitingBadgeStyle}>Awaiting decision</span>
);

const TransactionFields: React.FC<TransactionFieldsProps> = ({ transaction }) => {
  const evalTime = computeEvaluationTime(transaction.created_at, transaction.updated_at);
  const isFinalized = transaction.status === "APPROVED" || transaction.status === "DECLINED";

  return (
    <Card>
      <h3 style={{ marginBottom: "16px" }}>Details</h3>
      <FieldRow label="Status"><StatusBadge status={transaction.status} /></FieldRow>
      <FieldRow label="Amount">{formatCurrency(transaction.amount_in_cents, transaction.currency)}</FieldRow>
      <FieldRow label="Currency">{transaction.currency}</FieldRow>
      <FieldRow label="Payment Method">{transaction.payment_method}</FieldRow>
      <FieldRow label="Customer ID">{transaction.customer_id}</FieldRow>
      <FieldRow label="Customer Name">{transaction.customer_name}</FieldRow>
      <FieldRow label="Customer Email">{transaction.customer_email}</FieldRow>
      <FieldRow label="Customer Phone">{transaction.customer_phone}</FieldRow>
      <FieldRow label="Customer IP">{transaction.customer_ip_address}</FieldRow>
      <FieldRow label="Created At">{formatDate(transaction.created_at)}</FieldRow>
      <FieldRow label="Updated At">{formatDate(transaction.updated_at)}</FieldRow>
      <FieldRow label="Finalized At">
        {isFinalized && transaction.finalized_at ? formatDate(transaction.finalized_at) : <AwaitingBadge />}
      </FieldRow>
      <FieldRow label="Finalization Latency">
        {isFinalized ? <LatencyBadge latencyMs={transaction.finalization_latency_ms} /> : <AwaitingBadge />}
      </FieldRow>
      <FieldRow label="Evaluation Time" last>{evalTime} ms</FieldRow>
    </Card>
  );
};

export default React.memo(TransactionFields);
