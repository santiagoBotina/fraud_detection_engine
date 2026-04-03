import React from "react";
import type { Transaction } from "../types";
import StatusBadge from "./StatusBadge";
import FieldRow from "./ui/FieldRow";
import Card from "./ui/Card";
import { formatCurrency, formatDate, computeEvaluationTime } from "../utils/formatters";

interface TransactionFieldsProps {
  transaction: Transaction;
}

const TransactionFields: React.FC<TransactionFieldsProps> = ({ transaction }) => {
  const evalTime = computeEvaluationTime(transaction.created_at, transaction.updated_at);

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
      <FieldRow label="Evaluation Time" last>{evalTime} ms</FieldRow>
    </Card>
  );
};

export default React.memo(TransactionFields);
