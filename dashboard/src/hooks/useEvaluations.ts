import { useState, useCallback, useEffect } from "react";
import { fetchEvaluations, type RuleEvaluationResult } from "../api/evaluations";
import { ApiError } from "../api/errors";

interface UseEvaluationsResult {
  evaluations: RuleEvaluationResult[];
  loading: boolean;
  error: string | null;
  retry: () => void;
}

export function useEvaluations(id: string | undefined): UseEvaluationsResult {
  const [evaluations, setEvaluations] = useState<RuleEvaluationResult[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async () => {
    if (!id) return;
    setLoading(true);
    setError(null);
    try {
      const response = await fetchEvaluations(id);
      setEvaluations(response.data);
    } catch (e) {
      if (e instanceof ApiError) {
        setError("Unable to load rule evaluations");
      }
    } finally {
      setLoading(false);
    }
  }, [id]);

  useEffect(() => { load(); }, [load]);

  return { evaluations, loading, error, retry: load };
}
