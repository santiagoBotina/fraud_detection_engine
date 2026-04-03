import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import ScoreIndicator, { getScoreColor, getScoreLabel, getScoreDescription } from "./ScoreIndicator";

describe("getScoreColor", () => {
  it("returns green for scores below 30", () => {
    expect(getScoreColor(0)).toBe("green");
    expect(getScoreColor(15)).toBe("green");
    expect(getScoreColor(29)).toBe("green");
  });

  it("returns yellow for scores 30-69", () => {
    expect(getScoreColor(30)).toBe("yellow");
    expect(getScoreColor(50)).toBe("yellow");
    expect(getScoreColor(69)).toBe("yellow");
  });

  it("returns red for scores 70 and above", () => {
    expect(getScoreColor(70)).toBe("red");
    expect(getScoreColor(85)).toBe("red");
    expect(getScoreColor(100)).toBe("red");
  });
});

describe("ScoreIndicator (simple)", () => {
  it("renders score with green styling for low scores", () => {
    render(<ScoreIndicator score={10} />);
    const el = screen.getByText("10");
    expect(el).toBeInTheDocument();
    expect(el.style.backgroundColor).toBe("rgb(163, 217, 165)");
  });

  it("renders score with yellow styling for medium scores", () => {
    render(<ScoreIndicator score={50} />);
    const el = screen.getByText("50");
    expect(el.style.backgroundColor).toBe("rgb(245, 216, 154)");
  });

  it("renders score with red styling for high scores", () => {
    render(<ScoreIndicator score={85} />);
    const el = screen.getByText("85");
    expect(el.style.backgroundColor).toBe("rgb(245, 163, 163)");
  });
});

describe("ScoreIndicator (detailed)", () => {
  it("renders score label and description", () => {
    render(<ScoreIndicator score={42} detailed />);
    expect(screen.getByText("42")).toBeInTheDocument();
    expect(screen.getByText("Moderate Risk")).toBeInTheDocument();
    expect(screen.getByText(/Some risk indicators/)).toBeInTheDocument();
  });

  it("renders progress bar", () => {
    render(<ScoreIndicator score={75} detailed />);
    const bar = screen.getByTestId("score-bar");
    expect(bar.style.width).toBe("75%");
  });
});

describe("getScoreLabel", () => {
  it("returns correct labels for score ranges", () => {
    expect(getScoreLabel(5)).toBe("Very Low Risk");
    expect(getScoreLabel(25)).toBe("Low Risk");
    expect(getScoreLabel(42)).toBe("Moderate Risk");
    expect(getScoreLabel(60)).toBe("Elevated Risk");
    expect(getScoreLabel(75)).toBe("High Risk");
    expect(getScoreLabel(90)).toBe("Critical Risk");
  });
});

describe("getScoreDescription", () => {
  it("returns non-empty descriptions for all ranges", () => {
    expect(getScoreDescription(5).length).toBeGreaterThan(0);
    expect(getScoreDescription(25).length).toBeGreaterThan(0);
    expect(getScoreDescription(42).length).toBeGreaterThan(0);
    expect(getScoreDescription(60).length).toBeGreaterThan(0);
    expect(getScoreDescription(75).length).toBeGreaterThan(0);
    expect(getScoreDescription(90).length).toBeGreaterThan(0);
  });
});
