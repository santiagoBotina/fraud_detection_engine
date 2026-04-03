import React from "react";

interface ErrorBannerProps {
  message: string;
  onRetry: () => void;
}

const bannerStyle: React.CSSProperties = {
  display: "flex",
  alignItems: "center",
  justifyContent: "space-between",
  gap: "12px",
  padding: "12px 16px",
  backgroundColor: "var(--color-error-bg, #fef0f0)",
  border: "1px solid var(--color-error-border, #f5c6c6)",
  borderRadius: "8px",
  color: "var(--color-error-text, #8b2525)",
  marginBottom: "16px",
  fontFamily: "'Inter', sans-serif",
  fontSize: "0.875rem",
};

const buttonStyle: React.CSSProperties = {
  padding: "6px 16px",
  backgroundColor: "var(--color-error-btn, #e06060)",
  color: "#fff",
  border: "none",
  borderRadius: "6px",
  cursor: "pointer",
  fontWeight: 600,
  fontSize: "0.8rem",
  fontFamily: "'Inter', sans-serif",
  whiteSpace: "nowrap",
  transition: "background-color 150ms ease",
};

const ErrorBanner: React.FC<ErrorBannerProps> = ({ message, onRetry }) => {
  return (
    <div role="alert" style={bannerStyle}>
      <span>{message}</span>
      <button style={buttonStyle} onClick={onRetry}>
        Retry
      </button>
    </div>
  );
};

export default ErrorBanner;
