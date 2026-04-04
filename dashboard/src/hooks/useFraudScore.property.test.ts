import { expect, describe, it, vi, beforeEach } from "vitest";
import fc from "fast-check";
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

// Feature: dashboard-pagination-metrics, Property 8: Fraud score hook classifies error status codes correctly
describe("Property 8: Fraud score hook classifies error status codes correctly", () => {
  it("404 status sets notFound=true and error=null", async () => {
    // **Validates: Requirements 7.3, 7.4**
    mockFetch.mockRejectedValueOnce(new ApiError(404, "Not Found"));

    const { result } = renderHook(() => useFraudScore("txn_test"));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.notFound).toBe(true);
    expect(result.current.error).toBeNull();
  });

  it("non-404 error status codes set notFound=false and error message", () => {
    // **Validates: Requirements 7.3, 7.4**
    const errorStatusArb = fc.integer({ min: 400, max: 599 }).filter((s) => s !== 404);

    fc.assert(
      fc.asyncProperty(errorStatusArb, async (status) => {
        mockFetch.mockReset();
        mockFetch.mockRejectedValueOnce(new ApiError(status, `Error ${status}`));

        const { result, unmount } = renderHook(() => useFraudScore("txn_test"));

        await waitFor(() => {
          expect(result.current.loading).toBe(false);
        });

        const pass =
          result.current.notFound === false &&
          result.current.error === "Unable to load fraud score";

        unmount();
        return pass;
      }),
      { numRuns: 100 }
    );
  });

  it("404 always sets notFound=true across random transaction IDs", () => {
    // **Validates: Requirements 7.3, 7.4**
    const txnIdArb = fc.string({ minLength: 1, maxLength: 50 }).filter((s) => s.trim().length > 0);

    fc.assert(
      fc.asyncProperty(txnIdArb, async (txnId) => {
        mockFetch.mockReset();
        mockFetch.mockRejectedValueOnce(new ApiError(404, "Not Found"));

        const { result, unmount } = renderHook(() => useFraudScore(txnId));

        await waitFor(() => {
          expect(result.current.loading).toBe(false);
        });

        const pass =
          result.current.notFound === true &&
          result.current.error === null;

        unmount();
        return pass;
      }),
      { numRuns: 100 }
    );
  });
});
