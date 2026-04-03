import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor, fireEvent } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import TransactionList from "./TransactionList";
import * as txApi from "../api/transactions";

vi.mock("../api/transactions", () => ({
  fetchTransactions: vi.fn(),
}));

const mockFetchTransactions = vi.mocked(txApi.fetchTransactions);

function renderWithRouter() {
  return render(
    <MemoryRouter>
      <TransactionList />
    </MemoryRouter>
  );
}

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

beforeEach(() => {
  mockFetchTransactions.mockReset();
});

describe("TransactionList", () => {
  it("shows skeleton rows while loading", () => {
    mockFetchTransactions.mockReturnValue(new Promise(() => {}));
    renderWithRouter();
    expect(screen.getAllByTestId("skeleton-row").length).toBeGreaterThan(0);
  });

  it("renders transaction data in the table", async () => {
    mockFetchTransactions.mockResolvedValueOnce({
      data: [sampleTransaction],
      next_cursor: null,
    });

    renderWithRouter();

    await waitFor(() => {
      expect(screen.getByText("txn_001")).toBeInTheDocument();
    });

    expect(screen.getByText("Jane Doe")).toBeInTheDocument();
    expect(screen.getAllByText("CARD").length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByText("USD").length).toBeGreaterThanOrEqual(1);
    expect(screen.getByText("APPROVED")).toBeInTheDocument();
  });

  it("shows error banner when fetch fails", async () => {
    mockFetchTransactions.mockRejectedValueOnce(new Error("Network error"));

    renderWithRouter();

    await waitFor(() => {
      expect(
        screen.getByText("Unable to connect to Transaction service")
      ).toBeInTheDocument();
    });

    expect(screen.getByRole("button", { name: "Retry" })).toBeInTheDocument();
  });

  it("retries fetch when retry button is clicked", async () => {
    mockFetchTransactions.mockRejectedValueOnce(new Error("Network error"));

    renderWithRouter();

    await waitFor(() => {
      expect(
        screen.getByText("Unable to connect to Transaction service")
      ).toBeInTheDocument();
    });

    mockFetchTransactions.mockResolvedValueOnce({
      data: [sampleTransaction],
      next_cursor: null,
    });

    fireEvent.click(screen.getByRole("button", { name: "Retry" }));

    await waitFor(() => {
      expect(screen.getByText("txn_001")).toBeInTheDocument();
    });

    expect(mockFetchTransactions).toHaveBeenCalledTimes(2);
  });

  it("shows Load More button when next_cursor exists", async () => {
    mockFetchTransactions.mockResolvedValueOnce({
      data: [sampleTransaction],
      next_cursor: "cursor_abc",
    });

    renderWithRouter();

    await waitFor(() => {
      expect(screen.getByText("txn_001")).toBeInTheDocument();
    });

    expect(
      screen.getByRole("button", { name: "Load More" })
    ).toBeInTheDocument();
  });

  it("does not show Load More when next_cursor is null", async () => {
    mockFetchTransactions.mockResolvedValueOnce({
      data: [sampleTransaction],
      next_cursor: null,
    });

    renderWithRouter();

    await waitFor(() => {
      expect(screen.getByText("txn_001")).toBeInTheDocument();
    });

    expect(screen.queryByRole("button", { name: "Load More" })).toBeNull();
  });

  it("appends transactions on Load More click", async () => {
    const txn2: txApi.Transaction = {
      ...sampleTransaction,
      id: "txn_002",
      customer_name: "John Smith",
    };

    mockFetchTransactions.mockResolvedValueOnce({
      data: [sampleTransaction],
      next_cursor: "cursor_abc",
    });

    renderWithRouter();

    await waitFor(() => {
      expect(screen.getByText("txn_001")).toBeInTheDocument();
    });

    mockFetchTransactions.mockResolvedValueOnce({
      data: [txn2],
      next_cursor: null,
    });

    fireEvent.click(screen.getByRole("button", { name: "Load More" }));

    await waitFor(() => {
      expect(screen.getByText("txn_002")).toBeInTheDocument();
    });

    // Both transactions should be visible
    expect(screen.getByText("txn_001")).toBeInTheDocument();
    expect(screen.getByText("John Smith")).toBeInTheDocument();
    expect(mockFetchTransactions).toHaveBeenCalledTimes(2);
  });

  it("renders transaction ID as a link to detail page", async () => {
    mockFetchTransactions.mockResolvedValueOnce({
      data: [sampleTransaction],
      next_cursor: null,
    });

    renderWithRouter();

    await waitFor(() => {
      expect(screen.getByText("txn_001")).toBeInTheDocument();
    });

    const link = screen.getByText("txn_001").closest("a");
    expect(link).toHaveAttribute("href", "/transactions/txn_001");
  });

  it("renders refresh button and refreshes data on click", async () => {
    mockFetchTransactions.mockResolvedValueOnce({
      data: [sampleTransaction],
      next_cursor: null,
    });

    renderWithRouter();

    await waitFor(() => {
      expect(screen.getByText("txn_001")).toBeInTheDocument();
    });

    const txn2: txApi.Transaction = {
      ...sampleTransaction,
      id: "txn_new",
      customer_name: "New Person",
    };

    mockFetchTransactions.mockResolvedValueOnce({
      data: [txn2],
      next_cursor: null,
    });

    fireEvent.click(screen.getByRole("button", { name: /refresh/i }));

    await waitFor(() => {
      expect(screen.getByText("txn_new")).toBeInTheDocument();
    });
  });

  it("renders auto-refresh select with options", async () => {
    mockFetchTransactions.mockResolvedValueOnce({
      data: [sampleTransaction],
      next_cursor: null,
    });

    renderWithRouter();

    await waitFor(() => {
      expect(screen.getByText("txn_001")).toBeInTheDocument();
    });

    const select = screen.getByLabelText("Auto-refresh interval");
    expect(select).toBeInTheDocument();
    expect(select).toHaveValue("0");
  });
});
