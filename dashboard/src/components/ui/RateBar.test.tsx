import { describe, it, expect } from "vitest";
import { render } from "@testing-library/react";
import RateBar from "./RateBar";

describe("RateBar", () => {
  it("renders with correct width percentage", () => {
    const { container } = render(<RateBar rate={75} color="#a3d9a5" />);
    // The inner bar is the second div (child of the track div)
    const innerBar = container.querySelector("div > div > div") as HTMLElement;
    expect(innerBar.style.width).toBe("75%");
  });

  it("caps width at 100%", () => {
    const { container } = render(<RateBar rate={150} color="#a3d9a5" />);
    const innerBar = container.querySelector("div > div > div") as HTMLElement;
    expect(innerBar.style.width).toBe("100%");
  });

  it("applies the given color", () => {
    const { container } = render(<RateBar rate={50} color="#f5a3a3" />);
    const innerBar = container.querySelector("div > div > div") as HTMLElement;
    expect(innerBar.style.backgroundColor).toBe("rgb(245, 163, 163)");
  });
});
