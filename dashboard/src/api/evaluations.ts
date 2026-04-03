export interface RuleEvaluationResult {
  transaction_id: string;
  rule_id: string;
  rule_name: string;
  condition_field: string;
  condition_operator: string;
  condition_value: string;
  actual_field_value: string;
  matched: boolean;
  result_status: string;
  evaluated_at: string;
  priority: number;
}

export interface EvaluationsResponse {
  data: RuleEvaluationResult[];
}

export interface Rule {
  rule_id: string;
  rule_name: string;
  condition_field: string;
  condition_operator: string;
  condition_value: string;
  result_status: string;
  priority: number;
  is_active: boolean;
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
    throw new Error(
      `Failed to fetch evaluations for ${transactionId}: ${response.status} ${response.statusText}`
    );
  }
  return response.json() as Promise<EvaluationsResponse>;
}

export async function fetchRules(): Promise<RulesResponse> {
  const response = await fetch(`${BASE_URL}/rules`);
  if (!response.ok) {
    throw new Error(
      `Failed to fetch rules: ${response.status} ${response.statusText}`
    );
  }
  return response.json() as Promise<RulesResponse>;
}
