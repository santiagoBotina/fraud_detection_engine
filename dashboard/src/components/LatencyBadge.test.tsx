import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import fc from "fast-check";
import LatencyBadge from "./LatencyBadge";

// Feature: transaction-finalization-latency, Property 8: LatencyBadge renders null for undefined latency
describe("Property 8: LatencyBadge renders null for undefined latency", () => {
  it("renders nothing when latencyMs is undefined", () => {
    // **Validates: Requirements 6.3, 7.3**
    fc.assert(
      fc.property(
        fc.constant(undefined),
        () => {
          const { container } = render(<LatencyBadge />);
          expect(container.innerHTML).toBe("");
        }
      ),
      { numRuns: 100 }
    );
  });

  it("renders nothing for any nullish latency value", () => {
    // **Validates: Requirements 6.3, 7.3**
    const nullishArb = fc.constantFrom(undefined, null);

    fc.assert(
      fc.property(nullishArb, (value) => {
        const { container } = render(
          <LatencyBadge latencyMs={value as undefined} />
        );
        expect(container.innerHTML).toBe("");
      }),
      { numRuns: 100 }
    );
  });
});

// Unit tests for LatencyBadge component
// Validates: Requirements 6.1, 6.2, 6.3, 7.2, 7.3
describe("LatencyBadge unit tests", () => {
  it("renders green badge with '1.0s' for 1000ms (LOW tier)", () => {
    render(<LatencyBadge latencyMs={1000} />);
    const badge = screen.getByText("1.0s");
    expect(badge).toBeInTheDocument();
    expect(badge.style.backgroundColor).toBe("rgb(163, 217, 165)");
  });

  it("renders yellow badge with '3.0s' for 3000ms (MEDIUM tier)", () => {
    render(<LatencyBadge latencyMs={3000} />);
    const badge = screen.getByText("3.0s");
    expect(badge).toBeInTheDocument();
    expect(badge.style.backgroundColor).toBe("rgb(245, 216, 154)");
  });

  it("renders red badge with '6.0s' for 6000ms (HIGH tier)", () => {
    render(<LatencyBadge latencyMs={6000} />);
    const badge = screen.getByText("6.0s");
    expect(badge).toBeInTheDocument();
    expect(badge.style.backgroundColor).toBe("rgb(245, 163, 163)");
  });

  it("renders nothing when latencyMs is undefined", () => {
    const { container } = render(<LatencyBadge />);
    expect(container.innerHTML).toBe("");
  });
});
