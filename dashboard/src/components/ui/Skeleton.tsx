import React from "react";

interface SkeletonProps {
  width?: string;
  height?: string;
}

const Skeleton: React.FC<SkeletonProps> = ({ width = "100%", height = "16px" }) => (
  <div className="skeleton" style={{ width, height }}>&nbsp;</div>
);

export default React.memo(Skeleton);
