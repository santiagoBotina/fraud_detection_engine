import { useState, useEffect, useCallback } from "react";

export type Theme = "light" | "dark";

const STORAGE_KEY = "theme";

/** Read theme from localStorage, defaulting to "light". */
export function getTheme(): Theme {
  try {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (stored === "light" || stored === "dark") return stored;
  } catch {
    // localStorage unavailable
  }
  return "light";
}

/** Write theme to localStorage. */
export function setTheme(theme: Theme): void {
  try {
    localStorage.setItem(STORAGE_KEY, theme);
  } catch {
    // localStorage unavailable
  }
}

function applyTheme(theme: Theme) {
  // Set data-theme attribute for CSS variable switching
  document.documentElement.setAttribute("data-theme", theme);
  // Keep direct body styles for test compatibility
  if (theme === "dark") {
    document.body.style.backgroundColor = "#1A1A2E";
    document.body.style.color = "#e8e8f0";
  } else {
    document.body.style.backgroundColor = "#FFFFFF";
    document.body.style.color = "#1e1e2e";
  }
}

export function useTheme() {
  const [theme, setThemeState] = useState<Theme>(getTheme);

  useEffect(() => {
    applyTheme(theme);
    setTheme(theme);
  }, [theme]);

  const toggleTheme = useCallback(() => {
    setThemeState((prev) => (prev === "light" ? "dark" : "light"));
  }, []);

  return { theme, toggleTheme } as const;
}
