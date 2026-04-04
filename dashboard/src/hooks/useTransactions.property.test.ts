import { expect, describe, it, vi, beforeEach } from "vitest";
import fc from "fast-check";
import { renderHook, waitFor, act } from "@testing-library/react";
import { useTransactions } from "./useTransactions";
import * as txnApi from "../api/transactions";

vi.mock("../api/transactions", () => ({
  fetchTransactions: vi.fn(),
  fetchTransaction: vi.fn(),
}));

const mockFetchTransactions = vi.mocked(txnApi.fetchTransactions);

beforeEach(() => {
  mockFetchTransactions.mockReset();
});

function makeTxn(id: string) {
  return {
    id,
    amount_in_cents: 1000,
    currency: "USD",
    payment_method: "CARD",
    customer_id: "cust_1",
    customer_name: "Test",
    customer_email: "test@test.com",
    customer_phone: "555-0100",
    customer_ip_address: "127.0.0.1",
    status: "APPROVED",
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  };
}

describe("Property 1: Pagination controls reflect navigation state", () => {
  it(
    "on page 1, hasPreviousPage is always false and hasNextPage matches cursor presence",
    () => {
      const cursorArb = fc.option(
        fc.string({ minLength: 1, maxLength: 30 }),
        { nil: null }
      );

      return fc.assert(
        fc.asyncProperty(cursorArb, async (cursor) => {
          mockFetchTransactions.mockReset();
          mockFetchTransactions.mockResolvedValueOnce({
            data: [makeTxn("txn_1")],
            next_cursor: cursor,
          });

          const { result, unmount } = renderHook(() => useTransactions());

          await waitFor(() => {
            expect(result.current.loading).toBe(false);
          });

          const pass =
            result.current.page === 1 &&
            result.current.hasPreviousPage === false &&
            result.current.hasNextPage === (cursor !== null);

          unmount();
          return pass;
        }),
        { numRuns: 100 }
      );
    },
    30_000
  );

  it(
    "after navigating forward, page increments and hasPreviousPage becomes true",
    () => {
      const stateArb = fc.record({
        targetPage: fc.integer({ min: 2, max: 5 }),
        finalCursorPresent: fc.boolean(),
      });

      return fc.assert(
        fc.asyncProperty(stateArb, async ({ targetPage, finalCursorPresent }) => {
          mockFetchTransactions.mockReset();

          for (let i = 1; i < targetPage; i++) {
            mockFetchTransactions.mockResolvedValueOnce({
              data: [makeTxn(`txn_page_${i}`)],
              next_cursor: `cursor_${i}`,
            });
          }

          mockFetchTransactions.mockResolvedValueOnce({
            data: [makeTxn(`txn_page_${targetPage}`)],
            next_cursor: finalCursorPresent ? "cursor_next" : null,
          });

          const { result, unmount } = renderHook(() => useTransactions());

          // Wait for initial load (page 1)
          await waitFor(() => {
            expect(result.current.loading).toBe(false);
          });

          // Navigate forward to targetPage
          for (let i = 1; i < targetPage; i++) {
            act(() => {
              result.current.goNextPage();
            });
            await waitFor(() => {
              expect(result.current.loading).toBe(false);
            });
          }

          const pass =
            result.current.page === targetPage &&
            result.current.hasPreviousPage === true &&
            result.current.hasNextPage === finalCursorPresent;

          unmount();
          return pass;
        }),
        { numRuns: 100 }
      );
    },
    60_000
  );
});

// Feature: dashboard-pagination-metrics, Property 2: Page size change resets to first page
describe("Property 2: Page size change resets to first page", () => {
  it(
    "changing page size resets page to 1, clears cursor stack, and fetches with new limit",
    () => {
      // **Validates: Requirements 2.3, 2.4**
      const pageSizeArb = fc.constantFrom(20, 30, 50, 100);

      return fc.assert(
        fc.asyncProperty(pageSizeArb, async (newPageSize) => {
          mockFetchTransactions.mockReset();

          // Page 1: return a cursor so we can navigate forward
          mockFetchTransactions.mockResolvedValueOnce({
            data: [makeTxn("txn_p1")],
            next_cursor: "cursor_1",
          });

          const { result, unmount } = renderHook(() => useTransactions());

          // Wait for initial load (page 1)
          await waitFor(() => {
            expect(result.current.loading).toBe(false);
          });

          // Page 2: navigate forward to get into a non-initial state
          mockFetchTransactions.mockResolvedValueOnce({
            data: [makeTxn("txn_p2")],
            next_cursor: "cursor_2",
          });

          act(() => {
            result.current.goNextPage();
          });

          await waitFor(() => {
            expect(result.current.page).toBe(2);
          });

          // Now change page size — mock the re-fetch for the reset
          mockFetchTransactions.mockResolvedValueOnce({
            data: [makeTxn("txn_reset")],
            next_cursor: null,
          });

          act(() => {
            result.current.setPageSize(newPageSize);
          });

          await waitFor(() => {
            expect(result.current.loading).toBe(false);
          });

          // Verify: page reset to 1, no previous page
          const pageReset = result.current.page === 1;
          const noPrevious = result.current.hasPreviousPage === false;

          // Verify: fetchTransactions was called with the new page size and no cursor
          const lastCall =
            mockFetchTransactions.mock.calls[
              mockFetchTransactions.mock.calls.length - 1
            ];
          const calledWithNewLimit = lastCall[0] === newPageSize;
          const calledWithNoCursor =
            lastCall[1] === undefined || lastCall[1] === null;

          unmount();
          return pageReset && noPrevious && calledWithNewLimit && calledWithNoCursor;
        }),
        { numRuns: 100 }
      );
    },
    60_000
  );
});

// Feature: dashboard-pagination-metrics, Property 4: Local-first search avoids backend call
const mockFetchTransaction = vi.mocked(txnApi.fetchTransaction);

describe("Property 4: Local-first search avoids backend call", () => {
  it(
    "searching for an ID that exists in the current page does not call fetchTransaction",
    () => {
      // **Validates: Requirements 5.1, 5.2**
      const countArb = fc.integer({ min: 1, max: 10 });

      return fc.assert(
        fc.asyncProperty(countArb, async (count) => {
          mockFetchTransactions.mockReset();
          mockFetchTransaction.mockReset();

          // Build a page of transactions with simple IDs
          const txns = Array.from({ length: count }, (_, i) =>
            makeTxn(`txn_${i + 1}`)
          );

          mockFetchTransactions.mockResolvedValueOnce({
            data: txns,
            next_cursor: null,
          });

          const { result, unmount } = renderHook(() => useTransactions());

          await waitFor(() => {
            expect(result.current.loading).toBe(false);
          });

          // Pick the last transaction ID to search for (guaranteed to be in the page)
          const targetId = `txn_${count}`;

          act(() => {
            result.current.setSearchId(targetId);
          });

          // Wait for search result to appear
          await waitFor(() => {
            expect(result.current.transactions.length).toBe(1);
          });

          // fetchTransaction should NOT have been called — local match found
          const notCalled = mockFetchTransaction.mock.calls.length === 0;
          const matchFound =
            result.current.transactions.length === 1 &&
            result.current.transactions[0].id === targetId;

          unmount();
          return notCalled && matchFound;
        }),
        { numRuns: 100 }
      );
    },
    30_000
  );
});

// Feature: dashboard-pagination-metrics, Property 5: Backend search fallback for missing IDs
describe("Property 5: Backend search fallback for missing IDs", () => {
  it(
    "searching for an ID not in the current page calls fetchTransaction",
    () => {
      // **Validates: Requirements 5.3**
      const countArb = fc.integer({ min: 1, max: 10 });

      return fc.assert(
        fc.asyncProperty(countArb, async (count) => {
          mockFetchTransactions.mockReset();
          mockFetchTransaction.mockReset();

          // Build a page of transactions with IDs txn_1..txn_N
          const txns = Array.from({ length: count }, (_, i) =>
            makeTxn(`txn_${i + 1}`)
          );

          mockFetchTransactions.mockResolvedValueOnce({
            data: txns,
            next_cursor: null,
          });

          // The search ID will NOT be in the page
          const missingId = `txn_missing_${count}`;
          const missingTxn = makeTxn(missingId);

          mockFetchTransaction.mockResolvedValue({ data: missingTxn });

          const { result, unmount } = renderHook(() => useTransactions());

          await waitFor(() => {
            expect(result.current.loading).toBe(false);
          });

          act(() => {
            result.current.setSearchId(missingId);
          });

          // Wait for the backend search to complete
          await waitFor(() => {
            expect(result.current.searchLoading).toBe(false);
            expect(result.current.transactions.length).toBe(1);
          });

          // fetchTransaction SHOULD have been called with the missing ID
          const wasCalled = mockFetchTransaction.mock.calls.some(
            (call) => call[0] === missingId
          );

          unmount();
          return wasCalled;
        }),
        { numRuns: 100 }
      );
    },
    30_000
  );
});

// Feature: dashboard-pagination-metrics, Property 6: Clearing search restores paginated view
describe("Property 6: Clearing search restores paginated view", () => {
  it(
    "clearing the search input restores the original page data",
    () => {
      // **Validates: Requirements 5.7**
      const countArb = fc.integer({ min: 1, max: 10 });

      return fc.assert(
        fc.asyncProperty(countArb, async (count) => {
          mockFetchTransactions.mockReset();
          mockFetchTransaction.mockReset();

          // Build a page of transactions
          const txns = Array.from({ length: count }, (_, i) =>
            makeTxn(`txn_${i + 1}`)
          );

          mockFetchTransactions.mockResolvedValueOnce({
            data: txns,
            next_cursor: null,
          });

          const { result, unmount } = renderHook(() => useTransactions());

          await waitFor(() => {
            expect(result.current.loading).toBe(false);
          });

          // Verify original page data
          const originalIds = result.current.transactions.map((t) => t.id);

          // Search for a local match (first ID in the page)
          const searchTarget = `txn_1`;

          act(() => {
            result.current.setSearchId(searchTarget);
          });

          await waitFor(() => {
            expect(result.current.transactions.length).toBe(1);
          });

          // Now clear the search
          act(() => {
            result.current.setSearchId("");
          });

          await waitFor(() => {
            expect(result.current.transactions.length).toBe(count);
          });

          // Verify the original page data is restored
          const restoredIds = result.current.transactions.map((t) => t.id);
          const restored =
            restoredIds.length === originalIds.length &&
            restoredIds.every((id, idx) => id === originalIds[idx]);

          unmount();
          return restored;
        }),
        { numRuns: 100 }
      );
    },
    30_000
  );
});
