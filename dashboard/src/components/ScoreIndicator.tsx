import React from "react";

export function getScoreColor(score: number): string {
  if (score < 30) return "green";
  if (score < 70) return "yellow";
  return "red";
}

export function getScoreLabel(score: number): string {
  if (score < 20) return "Very Low Risk";
  if (score < 30) return "Low Risk";
  if (score < 50) return "Moderate Risk";
  if (score < 70) return "Elevated Risk";
  if (score < 85) return "High Risk";
  return "Critical Risk";
}

export function getScoreDescription(score: number): string {
  if (score < 20)
    return "This transaction shows minimal indicators of fraudulent activity. No further action is typically required.";
  if (score < 30)
    return "This transaction has a low probability of being fraudulent. Standard monitoring applies.";
  if (score < 50)
    return "Some risk indicators were detected. The transaction may warrant a closer look by an analyst.";
  if (score < 70)
    return "Multiple risk factors are present. This transaction should be reviewed before final approval.";
  if (score < 85)
    return "Significant fraud indicators detected. This transaction is likely fraudulent and should be investigated.";
  return "Extreme fraud risk. This transaction has strong indicators of fraudulent activity and should be blocked or escalated immediately.";
}

const colorMap: Record<string, React.CSSProperties> = {
  green: { backgroundColor: "#a3d9a5", color: "#1a5c2a", borderColor: "#a3d9a5" },
  yellow: { backgroundColor: "#f5d89a", color: "#6b4f10", borderColor: "#f5d89a" },
  red: { backgroundColor: "#f5a3a3", color: "#7c1d1d", borderColor: "#f5a3a3" },
};

const progressBarBg: Record<string, string> = {
  green: "#a3d9a5",
  yellow: "#f5d89a",
  red: "#f5a3a3",
};

interface ScoreIndicatorProps {
  score: number;
  detailed?: boolean;
}

const badgeStyle: React.CSSProperties = {
  display: "inline-flex",
  alignItems: "center",
  justifyContent: "center",
  padding: "6px 16px",
  borderRadius: "8px",
  fontWeight: 700,
  fontSize: "1.1rem",
  fontFamily: "'Inter', sans-serif",
  minWidth: "48px",
};

const ScoreIndicator: React.FC<ScoreIndicatorProps> = ({ score, detailed = false }) => {
  const color = getScoreColor(score);
  const styles = colorMap[color];

  if (!detailed) {
    return <span style={{ ...badgeStyle, ...styles }}>{score}</span>;
  }

  const label = getScoreLabel(score);
  const description = getScoreDescription(score);
  const barColor = progressBarBg[color];

  return (
    <div style={{ fontFamily: "'Inter', sans-serif" }}>
      <div style={{ display: "flex", alignItems: "center", gap: "16px", marginBottom: "16px" }}>
        <span style={{ ...badgeStyle, ...styles, fontSize: "1.5rem", padding: "10px 20px" }}>
          {score}
        </span>
        <div>
          <div style={{ fontWeight: 600, fontSize: "1rem", marginBottom: "2px" }}>{label}</div>
          <div style={{ fontSize: "0.8rem", color: "var(--color-text-muted)" }}>
            Score range: 0 (safe) — 100 (fraud)
          </div>
        </div>
      </div>

      {/* Progress bar */}
      <div
        style={{
          width: "100%",
          height: "8px",
          backgroundColor: "var(--color-border-muted, #ededf0)",
          borderRadius: "4px",
          overflow: "hidden",
          marginBottom: "12px",
        }}
      >
        <div
          data-testid="score-bar"
          style={{
            width: `${score}%`,
            height: "100%",
            backgroundColor: barColor,
            borderRadius: "4px",
            transition: "width 400ms ease",
          }}
        />
      </div>

      <p style={{ fontSize: "0.85rem", lineHeight: 1.6, color: "var(--color-text-muted)", margin: 0 }}>
        {description}
      </p>
    </div>
  );
};

export default ScoreIndicator;
