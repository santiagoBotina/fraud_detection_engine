import { useState, useCallback, useEffect } from "react";
import { fetchStats } from "../api/stats";
import type { TransactionStatsResponse } from "../types";

export interface UseStatsResult {
  stats: TransactionStatsResponse | null;
  loading: boolean;
  error: string | null;
  retry: () => void;
}

export function useStats(): UseStatsResult {
  const [stats, setStats] = useState<TransactionStatsResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const result = await fetchStats();
      setStats(result);
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "Unable to load stats");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { load(); }, [load]);

  return { stats, loading, error, retry: load };
}
