import { useEffect, useState } from "react";
import { api } from "../api/client";

type ConsentStatus = {
  couple_id?: string;
  granted: boolean;
  revoked: boolean;
  allowed_actions?: string[];
  granted_at?: string | null;
};

type GrantResponse = {
  couple_id?: string;
  person_a?: string;
  person_b?: string;
  granted: boolean;
  actions?: string[];
};

type RevokeResponse = {
  couple_id?: string;
  revoked: boolean;
  message?: string;
};

// handoffCodeFromURL extracts the handoff code from either the hash query
// (hash-routed dashboard) or the location search (direct deep link).
function handoffCodeFromURL(): string {
  const hash = window.location.hash || "";
  const qIdx = hash.indexOf("?");
  if (qIdx >= 0) {
    const ref = new URLSearchParams(hash.slice(qIdx + 1)).get("ref");
    if (ref) return ref;
  }
  const ref = new URLSearchParams(window.location.search).get("ref");
  return ref ?? "";
}

const CONSENT_ITEMS = [
  { id: "contact", label: "I consent to Neptune Radar contacting me about prenup services." },
  { id: "data", label: "I consent to Neptune Radar processing my data for this purpose." },
];

export function CelebrateView() {
  const handoffCode = handoffCodeFromURL();
  const [status, setStatus] = useState<ConsentStatus | null>(null);
  const [loading, setLoading] = useState(true);
  const [checked, setChecked] = useState<Record<string, boolean>>({ contact: false, data: false });
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [revoked, setRevoked] = useState(false);

  useEffect(() => {
    if (!handoffCode) {
      setLoading(false);
      return;
    }
    api
      .get<ConsentStatus>(`/api/consent/status/${encodeURIComponent(handoffCode)}`)
      .then((s) => setStatus(s))
      .catch(() => setError("We couldn't verify your consent status. Please try again."))
      .finally(() => setLoading(false));
  }, [handoffCode]);

  const allChecked = checked.contact && checked.data;

  const grant = () => {
    setBusy(true);
    setError("");
    api
      .post<GrantResponse>("/api/consent/grant", {
        handoff_code: handoffCode,
        consent_actions: ["postcard", "follow_up", "data_processing"],
      })
      .then((r) => setStatus({ granted: r.granted, revoked: false, allowed_actions: r.actions }))
      .catch(() => setError("Something went wrong recording your consent. Please try again."))
      .finally(() => setBusy(false));
  };

  const revoke = () => {
    setBusy(true);
    setError("");
    api
      .post<RevokeResponse>("/api/consent/revoke", { handoff_code: handoffCode })
      .then(() => {
        setRevoked(true);
        setStatus({ granted: false, revoked: true });
      })
      .catch(() => setError("Something went wrong processing your opt-out. Please try again."))
      .finally(() => setBusy(false));
  };

  if (!handoffCode) {
    return (
      <div className="celebrate-page">
        <div className="celebrate-hero">Welcome to Neptune Radar</div>
        <p className="celebrate-error">This link is missing a reference code. Please scan the QR code on your postcard.</p>
      </div>
    );
  }

  if (loading) {
    return (
      <div className="celebrate-page">
        <div className="celebrate-hero">Welcome to Neptune Radar</div>
        <p>Loading…</p>
      </div>
    );
  }

  // Opt-out confirmation state.
  if (revoked || status?.revoked) {
    return (
      <div className="celebrate-page">
        <div className="celebrate-hero">You're opted out</div>
        <p className="celebrate-success">
          Your consent has been revoked. We will not contact you again, and we have stopped
          processing your data for outreach purposes.
        </p>
      </div>
    );
  }

  // Consent already granted — show the welcome pitch.
  if (status?.granted) {
    return (
      <div className="celebrate-page">
        <div className="celebrate-hero">Welcome to Neptune Radar</div>
        <p>Here's what Neptune Radar can do for you:</p>
        <ul className="celebrate-list">
          <li>A prenup protects both partners — it's a conversation, not a conflict.</li>
          <li>We connect you with attorneys who specialize in modern, fair agreements.</li>
          <li>Start with a free, no-pressure conversation about your options.</li>
        </ul>
        <a
          className="celebrate-button celebrate-button--primary"
          href="https://app.meetneptune.com/chat"
        >
          Start a conversation
        </a>
        <button
          type="button"
          className="celebrate-button celebrate-button--secondary"
          onClick={() => setStatus({ granted: false, revoked: false })}
        >
          Manage your consent
        </button>
      </div>
    );
  }

  // Consent required — the capture form.
  return (
    <div className="celebrate-page">
      <div className="celebrate-hero">Congratulations on your engagement</div>
      <p>
        Neptune Radar found you because we believe you'd benefit from a prenup. We sent you a
        postcard.
      </p>
      <p>Before we continue, we'd like your consent to process your data:</p>
      <div className="celebrate-consent">
        {CONSENT_ITEMS.map((item) => (
          <label key={item.id} className="celebrate-checkbox">
            <input
              type="checkbox"
              checked={checked[item.id]}
              onChange={(e) => setChecked((c) => ({ ...c, [item.id]: e.target.checked }))}
            />
            <span className="celebrate-checkbox__box" aria-hidden />
            <span className="celebrate-checkbox__label">{item.label}</span>
          </label>
        ))}
      </div>
      <div className="celebrate-actions">
        <button
          type="button"
          className="celebrate-button celebrate-button--primary"
          disabled={!allChecked || busy}
          onClick={grant}
        >
          {busy ? "Processing…" : "I Consent"}
        </button>
        <button
          type="button"
          className="celebrate-button celebrate-button--secondary"
          disabled={busy}
          onClick={revoke}
        >
          Opt Out
        </button>
      </div>
      {error && <p className="celebrate-error">{error}</p>}
    </div>
  );
}

export default CelebrateView;
