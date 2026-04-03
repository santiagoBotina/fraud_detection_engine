import React, { useState, useEffect } from "react";
import { BrowserRouter, Routes, Route, useLocation } from "react-router-dom";
import TransactionList from "./pages/TransactionList";
import TransactionDetail from "./pages/TransactionDetail";
import { useTheme } from "./hooks/useTheme";

function FadeWrapper({ children }: { children: React.ReactNode }) {
  const location = useLocation();
  const [opacity, setOpacity] = useState(0);

  useEffect(() => {
    setOpacity(0);
    const frame = requestAnimationFrame(() => setOpacity(1));
    return () => cancelAnimationFrame(frame);
  }, [location.pathname]);

  return (
    <div style={{ opacity, transition: "opacity 300ms ease-in-out" }}>
      {children}
    </div>
  );
}

function AppContent() {
  const { theme, toggleTheme } = useTheme();

  return (
    <>
      <nav
        data-testid="navbar"
        style={{
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
        }}
      >
        <div style={{ display: "flex", alignItems: "center", gap: "10px" }}>
          <img
            src="/icons/logo.svg"
            alt="BastionIQ logo"
            style={{ width: "28px", height: "28px" }}
          />
          <span style={{ fontWeight: 700, fontSize: "16px", letterSpacing: "-0.02em" }}>
            BastionIQ
          </span>
        </div>
        <button
          onClick={toggleTheme}
          aria-label="Toggle theme"
          style={{
            padding: "6px 14px",
            borderRadius: "6px",
            border: "1px solid var(--color-border)",
            backgroundColor: "var(--color-bg)",
            color: "var(--color-text)",
            cursor: "pointer",
            fontWeight: 500,
            fontSize: "13px",
            fontFamily: "'Inter', sans-serif",
            transition: "all 150ms ease",
          }}
        >
          {theme === "light" ? "🌙 Dark" : "☀️ Light"}
        </button>
      </nav>
      <main style={{ paddingTop: "72px", padding: "72px 24px 48px" }}>
        <FadeWrapper>
          <Routes>
            <Route path="/" element={<TransactionList />} />
            <Route path="/transactions/:id" element={<TransactionDetail />} />
          </Routes>
        </FadeWrapper>
      </main>
    </>
  );
}

function App() {
  return (
    <BrowserRouter>
      <AppContent />
    </BrowserRouter>
  );
}

export default App;
