import { describe, it, expect, vi } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import PageSizeSelector from "./PageSizeSelector";

describe("PageSizeSelector", () => {
  it("renders all page size options", () => {
    render(<PageSizeSelector pageSize={20} onPageSizeChange={() => {}} />);
    const options = screen.getAllByRole("option");
    expect(options).toHaveLength(4);
    expect(options.map((o) => o.textContent)).toEqual(["20", "30", "50", "100"]);
  });

  it("selects the current pageSize value", () => {
    render(<PageSizeSelector pageSize={50} onPageSizeChange={() => {}} />);
    const select = screen.getByRole("combobox", { name: /page size/i });
    expect(select).toHaveValue("50");
  });

  it("defaults to 20 when pageSize is 20", () => {
    render(<PageSizeSelector pageSize={20} onPageSizeChange={() => {}} />);
    const select = screen.getByRole("combobox", { name: /page size/i });
    expect(select).toHaveValue("20");
  });

  it("calls onPageSizeChange with the selected numeric value", () => {
    const onChange = vi.fn();
    render(<PageSizeSelector pageSize={20} onPageSizeChange={onChange} />);
    fireEvent.change(screen.getByRole("combobox", { name: /page size/i }), {
      target: { value: "100" },
    });
    expect(onChange).toHaveBeenCalledWith(100);
  });

  it("has an accessible label", () => {
    render(<PageSizeSelector pageSize={20} onPageSizeChange={() => {}} />);
    expect(screen.getByLabelText(/page size/i)).toBeInTheDocument();
  });
});
