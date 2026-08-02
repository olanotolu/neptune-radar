const GROUPS: { title: string; rows: [string, string][] }[] = [
  {
    title: "Global",
    rows: [
      ["/", "Focus search"],
      ["g t", "Go to Today"],
      ["g w", "Go to Work"],
      ["g m", "Go to Map"],
      ["g s", "Go to Sources"],
      ["g f", "Go to Funnel"],
      ["?", "Show this help"],
      ["Esc", "Close this help"],
    ],
  },
  {
    title: "Work queue",
    rows: [
      ["j / k", "Move down / up the queue"],
      ["a", "Approve selected action"],
      ["i", "Ignore selected action"],
      ["Enter", "Open selected couple detail"],
    ],
  },
];

export function KeyboardHelp({ onClose }: { onClose: () => void }) {
  return (
    <div className="kbd-overlay" role="dialog" aria-label="Keyboard shortcuts" onClick={onClose}>
      <div className="kbd-modal" onClick={(e) => e.stopPropagation()}>
        <header className="kbd-modal__header">
          <h3>Keyboard shortcuts</h3>
          <p className="kbd-modal__subtitle">Move faster. Celebrate sooner.</p>
          <button type="button" className="btn btn--ghost btn--sm" onClick={onClose} aria-label="Close">
            Esc
          </button>
        </header>
        <div className="kbd-modal__groups">
          {GROUPS.map((g) => (
            <section key={g.title} className="kbd-modal__group">
              <h4 className="kbd-modal__group-title">{g.title}</h4>
              <dl className="kbd-modal__rows">
                {g.rows.map(([keys, desc]) => (
                  <div className="kbd-modal__row" key={keys}>
                    <dt className="kbd-modal__keys">
                      {keys.split(" ").map((k, i) => (
                        <kbd key={i} className="kbd">{k}</kbd>
                      ))}
                    </dt>
                    <dd className="kbd-modal__desc">{desc}</dd>
                  </div>
                ))}
              </dl>
            </section>
          ))}
        </div>
      </div>
    </div>
  );
}
