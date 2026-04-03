import React from "react";

interface PageWrapperProps {
  children: React.ReactNode;
  maxWidth?: string;
}

const PageWrapper: React.FC<PageWrapperProps> = ({ children, maxWidth = "920px" }) => (
  <div style={{ maxWidth, margin: "0 auto", fontFamily: "'Inter', sans-serif" }}>
    {children}
  </div>
);

export default React.memo(PageWrapper);
