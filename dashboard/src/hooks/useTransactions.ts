import { useState, useCallback, useEffect, useRef } from "react";
import {
  fetchTransactions,
  fetchTransaction,
  type Transaction,
} from "../api/transactions";
import { ApiError } from "../api/errors";

export interface UseTransactionsResult {
  transactions: Transaction[];
  loading: boolean;
  refreshing: boolean;
  error: string | null;
  page: number;
  pageSize: number;
  setPageSize: (size: number) => void;
  hasNextPage: boolean;
  hasPreviousPage: boolean;
  goNextPage: () => void;
  goPreviousPage: () => void;
  autoRefresh: number;
  setAutoRefresh: (interval: number) => void;
  retry: () => void;
  refresh: () => void;
  searchId: string;
  setSearchId: (id: string) => void;
  searchError: string | null;
  searchLoading: boolean;
}

const DEFAULT_PAGE_SIZE = 20;

export function useTransactions(): UseTransactionsResult {
  const [pageData, setPageData] = useState<Transaction[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [refreshing, setRefreshing] = useState(false);

  const [pageSize, setPageSizeState] = useState(DEFAULT_PAGE_SIZE);
  const [page, setPage] = useState(1);
  const [, setCursorStack] = useState<string[]>([]);
  const [currentCursor, setCurrentCursor] = useState<string | undefined>(
    undefined
  );
  const [nextCursor, setNextCursor] = useState<string | null>(null);
  const [fetchKey, setFetchKey] = useState(0);

  const [autoRefresh, setAutoRefresh] = useState(0);
  const [searchId, setSearchIdState] = useState("");
  const [searchResult, setSearchResult] = useState<Transaction[] | null>(null);
  const [searchError, setSearchError] = useState<string | null>(null);
  const [searchLoading, setSearchLoading] = useState(false);
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null);

  // Fetch a single page of transactions
  const fetchPage = useCallback(
    async (cursor: string | undefined, isRefresh: boolean) => {
      if (isRefresh) {
        setRefreshing(true);
      } else {
        setLoading(true);
      }
      setError(null);
      try {
        const response = await fetchTransactions(pageSize, cursor);
        setPageData(response.data);
        setNextCursor(response.next_cursor);
      } catch {
        setError("Unable to connect to Transaction service");
      } finally {
        if (isRefresh) {
          setRefreshing(false);
        } else {
          setLoading(false);
        }
      }
    },
    [pageSize]
  );

  // Initial load and re-fetch on cursor/pageSize/fetchKey change
  useEffect(() => {
    fetchPage(currentCursor, false);
  }, [fetchPage, currentCursor, fetchKey]);

  // Auto-refresh interval
  useEffect(() => {
    if (intervalRef.current) {
      clearInterval(intervalRef.current);
      intervalRef.current = null;
    }
    if (autoRefresh > 0) {
      intervalRef.current = setInterval(() => {
        fetchPage(currentCursor, true);
      }, autoRefresh * 1000);
    }
    return () => {
      if (intervalRef.current) clearInterval(intervalRef.current);
    };
  }, [autoRefresh, fetchPage, currentCursor]);

  const goNextPage = useCallback(() => {
    if (nextCursor) {
      setCursorStack((prev) => [
        ...prev,
        currentCursor ?? "",
      ]);
      setCurrentCursor(nextCursor);
      setPage((p) => p + 1);
    }
  }, [nextCursor, currentCursor]);

  const goPreviousPage = useCallback(() => {
    if (page > 1) {
      setCursorStack((prev) => {
        const newStack = [...prev];
        const prevCursor = newStack.pop();
        setCurrentCursor(prevCursor === "" ? undefined : prevCursor);
        return newStack;
      });
      setPage((p) => p - 1);
    }
  }, [page]);

  const setPageSize = useCallback((size: number) => {
    setPageSizeState(size);
    setPage(1);
    setCursorStack([]);
    setCurrentCursor(undefined);
    setNextCursor(null);
  }, []);

  const retry = useCallback(() => {
    setPage(1);
    setCursorStack([]);
    setCurrentCursor(undefined);
    setNextCursor(null);
    setFetchKey((k) => k + 1);
  }, []);

  const refresh = useCallback(() => {
    fetchPage(currentCursor, true);
  }, [fetchPage, currentCursor]);

  // Local-first search with backend fallback
  useEffect(() => {
    const trimmed = searchId.trim();
    if (!trimmed) {
      setSearchResult(null);
      setSearchError(null);
      setSearchLoading(false);
      return;
    }

    // Check current page first
    const localMatch = pageData.find(
      (t) => t.id.toLowerCase() === trimmed.toLowerCase()
    );
    if (localMatch) {
      setSearchResult([localMatch]);
      setSearchError(null);
      setSearchLoading(false);
      return;
    }

    // Backend fallback
    let cancelled = false;
    setSearchLoading(true);
    setSearchError(null);

    fetchTransaction(trimmed)
      .then((response) => {
        if (!cancelled) {
          setSearchResult([response.data]);
          setSearchError(null);
        }
      })
      .catch((err: unknown) => {
        if (!cancelled) {
          if (err instanceof ApiError && err.status === 404) {
            setSearchError("Transaction not found");
          } else {
            setSearchError(
              err instanceof Error ? err.message : "Search failed"
            );
          }
          setSearchResult([]);
        }
      })
      .finally(() => {
        if (!cancelled) {
          setSearchLoading(false);
        }
      });

    return () => {
      cancelled = true;
    };
  }, [searchId, pageData]);

  const setSearchId = useCallback((id: string) => {
    setSearchIdState(id);
  }, []);

  // When search is active, show search results; otherwise show page data
  const transactions = searchId.trim() ? (searchResult ?? []) : pageData;

  const hasNextPage = nextCursor !== null;
  const hasPreviousPage = page > 1;

  return {
    transactions,
    loading,
    refreshing,
    error,
    page,
    pageSize,
    setPageSize,
    hasNextPage,
    hasPreviousPage,
    goNextPage,
    goPreviousPage,
    autoRefresh,
    setAutoRefresh,
    retry,
    refresh,
    searchId,
    setSearchId,
    searchError,
    searchLoading,
  };
}
