interface EmptyStateProps {
  /** @deprecated Decorative icons removed — ignored. */
  icon?: string;
  title: string;
  message?: string;
  action?: { label: string; onClick: () => void };
  variant?: "default" | "success" | "warning" | "empty" | "info";
}

export function EmptyState({
  title,
  message,
  action,
  variant = "empty",
}: EmptyStateProps) {
  return (
    <div className={`empty-state-card empty-state-card--${variant}`} role="status">
      <h3 className="empty-state-card__title">{title}</h3>
      {message && <p className="empty-state-card__message">{message}</p>}
      {action && (
        <button type="button" className="btn btn--primary empty-state-card__action" onClick={action.onClick}>
          {action.label}
        </button>
      )}
    </div>
  );
}
