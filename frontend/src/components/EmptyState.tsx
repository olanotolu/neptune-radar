interface EmptyStateProps {
  icon?: string;
  title: string;
  message?: string;
  action?: { label: string; onClick: () => void };
  variant?: "default" | "success" | "warning" | "empty";
}

export function EmptyState({
  icon,
  title,
  message,
  action,
  variant = "empty",
}: EmptyStateProps) {
  return (
    <div className={`empty-state-card empty-state-card--${variant}`} role="status">
      <div className="empty-state-card__bg" aria-hidden />
      {icon && <div className="empty-state-card__icon" aria-hidden>{icon}</div>}
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
