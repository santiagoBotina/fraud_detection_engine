import { describe, it, expect, vi } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import PaginationControls from "./PaginationControls";

describe("PaginationControls", () => {
  const defaultProps = {
    page: 1,
    hasNextPage: true,
    hasPreviousPage: false,
    goNextPage: vi.fn(),
    goPreviousPage: vi.fn(),
  };

  it("displays the current page number", () => {
    render(<PaginationControls {...defaultProps} page={3} />);
    expect(screen.getByText("Page 3")).toBeInTheDocument();
  });

  it("disables Previous button when hasPreviousPage is false", () => {
    render(<PaginationControls {...defaultProps} hasPreviousPage={false} />);
    expect(screen.getByRole("button", { name: /previous/i })).toBeDisabled();
  });

  it("enables Previous button when hasPreviousPage is true", () => {
    render(<PaginationControls {...defaultProps} hasPreviousPage={true} page={2} />);
    expect(screen.getByRole("button", { name: /previous/i })).toBeEnabled();
  });

  it("disables Next button when hasNextPage is false", () => {
    render(<PaginationControls {...defaultProps} hasNextPage={false} />);
    expect(screen.getByRole("button", { name: /next/i })).toBeDisabled();
  });

  it("enables Next button when hasNextPage is true", () => {
    render(<PaginationControls {...defaultProps} hasNextPage={true} />);
    expect(screen.getByRole("button", { name: /next/i })).toBeEnabled();
  });

  it("calls goNextPage when Next button is clicked", () => {
    const goNext = vi.fn();
    render(<PaginationControls {...defaultProps} goNextPage={goNext} />);
    fireEvent.click(screen.getByRole("button", { name: /next/i }));
    expect(goNext).toHaveBeenCalledOnce();
  });

  it("calls goPreviousPage when Previous button is clicked", () => {
    const goPrev = vi.fn();
    render(
      <PaginationControls
        {...defaultProps}
        hasPreviousPage={true}
        page={2}
        goPreviousPage={goPrev}
      />
    );
    fireEvent.click(screen.getByRole("button", { name: /previous/i }));
    expect(goPrev).toHaveBeenCalledOnce();
  });

  it("does not call goPreviousPage when Previous is disabled and clicked", () => {
    const goPrev = vi.fn();
    render(
      <PaginationControls {...defaultProps} hasPreviousPage={false} goPreviousPage={goPrev} />
    );
    fireEvent.click(screen.getByRole("button", { name: /previous/i }));
    expect(goPrev).not.toHaveBeenCalled();
  });
});
