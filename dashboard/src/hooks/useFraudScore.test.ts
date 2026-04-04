import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import { useFraudScore } from "./useFraudScore";
import * as scoreApi from "../api/scores";
import { ApiError } from "../api/errors";

vi.mock("../api/scores", () => ({
  fetchScore: vi.fn(),
}));

const mockFetch = vi.mocked(scoreApi.fetchScore);

beforeEach(() => {
  mockFetch.mockReset();
});

describe("useFraudScore", () => {
  it("fetches score on mount", async () => {
    const score = { transaction_id: "txn_1", fraud_score: 42, calculated_at: "2025-01-15T10:30:02Z" };
    mockFetch.mockResolvedValueOnce(score);

    const { result } = renderHook(() => useFraudScore("txn_1"));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.score).toEqual(score);
    expect(result.current.notFound).toBe(false);
    expect(result.current.error).toBeNull();
  });

  it("sets notFound on 404", async () => {
    mockFetch.mockRejectedValueOnce(new ApiError(404, "Failed to fetch score for txn_1: 404 Not Found"));

    const { result } = renderHook(() => useFraudScore("txn_1"));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.notFound).toBe(true);
    expect(result.current.error).toBeNull();
  });

  it("sets error on non-404 failure", async () => {
    mockFetch.mockRejectedValueOnce(new ApiError(500, "Service unavailable"));

    const { result } = renderHook(() => useFraudScore("txn_1"));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.error).toBe("Unable to load fraud score");
    expect(result.current.notFound).toBe(false);
  });
});
