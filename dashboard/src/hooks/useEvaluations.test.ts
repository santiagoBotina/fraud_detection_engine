import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import { useEvaluations } from "./useEvaluations";
import * as evalApi from "../api/evaluations";
import { ApiError } from "../api/errors";

vi.mock("../api/evaluations", () => ({
  fetchEvaluations: vi.fn(),
}));

const mockFetch = vi.mocked(evalApi.fetchEvaluations);

beforeEach(() => {
  mockFetch.mockReset();
});

describe("useEvaluations", () => {
  it("fetches evaluations on mount", async () => {
    const evals = [{ rule_id: "r1", rule_name: "Test Rule" }] as evalApi.RuleEvaluationResult[];
    mockFetch.mockResolvedValueOnce({ data: evals });

    const { result } = renderHook(() => useEvaluations("txn_1"));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.evaluations).toEqual(evals);
    expect(result.current.error).toBeNull();
  });

  it("treats successful empty response as valid (no error)", async () => {
    mockFetch.mockResolvedValueOnce({ data: [] });

    const { result } = renderHook(() => useEvaluations("txn_1"));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.evaluations).toEqual([]);
    expect(result.current.error).toBeNull();
  });

  it("sets error on ApiError (HTTP failure)", async () => {
    mockFetch.mockRejectedValueOnce(new ApiError(500, "Internal Server Error"));

    const { result } = renderHook(() => useEvaluations("txn_1"));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.error).toBe("Unable to load rule evaluations");
  });

  it("does not set error for non-ApiError exceptions", async () => {
    mockFetch.mockRejectedValueOnce(new Error("network failure"));

    const { result } = renderHook(() => useEvaluations("txn_1"));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.error).toBeNull();
  });
});
