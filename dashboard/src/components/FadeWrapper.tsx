import React, { useState, useEffect } from "react";
import { useLocation } from "react-router-dom";

interface FadeWrapperProps {
  children: React.ReactNode;
}

const FadeWrapper: React.FC<FadeWrapperProps> = ({ children }) => {
  const location = useLocation();
  const [opacity, setOpacity] = useState(0);

  useEffect(() => {
    setOpacity(0);
    const frame = requestAnimationFrame(() => setOpacity(1));
    return () => cancelAnimationFrame(frame);
  }, [location.pathname]);

  return (
    <div style={{ opacity, transition: "opacity 300ms ease-in-out" }}>
      {children}
    </div>
  );
};

export default FadeWrapper;
