import type { Transaction, PaginatedResponse, SingleResponse } from "../types";

export type { Transaction };
export type TransactionsResponse = PaginatedResponse<Transaction>;
export type TransactionResponse = SingleResponse<Transaction>;

const BASE_URL = import.meta.env.VITE_TRANSACTION_API_URL ?? "http://localhost:3000";

export async function fetchTransactions(
  limit = 20,
  cursor?: string
): Promise<TransactionsResponse> {
  const params = new URLSearchParams({ limit: String(limit) });
  if (cursor) {
    params.set("cursor", cursor);
  }

  const response = await fetch(`${BASE_URL}/transactions?${params}`);
  if (!response.ok) {
    throw new Error(
      `Failed to fetch transactions: ${response.status} ${response.statusText}`
    );
  }
  return response.json() as Promise<TransactionsResponse>;
}

export async function fetchTransaction(id: string): Promise<TransactionResponse> {
  const response = await fetch(`${BASE_URL}/transactions/${encodeURIComponent(id)}`);
  if (!response.ok) {
    throw new Error(
      `Failed to fetch transaction ${id}: ${response.status} ${response.statusText}`
    );
  }
  return response.json() as Promise<TransactionResponse>;
}
