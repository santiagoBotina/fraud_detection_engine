import { describe, it, expect, vi } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import DashboardStats from "./DashboardStats";
import type { TransactionStatsResponse } from "../types";

function makeStats(overrides: Partial<TransactionStatsResponse> = {}): TransactionStatsResponse {
  return {
    today: 7,
    this_week: 33,
    this_month: 121,
    total: 200,
    approved: 142,
    declined: 37,
    pending: 21,
    payment_methods: { CARD: 113, BANK_TRANSFER: 58, CRYPTO: 29 },
    avg_latency_ms: 2500,
    finalized_count: 179,
    latency_low: 89,
    latency_medium: 53,
    latency_high: 37,
    ...overrides,
  };
}

describe("DashboardStats", () => {
  const noop = () => {};

  describe("rendering with stats data", () => {
    it("renders time bucket cards", () => {
      const stats = makeStats();
      render(<DashboardStats stats={stats} loading={false} error={null} onRetry={noop} />);

      expect(screen.getByText("Today")).toBeInTheDocument();
      expect(screen.getByText("7")).toBeInTheDocument();
      expect(screen.getByText("This Week")).toBeInTheDocument();
      expect(screen.getByText("33")).toBeInTheDocument();
      expect(screen.getByText("This Month")).toBeInTheDocument();
      expect(screen.getByText("121")).toBeInTheDocument();
      expect(screen.getByText("Total")).toBeInTheDocument();
      expect(screen.getByText("200")).toBeInTheDocument();
    });

    it("renders decision rates", () => {
      const stats = makeStats();
      render(<DashboardStats stats={stats} loading={false} error={null} onRetry={noop} />);

      expect(screen.getByText("Decision Rates")).toBeInTheDocument();
      expect(screen.getByText("Approved")).toBeInTheDocument();
      // 142/200 = 71%
      expect(screen.getByText("71% (142)")).toBeInTheDocument();
      expect(screen.getByText("Declined")).toBeInTheDocument();
      // 37/200 = 19%
      expect(screen.getByText("19% (37)")).toBeInTheDocument();
      expect(screen.getByText("Pending")).toBeInTheDocument();
      // 21/200 = 11%
      expect(screen.getByText("11% (21)")).toBeInTheDocument();
    });

    it("hides pending row when pending is 0", () => {
      const stats = makeStats({ pending: 0, approved: 163, declined: 37 });
      render(<DashboardStats stats={stats} loading={false} error={null} onRetry={noop} />);

      expect(screen.queryByText("Pending")).not.toBeInTheDocument();
    });

    it("renders payment methods sorted by count", () => {
      const stats = makeStats();
      render(<DashboardStats stats={stats} loading={false} error={null} onRetry={noop} />);

      expect(screen.getByText("Payment Methods")).toBeInTheDocument();
      expect(screen.getByText("CARD")).toBeInTheDocument();
      expect(screen.getByText("BANK_TRANSFER")).toBeInTheDocument();
      expect(screen.getByText("CRYPTO")).toBeInTheDocument();
    });

    it("renders latency section when finalized_count > 0", () => {
      const stats = makeStats();
      render(<DashboardStats stats={stats} loading={false} error={null} onRetry={noop} />);

      expect(screen.getByText("Finalization Latency")).toBeInTheDocument();
      expect(screen.getByText("2.5s")).toBeInTheDocument();
      expect(screen.getByText("avg across 179 finalized")).toBeInTheDocument();
    });

    it("hides latency section when finalized_count is 0", () => {
      const stats = makeStats({ finalized_count: 0, avg_latency_ms: 0, latency_low: 0, latency_medium: 0, latency_high: 0 });
      render(<DashboardStats stats={stats} loading={false} error={null} onRetry={noop} />);

      expect(screen.queryByText("Finalization Latency")).not.toBeInTheDocument();
    });
  });

  describe("null/loading/error states", () => {
    it("shows loading state when loading is true and stats is null", () => {
      render(<DashboardStats stats={null} loading={true} error={null} onRetry={noop} />);

      expect(screen.getByText("Loading stats…")).toBeInTheDocument();
    });

    it("shows loading state when stats is null even if loading is false", () => {
      render(<DashboardStats stats={null} loading={false} error={null} onRetry={noop} />);

      expect(screen.getByText("Loading stats…")).toBeInTheDocument();
    });

    it("shows ErrorBanner when error is set", () => {
      const onRetry = vi.fn();
      render(<DashboardStats stats={null} loading={false} error="Failed to load stats" onRetry={onRetry} />);

      expect(screen.getByRole("alert")).toBeInTheDocument();
      expect(screen.getByText("Failed to load stats")).toBeInTheDocument();
    });

    it("calls onRetry when retry button is clicked in error state", () => {
      const onRetry = vi.fn();
      render(<DashboardStats stats={null} loading={false} error="Failed to load stats" onRetry={onRetry} />);

      fireEvent.click(screen.getByText("Retry"));
      expect(onRetry).toHaveBeenCalledOnce();
    });

    it("error state takes precedence over loading", () => {
      render(<DashboardStats stats={null} loading={true} error="Server error" onRetry={noop} />);

      expect(screen.getByRole("alert")).toBeInTheDocument();
      expect(screen.queryByText("Loading stats…")).not.toBeInTheDocument();
    });
  });

  describe("edge cases", () => {
    it("returns null when stats.total is 0", () => {
      const stats = makeStats({ total: 0, approved: 0, declined: 0, pending: 0 });
      const { container } = render(<DashboardStats stats={stats} loading={false} error={null} onRetry={noop} />);

      expect(container.innerHTML).toBe("");
    });
  });
});
