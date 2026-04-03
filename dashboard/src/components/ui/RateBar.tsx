import React from "react";

interface RateBarProps {
  rate: number;
  color: string;
}

const trackStyle: React.CSSProperties = {
  height: "8px",
  backgroundColor: "var(--color-border-muted, #ededf0)",
  borderRadius: "4px",
  overflow: "hidden",
  marginTop: "8px",
};

const RateBar: React.FC<RateBarProps> = ({ rate, color }) => (
  <div style={trackStyle}>
    <div
      style={{
        width: `${Math.min(rate, 100)}%`,
        height: "100%",
        backgroundColor: color,
        borderRadius: "4px",
        transition: "width 400ms ease",
      }}
    />
  </div>
);

export default React.memo(RateBar);
