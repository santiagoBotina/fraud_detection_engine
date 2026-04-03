import React from "react";
import type { Theme } from "../hooks/useTheme";
import Button from "./ui/Button";

interface NavbarProps {
  theme: Theme;
  onToggleTheme: () => void;
}

const navStyle: React.CSSProperties = {
  position: "fixed",
  top: 0,
  left: 0,
  right: 0,
  height: "56px",
  display: "flex",
  alignItems: "center",
  justifyContent: "space-between",
  padding: "0 24px",
  backgroundColor: "var(--color-surface)",
  borderBottom: "1px solid var(--color-border)",
  zIndex: 1000,
  backdropFilter: "blur(8px)",
  fontFamily: "'Inter', sans-serif",
};

const logoStyle: React.CSSProperties = {
  display: "flex",
  alignItems: "center",
  gap: "10px",
};

const titleStyle: React.CSSProperties = {
  fontWeight: 700,
  fontSize: "16px",
  letterSpacing: "-0.02em",
};

const Navbar: React.FC<NavbarProps> = ({ theme, onToggleTheme }) => (
  <nav data-testid="navbar" style={navStyle}>
    <div style={logoStyle}>
      <img src="/icons/logo.svg" alt="BastionIQ logo" style={{ width: "28px", height: "28px" }} />
      <span style={titleStyle}>BastionIQ</span>
    </div>
    <Button onClick={onToggleTheme} aria-label="Toggle theme">
      {theme === "light" ? "🌙 Dark" : "☀️ Light"}
    </Button>
  </nav>
);

export default React.memo(Navbar);
