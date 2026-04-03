export function formatCurrency(amountInCents: number, currency: string): string {
  const amount = amountInCents / 100;
  const currencyCode = currency.toUpperCase();

  try {
    return new Intl.NumberFormat("en-US", {
      style: "currency",
      currency: currencyCode,
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
