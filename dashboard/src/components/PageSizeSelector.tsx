import React from "react";

interface PageSizeSelectorProps {
  pageSize: number;
  onPageSizeChange: (size: number) => void;
}

const PAGE_SIZE_OPTIONS = [20, 30, 50, 100];

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

const PageSizeSelector: React.FC<PageSizeSelectorProps> = ({
  pageSize,
  onPageSizeChange,
}) => (
  <label
    style={{
      display: "flex",
      alignItems: "center",
      gap: "6px",
      fontSize: "0.8rem",
      color: "var(--color-text-muted)",
    }}
  >
    Rows per page:
    <select
      style={selectStyle}
      value={pageSize}
      onChange={(e) => onPageSizeChange(Number(e.target.value))}
      aria-label="Page size"
    >
      {PAGE_SIZE_OPTIONS.map((size) => (
        <option key={size} value={size}>
          {size}
        </option>
      ))}
    </select>
  </label>
);

export default React.memo(PageSizeSelector);
