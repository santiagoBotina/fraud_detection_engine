import type { RuleEvaluationResult, Rule } from "../types";
import { ApiError } from "./errors";

export type { RuleEvaluationResult, Rule };

export interface EvaluationsResponse {
  data: RuleEvaluationResult[];
}

export interface RulesResponse {
  data: Rule[];
}

const BASE_URL = import.meta.env.VITE_DECISION_API_URL ?? "http://localhost:3001";

export async function fetchEvaluations(
  transactionId: string
): Promise<EvaluationsResponse> {
  const response = await fetch(
    `${BASE_URL}/evaluations/${encodeURIComponent(transactionId)}`
  );
  if (!response.ok) {
    throw new ApiError(
      response.status,
      `Failed to fetch evaluations for ${transactionId}: ${response.status} ${response.statusText}`
    );
  }
  return response.json() as Promise<EvaluationsResponse>;
}

export async function fetchRules(): Promise<RulesResponse> {
  const response = await fetch(`${BASE_URL}/rules`);
  if (!response.ok) {
    throw new ApiError(
      response.status,
      `Failed to fetch rules: ${response.status} ${response.statusText}`
    );
  }
  return response.json() as Promise<RulesResponse>;
}
