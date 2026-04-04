import { describe, it, expect, vi, beforeEach } from "vitest";
import { fetchEvaluations, fetchRules } from "./evaluations";
import { ApiError } from "./errors";

const mockFetch = vi.fn();
vi.stubGlobal("fetch", mockFetch);

beforeEach(() => {
  mockFetch.mockReset();
});

describe("fetchEvaluations", () => {
  it("returns evaluation results for a transaction", async () => {
    const body = {
      data: [
        {
          transaction_id: "txn_1",
          rule_id: "rule-001",
          rule_name: "Block CRYPTO",
          condition_field: "payment_method",
          condition_operator: "EQUAL",
          condition_value: "CRYPTO",
          actual_field_value: "CARD",
          matched: false,
          result_status: "DECLINED",
          evaluated_at: "2025-01-15T10:30:01Z",
          priority: 1,
        },
      ],
    };
    mockFetch.mockResolvedValueOnce({ ok: true, json: () => Promise.resolve(body) });

    const result = await fetchEvaluations("txn_1");

    expect(mockFetch).toHaveBeenCalledOnce();
    const url = mockFetch.mock.calls[0][0] as string;
    expect(url).toContain("/evaluations/txn_1");
    expect(result).toEqual(body);
  });

  it("returns empty list when no evaluations exist", async () => {
    const body = { data: [] };
    mockFetch.mockResolvedValueOnce({ ok: true, json: () => Promise.resolve(body) });

    const result = await fetchEvaluations("txn_none");
    expect(result.data).toEqual([]);
  });

  it("throws ApiError with status on non-OK response", async () => {
    mockFetch.mockResolvedValueOnce({ ok: false, status: 500, statusText: "Internal Server Error" });

    await expect(fetchEvaluations("txn_1")).rejects.toSatisfy((err: unknown) => {
      return err instanceof ApiError && err.status === 500 &&
        err.message === "Failed to fetch evaluations for txn_1: 500 Internal Server Error";
    });
  });
});

describe("fetchRules", () => {
  it("returns all rules", async () => {
    const body = {
      data: [
        {
          rule_id: "rule-001",
          rule_name: "Block CRYPTO",
          condition_field: "payment_method",
          condition_operator: "EQUAL",
          condition_value: "CRYPTO",
          result_status: "DECLINED",
          priority: 1,
          is_active: true,
        },
      ],
    };
    mockFetch.mockResolvedValueOnce({ ok: true, json: () => Promise.resolve(body) });

    const result = await fetchRules();

    expect(mockFetch).toHaveBeenCalledOnce();
    const url = mockFetch.mock.calls[0][0] as string;
    expect(url).toContain("/rules");
    expect(result).toEqual(body);
  });

  it("throws ApiError with status on non-OK response", async () => {
    mockFetch.mockResolvedValueOnce({ ok: false, status: 503, statusText: "Service Unavailable" });

    await expect(fetchRules()).rejects.toSatisfy((err: unknown) => {
      return err instanceof ApiError && err.status === 503 &&
        err.message === "Failed to fetch rules: 503 Service Unavailable";
    });
  });
});
