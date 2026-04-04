import React from "react";
import { formatLatency, getLatencyTier, getLatencyColor } from "../utils/formatters";

interface LatencyBadgeProps {
  latencyMs?: number;
}

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

const LatencyBadge: React.FC<LatencyBadgeProps> = ({ latencyMs }) => {
  if (latencyMs == null) {
    return null;
  }

  const tier = getLatencyTier(latencyMs);
  const backgroundColor = getLatencyColor(tier);

  return (
    <span style={{ ...baseStyle, backgroundColor, color: "#1a1a2e" }}>
      {formatLatency(latencyMs)}
    </span>
  );
};

export default React.memo(LatencyBadge);
