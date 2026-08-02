interface LoadingStateProps {
  message?: string;
  variant?: "skeleton" | "spinner" | "dots";
}

export function LoadingState({ message = "Loading…", variant = "spinner" }: LoadingStateProps) {
  if (variant === "skeleton") {
    return (
      <div className="loading-skeleton" role="status" aria-label={message}>
        <div className="loading-skeleton__row" />
        <div className="loading-skeleton__row" />
        <div className="loading-skeleton__row" />
        <span className="loading-skeleton__label">{message}</span>
      </div>
    );
  }

  if (variant === "dots") {
    return (
      <div className="loading-dots" role="status" aria-label={message}>
        <span className="loading-dots__dot" />
        <span className="loading-dots__dot" />
        <span className="loading-dots__dot" />
        <span className="loading-dots__label">{message}</span>
      </div>
    );
  }

  return (
    <div className="loading-spinner" role="status" aria-label={message}>
      <div className="loading-spinner__ring" aria-hidden />
      <span className="loading-spinner__label">{message}</span>
    </div>
  );
}
