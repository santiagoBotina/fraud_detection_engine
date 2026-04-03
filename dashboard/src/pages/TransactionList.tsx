import React, { useCallback } from "react";
import { useTransactions } from "../hooks/useTransactions";
import TransactionToolbar from "../components/TransactionToolbar";
import TransactionTable from "../components/TransactionTable";
import DashboardStats from "../components/DashboardStats";
import ErrorBanner from "../components/ErrorBanner";
import Button from "../components/ui/Button";
import PageWrapper from "../components/ui/PageWrapper";

const TransactionList: React.FC = () => {
  const {
    transactions,
    nextCursor,
    loading,
    loadingMore,
    refreshing,
    error,
    autoRefresh,
    setAutoRefresh,
    loadMore,
    retry,
    refresh,
  } = useTransactions();

  const handleAutoRefreshChange = useCallback(
    (value: number) => setAutoRefresh(value),
    [setAutoRefresh]
  );

  return (
    <PageWrapper maxWidth="1200px">
      <TransactionToolbar
        autoRefresh={autoRefresh}
        onAutoRefreshChange={handleAutoRefreshChange}
        refreshing={refreshing}
        onRefresh={refresh}
      />

      {error ? <ErrorBanner message={error} onRetry={retry} /> : null}

      {!loading && transactions.length > 0 ? (
        <DashboardStats transactions={transactions} />
      ) : null}

      <TransactionTable transactions={transactions} loading={loading} />

      {!loading && nextCursor ? (
        <Button
          loading={loadingMore}
          onClick={loadMore}
          style={{ display: "block", margin: "20px auto", padding: "8px 28px" }}
        >
          {loadingMore ? "Loading…" : "Load More"}
        </Button>
      ) : null}
    </PageWrapper>
  );
};

export default TransactionList;
