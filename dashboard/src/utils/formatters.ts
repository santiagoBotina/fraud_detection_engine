export function formatCurrency(amountInCents: number, currency: string): string {
  const amount = amountInCents / 100;
  const currencyCode = currency.toUpperCase();

  try {
    return new Intl.NumberFormat("en-US", {
      style: "currency",
      currency: currencyCode,
      currencyDisplay: "narrowSymbol",
      minimumFractionDigits: 2,
      maximumFractionDigits: 2,
    }).format(amount);
  } catch {
    return `${amount.toFixed(2)} ${currencyCode}`;
  }
}

export function formatDate(isoString: string): string {
  const date = new Date(isoString);
  return date.toLocaleString("en-US", {
    year: "numeric",
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  });
}

export function computeEvaluationTime(
  createdAt: string,
  updatedAt: string
): number {
  return new Date(updatedAt).getTime() - new Date(createdAt).getTime();
}

export function formatLatency(ms: number): string {
  const seconds = ms / 1000;
  return `${seconds.toFixed(1)}s`;
}

export type LatencyTier = "LOW" | "MEDIUM" | "HIGH";

export function getLatencyTier(ms: number): LatencyTier {
  if (ms <= 2000) return "LOW";
  if (ms <= 5000) return "MEDIUM";
  return "HIGH";
}

export function getLatencyColor(tier: string): string {
  switch (tier) {
    case "LOW":
      return "#a3d9a5";
    case "MEDIUM":
      return "#f5d89a";
    case "HIGH":
      return "#f5a3a3";
    default:
      return "#a3d9a5";
  }
}
