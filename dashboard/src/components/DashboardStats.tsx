import React, { useMemo } from "react";
import type { Transaction } from "../types";
import RateBar from "./ui/RateBar";

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

const rateRowStyle: React.CSSProperties = {
  display: "flex",
  justifyContent: "space-between",
  fontSize: "0.85rem",
  marginBottom: "2px",
};

function getTimeBuckets(transactions: Transaction[]) {
  const now = Date.now();
  const dayAgo = now - 86_400_000;
  const weekAgo = now - 604_800_000;
  const monthAgo = now - 2_592_000_000;

  let today = 0;
  let week = 0;
  let month = 0;

  for (const txn of transactions) {
    const d = new Date(txn.created_at).getTime();
    if (d >= dayAgo) today++;
    if (d >= weekAgo) week++;
    if (d >= monthAgo) month++;
  }

  return { today, week, month };
}

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

function computeStats(transactions: Transaction[]) {
  const total = transactions.length;
  let approved = 0;
  let declined = 0;
  let pending = 0;
  const methodCounts: Record<string, number> = {};

  // js-combine-iterations: single pass for all counts
  for (const txn of transactions) {
    if (txn.status === "APPROVED") approved++;
    else if (txn.status === "DECLINED") declined++;
    else if (txn.status === "PENDING") pending++;
    methodCounts[txn.payment_method] = (methodCounts[txn.payment_method] || 0) + 1;
  }

  const approvalRate = total > 0 ? Math.round((approved / total) * 100) : 0;
  const declineRate = total > 0 ? Math.round((declined / total) * 100) : 0;
  const pendingRate = total > 0 ? Math.round((pending / total) * 100) : 0;
  const buckets = getTimeBuckets(transactions);

  const topMethods = Object.entries(methodCounts)
    .sort((a, b) => b[1] - a[1])
    .slice(0, 4);

  return { total, approved, declined, pending, approvalRate, declineRate, pendingRate, buckets, topMethods };
}

const DashboardStats: React.FC<DashboardStatsProps> = ({ transactions }) => {
  const stats = useMemo(() => computeStats(transactions), [transactions]);

  if (stats.total === 0) return null;

  return (
    <div style={{ marginBottom: "28px", fontFamily: "'Inter', sans-serif" }}>
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

      <div style={{ display: "flex", gap: "16px", flexWrap: "wrap" }}>
        <div style={{ ...cardStyle, flex: "1 1 300px" }}>
          <div style={statLabel}>Decision Rates</div>
          <RateDisplay label="Approved" rate={stats.approvalRate} count={stats.approved} color="#a3d9a5" textColor="#1a5c2a" />
          <RateDisplay label="Declined" rate={stats.declineRate} count={stats.declined} color="#f5a3a3" textColor="#7c1d1d" />
          {stats.pending > 0 ? (
            <RateDisplay label="Pending" rate={stats.pendingRate} count={stats.pending} color="#f5d89a" textColor="#6b4f10" />
          ) : null}
        </div>

        <div style={{ ...cardStyle, flex: "1 1 300px" }}>
          <div style={statLabel}>Payment Methods</div>
          <div style={{ marginTop: "12px" }}>
            {stats.topMethods.length > 0 ? (
              stats.topMethods.map(([method, count]) => {
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
      </div>
    </div>
  );
};

export default React.memo(DashboardStats);
