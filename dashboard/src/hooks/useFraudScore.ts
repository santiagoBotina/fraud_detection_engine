import { useState, useCallback, useEffect } from "react";
import { fetchScore, type FraudScore } from "../api/scores";

interface UseFraudScoreResult {
  score: FraudScore | null;
  notFound: boolean;
  loading: boolean;
  error: string | null;
  retry: () => void;
}

export function useFraudScore(id: string | undefined): UseFraudScoreResult {
  const [score, setScore] = useState<FraudScore | null>(null);
  const [notFound, setNotFound] = useState(false);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async () => {
    if (!id) return;
    setLoading(true);
    setError(null);
    setNotFound(false);
    try {
      const result = await fetchScore(id);
      setScore(result);
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : "";
      if (message.includes("404")) {
        setNotFound(true);
      } else {
        setError("Unable to load fraud score");
      }
    } finally {
      setLoading(false);
    }
  }, [id]);

  useEffect(() => { load(); }, [load]);

  return { score, notFound, loading, error, retry: load };
}
