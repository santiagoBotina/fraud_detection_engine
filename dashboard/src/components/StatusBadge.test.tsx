import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import StatusBadge from "./StatusBadge";

describe("StatusBadge", () => {
  it("renders APPROVED with correct CSS variable", () => {
    render(<StatusBadge status="APPROVED" />);
    const badge = screen.getByText("APPROVED");
    expect(badge).toBeInTheDocument();
    expect(badge.style.backgroundColor).toBe("var(--color-approved)");
  });

  it("renders DECLINED with correct CSS variable", () => {
    render(<StatusBadge status="DECLINED" />);
    const badge = screen.getByText("DECLINED");
    expect(badge).toBeInTheDocument();
    expect(badge.style.backgroundColor).toBe("var(--color-declined)");
  });

  it("renders PENDING with correct CSS variable", () => {
    render(<StatusBadge status="PENDING" />);
    const badge = screen.getByText("PENDING");
    expect(badge).toBeInTheDocument();
    expect(badge.style.backgroundColor).toBe("var(--color-pending)");
  });

  it("renders unknown status with gray background", () => {
    render(<StatusBadge status="UNKNOWN" />);
    const badge = screen.getByText("UNKNOWN");
    expect(badge).toBeInTheDocument();
    expect(badge.style.backgroundColor).toBe("rgb(212, 212, 220)");
  });

  it("applies pill styling", () => {
    render(<StatusBadge status="APPROVED" />);
    const badge = screen.getByText("APPROVED");
    expect(badge.style.borderRadius).toBe("9999px");
    expect(badge.style.textTransform).toBe("uppercase");
  });
});
