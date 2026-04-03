import React from "react";

interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: "default" | "error";
  loading?: boolean;
}

const baseStyle: React.CSSProperties = {
  padding: "6px 16px",
  borderRadius: "8px",
  cursor: "pointer",
  fontWeight: 600,
  fontSize: "0.8rem",
  fontFamily: "'Inter', sans-serif",
  transition: "all 150ms ease",
  whiteSpace: "nowrap",
};

const variants: Record<string, React.CSSProperties> = {
  default: {
    border: "1px solid var(--color-border)",
    backgroundColor: "var(--color-surface)",
    color: "var(--color-text)",
    boxShadow: "var(--shadow-sm)",
  },
  error: {
    border: "none",
    backgroundColor: "var(--color-error-btn, #e06060)",
    color: "#fff",
  },
};

const Button: React.FC<ButtonProps> = ({
  variant = "default",
  loading,
  disabled,
  style,
  children,
  ...rest
}) => {
  const mergedStyle: React.CSSProperties = {
    ...baseStyle,
    ...variants[variant],
    ...(loading || disabled ? { opacity: 0.6, cursor: "not-allowed" } : {}),
    ...style,
  };

  return (
    <button style={mergedStyle} disabled={loading || disabled} {...rest}>
      {children}
    </button>
  );
};

export default React.memo(Button);
