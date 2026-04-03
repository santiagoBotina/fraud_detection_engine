import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import { useTransaction } from "./useTransaction";
import * as txApi from "../api/transactions";

vi.mock("../api/transactions", () => ({
  fetchTransaction: vi.fn(),
}));

const mockFetch = vi.mocked(txApi.fetchTransaction);

beforeEach(() => {
  mockFetch.mockReset();
});

describe("useTransaction", () => {
  it("fetches transaction on mount", async () => {
    const txn = { id: "txn_1", amount_in_cents: 1000 } as txApi.Transaction;
    mockFetch.mockResolvedValueOnce({ data: txn });

    const { result } = renderHook(() => useTransaction("txn_1"));

    expect(result.current.loading).toBe(true);

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.transaction).toEqual(txn);
    expect(result.current.error).toBeNull();
  });

  it("sets error on failure", async () => {
    mockFetch.mockRejectedValueOnce(new Error("Network error"));

    const { result } = renderHook(() => useTransaction("txn_1"));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.error).toBe("Unable to connect to Transaction service");
    expect(result.current.transaction).toBeNull();
  });

  it("does nothing when id is undefined", () => {
    const { result } = renderHook(() => useTransaction(undefined));
    expect(result.current.loading).toBe(true);
    expect(mockFetch).not.toHaveBeenCalled();
  });
});
