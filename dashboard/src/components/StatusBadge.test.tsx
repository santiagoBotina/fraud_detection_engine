import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import StatusBadge from "./StatusBadge";

describe("StatusBadge", () => {
  it("renders APPROVED with green background", () => {
    render(<StatusBadge status="APPROVED" />);
    const badge = screen.getByText("APPROVED");
    expect(badge).toBeInTheDocument();
    expect(badge.style.backgroundColor).toBe("rgb(163, 217, 165)");
  });

  it("renders DECLINED with red background", () => {
    render(<StatusBadge status="DECLINED" />);
    const badge = screen.getByText("DECLINED");
    expect(badge).toBeInTheDocument();
    expect(badge.style.backgroundColor).toBe("rgb(245, 163, 163)");
  });

  it("renders PENDING with yellow background", () => {
    render(<StatusBadge status="PENDING" />);
    const badge = screen.getByText("PENDING");
    expect(badge).toBeInTheDocument();
    expect(badge.style.backgroundColor).toBe("rgb(245, 216, 154)");
  });

  it("renders unknown status with gray background", () => {
    render(<StatusBadge status="UNKNOWN" />);
    const badge = screen.getByText("UNKNOWN");
    expect(badge).toBeInTheDocument();
    expect(badge.style.backgroundColor).toBe("rgb(212, 212, 220)");
  });
});
