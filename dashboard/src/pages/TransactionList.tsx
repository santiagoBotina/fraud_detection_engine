import React, { useCallback, useEffect, useRef } from "react";
import { useTransactions } from "../hooks/useTransactions";
import { useStats } from "../hooks/useStats";
import DashboardStats from "../components/DashboardStats";
import TransactionToolbar from "../components/TransactionToolbar";
import TransactionTable from "../components/TransactionTable";
import PaginationControls from "../components/PaginationControls";
import PageSizeSelector from "../components/PageSizeSelector";
import ErrorBanner from "../components/ErrorBanner";
import PageWrapper from "../components/ui/PageWrapper";

const inputStyle: React.CSSProperties = {
  padding: "8px 12px",
  border: "1px solid var(--color-border)",
  borderRadius: "8px",
  backgroundColor: "var(--color-surface)",
  color: "var(--color-text)",
  fontSize: "0.875rem",
  fontFamily: "'Inter', sans-serif",
  width: "100%",
  maxWidth: "400px",
  boxSizing: "border-box",
};

const TransactionList: React.FC = () => {
  const {
    transactions,
    loading,
    refreshing,
    error,
    page,
    pageSize,
    setPageSize,
    hasNextPage,
    hasPreviousPage,
    goNextPage,
    goPreviousPage,
    autoRefresh,
    setAutoRefresh,
    retry,
    refresh,
    searchId,
    setSearchId,
    searchError,
    searchLoading,
  } = useTransactions();

  const statsHook = useStats();
  const statsRetryRef = useRef(statsHook.retry);
  statsRetryRef.current = statsHook.retry;

  // Wire auto-refresh to also refresh stats
  const handleRefresh = useCallback(() => {
    refresh();
    statsRetryRef.current();
  }, [refresh]);

  const handleAutoRefreshChange = useCallback(
    (value: number) => setAutoRefresh(value),
    [setAutoRefresh]
  );

  // When auto-refresh triggers a refresh, also refresh stats
  const prevRefreshingRef = useRef(refreshing);
  useEffect(() => {
    if (refreshing && !prevRefreshingRef.current) {
      statsRetryRef.current();
    }
    prevRefreshingRef.current = refreshing;
  }, [refreshing]);

  return (
    <PageWrapper maxWidth="1200px">
      <DashboardStats
        stats={statsHook.stats}
        loading={statsHook.loading}
        error={statsHook.error}
        onRetry={statsHook.retry}
      />

      <TransactionToolbar
        autoRefresh={autoRefresh}
        onAutoRefreshChange={handleAutoRefreshChange}
        refreshing={refreshing}
        onRefresh={handleRefresh}
      />

      <div style={{ display: "flex", alignItems: "center", gap: "16px", marginBottom: "16px" }}>
        <input
          type="text"
          placeholder="Search by Transaction ID…"
          value={searchId}
          onChange={(e) => setSearchId(e.target.value)}
          style={inputStyle}
          aria-label="Search by transaction ID"
        />
        <PageSizeSelector pageSize={pageSize} onPageSizeChange={setPageSize} />
      </div>

      {error ? <ErrorBanner message={error} onRetry={retry} /> : null}

      {searchError ? (
        <div style={{ color: "var(--color-error, red)", marginBottom: "16px", fontSize: "0.875rem" }}>
          {searchError}
        </div>
      ) : null}

      {searchLoading ? (
        <div style={{ color: "var(--color-text-muted)", marginBottom: "16px", fontSize: "0.875rem" }}>
          Searching…
        </div>
      ) : null}

      <TransactionTable transactions={transactions} loading={loading} />

      <PaginationControls
        page={page}
        hasNextPage={hasNextPage}
        hasPreviousPage={hasPreviousPage}
        goNextPage={goNextPage}
        goPreviousPage={goPreviousPage}
      />
    </PageWrapper>
  );
};

export default TransactionList;
