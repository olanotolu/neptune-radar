export function ConfidenceBar({ value, label }: { value: number; label?: string }) {
  const pct = Math.round(Math.max(0, Math.min(1, value)) * 100);
  const tone = pct >= 75 ? "high" : pct >= 60 ? "mid" : "low";
  return (
    <div className="confidence-bar">
      {label && <div className="confidence-bar__label">{label}</div>}
      <div className="confidence-bar__track">
        <div className={`confidence-bar__fill confidence-bar__fill--${tone}`} style={{ width: `${pct}%` }} />
      </div>
      <div className="confidence-bar__value">{pct}%</div>
    </div>
  );
}
