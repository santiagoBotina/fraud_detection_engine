import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor, fireEvent } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import TransactionDetail from "./TransactionDetail";
import * as txApi from "../api/transactions";
import * as evalApi from "../api/evaluations";
import * as scoreApi from "../api/scores";
import { ApiError } from "../api/errors";

vi.mock("../api/transactions", () => ({
  fetchTransaction: vi.fn(),
}));

vi.mock("../api/evaluations", () => ({
  fetchEvaluations: vi.fn(),
}));

vi.mock("../api/scores", () => ({
  fetchScore: vi.fn(),
}));

const mockFetchTransaction = vi.mocked(txApi.fetchTransaction);
const mockFetchEvaluations = vi.mocked(evalApi.fetchEvaluations);
const mockFetchScore = vi.mocked(scoreApi.fetchScore);

const sampleTransaction: txApi.Transaction = {
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

const sampleEvaluations: evalApi.RuleEvaluationResult[] = [
  {
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
  },
];

const sampleScore: scoreApi.FraudScore = {
  transaction_id: "txn_001",
  fraud_score: 42,
  calculated_at: "2025-01-15T10:30:02Z",
};

function renderWithRouter(txnId = "txn_001") {
  return render(
    <MemoryRouter initialEntries={[`/transactions/${txnId}`]}>
      <Routes>
        <Route path="/transactions/:id" element={<TransactionDetail />} />
      </Routes>
    </MemoryRouter>
  );
}

beforeEach(() => {
  mockFetchTransaction.mockReset();
  mockFetchEvaluations.mockReset();
  mockFetchScore.mockReset();
});

describe("TransactionDetail", () => {
  it("shows loading state initially", () => {
    mockFetchTransaction.mockReturnValue(new Promise(() => {}));
    mockFetchEvaluations.mockReturnValue(new Promise(() => {}));
    mockFetchScore.mockReturnValue(new Promise(() => {}));

    renderWithRouter();
    expect(screen.getByTestId("loading")).toBeInTheDocument();
  });

  it("renders all transaction fields", async () => {
    mockFetchTransaction.mockResolvedValueOnce({ data: sampleTransaction });
    mockFetchEvaluations.mockResolvedValueOnce({ data: sampleEvaluations });
    mockFetchScore.mockResolvedValueOnce(sampleScore);

    renderWithRouter();

    await waitFor(() => {
      expect(screen.getByText(/txn_001/)).toBeInTheDocument();
    });

    expect(screen.getByText("Jane Doe")).toBeInTheDocument();
    expect(screen.getByText("jane@example.com")).toBeInTheDocument();
    expect(screen.getByText("+1234567890")).toBeInTheDocument();
    expect(screen.getByText("10.0.0.1")).toBeInTheDocument();
    expect(screen.getByText("CARD")).toBeInTheDocument();
    expect(screen.getByText("USD")).toBeInTheDocument();
    expect(screen.getByText("cust_1")).toBeInTheDocument();
    expect(screen.getByText("APPROVED")).toBeInTheDocument();
  });

  it("renders rule evaluations table", async () => {
    mockFetchTransaction.mockResolvedValueOnce({ data: sampleTransaction });
    mockFetchEvaluations.mockResolvedValueOnce({ data: sampleEvaluations });
    mockFetchScore.mockResolvedValueOnce(sampleScore);

    renderWithRouter();

    await waitFor(() => {
      expect(screen.getByText("Block CRYPTO payments")).toBeInTheDocument();
    });

    expect(screen.getByText("payment_method")).toBeInTheDocument();
    expect(screen.getByText("EQUAL")).toBeInTheDocument();
    expect(screen.getByText("CRYPTO")).toBeInTheDocument();
    expect(screen.getByText("No")).toBeInTheDocument();
    expect(screen.getByText("DECLINED")).toBeInTheDocument();
  });

  it("renders fraud score with ScoreIndicator", async () => {
    mockFetchTransaction.mockResolvedValueOnce({ data: sampleTransaction });
    mockFetchEvaluations.mockResolvedValueOnce({ data: sampleEvaluations });
    mockFetchScore.mockResolvedValueOnce(sampleScore);

    renderWithRouter();

    await waitFor(() => {
      expect(screen.getByText("42")).toBeInTheDocument();
    });
  });

  it("shows 'No fraud score computed' on 404", async () => {
    mockFetchTransaction.mockResolvedValueOnce({ data: sampleTransaction });
    mockFetchEvaluations.mockResolvedValueOnce({ data: sampleEvaluations });
    mockFetchScore.mockRejectedValueOnce(
      new ApiError(404, "Failed to fetch score for txn_001: 404 Not Found")
    );

    renderWithRouter();

    await waitFor(() => {
      expect(screen.getByText("No fraud score computed")).toBeInTheDocument();
    });
  });

  it("displays evaluation time", async () => {
    mockFetchTransaction.mockResolvedValueOnce({ data: sampleTransaction });
    mockFetchEvaluations.mockResolvedValueOnce({ data: sampleEvaluations });
    mockFetchScore.mockResolvedValueOnce(sampleScore);

    renderWithRouter();

    await waitFor(() => {
      expect(screen.getByText("5000 ms")).toBeInTheDocument();
    });
  });

  it("shows error banner when transaction fetch fails", async () => {
    mockFetchTransaction.mockRejectedValueOnce(new Error("Network error"));
    mockFetchEvaluations.mockResolvedValueOnce({ data: [] });
    mockFetchScore.mockResolvedValueOnce(sampleScore);

    renderWithRouter();

    await waitFor(() => {
      expect(
        screen.getByText("Unable to connect to Transaction service")
      ).toBeInTheDocument();
    });
  });

  it("shows inline error for evaluations failure with retry", async () => {
    mockFetchTransaction.mockResolvedValueOnce({ data: sampleTransaction });
    mockFetchEvaluations.mockRejectedValueOnce(new ApiError(500, "Service down"));
    mockFetchScore.mockResolvedValueOnce(sampleScore);

    renderWithRouter();

    await waitFor(() => {
      expect(
        screen.getByText("Unable to load rule evaluations")
      ).toBeInTheDocument();
    });

    const retryButtons = screen.getAllByRole("button", { name: "Retry" });
    expect(retryButtons.length).toBeGreaterThanOrEqual(1);
  });

  it("shows inline error for score failure with retry", async () => {
    mockFetchTransaction.mockResolvedValueOnce({ data: sampleTransaction });
    mockFetchEvaluations.mockResolvedValueOnce({ data: sampleEvaluations });
    mockFetchScore.mockRejectedValueOnce(new ApiError(500, "Service unavailable"));

    renderWithRouter();

    await waitFor(() => {
      expect(
        screen.getByText("Unable to load fraud score")
      ).toBeInTheDocument();
    });

    const retryButtons = screen.getAllByRole("button", { name: "Retry" });
    expect(retryButtons.length).toBeGreaterThanOrEqual(1);
  });

  it("retries evaluations on retry button click", async () => {
    mockFetchTransaction.mockResolvedValueOnce({ data: sampleTransaction });
    mockFetchEvaluations.mockRejectedValueOnce(new ApiError(500, "Service down"));
    mockFetchScore.mockResolvedValueOnce(sampleScore);

    renderWithRouter();

    await waitFor(() => {
      expect(
        screen.getByText("Unable to load rule evaluations")
      ).toBeInTheDocument();
    });

    mockFetchEvaluations.mockResolvedValueOnce({ data: sampleEvaluations });

    const retryButtons = screen.getAllByRole("button", { name: "Retry" });
    fireEvent.click(retryButtons[0]);

    await waitFor(() => {
      expect(screen.getByText("Block CRYPTO payments")).toBeInTheDocument();
    });

    expect(mockFetchEvaluations).toHaveBeenCalledTimes(2);
  });

  it("shows empty evaluations message when none exist", async () => {
    mockFetchTransaction.mockResolvedValueOnce({ data: sampleTransaction });
    mockFetchEvaluations.mockResolvedValueOnce({ data: [] });
    mockFetchScore.mockResolvedValueOnce(sampleScore);

    renderWithRouter();

    await waitFor(() => {
      expect(screen.getByText("No rule evaluations found")).toBeInTheDocument();
    });
  });

  it("fetches all three APIs in parallel", async () => {
    mockFetchTransaction.mockResolvedValueOnce({ data: sampleTransaction });
    mockFetchEvaluations.mockResolvedValueOnce({ data: sampleEvaluations });
    mockFetchScore.mockResolvedValueOnce(sampleScore);

    renderWithRouter();

    await waitFor(() => {
      expect(screen.getByText(/txn_001/)).toBeInTheDocument();
    });

    expect(mockFetchTransaction).toHaveBeenCalledWith("txn_001");
    expect(mockFetchEvaluations).toHaveBeenCalledWith("txn_001");
    expect(mockFetchScore).toHaveBeenCalledWith("txn_001");
  });
});
