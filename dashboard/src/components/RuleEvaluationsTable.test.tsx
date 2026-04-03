import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import RuleEvaluationsTable from "./RuleEvaluationsTable";
import type { RuleEvaluationResult } from "../types";

const sampleEval: RuleEvaluationResult = {
  transaction_id: "txn_001",
  rule_id: "rule-001",
  rule_name: "Block CRYPTO payments",
  condition_field: "payment_method",
  condition_operator: "EQUAL",
  condition_value: "CRYPTO",
  actual_field_value: "CARD",
  matched: false,
  result_status: "DECLINED",
  evaluated_at: "2025-01-15T10:30:01Z",
  priority: 1,
};

describe("RuleEvaluationsTable", () => {
  it("shows loading state", () => {
    render(
      <RuleEvaluationsTable evaluations={[]} loading={true} error={null} onRetry={() => {}} />
    );
    expect(screen.getByTestId("eval-loading")).toBeInTheDocument();
  });

  it("shows error with retry", () => {
    const onRetry = vi.fn();
    render(
      <RuleEvaluationsTable evaluations={[]} loading={false} error="Unable to load rule evaluations" onRetry={onRetry} />
    );
    expect(screen.getByText("Unable to load rule evaluations")).toBeInTheDocument();
  });

  it("shows empty state when no evaluations", () => {
    render(
      <RuleEvaluationsTable evaluations={[]} loading={false} error={null} onRetry={() => {}} />
    );
    expect(screen.getByText("No rule evaluations found")).toBeInTheDocument();
  });

  it("renders evaluation rows", () => {
    render(
      <RuleEvaluationsTable evaluations={[sampleEval]} loading={false} error={null} onRetry={() => {}} />
    );
    expect(screen.getByText("Block CRYPTO payments")).toBeInTheDocument();
    expect(screen.getByText("payment_method")).toBeInTheDocument();
    expect(screen.getByText("EQUAL")).toBeInTheDocument();
    expect(screen.getByText("CRYPTO")).toBeInTheDocument();
    expect(screen.getByText("No")).toBeInTheDocument();
    expect(screen.getByText("DECLINED")).toBeInTheDocument();
  });
});
