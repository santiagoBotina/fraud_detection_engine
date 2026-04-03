import React from "react";
import type { RuleEvaluationResult } from "../types";
import Card from "./ui/Card";
import EmptyState from "./ui/EmptyState";
import ErrorBanner from "./ErrorBanner";

interface RuleEvaluationsTableProps {
  evaluations: RuleEvaluationResult[];
  loading: boolean;
  error: string | null;
  onRetry: () => void;
}

const tableStyle: React.CSSProperties = {
  width: "100%",
  minWidth: "650px",
  borderCollapse: "collapse",
  fontSize: "0.875rem",
  tableLayout: "auto",
};

const thStyle: React.CSSProperties = {
  textAlign: "left",
  padding: "10px 16px",
  borderBottom: "1px solid var(--color-border)",
  fontWeight: 600,
  fontSize: "0.75rem",
  textTransform: "uppercase",
  letterSpacing: "0.04em",
  color: "var(--color-text-muted)",
  whiteSpace: "nowrap",
};

const tdStyle: React.CSSProperties = {
  padding: "10px 16px",
  borderBottom: "1px solid var(--color-border-muted)",
  whiteSpace: "nowrap",
};

const condFieldStyle: React.CSSProperties = {
  display: "inline-block",
  padding: "2px 6px",
  borderRadius: "4px",
  backgroundColor: "#dbeafe",
  color: "#1e40af",
  fontFamily: "monospace",
  fontSize: "0.78rem",
  fontWeight: 500,
};

const condOperatorStyle: React.CSSProperties = {
  display: "inline-block",
  padding: "2px 6px",
  borderRadius: "4px",
  backgroundColor: "#f3e8ff",
  color: "#6b21a8",
  fontFamily: "monospace",
  fontSize: "0.78rem",
  fontWeight: 600,
};

const condValueStyle: React.CSSProperties = {
  display: "inline-block",
  padding: "2px 6px",
  borderRadius: "4px",
  backgroundColor: "#fef3c7",
  color: "#92400e",
  fontFamily: "monospace",
  fontSize: "0.78rem",
  fontWeight: 500,
};

const priorityBadgeStyle: React.CSSProperties = {
  display: "inline-flex",
  alignItems: "center",
  justifyContent: "center",
  width: "28px",
  height: "28px",
  borderRadius: "50%",
  backgroundColor: "var(--color-border-muted, #ededf0)",
  color: "var(--color-text-muted)",
  fontSize: "0.75rem",
  fontWeight: 700,
  fontFamily: "'Inter', sans-serif",
};

const matchDotStyle = (matched: boolean): React.CSSProperties => ({
  display: "inline-block",
  width: "8px",
  height: "8px",
  borderRadius: "50%",
  backgroundColor: matched ? "#a3d9a5" : "#f5a3a3",
  marginRight: "6px",
});

const RuleEvaluationsTable: React.FC<RuleEvaluationsTableProps> = ({
  evaluations,
  loading,
  error,
  onRetry,
}) => (
  <>
    <h3>Rule Evaluations</h3>
    {error ? <ErrorBanner message={error} onRetry={onRetry} /> : null}
    {loading ? <div data-testid="eval-loading">Loading evaluations…</div> : null}
    {!loading && !error && evaluations.length === 0 ? (
      <EmptyState message="No rule evaluations found" />
    ) : null}
    {!loading && !error && evaluations.length > 0 ? (
      <Card scrollable>
        <table style={tableStyle}>
          <thead>
            <tr>
              <th style={thStyle}>Priority</th>
              <th style={thStyle}>Rule Name</th>
              <th style={thStyle}>Condition</th>
              <th style={thStyle}>Matched</th>
              <th style={thStyle}>Result Status</th>
            </tr>
          </thead>
          <tbody>
            {evaluations.map((ev) => (
              <tr key={ev.rule_id}>
                <td style={{ ...tdStyle, textAlign: "center" }}>
                  <span style={priorityBadgeStyle}>{ev.priority}</span>
                </td>
                <td style={{ ...tdStyle, fontWeight: 500 }}>{ev.rule_name}</td>
                <td style={tdStyle}>
                  <span style={condFieldStyle}>{ev.condition_field}</span>{" "}
                  <span style={condOperatorStyle}>{ev.condition_operator}</span>{" "}
                  <span style={condValueStyle}>{ev.condition_value}</span>
                </td>
                <td style={tdStyle}>
                  <span style={matchDotStyle(ev.matched)} />
                  {ev.matched ? "Yes" : "No"}
                </td>
                <td style={tdStyle}>{ev.result_status}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </Card>
    ) : null}
  </>
);

export default React.memo(RuleEvaluationsTable);
