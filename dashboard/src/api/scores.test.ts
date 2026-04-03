import { describe, it, expect, vi, beforeEach } from "vitest";
import { fetchScore } from "./scores";

const mockFetch = vi.fn();
vi.stubGlobal("fetch", mockFetch);

beforeEach(() => {
  mockFetch.mockReset();
});

describe("fetchScore", () => {
  it("returns fraud score for a transaction", async () => {
    const body = {
      transaction_id: "txn_1",
      fraud_score: 42,
      calculated_at: "2025-01-15T10:30:02Z",
    };
    mockFetch.mockResolvedValueOnce({ ok: true, json: () => Promise.resolve(body) });

    const result = await fetchScore("txn_1");

    expect(mockFetch).toHaveBeenCalledOnce();
    const url = mockFetch.mock.calls[0][0] as string;
    expect(url).toContain("/scores/txn_1");
    expect(result).toEqual(body);
  });

  it("encodes the transaction id in the URL", async () => {
    const body = { transaction_id: "txn/special", fraud_score: 10, calculated_at: "2025-01-15T10:30:02Z" };
    mockFetch.mockResolvedValueOnce({ ok: true, json: () => Promise.resolve(body) });

    await fetchScore("txn/special");

    const url = mockFetch.mock.calls[0][0] as string;
    expect(url).toContain("/scores/txn%2Fspecial");
  });

  it("throws on 404 response", async () => {
    mockFetch.mockResolvedValueOnce({ ok: false, status: 404, statusText: "Not Found" });

    await expect(fetchScore("missing")).rejects.toThrow(
      "Failed to fetch score for missing: 404 Not Found"
    );
  });

  it("throws on 500 response", async () => {
    mockFetch.mockResolvedValueOnce({ ok: false, status: 500, statusText: "Internal Server Error" });

    await expect(fetchScore("txn_1")).rejects.toThrow(
      "Failed to fetch score for txn_1: 500 Internal Server Error"
    );
  });
});
