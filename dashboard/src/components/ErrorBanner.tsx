import React from "react";
import Button from "./ui/Button";

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

const ErrorBanner: React.FC<ErrorBannerProps> = ({ message, onRetry }) => (
  <div role="alert" style={bannerStyle}>
    <span>{message}</span>
    <Button variant="error" onClick={onRetry}>
      Retry
    </Button>
  </div>
);

export default React.memo(ErrorBanner);
