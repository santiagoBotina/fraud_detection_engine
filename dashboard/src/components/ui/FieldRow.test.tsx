import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import FieldRow from "./FieldRow";

describe("FieldRow", () => {
  it("renders label and value", () => {
    render(<FieldRow label="Name">Jane Doe</FieldRow>);
    expect(screen.getByText("Name")).toBeInTheDocument();
    expect(screen.getByText("Jane Doe")).toBeInTheDocument();
  });

  it("has border-bottom by default", () => {
    const { container } = render(<FieldRow label="Test">Value</FieldRow>);
    expect(container.firstChild).toHaveStyle({ borderBottom: "1px solid var(--color-border-muted)" });
  });

  it("removes border-bottom when last", () => {
    const { container } = render(<FieldRow label="Test" last>Value</FieldRow>);
    expect(container.firstChild).toHaveStyle({ borderBottom: "none" });
  });
});
