import React from "react";
import Button from "./ui/Button";

interface TransactionToolbarProps {
  autoRefresh: number;
  onAutoRefreshChange: (value: number) => void;
  refreshing: boolean;
  onRefresh: () => void;
}

const AUTO_REFRESH_OPTIONS = [
  { label: "Off", value: 0 },
  { label: "10s", value: 10 },
  { label: "30s", value: 30 },
  { label: "1min", value: 60 },
  { label: "5min", value: 300 },
];

const selectStyle: React.CSSProperties = {
  padding: "6px 10px",
  border: "1px solid var(--color-border)",
  borderRadius: "8px",
  backgroundColor: "var(--color-surface)",
  color: "var(--color-text)",
  fontSize: "0.8rem",
  fontFamily: "'Inter', sans-serif",
  cursor: "pointer",
};

const TransactionToolbar: React.FC<TransactionToolbarProps> = ({
  autoRefresh,
  onAutoRefreshChange,
  refreshing,
  onRefresh,
}) => (
  <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", marginBottom: "16px" }}>
    <h2 style={{ margin: 0 }}>Transactions</h2>
    <div style={{ display: "flex", alignItems: "center", gap: "10px" }}>
      <label style={{ display: "flex", alignItems: "center", gap: "6px", fontSize: "0.8rem", color: "var(--color-text-muted)" }}>
        Auto-refresh:
        <select
          style={selectStyle}
          value={autoRefresh}
          onChange={(e) => onAutoRefreshChange(Number(e.target.value))}
          aria-label="Auto-refresh interval"
        >
          {AUTO_REFRESH_OPTIONS.map((opt) => (
            <option key={opt.value} value={opt.value}>{opt.label}</option>
          ))}
        </select>
      </label>
      <Button loading={refreshing} onClick={onRefresh} aria-label="Refresh transactions">
        {refreshing ? "Refreshing…" : "↻ Refresh"}
      </Button>
    </div>
  </div>
);

export default React.memo(TransactionToolbar);
