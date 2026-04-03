import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import { useEvaluations } from "./useEvaluations";
import * as evalApi from "../api/evaluations";

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

  it("sets error on failure", async () => {
    mockFetch.mockRejectedValueOnce(new Error("fail"));

    const { result } = renderHook(() => useEvaluations("txn_1"));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.error).toBe("Unable to load rule evaluations");
  });
});
