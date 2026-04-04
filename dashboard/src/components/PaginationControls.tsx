import React from "react";

interface PaginationControlsProps {
  page: number;
  hasNextPage: boolean;
  hasPreviousPage: boolean;
  goNextPage: () => void;
  goPreviousPage: () => void;
}

const buttonStyle: React.CSSProperties = {
  padding: "6px 16px",
  borderRadius: "8px",
  fontWeight: 600,
  fontSize: "0.8rem",
  fontFamily: "'Inter', sans-serif",
  cursor: "pointer",
  border: "1px solid var(--color-border)",
  backgroundColor: "var(--color-surface)",
  color: "var(--color-text)",
  transition: "all 150ms ease",
};

const disabledStyle: React.CSSProperties = {
  opacity: 0.5,
  cursor: "not-allowed",
};

const PaginationControls: React.FC<PaginationControlsProps> = ({
  page,
  hasNextPage,
  hasPreviousPage,
  goNextPage,
  goPreviousPage,
}) => (
  <div
    style={{
      display: "flex",
      alignItems: "center",
      justifyContent: "center",
      gap: "12px",
      marginTop: "16px",
    }}
  >
    <button
      style={{ ...buttonStyle, ...(hasPreviousPage ? {} : disabledStyle) }}
      disabled={!hasPreviousPage}
      onClick={goPreviousPage}
      aria-label="Previous page"
    >
      Previous
    </button>
    <span
      style={{
        fontSize: "0.85rem",
        fontFamily: "'Inter', sans-serif",
        color: "var(--color-text)",
      }}
    >
      Page {page}
    </span>
    <button
      style={{ ...buttonStyle, ...(hasNextPage ? {} : disabledStyle) }}
      disabled={!hasNextPage}
      onClick={goNextPage}
      aria-label="Next page"
    >
      Next
    </button>
  </div>
);

export default React.memo(PaginationControls);
