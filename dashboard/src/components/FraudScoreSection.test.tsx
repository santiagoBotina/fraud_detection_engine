import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import FraudScoreSection from "./FraudScoreSection";

describe("FraudScoreSection", () => {
  it("shows loading state", () => {
    render(
      <FraudScoreSection score={null} notFound={false} loading={true} error={null} onRetry={() => {}} />
    );
    expect(screen.getByTestId("score-loading")).toBeInTheDocument();
  });

  it("shows error with retry", () => {
    const onRetry = vi.fn();
    render(
      <FraudScoreSection score={null} notFound={false} loading={false} error="Unable to load fraud score" onRetry={onRetry} />
    );
    expect(screen.getByText("Unable to load fraud score")).toBeInTheDocument();
    expect(screen.getByText("Retry")).toBeInTheDocument();
  });

  it("shows not found message", () => {
    render(
      <FraudScoreSection score={null} notFound={true} loading={false} error={null} onRetry={() => {}} />
    );
    expect(screen.getByText("No fraud score computed")).toBeInTheDocument();
  });

  it("renders score when available", () => {
    const score = { transaction_id: "txn_1", fraud_score: 42, calculated_at: "2025-01-15T10:30:02Z" };
    render(
      <FraudScoreSection score={score} notFound={false} loading={false} error={null} onRetry={() => {}} />
    );
    expect(screen.getByText("42")).toBeInTheDocument();
  });
});
