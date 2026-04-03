import React from "react";

interface StatusBadgeProps {
  status: string;
}

const statusStyles: Record<string, React.CSSProperties> = {
  APPROVED: {
    backgroundColor: "#a3d9a5",
    color: "#1a5c2a",
  },
  DECLINED: {
    backgroundColor: "#f5a3a3",
    color: "#7c1d1d",
  },
  PENDING: {
    backgroundColor: "#f5d89a",
    color: "#6b4f10",
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

const StatusBadge: React.FC<StatusBadgeProps> = ({ status }) => {
  const colorStyle = statusStyles[status.toUpperCase()] ?? {
    backgroundColor: "#d4d4dc",
    color: "#4a4a5a",
  };

  return (
    <span style={{ ...baseStyle, ...colorStyle }}>{status}</span>
  );
};

export default StatusBadge;
