import { describe, it, beforeEach } from "vitest";
import fc from "fast-check";
import { getTheme, setTheme } from "./useTheme";

// Feature: fraud-analyst-dashboard, Property 13: Theme preference persistence round-trip
describe("Property 13: Theme preference persistence round-trip", () => {
  beforeEach(() => {
    localStorage.clear();
  });

  it("persisting a theme to localStorage and reading back returns the same value", () => {
    // Validates: Requirements 9.3, 9.4
    const themeArb = fc.constantFrom("light" as const, "dark" as const);

    fc.assert(
      fc.property(themeArb, (theme) => {
        localStorage.clear();
        setTheme(theme);
        const restored = getTheme();
        return restored === theme;
      }),
      { numRuns: 100 }
    );
  });
});
