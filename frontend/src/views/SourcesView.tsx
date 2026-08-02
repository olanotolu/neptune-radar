import { useEffect, useMemo, useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import {
  useAddSource,
  useApproveAction,
  useBulkScan,
  useEnrichSource,
  useIgnoreAction,
  usePatchSourceLocation,
  useReactivateSource,
  useRemoveSource,
  useScanJob,
  useSourcePosts,
  useSources,
  useStartScanJob,
  useSuppressVendorPairs,
} from "../api/hooks";
import { mediaURL } from "../api/media";
import { useToast } from "../components/Toast";
import type { ScannedCouple, ScanJob, SourcePost, SourceScanResult, WatchedSource } from "../api/types";

function formatCount(n?: number): string | null {
  if (n == null) return null;
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`;
  if (n >= 1_000) return `${(n / 1_000).toFixed(1)}K`;
  return String(n);
}

function formatAge(iso?: string): string | null {
  if (!iso) return null;
  const ms = Date.now() - new Date(iso).getTime();
  if (ms < 0) return "just now";
  const m = Math.floor(ms / 60000);
  if (m < 60) return `${m}m ago`;
  const h = Math.floor(m / 60);
  if (h < 48) return `${h}h ago`;
  return `${Math.floor(h / 24)}d ago`;
}

const SOURCE_CLASSES = [
  "engagement_photographer",
  "proposal_planner",
  "wedding_planner",
  "wedding_venue",
  "jeweler",
  "wedding_publication",
  "registry_provider",
  "bridal_boutique",
] as const;

const CATEGORY_LABEL: Record<string, string> = {
  engagement_photographer: "Photographers",
  proposal_planner: "Proposal planners",
  wedding_planner: "Wedding planners",
  wedding_venue: "Venues",
  jeweler: "Jewelers",
  wedding_publication: "Publications",
  registry_provider: "Registries",
  bridal_boutique: "Boutiques",
};

function initials(name: string): string {
  return name.trim().slice(0, 1).toUpperCase() || "?";
}

function locLabel(s: WatchedSource): string | null {
  if (s.city && s.state) return `${s.city}, ${s.state}`;
  if (s.city) return s.city;
  if (s.state) return s.state;
  return null;
}

function SourceCard({
  source,
  selected,
  onSelect,
  onRemove,
  onReactivate,
  removing,
  reactivating,
}: {
  source: WatchedSource;
  selected: boolean;
  onSelect: () => void;
  onRemove: () => void;
  onReactivate: () => void;
  removing: boolean;
  reactivating: boolean;
}) {
  const displayName = source.full_name || source.handle;
  const loc = locLabel(source);
  const scanAge = formatAge(source.last_scanned_at);
  const postAge = formatAge(source.last_post_at);
  return (
    <div
      className={`kanban-card ${source.active ? "" : "kanban-card--inactive"} ${selected ? "kanban-card--selected" : ""} ${source.scan_mode === "monitor_only" ? "kanban-card--monitor" : ""}`}
      onClick={onSelect}
      role="button"
      tabIndex={0}
      onKeyDown={(e) => {
        if (e.key === "Enter" || e.key === " ") {
          e.preventDefault();
          onSelect();
        }
      }}
    >
      <div className="kanban-card__top">
        {source.profile_pic_url ? (
          <img className="kanban-card__avatar" src={mediaURL(source.profile_pic_url)} alt="" referrerPolicy="no-referrer" />
        ) : (
          <span className="kanban-card__avatar kanban-card__avatar--fallback">{initials(displayName)}</span>
        )}
        <div className="kanban-card__identity">
          <span className="kanban-card__name" title={displayName}>
            {displayName}
          </span>
          <a
            className="kanban-card__handle"
            href={`https://instagram.com/${source.handle}`}
            target="_blank"
            rel="noreferrer"
            onClick={(e) => e.stopPropagation()}
          >
            @{source.handle}
          </a>
        </div>
      </div>

      {source.follower_count != null ? (
        <div className="kanban-card__stats">
          <span>
            <strong>{formatCount(source.follower_count)}</strong> followers
          </span>
          <span>
            <strong>{formatCount(source.post_count)}</strong> posts
          </span>
        </div>
      ) : (
        <div className="kanban-card__stats kanban-card__stats--pending">No profile check yet</div>
      )}

      {loc && <div className="kanban-card__loc">📍 {loc}</div>}
      <div className="kanban-card__meta-row">
        <span>{source.posts_stored ?? 0} stored</span>
        {postAge && <span>post {postAge}</span>}
        {scanAge && <span>scan {scanAge}</span>}
      </div>
      {source.stale && <div className="kanban-card__stale">Stale · needs scan</div>}
      {source.scan_mode === "monitor_only" && (
        <div className="kanban-card__monitor-pill">Monitor only</div>
      )}

      <div className="kanban-card__footer">
        <span className={`kanban-card__pill ${source.active ? "kanban-card__pill--watching" : "kanban-card__pill--paused"}`}>
          {source.active ? "Watching" : "Paused"}
        </span>
        {source.active ? (
          <button
            className="kanban-card__remove"
            onClick={(e) => {
              e.stopPropagation();
              onRemove();
            }}
            disabled={removing}
          >
            Stop
          </button>
        ) : (
          <button
            className="kanban-card__remove"
            onClick={(e) => {
              e.stopPropagation();
              onReactivate();
            }}
            disabled={reactivating}
          >
            Resume
          </button>
        )}
      </div>
    </div>
  );
}

type PostFilter = "people" | "all";

function looksLikePeoplePost(p: SourcePost): boolean {
  const cap = (p.caption || "").toLowerCase();
  if (/styled shoot|editorial|workshop|#ad\b|sponsored/.test(cap)) return false;
  const tags = p.tags || [];
  // at least one non-business-ish tag heuristic on client
  const biz = /photo|studio|farm|event|venue|design|wedding|planner|florist|luminary|band|music/;
  const people = tags.filter((t) => !biz.test(t.toLowerCase()));
  if (people.length >= 2) return true;
  if (/\b(and|&)\b/.test(cap) && /(engaged|yes|fiancé|fiance|proposal|capturing|moments)/.test(cap)) return true;
  return people.length >= 1 && tags.length <= 6;
}

function PostTile({ post }: { post: SourcePost }) {
  return (
    <article className="source-post">
      {post.image_url ? (
        <a href={post.url || post.image_url} target="_blank" rel="noreferrer" className="source-post__media">
          <img src={mediaURL(post.image_url)} alt="" referrerPolicy="no-referrer" />
        </a>
      ) : (
        <div className="source-post__media source-post__media--empty">No image</div>
      )}
      <div className="source-post__body">
        <p className="source-post__caption">{post.caption || "—"}</p>
        {(post.tags?.length ?? 0) > 0 && (
          <div className="source-post__tags">
            {post.tags!.slice(0, 8).map((t) => (
              <span key={t} className="source-post__tag">
                @{t}
              </span>
            ))}
          </div>
        )}
        <div className="source-post__meta">
          {post.location && <span>{post.location}</span>}
          <span>{new Date(post.observed_at).toLocaleString()}</span>
          {post.url && (
            <a href={post.url} target="_blank" rel="noreferrer">
              Open
            </a>
          )}
        </div>
      </div>
    </article>
  );
}

function qualityBadge(couple: ScannedCouple): { text: string; cls: string } {
  const q = couple.quality ?? 0;
  const label = couple.quality_label ?? "weak";
  if (label === "strong_couple" || q >= 75) return { text: `Strong · ${q}`, cls: "scan-q scan-q--strong" };
  if (label === "likely_couple" || q >= 55) return { text: `Likely · ${q}`, cls: "scan-q scan-q--likely" };
  return { text: `Weak · ${q}`, cls: "scan-q scan-q--weak" };
}

function CoupleScanCard({
  couple,
  onApprove,
  onIgnore,
  busy,
}: {
  couple: ScannedCouple;
  onApprove?: () => void;
  onIgnore?: () => void;
  busy?: boolean;
}) {
  const badge = qualityBadge(couple);
  return (
    <div className={`scan-couple ${couple.quality_label === "strong_couple" ? "scan-couple--strong" : ""}`}>
      {couple.image_url && (
        <a className="scan-couple__thumb" href={couple.post_url || couple.image_url} target="_blank" rel="noreferrer">
          <img src={mediaURL(couple.image_url)} alt="" referrerPolicy="no-referrer" />
        </a>
      )}
      <div className="scan-couple__body">
        <div className="scan-couple__pair">
          <strong>@{couple.handle_a}</strong>
          <span className="prospect-card__amp">&</span>
          <strong>@{couple.handle_b}</strong>
          <span className={badge.cls}>{badge.text}</span>
        </div>
        {couple.caption && <p className="scan-couple__caption">{couple.caption}</p>}
        <div className="scan-couple__meta">
          {couple.has_people_shot && <span className="scan-q scan-q--strong">people</span>}
          {couple.tags?.slice(0, 3).map((t) => (
            <span key={t} className="source-post__tag">
              @{t}
            </span>
          ))}
          {couple.vendor_tags && couple.vendor_tags.length > 0 && (
            <span className="scan-couple__vendors">+{couple.vendor_tags.length} vendors hidden</span>
          )}
        </div>
        {couple.action_id && onApprove && onIgnore && (
          <div className="prospect-card__actions">
            <button className="btn btn--primary btn--sm" disabled={busy} onClick={onApprove}>
              Approve
            </button>
            <button className="btn btn--ghost btn--sm" disabled={busy} onClick={onIgnore}>
              Ignore
            </button>
          </div>
        )}
      </div>
    </div>
  );
}

function ScanProgress({ job }: { job: ScanJob }) {
  const steps = [
    { id: "queued", label: "Queued" },
    { id: "profile", label: "Profile" },
    { id: "posts", label: "Posts" },
    { id: "scanning", label: "Scanning" },
    { id: "done", label: "Done" },
  ];
  const stepIdx = Math.max(
    0,
    steps.findIndex((s) => job.step === s.id || (job.status === "done" && s.id === "done")),
  );
  return (
    <div className="scan-progress">
      <div className="scan-progress__bar">
        <div className="scan-progress__fill" style={{ width: `${job.progress}%` }} />
      </div>
      <div className="scan-progress__meta">
        <span>
          {job.status === "running" || job.status === "queued" ? "Running" : job.status} · {job.progress}%
        </span>
        <span>{job.message || job.step}</span>
      </div>
      <div className="scan-progress__steps">
        {steps.map((s, i) => (
          <span key={s.id} className={`scan-progress__step ${i <= stepIdx ? "scan-progress__step--on" : ""}`}>
            {s.label}
          </span>
        ))}
      </div>
    </div>
  );
}

function SourceDetail({
  source,
  onClose,
  onSourceUpdated,
  onScanDone,
}: {
  source: WatchedSource;
  onClose: () => void;
  onSourceUpdated: () => void;
  onScanDone?: () => void;
}) {
  const { data: posts, error, isLoading, refetch: refetchPosts } = useSourcePosts(source.handle);
  const startJob = useStartScanJob();
  const [jobId, setJobId] = useState<string | null>(null);
  const { data: job } = useScanJob(jobId ?? undefined);
  const enrich = useEnrichSource();
  const patchLoc = usePatchSourceLocation();
  const approve = useApproveAction();
  const ignore = useIgnoreAction();
  const toast = useToast();
  const qc = useQueryClient();
  const [scanResult, setScanResult] = useState<SourceScanResult | null>(null);
  const [city, setCity] = useState(source.city ?? "");
  const [state, setState] = useState(source.state ?? "");
  const [postFilter, setPostFilter] = useState<PostFilter>("people");
  const displayName = source.full_name || source.handle;
  const loc = locLabel(source);
  const monitorOnly = source.scan_mode === "monitor_only";

  useEffect(() => {
    setCity(source.city ?? "");
    setState(source.state ?? "");
    setScanResult(null);
    setJobId(null);
    setPostFilter("people");
  }, [source.handle, source.city, source.state]);

  useEffect(() => {
    if (!job) return;
    if (job.status === "done" && job.result) {
      setScanResult(job.result);
      refetchPosts();
      onSourceUpdated();
      qc.invalidateQueries({ queryKey: ["sources"] });
      const n = job.result.couples?.length ?? 0;
      const a = job.result.actions_created ?? 0;
      toast.push(`Scan @${source.handle}: ${n} couples, ${a} approvals`, n || a ? "ok" : "info");
    }
    if (job.status === "failed") {
      toast.push(job.error || "Scan failed", "err");
    }
  }, [job?.status, job?.id]); // eslint-disable-line react-hooks/exhaustive-deps

  const runAgent = () => {
    setScanResult(null);
    startJob.mutate(
      { handle: source.handle, limit: 15 },
      {
        onSuccess: (res) => {
          if (res.job_id) setJobId(res.job_id);
          else toast.push("No job id returned", "err");
        },
        onError: (e) => toast.push((e as Error).message, "err"),
      },
    );
  };

  const filteredPosts = useMemo(() => {
    const list = posts ?? [];
    if (postFilter === "all") return list;
    return list.filter(looksLikePeoplePost);
  }, [posts, postFilter]);

  const running = job?.status === "running" || job?.status === "queued" || startJob.isPending;

  return (
    <aside className="source-detail">
      <div className="source-detail__header">
        <div className="source-detail__identity">
          {source.profile_pic_url ? (
            <img className="source-detail__avatar" src={mediaURL(source.profile_pic_url)} alt="" referrerPolicy="no-referrer" />
          ) : (
            <span className="source-detail__avatar source-detail__avatar--fallback">{initials(displayName)}</span>
          )}
          <div>
            <h3 className="source-detail__name">{displayName}</h3>
            <a href={`https://instagram.com/${source.handle}`} target="_blank" rel="noreferrer">
              @{source.handle}
            </a>
            <div className="source-detail__stats">
              {source.follower_count != null && <span>{formatCount(source.follower_count)} followers</span>}
              {source.post_count != null && <span>{formatCount(source.post_count)} posts</span>}
              {loc && <span>📍 {loc}</span>}
            </div>
            <div className="source-detail__stats">
              <span>{source.posts_stored ?? 0} stored</span>
              {source.last_scanned_at && <span>scanned {formatAge(source.last_scanned_at)}</span>}
              {source.last_post_at && <span>last post {formatAge(source.last_post_at)}</span>}
            </div>
          </div>
        </div>
        <button type="button" className="btn btn--ghost" onClick={onClose}>
          Close
        </button>
      </div>

      <form
        className="source-loc-form"
        onSubmit={(e) => {
          e.preventDefault();
          patchLoc.mutate(
            { handle: source.handle, city: city.trim(), state: state.trim().toUpperCase() },
            { onSuccess: () => onSourceUpdated() },
          );
        }}
      >
        <input className="feed-filter" placeholder="City" value={city} onChange={(e) => setCity(e.target.value)} />
        <input
          className="feed-filter source-loc-form__state"
          placeholder="ST"
          maxLength={2}
          value={state}
          onChange={(e) => setState(e.target.value.toUpperCase())}
        />
        <button className="btn btn--ghost btn--sm" type="submit" disabled={patchLoc.isPending}>
          Save location
        </button>
        <button
          className="btn btn--ghost btn--sm"
          type="button"
          disabled={enrich.isPending}
          onClick={() =>
            enrich.mutate(source.handle, {
              onSuccess: () => {
                onSourceUpdated();
                toast.push("Profile refreshed", "ok");
              },
              onError: (e) => toast.push((e as Error).message, "err"),
            })
          }
        >
          {enrich.isPending ? "…" : "Refresh profile"}
        </button>
      </form>

      <div className="source-agent-bar">
        {monitorOnly ? (
          <>
            <p className="source-detail__hint" style={{ margin: 0 }}>
              <strong>Monitor only</strong> — venues/jewelers provide context, not couple discovery. Use a photographer
              source to find tagged couples.
            </p>
            <button className="btn btn--ghost" type="button" onClick={runAgent} disabled={running}>
              {running ? "Running…" : "Force scan anyway"}
            </button>
          </>
        ) : (
          <>
            <button className="btn btn--primary" type="button" onClick={runAgent} disabled={running}>
              {running ? "Running agent…" : "Run agent — find tagged couples"}
            </button>
            <p className="source-detail__hint" style={{ margin: 0 }}>
              Pulls posts via Bright Data, drops vendor tags, ranks real couples. Approve on Work.
            </p>
          </>
        )}
      </div>

      {job && (job.status === "running" || job.status === "queued") && <ScanProgress job={job} />}
      {job?.status === "failed" && <div className="empty-state">{job.error || "Scan failed"}</div>}

      {scanResult && (
        <div className="scan-results">
          <div className="scan-results__summary">
            <strong>Scan complete</strong>
            <span>
              {scanResult.posts_fetched} fetched · {scanResult.posts_processed} new · {scanResult.duplicates} dupes ·{" "}
              {scanResult.couples?.length ?? 0} couples · {scanResult.actions_created} approvals
            </span>
            <span className="scan-results__ms">{scanResult.duration_ms}ms</span>
          </div>
          <div className="scan-results__actions">
            <button type="button" className="btn btn--primary btn--sm" onClick={() => onScanDone?.()}>
              Open in Work
            </button>
          </div>
          <p className="scan-results__hint">
            Ranked by couple quality. Business tags (venues, planners) are hidden. Styled shoots excluded.
          </p>
          {(scanResult.couples ?? []).length === 0 ? (
            <div className="empty-state">No person-tag couples in this batch.</div>
          ) : (
            <div className="scan-couple-list">
              {scanResult.couples.map((c, i) => (
                <CoupleScanCard
                  key={`${c.handle_a}-${c.handle_b}-${i}`}
                  couple={c}
                  busy={approve.isPending || ignore.isPending}
                  onApprove={
                    c.action_id
                      ? () =>
                          approve.mutate(c.action_id!, {
                            onSuccess: () => toast.push("Approved", "ok"),
                          })
                      : undefined
                  }
                  onIgnore={
                    c.action_id
                      ? () =>
                          ignore.mutate(c.action_id!, {
                            onSuccess: () => toast.push("Ignored", "info"),
                          })
                      : undefined
                  }
                />
              ))}
            </div>
          )}
        </div>
      )}

      <div className="source-detail__section-row">
        <h4 className="source-detail__section">Stored posts</h4>
        <div className="work-filters" style={{ marginBottom: 0 }}>
          <button
            type="button"
            className={`work-filter ${postFilter === "people" ? "work-filter--active" : ""}`}
            onClick={() => setPostFilter("people")}
          >
            People
          </button>
          <button
            type="button"
            className={`work-filter ${postFilter === "all" ? "work-filter--active" : ""}`}
            onClick={() => setPostFilter("all")}
          >
            All
          </button>
        </div>
      </div>
      {isLoading && <div className="empty-state">Loading posts…</div>}
      {error && <div className="empty-state">{(error as Error).message}</div>}
      {!isLoading && !error && filteredPosts.length === 0 && (
        <div className="empty-state">
          {postFilter === "people"
            ? "No people-forward posts stored. Run agent or switch to All."
            : "No posts yet — Run agent to pull this feed."}
        </div>
      )}
      <div className="source-post-grid">
        {filteredPosts.map((p) => (
          <PostTile key={p.id} post={p} />
        ))}
      </div>
    </aside>
  );
}

export function SourcesView({
  initialHandle,
  onOpenHandle,
  onScanDone,
}: {
  initialHandle?: string;
  onOpenHandle?: (handle: string) => void;
  onScanDone?: () => void;
} = {}) {
  const { data: sources, error, refetch } = useSources(false);
  const addSource = useAddSource();
  const removeSource = useRemoveSource();
  const reactivateSource = useReactivateSource();
  const bulk = useBulkScan();
  const suppressVendors = useSuppressVendorPairs();
  const toast = useToast();
  const [handle, setHandle] = useState("");
  const [sourceClass, setSourceClass] = useState<string>(SOURCE_CLASSES[0]);
  const [city, setCity] = useState("");
  const [state, setState] = useState("");
  const [selectedHandle, setSelectedHandle] = useState<string | null>(initialHandle ?? null);
  const [bulkJobId, setBulkJobId] = useState<string | null>(null);
  const { data: bulkJob } = useScanJob(bulkJobId ?? undefined);

  useEffect(() => {
    if (initialHandle) setSelectedHandle(initialHandle);
  }, [initialHandle]);

  useEffect(() => {
    if (bulkJob?.status === "done") {
      toast.push(
        `Bulk scan done · ${(bulkJob.results || []).length} sources`,
        "ok",
      );
      refetch();
      setBulkJobId(null);
    }
    if (bulkJob?.status === "failed") {
      toast.push(bulkJob.error || "Bulk scan failed", "err");
      setBulkJobId(null);
    }
  }, [bulkJob?.status]); // eslint-disable-line react-hooks/exhaustive-deps

  const submit = (e: React.FormEvent) => {
    e.preventDefault();
    const h = handle.trim().replace(/^@/, "");
    if (!h) return;
    addSource.mutate(
      {
        handle: h,
        source_class: sourceClass,
        city: city.trim() || undefined,
        state: state.trim().toUpperCase() || undefined,
      },
      {
        onSuccess: (src) => {
          setHandle("");
          setCity("");
          setState("");
          setSelectedHandle(src.handle);
          onOpenHandle?.(src.handle);
          toast.push(`Watching @${src.handle}`, "ok");
        },
        onError: (err) => toast.push((err as Error).message, "err"),
      },
    );
  };

  const byCategory = new Map<string, WatchedSource[]>();
  for (const c of SOURCE_CLASSES) byCategory.set(c, []);
  for (const s of sources ?? []) {
    (byCategory.get(s.source_class) ?? byCategory.set(s.source_class, []).get(s.source_class)!).push(s);
  }

  const total = sources?.length ?? 0;
  const watching = sources?.filter((s) => s.active).length ?? 0;
  const withLoc = sources?.filter((s) => s.city || s.state).length ?? 0;
  const staleN = sources?.filter((s) => s.stale).length ?? 0;
  const selected = sources?.find((s) => s.handle === selectedHandle) ?? null;

  const selectHandle = (h: string) => {
    setSelectedHandle(h);
    onOpenHandle?.(h);
  };

  return (
    <div className={`view view--sources ${selected ? "view--sources-split" : ""}`}>
      <div className="sources-main">
        <div className="sources-header">
          <div>
            <h2 className="view__title">Radar sources</h2>
            <p className="view__subtitle">
              Wedding vendors, publications, and venues we monitor for new couples.
            </p>
          </div>
          <div className="sources-header__stats">
            <div className="sources-header__stat">
              <span className="sources-header__stat-value">{total}</span>
              <span className="sources-header__stat-label">total</span>
            </div>
            <div className="sources-header__stat">
              <span className="sources-header__stat-value">{watching}</span>
              <span className="sources-header__stat-label">watching</span>
            </div>
            <div className="sources-header__stat">
              <span className="sources-header__stat-value">{withLoc}</span>
              <span className="sources-header__stat-label">with location</span>
            </div>
            <div className="sources-header__stat">
              <span className="sources-header__stat-value">{staleN}</span>
              <span className="sources-header__stat-label">stale</span>
            </div>
          </div>
        </div>

        <div className="sources-toolbar">
          <button
            type="button"
            className="btn btn--primary"
            disabled={bulk.isPending || !!bulkJobId}
            onClick={() =>
              bulk.mutate(
                {
                  stale_only: true,
                  classes: ["engagement_photographer", "proposal_planner"],
                  posts_per_source: 12,
                },
                {
                  onSuccess: (r) => {
                    if (!r.job_id) {
                      toast.push(r.message || "Nothing to scan", "info");
                      return;
                    }
                    setBulkJobId(r.job_id);
                    toast.push(`Bulk scan started · ${r.count} sources`, "ok");
                  },
                  onError: (e) => toast.push((e as Error).message, "err"),
                },
              )
            }
          >
            {bulkJobId ? "Bulk scanning…" : "Scan all stale photographers"}
          </button>
          <button
            type="button"
            className="btn btn--ghost"
            disabled={suppressVendors.isPending}
            onClick={() =>
              suppressVendors.mutate(undefined, {
                onSuccess: (r) => toast.push(`Suppressed ${r.suppressed} vendor-vendor pairs`, "ok"),
                onError: (e) => toast.push((e as Error).message, "err"),
              })
            }
          >
            Clean vendor-vendor pairs
          </button>
          <button type="button" className="btn btn--ghost" onClick={() => onScanDone?.()}>
            Open Work
          </button>
        </div>

        {bulkJobId && bulkJob && (bulkJob.status === "running" || bulkJob.status === "queued") && (
          <div className="scan-progress" style={{ marginBottom: 16 }}>
            <ScanProgress job={bulkJob} />
          </div>
        )}

        <form className="source-form source-form--geo" onSubmit={submit}>
          <input
            className="feed-filter"
            placeholder="@vendorhandle"
            value={handle}
            onChange={(e) => setHandle(e.target.value)}
          />
          <select value={sourceClass} onChange={(e) => setSourceClass(e.target.value)}>
            {SOURCE_CLASSES.map((c) => (
              <option key={c} value={c}>
                {CATEGORY_LABEL[c]}
              </option>
            ))}
          </select>
          <input className="feed-filter" placeholder="City" value={city} onChange={(e) => setCity(e.target.value)} />
          <input
            className="feed-filter source-loc-form__state"
            placeholder="ST"
            maxLength={2}
            value={state}
            onChange={(e) => setState(e.target.value.toUpperCase())}
          />
          <button className="btn btn--primary" type="submit" disabled={addSource.isPending}>
            Add source
          </button>
        </form>
        {addSource.error && <div className="empty-state">{(addSource.error as Error).message}</div>}
        {error && <div className="empty-state">{(error as Error).message}</div>}

        {!sources || sources.length === 0 ? (
          <div className="empty-state">No watched sources yet.</div>
        ) : (
          <div className="kanban-board">
            {SOURCE_CLASSES.map((c) => {
              const rows = byCategory.get(c) ?? [];
              if (rows.length === 0) return null; // hide empty categories
              return (
                <div className="kanban-column" key={c}>
                  <div className="kanban-column__header">
                    <span className="kanban-column__title">{CATEGORY_LABEL[c]}</span>
                    <span className="kanban-column__count">{rows.length}</span>
                  </div>
                  <div className="kanban-column__cards">
                    {rows.map((s) => (
                      <SourceCard
                        key={s.id}
                        source={s}
                        selected={selectedHandle === s.handle}
                        onSelect={() => selectHandle(s.handle)}
                        onRemove={() => removeSource.mutate(s.handle)}
                        onReactivate={() => reactivateSource.mutate(s)}
                        removing={removeSource.isPending}
                        reactivating={reactivateSource.isPending}
                      />
                    ))}
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </div>
      {selected && (
        <SourceDetail
          source={selected}
          onClose={() => setSelectedHandle(null)}
          onSourceUpdated={() => refetch()}
          onScanDone={onScanDone}
        />
      )}
    </div>
  );
}
