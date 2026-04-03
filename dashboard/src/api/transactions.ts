export interface Transaction {
  id: string;
  amount_in_cents: number;
  currency: string;
  payment_method: string;
  customer_id: string;
  customer_name: string;
  customer_email: string;
  customer_phone: string;
  customer_ip_address: string;
  status: string;
  created_at: string;
  updated_at: string;
}

export interface TransactionsResponse {
  data: Transaction[];
  next_cursor: string | null;
}

export interface TransactionResponse {
  data: Transaction;
}

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
