import { useState, useCallback, useEffect, useRef } from "react";
import { fetchTransactions, type Transaction } from "../api/transactions";

interface UseTransactionsResult {
  transactions: Transaction[];
  nextCursor: string | null;
  loading: boolean;
  loadingMore: boolean;
  refreshing: boolean;
  error: string | null;
  autoRefresh: number;
  setAutoRefresh: (interval: number) => void;
  loadMore: () => void;
  retry: () => void;
  refresh: () => void;
}

export function useTransactions(): UseTransactionsResult {
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [nextCursor, setNextCursor] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [refreshing, setRefreshing] = useState(false);
  const [autoRefresh, setAutoRefresh] = useState(0);
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const load = useCallback(async (cursor?: string) => {
    const isInitial = !cursor;
    if (isInitial) {
      setLoading(true);
    } else {
      setLoadingMore(true);
    }
    setError(null);

    try {
      const response = await fetchTransactions(20, cursor);
      setTransactions((prev) =>
        isInitial ? response.data : [...prev, ...response.data]
      );
      setNextCursor(response.next_cursor ?? null);
    } catch {
      setError("Unable to connect to Transaction service");
    } finally {
      setLoading(false);
      setLoadingMore(false);
    }
  }, []);

  const refresh = useCallback(async () => {
    setRefreshing(true);
    setError(null);
    try {
      const response = await fetchTransactions(20);
      setTransactions(response.data);
      setNextCursor(response.next_cursor ?? null);
    } catch {
      setError("Unable to connect to Transaction service");
    } finally {
      setRefreshing(false);
    }
  }, []);

  useEffect(() => { load(); }, [load]);

  // Auto-refresh interval
  useEffect(() => {
    if (intervalRef.current) {
      clearInterval(intervalRef.current);
      intervalRef.current = null;
    }
    if (autoRefresh > 0) {
      intervalRef.current = setInterval(refresh, autoRefresh * 1000);
    }
    return () => {
      if (intervalRef.current) clearInterval(intervalRef.current);
    };
  }, [autoRefresh, refresh]);

  const loadMore = useCallback(() => {
    if (nextCursor) load(nextCursor);
  }, [nextCursor, load]);

  const retry = useCallback(() => {
    setTransactions([]);
    setNextCursor(null);
    load();
  }, [load]);

  return {
    transactions,
    nextCursor,
    loading,
    loadingMore,
    refreshing,
    error,
    autoRefresh,
    setAutoRefresh,
    loadMore,
    retry,
    refresh,
  };
}
