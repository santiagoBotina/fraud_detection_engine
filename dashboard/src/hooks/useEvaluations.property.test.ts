// Feature: dashboard-pagination-metrics, Property 7
import { describe, it, expect, vi, beforeEach } from "vitest";
import fc from "fast-check";
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

// Feature: dashboard-pagination-metrics, Property 7: Evaluations hook classifies responses correctly
describe("Property 7: Evaluations hook classifies responses correctly", () => {
  it(
    "successful responses with non-empty data set evaluations and error=null",
    () => {
      // **Validates: Requirements 6.2, 6.4**
      const evalArb = fc.record({
        transaction_id: fc.string({ minLength: 1, maxLength: 20 }),
        rule_id: fc.string({ minLength: 1, maxLength: 20 }),
        rule_name: fc.string({ minLength: 1, maxLength: 30 }),
        condition_field: fc.string({ minLength: 1, maxLength: 20 }),
        condition_operator: fc.string({ minLength: 1, maxLength: 10 }),
        condition_value: fc.string({ minLength: 1, maxLength: 20 }),
        actual_field_value: fc.string({ minLength: 1, maxLength: 20 }),
        matched: fc.boolean(),
        result_status: fc.constantFrom("APPROVED", "DECLINED", "PENDING"),
        evaluated_at: fc.date().map((d) => d.toISOString()),
        priority: fc.integer({ min: 1, max: 100 }),
      });

      const nonEmptyEvalsArb = fc.array(evalArb, { minLength: 1, maxLength: 5 });

      return fc.assert(
        fc.asyncProperty(nonEmptyEvalsArb, async (evals) => {
          mockFetch.mockReset();
          mockFetch.mockResolvedValueOnce({ data: evals });

          const { result, unmount } = renderHook(() => useEvaluations("txn_test"));

          await waitFor(() => {
            expect(result.current.loading).toBe(false);
          });

          const pass =
            result.current.evaluations.length === evals.length &&
            result.current.error === null;

          unmount();
          return pass;
        }),
        { numRuns: 100 }
      );
    },
    30_000
  );

  it("successful responses with empty data set evaluations=[] and error=null", async () => {
    // **Validates: Requirements 6.2, 6.4**
    mockFetch.mockResolvedValueOnce({ data: [] });

    const { result } = renderHook(() => useEvaluations("txn_test"));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.evaluations).toEqual([]);
    expect(result.current.error).toBeNull();
  });

  it(
    "ApiError responses set error message for any HTTP error status code",
    () => {
      // **Validates: Requirements 6.2, 6.4**
      const errorStatusArb = fc.integer({ min: 400, max: 599 });

      return fc.assert(
        fc.asyncProperty(errorStatusArb, async (status) => {
          mockFetch.mockReset();
          mockFetch.mockRejectedValueOnce(new ApiError(status, `Error ${status}`));

          const { result, unmount } = renderHook(() => useEvaluations("txn_test"));

          await waitFor(() => {
            expect(result.current.loading).toBe(false);
          });

          const pass = result.current.error === "Unable to load rule evaluations";

          unmount();
          return pass;
        }),
        { numRuns: 100 }
      );
    },
    30_000
  );
});
