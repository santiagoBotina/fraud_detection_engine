import { render, screen, fireEvent } from "@testing-library/react";
import { describe, it, expect, beforeEach, vi } from "vitest";
import App from "./App";

// Mock the page components to isolate App routing/layout tests
vi.mock("./pages/TransactionList", () => ({
  default: () => <div>TransactionList Page</div>,
}));
vi.mock("./pages/TransactionDetail", () => ({
  default: () => <div>TransactionDetail Page</div>,
}));

beforeEach(() => {
  localStorage.clear();
  document.body.style.backgroundColor = "";
  document.body.style.color = "";
  window.history.pushState({}, "", "/");
});

describe("App", () => {
  it("renders the navbar with app title", () => {
    render(<App />);
    expect(screen.getByText("BastionIQ")).toBeInTheDocument();
  });

  it("renders the theme toggle button", () => {
    render(<App />);
    expect(screen.getByRole("button", { name: /toggle theme/i })).toBeInTheDocument();
  });

  it("renders TransactionList at /", () => {
    render(<App />);
    expect(screen.getByText("TransactionList Page")).toBeInTheDocument();
  });

  it("renders TransactionDetail at /transactions/:id", () => {
    window.history.pushState({}, "", "/transactions/txn_123");
    render(<App />);
    expect(screen.getByText("TransactionDetail Page")).toBeInTheDocument();
  });

  it("toggles theme from light to dark", () => {
    render(<App />);
    const btn = screen.getByRole("button", { name: /toggle theme/i });
    expect(btn.textContent).toContain("Dark");

    fireEvent.click(btn);
    expect(btn.textContent).toContain("Light");
    expect(localStorage.getItem("theme")).toBe("dark");
    expect(document.body.style.backgroundColor).toBe("rgb(26, 26, 46)");
  });

  it("restores dark theme from localStorage", () => {
    localStorage.setItem("theme", "dark");
    render(<App />);
    const btn = screen.getByRole("button", { name: /toggle theme/i });
    expect(btn.textContent).toContain("Light");
    expect(document.body.style.backgroundColor).toBe("rgb(26, 26, 46)");
  });

  it("defaults to light theme", () => {
    render(<App />);
    expect(document.body.style.backgroundColor).toBe("rgb(255, 255, 255)");
    expect(document.body.style.color).toBe("rgb(30, 30, 46)");
  });

  it("has a fixed navbar", () => {
    render(<App />);
    const navbar = screen.getByTestId("navbar");
    expect(navbar.style.position).toBe("fixed");
  });
});
