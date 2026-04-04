import { describe, it, expect } from "vitest";
import {
  formatCurrency,
  formatDate,
  computeEvaluationTime,
  formatLatency,
  getLatencyTier,
  getLatencyColor,
} from "./formatters";

describe("formatCurrency", () => {
  it("formats USD cents to dollars", () => {
    expect(formatCurrency(15000, "USD")).toBe("$150.00");
  });

  it("formats zero cents", () => {
    expect(formatCurrency(0, "USD")).toBe("$0.00");
  });

  it("formats single cent", () => {
    expect(formatCurrency(1, "USD")).toBe("$0.01");
  });

  it("formats EUR currency", () => {
    const result = formatCurrency(9999, "EUR");
    expect(result).toContain("99.99");
  });

  it("formats COP currency", () => {
    const result = formatCurrency(500000, "COP");
    expect(result).toContain("5,000.00");
  });
});

describe("formatDate", () => {
  it("formats ISO 8601 string to human-readable date", () => {
    const result = formatDate("2025-01-15T10:30:00Z");
    expect(result).toBeTruthy();
    expect(result.length).toBeGreaterThan(0);
    expect(result).toContain("2025");
    expect(result).toContain("Jan");
  });

  it("returns a non-empty string for any valid ISO date", () => {
    const result = formatDate("2023-06-01T00:00:00Z");
    expect(result.length).toBeGreaterThan(0);
  });
});

describe("computeEvaluationTime", () => {
  it("returns difference in milliseconds", () => {
    const created = "2025-01-15T10:30:00Z";
    const updated = "2025-01-15T10:30:05Z";
    expect(computeEvaluationTime(created, updated)).toBe(5000);
  });

  it("returns 0 when timestamps are equal", () => {
    const ts = "2025-01-15T10:30:00Z";
    expect(computeEvaluationTime(ts, ts)).toBe(0);
  });

  it("returns negative when updatedAt is before createdAt", () => {
    const created = "2025-01-15T10:30:05Z";
    const updated = "2025-01-15T10:30:00Z";
    expect(computeEvaluationTime(created, updated)).toBe(-5000);
  });
});

describe("formatLatency", () => {
  it("formats 1200ms as 1.2s", () => {
    expect(formatLatency(1200)).toBe("1.2s");
  });

  it("formats 3500ms as 3.5s", () => {
    expect(formatLatency(3500)).toBe("3.5s");
  });

  it("formats 500ms as 0.5s", () => {
    expect(formatLatency(500)).toBe("0.5s");
  });
});

describe("getLatencyTier", () => {
  it("returns LOW for 2000ms (boundary)", () => {
    expect(getLatencyTier(2000)).toBe("LOW");
  });

  it("returns MEDIUM for 2001ms (just above LOW)", () => {
    expect(getLatencyTier(2001)).toBe("MEDIUM");
  });

  it("returns MEDIUM for 5000ms (boundary)", () => {
    expect(getLatencyTier(5000)).toBe("MEDIUM");
  });

  it("returns HIGH for 5001ms (just above MEDIUM)", () => {
    expect(getLatencyTier(5001)).toBe("HIGH");
  });
});

describe("getLatencyColor", () => {
  it('returns green for LOW tier', () => {
    expect(getLatencyColor("LOW")).toBe("#a3d9a5");
  });

  it('returns yellow for MEDIUM tier', () => {
    expect(getLatencyColor("MEDIUM")).toBe("#f5d89a");
  });

  it('returns red for HIGH tier', () => {
    expect(getLatencyColor("HIGH")).toBe("#f5a3a3");
  });
});
