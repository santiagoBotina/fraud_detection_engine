import { describe, it, expect, vi, beforeEach } from "vitest";
import { fetchTransactions, fetchTransaction } from "./transactions";
import { ApiError } from "./errors";

const mockFetch = vi.fn();
vi.stubGlobal("fetch", mockFetch);

beforeEach(() => {
  mockFetch.mockReset();
});

describe("fetchTransactions", () => {
  it("returns paginated transactions with default limit", async () => {
    const body = {
      data: [{ id: "txn_1", amount_in_cents: 1000, currency: "USD", status: "APPROVED" }],
      next_cursor: "abc123",
    };
    mockFetch.mockResolvedValueOnce({ ok: true, json: () => Promise.resolve(body) });

    const result = await fetchTransactions();

    expect(mockFetch).toHaveBeenCalledOnce();
    const url = mockFetch.mock.calls[0][0] as string;
    expect(url).toContain("/transactions?");
    expect(url).toContain("limit=20");
    expect(result).toEqual(body);
  });

  it("passes limit and cursor as query params", async () => {
    const body = { data: [], next_cursor: null };
    mockFetch.mockResolvedValueOnce({ ok: true, json: () => Promise.resolve(body) });

    await fetchTransactions(50, "cursor_token");

    const url = mockFetch.mock.calls[0][0] as string;
    expect(url).toContain("limit=50");
    expect(url).toContain("cursor=cursor_token");
  });

  it("throws ApiError on non-OK response", async () => {
    mockFetch.mockResolvedValueOnce({ ok: false, status: 500, statusText: "Internal Server Error" });

    await expect(fetchTransactions()).rejects.toThrow(ApiError);
    await mockFetch.mockResolvedValueOnce({ ok: false, status: 500, statusText: "Internal Server Error" });
    try {
      await fetchTransactions();
    } catch (e) {
      expect(e).toBeInstanceOf(ApiError);
      expect((e as ApiError).status).toBe(500);
    }
  });
});

describe("fetchTransaction", () => {
  it("returns a single transaction by id", async () => {
    const body = { data: { id: "txn_1", amount_in_cents: 5000, currency: "EUR" } };
    mockFetch.mockResolvedValueOnce({ ok: true, json: () => Promise.resolve(body) });

    const result = await fetchTransaction("txn_1");

    expect(mockFetch).toHaveBeenCalledOnce();
    const url = mockFetch.mock.calls[0][0] as string;
    expect(url).toContain("/transactions/txn_1");
    expect(result).toEqual(body);
  });

  it("encodes the transaction id in the URL", async () => {
    const body = { data: { id: "txn/special" } };
    mockFetch.mockResolvedValueOnce({ ok: true, json: () => Promise.resolve(body) });

    await fetchTransaction("txn/special");

    const url = mockFetch.mock.calls[0][0] as string;
    expect(url).toContain("/transactions/txn%2Fspecial");
  });

  it("throws ApiError on 404 response", async () => {
    mockFetch.mockResolvedValueOnce({ ok: false, status: 404, statusText: "Not Found" });

    await expect(fetchTransaction("missing")).rejects.toThrow(ApiError);
    mockFetch.mockResolvedValueOnce({ ok: false, status: 404, statusText: "Not Found" });
    try {
      await fetchTransaction("missing");
    } catch (e) {
      expect(e).toBeInstanceOf(ApiError);
      expect((e as ApiError).status).toBe(404);
    }
  });
});
