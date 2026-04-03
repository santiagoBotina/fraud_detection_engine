import React, { useState, useEffect, useCallback, useRef } from "react";
import { Link } from "react-router-dom";
import { fetchTransactions, Transaction } from "../api/transactions";
import StatusBadge from "../components/StatusBadge";
import ErrorBanner from "../components/ErrorBanner";
import DashboardStats from "../components/DashboardStats";
import { formatCurrency, formatDate } from "../utils/formatters";

const cardStyle: React.CSSProperties = {
  backgroundColor: "var(--color-surface)",
  border: "1px solid var(--color-border)",
  borderRadius: "12px",
  boxShadow: "var(--shadow-sm)",
  overflow: "hidden",
};

const tableStyle: React.CSSProperties = {
  width: "100%",
  borderCollapse: "collapse",
  fontFamily: "'Inter', sans-serif",
  fontSize: "0.875rem",
};

const thStyle: React.CSSProperties = {
  textAlign: "left",
  padding: "12px 16px",
  borderBottom: "1px solid var(--color-border)",
  fontWeight: 600,
  fontSize: "0.75rem",
  textTransform: "uppercase",
  letterSpacing: "0.04em",
  color: "var(--color-text-muted)",
};

const tdStyle: React.CSSProperties = {
  padding: "12px 16px",
  borderBottom: "1px solid var(--color-border-muted)",
  color: "var(--color-text)",
};

const btnStyle: React.CSSProperties = {
  padding: "6px 16px",
  border: "1px solid var(--color-border)",
  borderRadius: "8px",
  backgroundColor: "var(--color-surface)",
  color: "var(--color-text)",
  cursor: "pointer",
  fontWeight: 500,
  fontSize: "0.8rem",
  fontFamily: "'Inter', sans-serif",
  transition: "all 150ms ease",
  boxShadow: "var(--shadow-sm)",
};

const selectStyle: React.CSSProperties = {
  padding: "6px 10px",
  border: "1px solid var(--color-border)",
  borderRadius: "8px",
  backgroundColor: "var(--color-surface)",
  color: "var(--color-text)",
  fontSize: "0.8rem",
  fontFamily: "'Inter', sans-serif",
  cursor: "pointer",
};

const AUTO_REFRESH_OPTIONS = [
  { label: "Off", value: 0 },
  { label: "10s", value: 10 },
  { label: "30s", value: 30 },
  { label: "1min", value: 60 },
  { label: "5min", value: 300 },
];

function SkeletonRows({ count }: { count: number }) {
  return (
    <>
      {Array.from({ length: count }, (_, i) => (
        <tr key={`skeleton-${i}`} data-testid="skeleton-row">
          {Array.from({ length: 7 }, (_, j) => (
            <td key={j} style={tdStyle}>
              <div className="skeleton" style={{ height: "16px" }}>&nbsp;</div>
            </td>
          ))}
        </tr>
      ))}
    </>
  );
}

const TransactionList: React.FC = () => {
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [nextCursor, setNextCursor] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [refreshing, setRefreshing] = useState(false);
  const [autoRefresh, setAutoRefresh] = useState(0);
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const loadTransactions = useCallback(async (cursor?: string) => {
    const isInitial = !cursor;
    if (isInitial) {
      setLoading(true);
    } else {
      setLoadingMore(true);
    }
    setError(null);

    try {
      const response = await fetchTransactions(20, cursor);
      setTransactions((prev) =>
        isInitial ? response.data : [...prev, ...response.data]
      );
      setNextCursor(response.next_cursor ?? null);
    } catch {
      setError("Unable to connect to Transaction service");
    } finally {
      setLoading(false);
      setLoadingMore(false);
    }
  }, []);

  useEffect(() => {
    loadTransactions();
  }, [loadTransactions]);

  // Auto-refresh interval
  useEffect(() => {
    if (intervalRef.current) {
      clearInterval(intervalRef.current);
      intervalRef.current = null;
    }
    if (autoRefresh > 0) {
      intervalRef.current = setInterval(() => {
        handleRefresh();
      }, autoRefresh * 1000);
    }
    return () => {
      if (intervalRef.current) clearInterval(intervalRef.current);
    };
  }, [autoRefresh]); // eslint-disable-line react-hooks/exhaustive-deps

  const handleLoadMore = () => {
    if (nextCursor) {
      loadTransactions(nextCursor);
    }
  };

  const handleRetry = () => {
    setTransactions([]);
    setNextCursor(null);
    loadTransactions();
  };

  const handleRefresh = async () => {
    setRefreshing(true);
    setError(null);
    try {
      const response = await fetchTransactions(20);
      setTransactions(response.data);
      setNextCursor(response.next_cursor ?? null);
    } catch {
      setError("Unable to connect to Transaction service");
    } finally {
      setRefreshing(false);
    }
  };

  return (
    <div style={{ maxWidth: "1200px", margin: "0 auto" }}>
      <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", marginBottom: "16px" }}>
        <h2 style={{ margin: 0 }}>Transactions</h2>
        <div style={{ display: "flex", alignItems: "center", gap: "10px" }}>
          <label style={{ display: "flex", alignItems: "center", gap: "6px", fontSize: "0.8rem", color: "var(--color-text-muted)" }}>
            Auto-refresh:
            <select
              style={selectStyle}
              value={autoRefresh}
              onChange={(e) => setAutoRefresh(Number(e.target.value))}
              aria-label="Auto-refresh interval"
            >
              {AUTO_REFRESH_OPTIONS.map((opt) => (
                <option key={opt.value} value={opt.value}>{opt.label}</option>
              ))}
            </select>
          </label>
          <button
            style={{ ...btnStyle, opacity: refreshing ? 0.6 : 1 }}
            onClick={handleRefresh}
            disabled={refreshing}
            aria-label="Refresh transactions"
          >
            {refreshing ? "Refreshing…" : "↻ Refresh"}
          </button>
        </div>
      </div>

      {error && <ErrorBanner message={error} onRetry={handleRetry} />}

      {!loading && transactions.length > 0 && (
        <DashboardStats transactions={transactions} />
      )}

      <div style={cardStyle}>
        <table style={tableStyle}>
          <thead>
            <tr>
              <th style={thStyle}>Transaction ID</th>
              <th style={thStyle}>Amount</th>
              <th style={thStyle}>Currency</th>
              <th style={thStyle}>Payment Method</th>
              <th style={thStyle}>Customer Name</th>
              <th style={thStyle}>Status</th>
              <th style={thStyle}>Created At</th>
            </tr>
          </thead>
          <tbody>
            {loading ? (
              <SkeletonRows count={5} />
            ) : (
              transactions.map((txn) => (
                <tr key={txn.id} style={{ transition: "background-color 100ms ease" }}>
                  <td style={tdStyle}>
                    <Link to={`/transactions/${txn.id}`} style={{ fontWeight: 500 }}>{txn.id}</Link>
                  </td>
                  <td style={{ ...tdStyle, fontVariantNumeric: "tabular-nums" }}>
                    {formatCurrency(txn.amount_in_cents, txn.currency)}
                  </td>
                  <td style={tdStyle}>{txn.currency}</td>
                  <td style={tdStyle}>{txn.payment_method}</td>
                  <td style={tdStyle}>{txn.customer_name}</td>
                  <td style={tdStyle}><StatusBadge status={txn.status} /></td>
                  <td style={{ ...tdStyle, color: "var(--color-text-muted)" }}>{formatDate(txn.created_at)}</td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {!loading && nextCursor && (
        <button
          style={{ ...btnStyle, display: "block", margin: "20px auto", padding: "8px 28px" }}
          onClick={handleLoadMore}
          disabled={loadingMore}
        >
          {loadingMore ? "Loading…" : "Load More"}
        </button>
      )}
    </div>
  );
};

export default TransactionList;
