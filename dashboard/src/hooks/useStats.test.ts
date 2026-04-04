import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, waitFor, act } from "@testing-library/react";
import { useStats } from "./useStats";
import * as statsApi from "../api/stats";
import type { TransactionStatsResponse } from "../types";

vi.mock("../api/stats", () => ({
  fetchStats: vi.fn(),
}));

const mockFetch = vi.mocked(statsApi.fetchStats);

const sampleStats: TransactionStatsResponse = {
  today: 5,
  this_week: 30,
  this_month: 120,
  total: 500,
  approved: 300,
  declined: 150,
  pending: 50,
  payment_methods: { CARD: 200, BANK_TRANSFER: 200, CRYPTO: 100 },
  avg_latency_ms: 1500,
  finalized_count: 450,
  latency_low: 200,
  latency_medium: 150,
  latency_high: 100,
};

beforeEach(() => {
  mockFetch.mockReset();
});

describe("useStats", () => {
  it("fetches stats on mount and sets state correctly", async () => {
    mockFetch.mockResolvedValueOnce(sampleStats);

    const { result } = renderHook(() => useStats());

    expect(result.current.loading).toBe(true);

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.stats).toEqual(sampleStats);
    expect(result.current.error).toBeNull();
  });

  it("sets error state when fetch fails", async () => {
    mockFetch.mockRejectedValueOnce(new Error("Network error"));

    const { result } = renderHook(() => useStats());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.error).toBe("Network error");
    expect(result.current.stats).toBeNull();
  });

  it("retries after error and succeeds", async () => {
    mockFetch.mockRejectedValueOnce(new Error("Server down"));

    const { result } = renderHook(() => useStats());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.error).toBe("Server down");
    expect(result.current.stats).toBeNull();

    mockFetch.mockResolvedValueOnce(sampleStats);

    await act(async () => {
      result.current.retry();
    });

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.stats).toEqual(sampleStats);
    expect(result.current.error).toBeNull();
  });
});
