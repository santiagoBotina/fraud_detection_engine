import React from "react";
import Card from "./Card";

interface EmptyStateProps {
  message: string;
}

const EmptyState: React.FC<EmptyStateProps> = ({ message }) => (
  <Card centered muted>
    {message}
  </Card>
);

export default React.memo(EmptyState);
