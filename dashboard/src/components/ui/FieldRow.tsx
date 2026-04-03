import React from "react";

interface FieldRowProps {
  label: string;
  children: React.ReactNode;
  last?: boolean;
}

const rowStyle: React.CSSProperties = {
  display: "flex",
  padding: "8px 0",
  fontSize: "0.875rem",
};

const labelStyle: React.CSSProperties = {
  fontWeight: 600,
  width: "180px",
  flexShrink: 0,
  color: "var(--color-text-muted)",
  fontSize: "0.8rem",
  textTransform: "uppercase",
  letterSpacing: "0.03em",
};

const FieldRow: React.FC<FieldRowProps> = ({ label, children, last }) => (
  <div style={{ ...rowStyle, borderBottom: last ? "none" : "1px solid var(--color-border-muted)" }}>
    <span style={labelStyle}>{label}</span>
    <span>{children}</span>
  </div>
);

export default React.memo(FieldRow);
