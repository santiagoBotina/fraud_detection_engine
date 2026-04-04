import React from "react";
import type { TransactionStatsResponse } from "../types";
import { formatLatency, getLatencyColor } from "../utils/formatters";
import RateBar from "./ui/RateBar";
import ErrorBanner from "./ErrorBanner";

interface DashboardStatsProps {
  stats: TransactionStatsResponse | null;
  loading: boolean;
  error: string | null;
  onRetry: () => void;
}

const cardStyle: React.CSSProperties = {
  backgroundColor: "var(--color-surface)",
  border: "1px solid var(--color-border)",
  borderRadius: "12px",
  boxShadow: "var(--shadow-sm)",
  padding: "20px",
  flex: "1 1 0",
  minWidth: "200px",
};

const statValue: React.CSSProperties = {
  fontSize: "1.8rem",
  fontWeight: 700,
  letterSpacing: "-0.02em",
  lineHeight: 1.2,
};

const statLabel: React.CSSProperties = {
  fontSize: "0.75rem",
  fontWeight: 600,
  textTransform: "uppercase",
  letterSpacing: "0.04em",
  color: "var(--color-text-muted)",
  marginBottom: "4px",
};

const rateRowStyle: React.CSSProperties = {
  display: "flex",
  justifyContent: "space-between",
  fontSize: "0.85rem",
  marginBottom: "2px",
};

interface RateDisplayProps {
  label: string;
  rate: number;
  count: number;
  color: string;
  textColor: string;
}

function RateDisplay({ label, rate, count, color, textColor }: RateDisplayProps) {
  return (
    <div style={{ marginTop: "12px" }}>
      <div style={rateRowStyle}>
        <span>{label}</span>
        <span style={{ fontWeight: 600, color: textColor }}>{rate}% ({count})</span>
      </div>
      <RateBar rate={rate} color={color} />
    </div>
  );
}

const DashboardStats: React.FC<DashboardStatsProps> = ({ stats, loading, error, onRetry }) => {
  if (error) {
    return (
      <div style={{ marginBottom: "28px", fontFamily: "'Inter', sans-serif" }}>
        <ErrorBanner message={error} onRetry={onRetry} />
      </div>
    );
  }

  if (loading || !stats) {
    return (
      <div style={{ marginBottom: "28px", fontFamily: "'Inter', sans-serif" }}>
        <div style={{ color: "var(--color-text-muted)", padding: "20px", textAlign: "center" }}>
          Loading stats…
        </div>
      </div>
    );
  }

  if (stats.total === 0) return null;

  const approvalRate = Math.round((stats.approved / stats.total) * 100);
  const declineRate = Math.round((stats.declined / stats.total) * 100);
  const pendingRate = Math.round((stats.pending / stats.total) * 100);

  const topMethods = Object.entries(stats.payment_methods)
    .sort((a, b) => b[1] - a[1])
    .slice(0, 4);

  return (
    <div style={{ marginBottom: "28px", fontFamily: "'Inter', sans-serif" }}>
      <div style={{ display: "flex", gap: "16px", flexWrap: "wrap", marginBottom: "16px" }}>
        <div style={cardStyle}>
          <div style={statLabel}>Today</div>
          <div style={statValue}>{stats.today}</div>
        </div>
        <div style={cardStyle}>
          <div style={statLabel}>This Week</div>
          <div style={statValue}>{stats.this_week}</div>
        </div>
        <div style={cardStyle}>
          <div style={statLabel}>This Month</div>
          <div style={statValue}>{stats.this_month}</div>
        </div>
        <div style={cardStyle}>
          <div style={statLabel}>Total</div>
          <div style={statValue}>{stats.total}</div>
        </div>
      </div>

      <div style={{ display: "flex", gap: "16px", flexWrap: "wrap" }}>
        <div style={{ ...cardStyle, flex: "1 1 300px" }}>
          <div style={statLabel}>Decision Rates</div>
          <RateDisplay label="Approved" rate={approvalRate} count={stats.approved} color="#a3d9a5" textColor="#1a5c2a" />
          <RateDisplay label="Declined" rate={declineRate} count={stats.declined} color="#f5a3a3" textColor="#7c1d1d" />
          {stats.pending > 0 ? (
            <RateDisplay label="Pending" rate={pendingRate} count={stats.pending} color="#f5d89a" textColor="#6b4f10" />
          ) : null}
        </div>

        <div style={{ ...cardStyle, flex: "1 1 300px" }}>
          <div style={statLabel}>Payment Methods</div>
          <div style={{ marginTop: "12px" }}>
            {topMethods.length > 0 ? (
              topMethods.map(([method, count]) => {
                const pct = Math.round((count / stats.total) * 100);
                return (
                  <div key={method} style={{ marginBottom: "10px" }}>
                    <div style={rateRowStyle}>
                      <span>{method}</span>
                      <span style={{ fontWeight: 600, color: "var(--color-text-muted)" }}>{pct}% ({count})</span>
                    </div>
                    <RateBar rate={pct} color="var(--color-primary, #7c6fea)" />
                  </div>
                );
              })
            ) : (
              <div style={{ color: "var(--color-text-muted)", fontSize: "0.85rem" }}>No data</div>
            )}
          </div>
        </div>

        {stats.finalized_count > 0 && (
          <div style={{ ...cardStyle, flex: "1 1 300px" }}>
            <div style={statLabel}>Finalization Latency</div>
            <div style={{ ...statValue, marginTop: "4px", marginBottom: "8px" }}>
              {formatLatency(stats.avg_latency_ms)}
            </div>
            <div style={{ fontSize: "0.75rem", color: "var(--color-text-muted)", marginBottom: "8px" }}>
              avg across {stats.finalized_count} finalized
            </div>
            <RateDisplay
              label="Low"
              rate={Math.round((stats.latency_low / stats.finalized_count) * 100)}
              count={stats.latency_low}
              color={getLatencyColor("LOW")}
              textColor="#1a5c2a"
            />
            <RateDisplay
              label="Medium"
              rate={Math.round((stats.latency_medium / stats.finalized_count) * 100)}
              count={stats.latency_medium}
              color={getLatencyColor("MEDIUM")}
              textColor="#6b4f10"
            />
            <RateDisplay
              label="High"
              rate={Math.round((stats.latency_high / stats.finalized_count) * 100)}
              count={stats.latency_high}
              color={getLatencyColor("HIGH")}
              textColor="#7c1d1d"
            />
          </div>
        )}
      </div>
    </div>
  );
};

export default React.memo(DashboardStats);
