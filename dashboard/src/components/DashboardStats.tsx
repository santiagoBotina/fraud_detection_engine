import React, { useMemo } from "react";
import { Transaction } from "../api/transactions";

interface DashboardStatsProps {
  transactions: Transaction[];
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

const barTrack: React.CSSProperties = {
  height: "8px",
  backgroundColor: "var(--color-border-muted, #ededf0)",
  borderRadius: "4px",
  overflow: "hidden",
  marginTop: "8px",
};

function RateBar({ rate, color }: { rate: number; color: string }) {
  return (
    <div style={barTrack}>
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
}

function getTimeBuckets(transactions: Transaction[]) {
  const now = new Date();
  const dayAgo = new Date(now.getTime() - 24 * 60 * 60 * 1000);
  const weekAgo = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000);
  const monthAgo = new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000);

  let today = 0;
  let week = 0;
  let month = 0;

  for (const txn of transactions) {
    const d = new Date(txn.created_at);
    if (d >= dayAgo) today++;
    if (d >= weekAgo) week++;
    if (d >= monthAgo) month++;
  }

  return { today, week, month };
}

const DashboardStats: React.FC<DashboardStatsProps> = ({ transactions }) => {
  const stats = useMemo(() => {
    const total = transactions.length;
    const approved = transactions.filter((t) => t.status === "APPROVED").length;
    const declined = transactions.filter((t) => t.status === "DECLINED").length;
    const pending = transactions.filter((t) => t.status === "PENDING").length;
    const approvalRate = total > 0 ? Math.round((approved / total) * 100) : 0;
    const declineRate = total > 0 ? Math.round((declined / total) * 100) : 0;
    const buckets = getTimeBuckets(transactions);

    // Payment method distribution
    const methodCounts: Record<string, number> = {};
    for (const txn of transactions) {
      methodCounts[txn.payment_method] = (methodCounts[txn.payment_method] || 0) + 1;
    }
    const topMethods = Object.entries(methodCounts)
      .sort((a, b) => b[1] - a[1])
      .slice(0, 4);

    return { total, approved, declined, pending, approvalRate, declineRate, buckets, topMethods };
  }, [transactions]);

  if (stats.total === 0) return null;

  return (
    <div style={{ marginBottom: "28px", fontFamily: "'Inter', sans-serif" }}>
      {/* Top row: volume cards */}
      <div style={{ display: "flex", gap: "16px", flexWrap: "wrap", marginBottom: "16px" }}>
        <div style={cardStyle}>
          <div style={statLabel}>Today</div>
          <div style={statValue}>{stats.buckets.today}</div>
        </div>
        <div style={cardStyle}>
          <div style={statLabel}>This Week</div>
          <div style={statValue}>{stats.buckets.week}</div>
        </div>
        <div style={cardStyle}>
          <div style={statLabel}>This Month</div>
          <div style={statValue}>{stats.buckets.month}</div>
        </div>
        <div style={cardStyle}>
          <div style={statLabel}>Total Loaded</div>
          <div style={statValue}>{stats.total}</div>
        </div>
      </div>

      {/* Bottom row: rates + payment methods */}
      <div style={{ display: "flex", gap: "16px", flexWrap: "wrap" }}>
        {/* Approval / Decline rates */}
        <div style={{ ...cardStyle, flex: "1 1 300px" }}>
          <div style={statLabel}>Decision Rates</div>
          <div style={{ marginTop: "12px" }}>
            <div style={{ display: "flex", justifyContent: "space-between", fontSize: "0.85rem", marginBottom: "2px" }}>
              <span>Approved</span>
              <span style={{ fontWeight: 600, color: "#1a5c2a" }}>{stats.approvalRate}% ({stats.approved})</span>
            </div>
            <RateBar rate={stats.approvalRate} color="#a3d9a5" />
          </div>
          <div style={{ marginTop: "12px" }}>
            <div style={{ display: "flex", justifyContent: "space-between", fontSize: "0.85rem", marginBottom: "2px" }}>
              <span>Declined</span>
              <span style={{ fontWeight: 600, color: "#7c1d1d" }}>{stats.declineRate}% ({stats.declined})</span>
            </div>
            <RateBar rate={stats.declineRate} color="#f5a3a3" />
          </div>
          {stats.pending > 0 && (
            <div style={{ marginTop: "12px" }}>
              <div style={{ display: "flex", justifyContent: "space-between", fontSize: "0.85rem", marginBottom: "2px" }}>
                <span>Pending</span>
                <span style={{ fontWeight: 600, color: "#6b4f10" }}>
                  {stats.total > 0 ? Math.round((stats.pending / stats.total) * 100) : 0}% ({stats.pending})
                </span>
              </div>
              <RateBar rate={stats.total > 0 ? (stats.pending / stats.total) * 100 : 0} color="#f5d89a" />
            </div>
          )}
        </div>

        {/* Payment method distribution */}
        <div style={{ ...cardStyle, flex: "1 1 300px" }}>
          <div style={statLabel}>Payment Methods</div>
          <div style={{ marginTop: "12px" }}>
            {stats.topMethods.map(([method, count]) => {
              const pct = stats.total > 0 ? Math.round((count / stats.total) * 100) : 0;
              return (
                <div key={method} style={{ marginBottom: "10px" }}>
                  <div style={{ display: "flex", justifyContent: "space-between", fontSize: "0.85rem", marginBottom: "2px" }}>
                    <span>{method}</span>
                    <span style={{ fontWeight: 600, color: "var(--color-text-muted)" }}>{pct}% ({count})</span>
                  </div>
                  <RateBar rate={pct} color="var(--color-primary, #7c6fea)" />
                </div>
              );
            })}
            {stats.topMethods.length === 0 && (
              <div style={{ color: "var(--color-text-muted)", fontSize: "0.85rem" }}>No data</div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};

export default DashboardStats;
