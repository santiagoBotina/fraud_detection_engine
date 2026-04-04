import type { TransactionStatsResponse } from "../types";
import { ApiError } from "./errors";

const BASE_URL = import.meta.env.VITE_TRANSACTION_API_URL ?? "http://localhost:3000";

export async function fetchStats(): Promise<TransactionStatsResponse> {
  const response = await fetch(`${BASE_URL}/transactions/stats`);
  if (!response.ok) {
    throw new ApiError(
      response.status,
      `Failed to fetch stats: ${response.status} ${response.statusText}`
    );
  }
  return response.json() as Promise<TransactionStatsResponse>;
}
