import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import fc from "fast-check";
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
  finalized_at: "2025-01-15T10:30:05Z",
  finalization_latency_ms: 5000,
};

const pendingTransaction: Transaction = {
  id: "txn_002",
  amount_in_cents: 8000,
  currency: "USD",
  payment_method: "CARD",
  customer_id: "cust_2",
  customer_name: "John Smith",
  customer_email: "john@example.com",
  customer_phone: "+9876543210",
  customer_ip_address: "10.0.0.2",
  status: "PENDING",
  created_at: "2025-01-15T11:00:00Z",
  updated_at: "2025-01-15T11:00:00Z",
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

  it("renders finalized_at and latency badge for finalized transactions", () => {
    render(<TransactionFields transaction={sampleTransaction} />);

    expect(screen.getByText("Finalized At")).toBeInTheDocument();
    expect(screen.getByText("Finalization Latency")).toBeInTheDocument();
    expect(screen.getByText("5.0s")).toBeInTheDocument();
  });

  it("renders 'Awaiting decision' for PENDING transactions", () => {
    render(<TransactionFields transaction={pendingTransaction} />);

    const awaitingTexts = screen.getAllByText("Awaiting decision");
    expect(awaitingTexts.length).toBe(2);
  });
});

// Feature: transaction-finalization-latency, Property 9: Graceful handling of missing finalized_at
describe("Property 9: Graceful handling of missing finalized_at", () => {
  it("renders without runtime errors for transactions with or without finalized_at", () => {
    // **Validates: Requirements 9.2**
    const statusArb = fc.constantFrom("PENDING", "APPROVED", "DECLINED");
    const currencyArb = fc.constantFrom("USD", "COP", "EUR");
    const paymentMethodArb = fc.constantFrom("CARD", "BANK_TRANSFER", "CRYPTO");

    const isoDateArb = fc
      .date({
        min: new Date("2020-01-01T00:00:00Z"),
        max: new Date("2099-12-31T23:59:59Z"),
      })
      .map((d) => d.toISOString());

    const transactionArb = fc.record({
      id: fc.uuid(),
      amount_in_cents: fc.integer({ min: 1, max: 100_000_000 }),
      currency: currencyArb,
      payment_method: paymentMethodArb,
      customer_id: fc.string({ minLength: 1, maxLength: 20 }),
      customer_name: fc.string({ minLength: 1, maxLength: 50 }),
      customer_email: fc.emailAddress(),
      customer_phone: fc.string({ minLength: 5, maxLength: 15 }),
      customer_ip_address: fc
        .tuple(
          fc.integer({ min: 0, max: 255 }),
          fc.integer({ min: 0, max: 255 }),
          fc.integer({ min: 0, max: 255 }),
          fc.integer({ min: 0, max: 255 })
        )
        .map(([a, b, c, d]) => `${a}.${b}.${c}.${d}`),
      status: statusArb,
      created_at: isoDateArb,
      updated_at: isoDateArb,
      finalized_at: fc.option(isoDateArb, { nil: undefined }),
      finalization_latency_ms: fc.option(
        fc.integer({ min: 0, max: 100_000 }),
        { nil: undefined }
      ),
    }) as fc.Arbitrary<Transaction>;

    fc.assert(
      fc.property(transactionArb, (txn) => {
        const { unmount } = render(<TransactionFields transaction={txn} />);
        // If render completes without throwing, the component handles the data gracefully
        unmount();
        return true;
      }),
      { numRuns: 100 }
    );
  });
});
