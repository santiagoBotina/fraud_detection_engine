import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import TransactionFields from "./TransactionFields";
import type { Transaction } from "../types";

const sampleTransaction: Transaction = {
  id: "txn_001",
  amount_in_cents: 15000,
  currency: "USD",
  payment_method: "CARD",
  customer_id: "cust_1",
  customer_name: "Jane Doe",
  customer_email: "jane@example.com",
  customer_phone: "+1234567890",
  customer_ip_address: "10.0.0.1",
  status: "APPROVED",
  created_at: "2025-01-15T10:30:00Z",
  updated_at: "2025-01-15T10:30:05Z",
};

describe("TransactionFields", () => {
  it("renders all transaction fields", () => {
    render(<TransactionFields transaction={sampleTransaction} />);

    expect(screen.getByText("APPROVED")).toBeInTheDocument();
    expect(screen.getByText("$150.00")).toBeInTheDocument();
    expect(screen.getByText("USD")).toBeInTheDocument();
    expect(screen.getByText("CARD")).toBeInTheDocument();
    expect(screen.getByText("cust_1")).toBeInTheDocument();
    expect(screen.getByText("Jane Doe")).toBeInTheDocument();
    expect(screen.getByText("jane@example.com")).toBeInTheDocument();
    expect(screen.getByText("+1234567890")).toBeInTheDocument();
    expect(screen.getByText("10.0.0.1")).toBeInTheDocument();
    expect(screen.getByText("5000 ms")).toBeInTheDocument();
  });

  it("renders the Details heading", () => {
    render(<TransactionFields transaction={sampleTransaction} />);
    expect(screen.getByText("Details")).toBeInTheDocument();
  });
});
