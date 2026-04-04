import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor, fireEvent } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import TransactionList from "./TransactionList";
import * as txApi from "../api/transactions";
import * as statsApi from "../api/stats";
import { ApiError } from "../api/errors";
import type { TransactionStatsResponse } from "../types";

vi.mock("../api/transactions", () => ({
  fetchTransactions: vi.fn(),
  fetchTransaction: vi.fn(),
}));

vi.mock("../api/stats", () => ({
  fetchStats: vi.fn(),
}));

const mockFetchTransactions = vi.mocked(txApi.fetchTransactions);
const mockFetchTransaction = vi.mocked(txApi.fetchTransaction);
const mockFetchStats = vi.mocked(statsApi.fetchStats);

const sampleStats: TransactionStatsResponse = {
  today: 5,
  this_week: 20,
  this_month: 80,
  total: 100,
  approved: 60,
  declined: 30,
  pending: 10,
  payment_methods: { CARD: 50, BANK_TRANSFER: 30, CRYPTO: 20 },
  avg_latency_ms: 1500,
  finalized_count: 90,
  latency_low: 50,
  latency_medium: 30,
  latency_high: 10,
};

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
  mockFetchTransaction.mockReset();
  mockFetchStats.mockReset();
  mockFetchStats.mockResolvedValue(sampleStats);
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
  });

  it("fetches only one page (does not load all pages)", async () => {
    mockFetchTransactions.mockResolvedValueOnce({
      data: [sampleTransaction],
      next_cursor: "cursor_abc",
    });

    renderWithRouter();

    await waitFor(() => {
      expect(screen.getByText("txn_001")).toBeInTheDocument();
    });

    // Only one fetch call — no automatic loading of subsequent pages
    expect(mockFetchTransactions).toHaveBeenCalledTimes(1);
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

  it("searches transactions by ID with local-first match", async () => {
    const txn2: txApi.Transaction = {
      ...sampleTransaction,
      id: "txn_002",
      customer_name: "John Smith",
    };

    mockFetchTransactions.mockResolvedValueOnce({
      data: [sampleTransaction, txn2],
      next_cursor: null,
    });

    renderWithRouter();

    await waitFor(() => {
      expect(screen.getByText("txn_001")).toBeInTheDocument();
      expect(screen.getByText("txn_002")).toBeInTheDocument();
    });

    const searchInput = screen.getByLabelText("Search by transaction ID");
    fireEvent.change(searchInput, { target: { value: "txn_002" } });

    await waitFor(() => {
      expect(screen.queryByText("txn_001")).toBeNull();
      expect(screen.getByText("txn_002")).toBeInTheDocument();
    });

    // No backend call since it was found locally
    expect(mockFetchTransaction).not.toHaveBeenCalled();
  });

  it("falls back to backend search when ID not found locally", async () => {
    mockFetchTransactions.mockResolvedValueOnce({
      data: [sampleTransaction],
      next_cursor: null,
    });

    const remoteTxn: txApi.Transaction = {
      ...sampleTransaction,
      id: "txn_remote",
      customer_name: "Remote Person",
    };

    mockFetchTransaction.mockResolvedValueOnce({ data: remoteTxn });

    renderWithRouter();

    await waitFor(() => {
      expect(screen.getByText("txn_001")).toBeInTheDocument();
    });

    const searchInput = screen.getByLabelText("Search by transaction ID");
    fireEvent.change(searchInput, { target: { value: "txn_remote" } });

    await waitFor(() => {
      expect(screen.getByText("txn_remote")).toBeInTheDocument();
    });

    expect(mockFetchTransaction).toHaveBeenCalledWith("txn_remote");
  });

  it("shows 'Transaction not found' on 404 search result", async () => {
    mockFetchTransactions.mockResolvedValueOnce({
      data: [sampleTransaction],
      next_cursor: null,
    });

    mockFetchTransaction.mockRejectedValueOnce(new ApiError(404, "Not Found"));

    renderWithRouter();

    await waitFor(() => {
      expect(screen.getByText("txn_001")).toBeInTheDocument();
    });

    const searchInput = screen.getByLabelText("Search by transaction ID");
    fireEvent.change(searchInput, { target: { value: "txn_missing" } });

    await waitFor(() => {
      expect(screen.getByText("Transaction not found")).toBeInTheDocument();
    });
  });

  it("renders pagination controls with Next button when cursor exists", async () => {
    mockFetchTransactions.mockResolvedValueOnce({
      data: [sampleTransaction],
      next_cursor: "cursor_abc",
    });

    renderWithRouter();

    await waitFor(() => {
      expect(screen.getByText("txn_001")).toBeInTheDocument();
    });

    const nextBtn = screen.getByRole("button", { name: "Next page" });
    expect(nextBtn).toBeInTheDocument();
    expect(nextBtn).toBeEnabled();

    // Previous should be disabled on page 1
    const prevBtn = screen.getByRole("button", { name: "Previous page" });
    expect(prevBtn).toBeDisabled();

    // Page number displayed
    expect(screen.getByText("Page 1")).toBeInTheDocument();
  });

  it("navigates to next page when Next button is clicked", async () => {
    const txnPage2: txApi.Transaction = {
      ...sampleTransaction,
      id: "txn_page2",
      customer_name: "Page Two Person",
    };

    // First page
    mockFetchTransactions.mockResolvedValueOnce({
      data: [sampleTransaction],
      next_cursor: "cursor_page2",
    });

    renderWithRouter();

    await waitFor(() => {
      expect(screen.getByText("txn_001")).toBeInTheDocument();
    });

    // Mock second page fetch
    mockFetchTransactions.mockResolvedValueOnce({
      data: [txnPage2],
      next_cursor: null,
    });

    fireEvent.click(screen.getByRole("button", { name: "Next page" }));

    await waitFor(() => {
      expect(screen.getByText("txn_page2")).toBeInTheDocument();
    });

    expect(screen.getByText("Page 2")).toBeInTheDocument();
    expect(screen.queryByText("txn_001")).not.toBeInTheDocument();
  });

  it("renders page size selector", async () => {
    mockFetchTransactions.mockResolvedValueOnce({
      data: [sampleTransaction],
      next_cursor: null,
    });

    renderWithRouter();

    await waitFor(() => {
      expect(screen.getByText("txn_001")).toBeInTheDocument();
    });

    const pageSizeSelect = screen.getByLabelText("Page size");
    expect(pageSizeSelect).toBeInTheDocument();
    expect(pageSizeSelect).toHaveValue("20");
  });

  it("changes page size and resets to page 1", async () => {
    const txnPage2: txApi.Transaction = {
      ...sampleTransaction,
      id: "txn_page2",
      customer_name: "Page Two Person",
    };

    // First page load
    mockFetchTransactions.mockResolvedValueOnce({
      data: [sampleTransaction],
      next_cursor: "cursor_page2",
    });

    renderWithRouter();

    await waitFor(() => {
      expect(screen.getByText("txn_001")).toBeInTheDocument();
    });

    // Navigate to page 2
    mockFetchTransactions.mockResolvedValueOnce({
      data: [txnPage2],
      next_cursor: null,
    });

    fireEvent.click(screen.getByRole("button", { name: "Next page" }));

    await waitFor(() => {
      expect(screen.getByText("Page 2")).toBeInTheDocument();
    });

    // Change page size — should reset to page 1
    mockFetchTransactions.mockResolvedValueOnce({
      data: [sampleTransaction],
      next_cursor: null,
    });

    fireEvent.change(screen.getByLabelText("Page size"), {
      target: { value: "50" },
    });

    await waitFor(() => {
      expect(screen.getByText("Page 1")).toBeInTheDocument();
    });
  });

  it("renders DashboardStats from useStats", async () => {
    mockFetchTransactions.mockResolvedValueOnce({
      data: [sampleTransaction],
      next_cursor: null,
    });

    renderWithRouter();

    await waitFor(() => {
      expect(screen.getByText("txn_001")).toBeInTheDocument();
    });

    // DashboardStats renders labels from the stats response
    expect(screen.getByText("Today")).toBeInTheDocument();
    expect(screen.getByText("Total")).toBeInTheDocument();
    expect(screen.getByText("This Week")).toBeInTheDocument();
    expect(screen.getByText("This Month")).toBeInTheDocument();

    // Verify stats values from sampleStats
    expect(screen.getByText("5")).toBeInTheDocument();   // today
    expect(screen.getAllByText("100").length).toBeGreaterThanOrEqual(1); // total (also appears in page size options)
  });

  it("shows stats error when stats fetch fails", async () => {
    mockFetchStats.mockReset();
    mockFetchStats.mockRejectedValueOnce(new Error("Stats service down"));

    mockFetchTransactions.mockResolvedValueOnce({
      data: [sampleTransaction],
      next_cursor: null,
    });

    renderWithRouter();

    await waitFor(() => {
      expect(screen.getByText("Stats service down")).toBeInTheDocument();
    });
  });
});
