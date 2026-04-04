// Shared domain types

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
  finalized_at?: string;
  finalization_latency_ms?: number;
}

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

export interface FraudScore {
  transaction_id: string;
  fraud_score: number;
  calculated_at: string;
}

export interface TransactionStatsResponse {
  today: number;
  this_week: number;
  this_month: number;
  total: number;
  approved: number;
  declined: number;
  pending: number;
  payment_methods: Record<string, number>;
  avg_latency_ms: number;
  finalized_count: number;
  latency_low: number;
  latency_medium: number;
  latency_high: number;
}

// API response wrappers
export interface PaginatedResponse<T> {
  data: T[];
  next_cursor: string | null;
}

export interface SingleResponse<T> {
  data: T;
}
