import { useEffect, useMemo, useState } from "react";
import {
  useApplyCandidate,
  useBuildCongratulateKit,
  useCoupleKit,
  useKitMarkMailed,
  useKitReadyToMail,
  useKits,
  usePatchKit,
  useRunDetective,
  useSendPostcard,
  useVerifyKitAddress,
} from "../api/hooks";
import { mediaURL } from "../api/media";
import type { CongratulateKit, KitStatus } from "../api/types";
import { useToast } from "../components/Toast";
import { getToken } from "../api/client";

const STATUS_LABEL: Record<KitStatus, string> = {
  draft: "Draft",
  ready_review: "Ready for review",
  address_verified: "Address verified",
  ready_to_mail: "Ready to mail",
  mailed: "Mailed",
  cancelled: "Cancelled",
};

function confPct(n: number): string {
  return `${Math.round((n || 0) * 100)}%`;
}

function KitCard({
  kit,
  selected,
  onSelect,
}: {
  kit: CongratulateKit;
  selected: boolean;
  onSelect: () => void;
}) {
  return (
    <button
      type="button"
      className={`kit-card ${selected ? "kit-card--selected" : ""} kit-card--${kit.status}`}
      onClick={onSelect}
    >
      <div className="kit-card__faces">
        {kit.profile_pic_a ? (
          <img src={mediaURL(kit.profile_pic_a)} alt="" referrerPolicy="no-referrer" />
        ) : (
          <span className="kit-card__fallback">{(kit.person_a_name || "?")[0]}</span>
        )}
        {kit.profile_pic_b ? (
          <img src={mediaURL(kit.profile_pic_b)} alt="" referrerPolicy="no-referrer" />
        ) : (
          <span className="kit-card__fallback">{(kit.person_b_name || "?")[0]}</span>
        )}
      </div>
      <div className="kit-card__body">
        <strong>
          {kit.person_a_name || kit.handle_a} & {kit.person_b_name || kit.handle_b}
        </strong>
        <span className="kit-card__meta">
          {kit.market_city
            ? `${kit.market_city}${kit.market_region ? `, ${kit.market_region}` : ""}`
            : "Market unknown"}
          {" · "}
          {STATUS_LABEL[kit.status]}
        </span>
      </div>
    </button>
  );
}

function KitWorkspace({ kit, onUpdated }: { kit: CongratulateKit; onUpdated: () => void }) {
  const patch = usePatchKit();
  const ready = useKitReadyToMail();
  const mailed = useKitMarkMailed();
  const detective = useRunDetective();
  const applyCand = useApplyCandidate();
  const verifyAddr = useVerifyKitAddress();
  const sendMail = useSendPostcard();
  const toast = useToast();
  const [line1, setLine1] = useState(kit.address_line1 ?? "");
  const [line2, setLine2] = useState(kit.address_line2 ?? "");
  const [city, setCity] = useState(kit.address_city ?? "");
  const [region, setRegion] = useState(kit.address_region ?? "");
  const [postal, setPostal] = useState(kit.address_postal ?? "");
  const [message, setMessage] = useState(kit.body_message ?? "");
  const [headline, setHeadline] = useState(kit.headline ?? "Congratulations");
  const [firstA, setFirstA] = useState(kit.first_name_a || kit.person_a_name || "");
  const [lastA, setLastA] = useState(kit.last_name_a || "");
  const [firstB, setFirstB] = useState(kit.first_name_b || kit.person_b_name || "");
  const [lastB, setLastB] = useState(kit.last_name_b || "");
  const [previewHTML, setPreviewHTML] = useState<string | null>(null);

  useEffect(() => {
    setLine1(kit.address_line1 ?? "");
    setLine2(kit.address_line2 ?? "");
    setCity(kit.address_city ?? "");
    setRegion(kit.address_region ?? "");
    setPostal(kit.address_postal ?? "");
    setMessage(kit.body_message ?? "");
    setHeadline(kit.headline ?? "Congratulations");
    setFirstA(kit.first_name_a || kit.person_a_name || "");
    setLastA(kit.last_name_a || "");
    setFirstB(kit.first_name_b || kit.person_b_name || "");
    setLastB(kit.last_name_b || "");
  }, [kit.id, kit.updated_at]);

  // Load postcard HTML with auth (iframe src can't send bearer).
  useEffect(() => {
    let cancelled = false;
    const base = import.meta.env.VITE_API_URL ?? "";
    fetch(`${base}/api/kits/${kit.id}/postcard`, {
      headers: { Authorization: `Bearer ${getToken()}` },
    })
      .then((r) => r.text())
      .then((html) => {
        if (!cancelled) setPreviewHTML(html);
      })
      .catch(() => {
        if (!cancelled) setPreviewHTML(kit.postcard_html || null);
      });
    return () => {
      cancelled = true;
    };
  }, [kit.id, kit.updated_at, kit.postcard_html]);

  const busy =
    patch.isPending ||
    ready.isPending ||
    mailed.isPending ||
    detective.isPending ||
    applyCand.isPending ||
    verifyAddr.isPending ||
    sendMail.isPending;

  const save = (verify: boolean) => {
    patch.mutate(
      {
        id: kit.id,
        address_line1: line1,
        address_line2: line2,
        address_city: city,
        address_region: region,
        address_postal: postal,
        headline,
        body_message: message,
        verify,
        verified_by: "operator",
      },
      {
        onSuccess: () => {
          toast.push(verify ? "Address verified" : "Kit saved", "ok");
          onUpdated();
        },
        onError: (e) => toast.push((e as Error).message, "err"),
      },
    );
  };

  const saveNamesThen = (next: () => void) => {
    patch.mutate(
      {
        id: kit.id,
        first_name_a: firstA.trim(),
        last_name_a: lastA.trim(),
        first_name_b: firstB.trim(),
        last_name_b: lastB.trim(),
        headline,
        body_message: message,
      },
      {
        onSuccess: () => next(),
        onError: (e) => toast.push((e as Error).message, "err"),
      },
    );
  };

  const runDetective = () => {
    if (!lastA.trim() && !lastB.trim()) {
      toast.push("Tip: add at least one last name for better street results", "info");
    }
    saveNamesThen(() => {
      detective.mutate(kit.id, {
        onSuccess: (k) => {
          const n = k.address_candidates?.length ?? 0;
          const streets = (k.address_candidates || []).filter((c) => c.line1).length;
          toast.push(
            streets
              ? `Detective found ${streets} street candidate${streets === 1 ? "" : "s"}`
              : n
                ? `Detective: ${n} market-level hit(s) — add last names + PDL/Melissa key for streets`
                : "Detective ran — no hits",
            streets ? "ok" : "info",
          );
          onUpdated();
        },
        onError: (e) => toast.push((e as Error).message, "err"),
      });
    });
  };

  return (
    <div className="kit-workspace">
      <header className="kit-workspace__hero">
        <div className="kit-workspace__intro">
          <p className="kit-workspace__eyebrow">Congratulate kit</p>
          <h2>
            {kit.person_a_name || kit.handle_a} <span className="kit-workspace__amp">&</span>{" "}
            {kit.person_b_name || kit.handle_b}
          </h2>
          <p className="kit-workspace__sub">
            Agent resolved <strong>first names</strong> for the card, gathered market + discovery
            evidence, and drafted a postcard. Street address is never invented — you verify, then
            mail.
          </p>
          <div className="kit-workspace__pills">
            <span className={`kit-pill kit-pill--${kit.status}`}>{STATUS_LABEL[kit.status]}</span>
            <span className="kit-pill kit-pill--names">
              Dear {kit.person_a_name || "…"} & {kit.person_b_name || "…"}
            </span>
            {kit.market_city && (
              <span className="kit-pill">
                ⌖ {kit.market_city}
                {kit.market_region ? `, ${kit.market_region}` : ""}
              </span>
            )}
            <span className="kit-pill">Address conf {confPct(kit.address_confidence)}</span>
            {kit.source_handle && <span className="kit-pill">via @{kit.source_handle}</span>}
          </div>
          {(kit.handle_a || kit.handle_b) && (
            <p className="kit-workspace__handles">
              Handles: {kit.handle_a && <>@{kit.handle_a}</>}
              {kit.handle_a && kit.handle_b ? " · " : ""}
              {kit.handle_b && <>@{kit.handle_b}</>}
            </p>
          )}
        </div>
        <div className="kit-workspace__faces">
          {kit.profile_pic_a && <img src={mediaURL(kit.profile_pic_a)} alt="" referrerPolicy="no-referrer" />}
          {kit.profile_pic_b && <img src={mediaURL(kit.profile_pic_b)} alt="" referrerPolicy="no-referrer" />}
        </div>
      </header>

      <div className="kit-workspace__grid">
        <section className="kit-panel kit-panel--postcard">
          <div className="kit-panel__head">
            <h3>Postcard preview</h3>
            <button
              type="button"
              className="btn btn--ghost btn--sm"
              onClick={() => {
                if (!previewHTML) return;
                // Blob URL, not document.write on a window with opener back
                // to this app: previewHTML is built from scraped Instagram
                // strings, and any script in it must not reach our origin
                // (the admin token lives in localStorage).
                const blob = new Blob([previewHTML], { type: "text/html" });
                const url = URL.createObjectURL(blob);
                const w = window.open(url, "_blank", "noopener");
                if (w) {
                  w.addEventListener("load", () => {
                    w.focus();
                    w.print();
                  });
                }
              }}
            >
              Print / PDF
            </button>
          </div>
          {previewHTML ? (
            // sandbox="" = all restrictions: no scripts, no same-origin.
            // srcdoc without sandbox inherits THIS origin — with scraped
            // content inside, that's an XSS path straight to localStorage.
            <iframe title="Postcard" className="kit-postcard-frame" sandbox="" srcDoc={previewHTML} />
          ) : (
            <div className="empty-state">Loading postcard…</div>
          )}
        </section>

        <div className="kit-workspace__side">
          <section className="kit-panel">
            <h3>Curated message</h3>
            <input
              className="feed-filter"
              value={headline}
              onChange={(e) => setHeadline(e.target.value)}
              placeholder="Headline"
            />
            <textarea
              className="kit-message"
              rows={6}
              value={message}
              onChange={(e) => setMessage(e.target.value)}
            />
          </section>

          <section className="kit-panel">
            <h3>Names for research</h3>
            <p className="kit-panel__hint">
              Postcard greets with <strong>first</strong> names. Address search needs{" "}
              <strong>first + last + city</strong> for real street hits. Fill last names if missing
              from Instagram, then Run detective.
            </p>
            <div className="kit-name-grid">
              <div>
                <label className="kit-label">Person A</label>
                <div className="kit-addr-form__row">
                  <input
                    className="feed-filter"
                    placeholder="First"
                    value={firstA}
                    onChange={(e) => setFirstA(e.target.value)}
                  />
                  <input
                    className="feed-filter"
                    placeholder="Last"
                    value={lastA}
                    onChange={(e) => setLastA(e.target.value)}
                    style={{ gridColumn: "span 2" }}
                  />
                </div>
              </div>
              <div>
                <label className="kit-label">Person B</label>
                <div className="kit-addr-form__row">
                  <input
                    className="feed-filter"
                    placeholder="First"
                    value={firstB}
                    onChange={(e) => setFirstB(e.target.value)}
                  />
                  <input
                    className="feed-filter"
                    placeholder="Last"
                    value={lastB}
                    onChange={(e) => setLastB(e.target.value)}
                    style={{ gridColumn: "span 2" }}
                  />
                </div>
              </div>
            </div>
            <button
              type="button"
              className="btn btn--ghost btn--sm"
              style={{ marginTop: 8 }}
              disabled={busy}
              onClick={() =>
                saveNamesThen(() => {
                  toast.push("Names saved", "ok");
                  onUpdated();
                })
              }
            >
              Save names
            </button>
          </section>

          <section className="kit-panel">
            <h3>Detective · mailing address</h3>
            <p className="kit-panel__hint">
              <strong>Run detective</strong> queries people-data APIs (PDL / Melissa when keys are set)
              using first + last + market. Pick a candidate, verify (Lob USPS), then send postcard.
            </p>
            <div className="kit-actions" style={{ marginBottom: 12 }}>
              <button type="button" className="btn btn--primary" disabled={busy} onClick={runDetective}>
                {detective.isPending ? "Running detective…" : "Run detective"}
              </button>
            </div>
            {(kit.address_candidates?.length ?? 0) > 0 && (
              <ul className="kit-candidates kit-candidates--pick">
                {kit.address_candidates!.map((c, i) => (
                  <li key={i}>
                    <button
                      type="button"
                      className="kit-cand-btn"
                      disabled={busy || !c.line1}
                      onClick={() =>
                        applyCand.mutate(
                          { id: kit.id, index: i },
                          {
                            onSuccess: (k) => {
                              setLine1(k.address_line1 ?? "");
                              setLine2(k.address_line2 ?? "");
                              setCity(k.address_city ?? "");
                              setRegion(k.address_region ?? "");
                              setPostal(k.address_postal ?? "");
                              toast.push(c.line1 ? "Candidate applied" : "City-only candidate", "ok");
                              onUpdated();
                            },
                            onError: (e) => toast.push((e as Error).message, "err"),
                          },
                        )
                      }
                    >
                      <strong>
                        {c.line1 || "(no street)"}
                        {c.line2 ? `, ${c.line2}` : ""}
                      </strong>
                      <span>
                        {[c.city, c.region, c.postal].filter(Boolean).join(", ")} ·{" "}
                        {Math.round((c.confidence || 0) * 100)}% · {c.source}
                      </span>
                      {c.note && <em>{c.note}</em>}
                    </button>
                  </li>
                ))}
              </ul>
            )}
            <div className="kit-addr-form">
              <input
                className="feed-filter"
                placeholder="Street address"
                value={line1}
                onChange={(e) => setLine1(e.target.value)}
              />
              <input
                className="feed-filter"
                placeholder="Apt / suite (optional)"
                value={line2}
                onChange={(e) => setLine2(e.target.value)}
              />
              <div className="kit-addr-form__row">
                <input
                  className="feed-filter"
                  placeholder="City"
                  value={city}
                  onChange={(e) => setCity(e.target.value)}
                />
                <input
                  className="feed-filter kit-addr-form__st"
                  placeholder="ST"
                  maxLength={2}
                  value={region}
                  onChange={(e) => setRegion(e.target.value.toUpperCase())}
                />
                <input
                  className="feed-filter"
                  placeholder="ZIP"
                  value={postal}
                  onChange={(e) => setPostal(e.target.value)}
                />
              </div>
            </div>
            <div className="kit-actions">
              <button type="button" className="btn btn--ghost" disabled={busy} onClick={() => save(false)}>
                Save draft
              </button>
              <button
                type="button"
                className="btn btn--primary"
                disabled={busy}
                onClick={() => {
                  // Save fields then USPS verify
                  patch.mutate(
                    {
                      id: kit.id,
                      address_line1: line1,
                      address_line2: line2,
                      address_city: city,
                      address_region: region,
                      address_postal: postal,
                      headline,
                      body_message: message,
                    },
                    {
                      onSuccess: () =>
                        verifyAddr.mutate(kit.id, {
                          onSuccess: () => {
                            toast.push("Address verified", "ok");
                            onUpdated();
                          },
                          onError: (e) => {
                            // fallback local verify if Lob missing
                            save(true);
                            toast.push((e as Error).message, "err");
                          },
                        }),
                      onError: (e) => toast.push((e as Error).message, "err"),
                    },
                  );
                }}
              >
                Verify address
              </button>
              {(kit.status === "address_verified" || kit.status === "ready_to_mail") && (
                <>
                  <button
                    type="button"
                    className="btn btn--ghost"
                    disabled={busy}
                    onClick={() =>
                      ready.mutate(kit.id, {
                        onSuccess: () => {
                          toast.push("Ready to mail", "ok");
                          onUpdated();
                        },
                        onError: (e) => toast.push((e as Error).message, "err"),
                      })
                    }
                  >
                    Ready to mail
                  </button>
                  <button
                    type="button"
                    className="btn btn--primary"
                    disabled={busy}
                    onClick={() => {
                      if (!confirm("Send physical postcard via Lob? Uses postage credits.")) return;
                      sendMail.mutate(kit.id, {
                        onSuccess: () => {
                          toast.push("Postcard sent (or queued)", "ok");
                          onUpdated();
                        },
                        onError: (e) => toast.push((e as Error).message, "err"),
                      });
                    }}
                  >
                    {sendMail.isPending ? "Sending…" : "Send postcard"}
                  </button>
                </>
              )}
              {kit.status === "ready_to_mail" && (
                <button
                  type="button"
                  className="btn btn--ghost"
                  disabled={busy}
                  onClick={() =>
                    mailed.mutate(kit.id, {
                      onSuccess: () => {
                        toast.push("Marked mailed", "ok");
                        onUpdated();
                      },
                      onError: (e) => toast.push((e as Error).message, "err"),
                    })
                  }
                >
                  Mark mailed (manual)
                </button>
              )}
            </div>
          </section>
        </div>
      </div>

      <div className="kit-workspace__lower">
        <section className="kit-panel">
          <h3>Research steps</h3>
          <ol className="kit-steps">
            {(kit.research_steps || []).map((s) => (
              <li key={s.id} className={`kit-step kit-step--${s.status}`}>
                <div className="kit-step__label">
                  <span className="kit-step__status">{s.status}</span>
                  {s.label}
                </div>
                <p>{s.detail}</p>
                {s.url && (
                  <a href={s.url} target="_blank" rel="noreferrer">
                    Open search
                  </a>
                )}
              </li>
            ))}
          </ol>
        </section>

        <section className="kit-panel">
          <h3>Evidence dossier</h3>
          <ul className="kit-evidence">
            {(kit.evidence || []).map((e) => (
              <li key={e}>{e}</li>
            ))}
          </ul>
          {(kit.bio_a || kit.bio_b) && (
            <div className="kit-bios">
              {kit.bio_a && (
                <p>
                  <strong>@{kit.handle_a}:</strong> {kit.bio_a}
                </p>
              )}
              {kit.bio_b && (
                <p>
                  <strong>@{kit.handle_b}:</strong> {kit.bio_b}
                </p>
              )}
            </div>
          )}
          {kit.discovery_caption && (
            <blockquote className="kit-caption">“{kit.discovery_caption}”</blockquote>
          )}
          {kit.discovery_post_url && (
            <a href={kit.discovery_post_url} target="_blank" rel="noreferrer">
              Open discovery post
            </a>
          )}
        </section>

        <section className="kit-panel">
          <h3>Research notes</h3>
          <pre className="kit-notes">{kit.research_notes || "—"}</pre>
          {(kit.address_candidates || []).length > 0 && (
            <>
              <h4 className="kit-panel__sub">Address candidates</h4>
              <ul className="kit-candidates">
                {kit.address_candidates!.map((c, i) => (
                  <li key={i}>
                    <strong>
                      {c.city}
                      {c.region ? `, ${c.region}` : ""}
                    </strong>{" "}
                    · {confPct(c.confidence)} · {c.source}
                    {c.note && <div className="kit-panel__hint">{c.note}</div>}
                  </li>
                ))}
              </ul>
            </>
          )}
        </section>
      </div>
    </div>
  );
}

export function CongratulateView({
  initialCoupleId,
}: {
  initialCoupleId?: string;
} = {}) {
  const { data: kits, refetch } = useKits();
  const build = useBuildCongratulateKit();
  const toast = useToast();
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [focusCouple, setFocusCouple] = useState(initialCoupleId ?? "");

  const { data: coupleKit } = useCoupleKit(initialCoupleId);

  useEffect(() => {
    if (initialCoupleId) setFocusCouple(initialCoupleId);
  }, [initialCoupleId]);

  useEffect(() => {
    if (coupleKit?.id) setSelectedId(coupleKit.id);
  }, [coupleKit?.id]);

  const selected = useMemo(
    () => (kits ?? []).find((k) => k.id === selectedId) ?? coupleKit ?? null,
    [kits, selectedId, coupleKit],
  );

  return (
    <div className="view view--kits">
      <header className="sources-header">
        <div>
          <h2 className="view__title">Congratulate</h2>
          <p className="view__subtitle">
            After a couple is found, the agent builds a kit: profiles, market location, research
            checklist for public records, and a print-ready postcard. Nothing mails without your
            verify step.
          </p>
        </div>
      </header>

      <div className="kit-layout">
        <aside className="kit-list">
          <div className="kit-list__build">
            <input
              className="feed-filter"
              placeholder="Couple ID (from Work card)"
              value={focusCouple}
              onChange={(e) => setFocusCouple(e.target.value)}
            />
            <button
              type="button"
              className="btn btn--primary"
              disabled={build.isPending || !focusCouple.trim()}
              onClick={() =>
                build.mutate(focusCouple.trim(), {
                  onSuccess: (kit) => {
                    toast.push("Kit built", "ok");
                    setSelectedId(kit.id);
                    refetch();
                  },
                  onError: (e) => toast.push((e as Error).message, "err"),
                })
              }
            >
              {build.isPending ? "Building…" : "Build kit"}
            </button>
          </div>
          <div className="kit-list__items">
            {(kits ?? []).length === 0 && (
              <div className="empty-state">No kits yet. Open Work → Congratulate on a couple.</div>
            )}
            {(kits ?? []).map((k) => (
              <KitCard
                key={k.id}
                kit={k}
                selected={selected?.id === k.id}
                onSelect={() => setSelectedId(k.id)}
              />
            ))}
          </div>
        </aside>
        <main className="kit-main">
          {selected ? (
            <KitWorkspace kit={selected} onUpdated={() => refetch()} />
          ) : (
            <div className="empty-state kit-main__empty">
              Select a kit or build one for a couple to see the postcard and research workspace.
            </div>
          )}
        </main>
      </div>
    </div>
  );
}
