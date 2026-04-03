import { useState, useCallback, useEffect } from "react";
import { fetchTransaction, type Transaction } from "../api/transactions";

interface UseTransactionResult {
  transaction: Transaction | null;
  loading: boolean;
  error: string | null;
  retry: () => void;
}

export function useTransaction(id: string | undefined): UseTransactionResult {
  const [transaction, setTransaction] = useState<Transaction | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async () => {
    if (!id) return;
    setLoading(true);
    setError(null);
    try {
      const response = await fetchTransaction(id);
      setTransaction(response.data);
    } catch {
      setError("Unable to connect to Transaction service");
    } finally {
      setLoading(false);
    }
  }, [id]);

  useEffect(() => { load(); }, [load]);

  return { transaction, loading, error, retry: load };
}
