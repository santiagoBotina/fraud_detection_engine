import { describe, it } from "vitest";
import fc from "fast-check";
import { formatCurrency, formatDate, computeEvaluationTime } from "./formatters";
import { getScoreColor } from "../components/ScoreIndicator";

// Feature: fraud-analyst-dashboard, Property 10: Currency and date formatting correctness
describe("Property 10: Currency and date formatting correctness", () => {
  it("currency formatter produces string with correct decimal amount", () => {
    // Validates: Requirements 7.1
    const currencyArb = fc.constantFrom("USD", "COP", "EUR");
    const amountArb = fc.integer({ min: 1, max: 100_000_000 });

    fc.assert(
      fc.property(amountArb, currencyArb, (amountInCents, currency) => {
        const result = formatCurrency(amountInCents, currency);
        const expectedDecimal = (amountInCents / 100).toFixed(2);
        // The formatted string must contain the decimal amount (ignoring thousands separators)
        const stripped = result.replace(/,/g, "");
        return stripped.includes(expectedDecimal);
      }),
      { numRuns: 100 }
    );
  });

  it("date formatter produces non-empty string for any valid ISO 8601 timestamp", () => {
    // Validates: Requirements 7.1
    const isoDateArb = fc.date({
      min: new Date("2000-01-01T00:00:00Z"),
      max: new Date("2099-12-31T23:59:59Z"),
    }).map((d) => d.toISOString());

    fc.assert(
      fc.property(isoDateArb, (isoString) => {
        const result = formatDate(isoString);
        return typeof result === "string" && result.length > 0;
      }),
      { numRuns: 100 }
    );
  });
});

// Feature: fraud-analyst-dashboard, Property 11: Fraud score color indicator mapping
describe("Property 11: Fraud score color indicator mapping", () => {
  it("returns correct color for any score in [0, 100]", () => {
    // Validates: Requirements 8.3
    const scoreArb = fc.integer({ min: 0, max: 100 });

    fc.assert(
      fc.property(scoreArb, (score) => {
        const color = getScoreColor(score);
        if (score < 30) return color === "green";
        if (score < 70) return color === "yellow";
        return color === "red";
      }),
      { numRuns: 100 }
    );
  });
});

// Feature: fraud-analyst-dashboard, Property 12: Evaluation time computation
describe("Property 12: Evaluation time computation", () => {
  it("returns non-negative duration equal to updatedAt - createdAt", () => {
    // Validates: Requirements 8.5
    const timestampPairArb = fc
      .date({
        min: new Date("2000-01-01T00:00:00Z"),
        max: new Date("2099-12-31T23:59:59Z"),
      })
      .chain((createdDate) =>
        fc
          .date({
            min: createdDate,
            max: new Date("2099-12-31T23:59:59Z"),
          })
          .map((updatedDate) => ({
            createdAt: createdDate.toISOString(),
            updatedAt: updatedDate.toISOString(),
            expectedDiff: updatedDate.getTime() - createdDate.getTime(),
          }))
      );

    fc.assert(
      fc.property(timestampPairArb, ({ createdAt, updatedAt, expectedDiff }) => {
        const result = computeEvaluationTime(createdAt, updatedAt);
        return result >= 0 && result === expectedDiff;
      }),
      { numRuns: 100 }
    );
  });
});
