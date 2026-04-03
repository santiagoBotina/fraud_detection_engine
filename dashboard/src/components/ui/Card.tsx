import React from "react";

interface CardProps {
  children: React.ReactNode;
  scrollable?: boolean;
  centered?: boolean;
  muted?: boolean;
  style?: React.CSSProperties;
}

const baseStyle: React.CSSProperties = {
  backgroundColor: "var(--color-surface)",
  border: "1px solid var(--color-border)",
  borderRadius: "12px",
  boxShadow: "var(--shadow-sm)",
  marginBottom: "24px",
};

const Card: React.FC<CardProps> = ({ children, scrollable, centered, muted, style }) => {
  const mergedStyle: React.CSSProperties = {
    ...baseStyle,
    ...(scrollable ? { overflowX: "auto" } : { padding: "24px" }),
    ...(centered ? { textAlign: "center" } : {}),
    ...(muted ? { color: "var(--color-text-muted)" } : {}),
    ...style,
  };

  return <div style={mergedStyle}>{children}</div>;
};

export default React.memo(Card);
