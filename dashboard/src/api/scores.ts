export interface FraudScore {
  transaction_id: string;
  fraud_score: number;
  calculated_at: string;
}

const BASE_URL = import.meta.env.VITE_FRAUD_SCORE_API_URL ?? "http://localhost:3002";

export async function fetchScore(transactionId: string): Promise<FraudScore> {
  const response = await fetch(
    `${BASE_URL}/scores/${encodeURIComponent(transactionId)}`
  );
  if (!response.ok) {
    throw new Error(
      `Failed to fetch score for ${transactionId}: ${response.status} ${response.statusText}`
    );
  }
  return response.json() as Promise<FraudScore>;
}
