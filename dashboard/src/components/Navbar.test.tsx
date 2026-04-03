import { describe, it, expect, vi } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import Navbar from "./Navbar";

describe("Navbar", () => {
  it("renders the app title", () => {
    render(<Navbar theme="light" onToggleTheme={() => {}} />);
    expect(screen.getByText("BastionIQ")).toBeInTheDocument();
  });

  it("renders the logo", () => {
    render(<Navbar theme="light" onToggleTheme={() => {}} />);
    expect(screen.getByAltText("BastionIQ logo")).toBeInTheDocument();
  });

  it("shows dark label in light mode", () => {
    render(<Navbar theme="light" onToggleTheme={() => {}} />);
    expect(screen.getByText(/Dark/)).toBeInTheDocument();
  });

  it("shows light label in dark mode", () => {
    render(<Navbar theme="dark" onToggleTheme={() => {}} />);
    expect(screen.getByText(/Light/)).toBeInTheDocument();
  });

  it("calls onToggleTheme when button is clicked", () => {
    const toggle = vi.fn();
    render(<Navbar theme="light" onToggleTheme={toggle} />);
    fireEvent.click(screen.getByRole("button", { name: /toggle theme/i }));
    expect(toggle).toHaveBeenCalledOnce();
  });

  it("has fixed positioning", () => {
    render(<Navbar theme="light" onToggleTheme={() => {}} />);
    const nav = screen.getByTestId("navbar");
    expect(nav.style.position).toBe("fixed");
  });
});
