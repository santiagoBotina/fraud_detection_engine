import React from "react";

interface StatusBadgeProps {
  status: string;
}

const statusStyles: Record<string, React.CSSProperties> = {
  APPROVED: {
    backgroundColor: "var(--color-approved)",
    color: "var(--color-approved-text)",
  },
  DECLINED: {
    backgroundColor: "var(--color-declined)",
    color: "var(--color-declined-text)",
  },
  PENDING: {
    backgroundColor: "var(--color-pending)",
    color: "var(--color-pending-text)",
  },
};

const baseStyle: React.CSSProperties = {
  display: "inline-block",
  padding: "3px 12px",
  borderRadius: "9999px",
  fontSize: "0.75rem",
  fontWeight: 600,
  fontFamily: "'Inter', sans-serif",
  letterSpacing: "0.02em",
  textTransform: "uppercase",
};

const defaultColorStyle: React.CSSProperties = {
  backgroundColor: "#d4d4dc",
  color: "#4a4a5a",
};

const StatusBadge: React.FC<StatusBadgeProps> = ({ status }) => {
  const colorStyle = statusStyles[status.toUpperCase()] ?? defaultColorStyle;

  return (
    <span style={{ ...baseStyle, ...colorStyle }}>{status}</span>
  );
};

export default React.memo(StatusBadge);
