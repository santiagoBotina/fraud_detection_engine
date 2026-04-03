import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import Card from "./Card";

describe("Card", () => {
  it("renders children", () => {
    render(<Card>Hello</Card>);
    expect(screen.getByText("Hello")).toBeInTheDocument();
  });

  it("applies padding by default", () => {
    const { container } = render(<Card>Content</Card>);
    expect(container.firstChild).toHaveStyle({ padding: "24px" });
  });

  it("applies overflow when scrollable", () => {
    const { container } = render(<Card scrollable>Table</Card>);
    expect(container.firstChild).toHaveStyle({ overflowX: "auto" });
  });

  it("centers text when centered", () => {
    const { container } = render(<Card centered>Centered</Card>);
    expect(container.firstChild).toHaveStyle({ textAlign: "center" });
  });

  it("applies muted color", () => {
    const { container } = render(<Card muted>Muted</Card>);
    expect(container.firstChild).toHaveStyle({ color: "var(--color-text-muted)" });
  });
});
