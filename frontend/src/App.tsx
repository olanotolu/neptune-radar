import { Suspense, lazy, useCallback, useEffect, useMemo, useState } from "react";
import "./App.css";
import { getToken, setToken, setUnauthorizedHandler } from "./api/client";
import { useIngestStatus, useOpsSummary, usePauseIngest, useResumeIngest } from "./api/hooks";
import { ToastProvider } from "./components/Toast";
import { FeedView } from "./views/FeedView";
import { CoupleGraphView } from "./views/CoupleGraphView";
import { CaseDetailView } from "./views/CaseDetailView";
import { AuditTrailView } from "./views/AuditTrailView";
import { SourcesView } from "./views/SourcesView";
import { WorkView } from "./views/WorkView";
import { TodayView } from "./views/TodayView";

// Map (d3 + both us-atlas topojson files) and Congratulate (postcard kit)
// are the two heaviest views — lazy so the main bundle doesn't carry them.
const MapView = lazy(() =>
  import("./views/MapView").then((m) => ({ default: m.MapView })),
);
const CongratulateView = lazy(() =>
  import("./views/CongratulateView").then((m) => ({ default: m.CongratulateView })),
);

type Route = {
  tab: string;
  sourceHandle?: string;
  filter?: "all" | "action" | "pics" | "loc";
  coupleId?: string;
};

function parseHash(): Route {
  const raw = (window.location.hash || "#/today").replace(/^#/, "") || "/today";
  const [pathPart, queryPart] = raw.split("?");
  const parts = pathPart.split("/").filter(Boolean);
  const qs = new URLSearchParams(queryPart || "");
  const tab = parts[0] || "today";
  const route: Route = { tab };
  if (tab === "sources" && parts[1]) route.sourceHandle = decodeURIComponent(parts[1]);
  if (tab === "work") {
    const f = qs.get("filter");
    if (f === "action" || f === "pics" || f === "loc" || f === "all") route.filter = f;
    const c = qs.get("couple");
    if (c) route.coupleId = c;
  }
  if (tab === "congratulate" || tab === "kits") {
    const c = qs.get("couple");
    if (c) route.coupleId = c;
  }
  return route;
}

function setHash(path: string) {
  const next = path.startsWith("#") ? path : `#${path.startsWith("/") ? path : `/${path}`}`;
  if (window.location.hash !== next) {
    window.location.hash = next;
  } else {
    window.dispatchEvent(new HashChangeEvent("hashchange"));
  }
}

const NAV: { id: string; label: string; path: string }[] = [
  { id: "today", label: "Today", path: "/today" },
  { id: "work", label: "Work", path: "/work?filter=action" },
  { id: "congratulate", label: "Congratulate", path: "/congratulate" },
  { id: "sources", label: "Sources", path: "/sources" },
  { id: "map", label: "Map", path: "/map" },
  { id: "feed", label: "Feed", path: "/feed" },
  { id: "graph", label: "Graph", path: "/graph" },
  { id: "case", label: "Cases", path: "/case" },
  { id: "audit", label: "System", path: "/audit" },
];

function TokenGate({ onSave }: { onSave: () => void }) {
  const [value, setValue] = useState("");
  return (
    <div className="token-gate">
      <h2>Neptune Radar</h2>
      <p>Internal operator console — enter the admin token (NEPTUNE_ADMIN_TOKEN).</p>
      <form
        onSubmit={(e) => {
          e.preventDefault();
          setToken(value.trim());
          onSave();
        }}
      >
        <input
          type="password"
          placeholder="admin token"
          value={value}
          onChange={(e) => setValue(e.target.value)}
          autoFocus
        />
        <button className="btn btn--primary" type="submit">
          Connect
        </button>
      </form>
    </div>
  );
}

function WatchTransport() {
  const { data: status, isLoading } = useIngestStatus();
  const { data: ops } = useOpsSummary();
  const pause = usePauseIngest();
  const resume = useResumeIngest();
  const busy = pause.isPending || resume.isPending;

  const paused = status?.paused ?? false;
  const providerOk = status?.provider_available ?? false;
  const running = status?.running ?? false;
  const used = status?.results_used_today ?? 0;
  const budget = status?.daily_budget ?? ops?.daily_budget ?? 0;
  const pct = budget > 0 ? Math.round((used / budget) * 100) : 0;

  let label = "…";
  let tone: "live" | "paused" | "idle" | "warn" = "idle";
  if (!isLoading && status) {
    if (paused) {
      label = "Paused";
      tone = "paused";
    } else if (!providerOk) {
      label = "No provider";
      tone = "idle";
    } else if (pct >= 95) {
      label = "Budget maxed";
      tone = "warn";
    } else if (running) {
      label = "Live";
      tone = "live";
    } else {
      label = "Idle";
      tone = "idle";
    }
  }

  return (
    <div className="watch-transport" title="Global radar — pause stops Apify spend">
      <span className={`watch-transport__pill watch-transport__pill--${tone}`}>
        <span className="watch-transport__dot" aria-hidden />
        {label}
      </span>
      <span className="watch-transport__meta">
        {used}
        {budget ? `/${budget}` : ""} · {pct}%
      </span>
      {ops && ops.pending_actions > 0 && (
        <button type="button" className="watch-transport__queue" onClick={() => setHash("/work?filter=action")}>
          {ops.pending_actions} queue
        </button>
      )}
      <button
        type="button"
        className={`watch-transport__btn ${paused ? "watch-transport__btn--play" : "watch-transport__btn--pause"}`}
        onClick={() => (paused ? resume.mutate() : pause.mutate())}
        disabled={busy || isLoading}
        aria-label={paused ? "Resume watch loop" : "Pause watch loop"}
      >
        {paused ? (
          <>
            <span className="watch-transport__icon">▶</span> Play
          </>
        ) : (
          <>
            <span className="watch-transport__icon">⏸</span> Pause
          </>
        )}
      </button>
    </div>
  );
}

function Shell() {
  const [route, setRoute] = useState<Route>(() => parseHash());

  useEffect(() => {
    const onHash = () => setRoute(parseHash());
    window.addEventListener("hashchange", onHash);
    if (!window.location.hash) setHash("/today");
    return () => window.removeEventListener("hashchange", onHash);
  }, []);

  const navigate = useCallback((path: string) => setHash(path), []);

  const body = useMemo(() => {
    switch (route.tab) {
      case "today":
        return <TodayView onNavigate={navigate} />;
      case "work":
      case "prospects":
      case "queue":
        return (
          <WorkView
            initialFilter={route.filter ?? "action"}
            focusCoupleId={route.coupleId}
            onCongratulate={(id) => navigate(`/congratulate?couple=${encodeURIComponent(id)}`)}
          />
        );
      case "congratulate":
      case "kits":
        return <CongratulateView initialCoupleId={route.coupleId} />;
      case "sources":
        return (
          <SourcesView
            initialHandle={route.sourceHandle}
            onOpenHandle={(h) => navigate(`/sources/${encodeURIComponent(h)}`)}
            onScanDone={() => navigate("/work?filter=action")}
          />
        );
      case "map":
        return <MapView />;
      case "feed":
        return <FeedView />;
      case "graph":
        return <CoupleGraphView />;
      case "case":
        return <CaseDetailView />;
      case "audit":
        return <AuditTrailView />;
      default:
        return <TodayView onNavigate={navigate} />;
    }
  }, [route, navigate]);

  return (
    <div className="app-shell">
      <header className="app-header">
        <div className="app-header__brand">
          <span className="app-header__logo">N</span>
          <div>
            <div className="app-header__title">Neptune Radar</div>
            <div className="app-header__subtitle">Operator workbench · relationship radar</div>
          </div>
        </div>
        <WatchTransport />
        <nav className="app-nav">
          {NAV.map((t) => (
            <button
              key={t.id}
              type="button"
              className={`app-nav__tab ${route.tab === t.id || (t.id === "work" && (route.tab === "prospects" || route.tab === "queue")) ? "app-nav__tab--active" : ""}`}
              onClick={() => navigate(t.path)}
            >
              {t.label}
            </button>
          ))}
        </nav>
      </header>
      <main className={`app-main ${route.tab === "work" || route.tab === "sources" || route.tab === "map" ? "app-main--wide" : ""}`}>
        <Suspense fallback={<div className="empty-state">Loading…</div>}>
          {body}
        </Suspense>
      </main>
    </div>
  );
}

export default function App() {
  const [authed, setAuthed] = useState(!!getToken());

  useEffect(() => {
    setUnauthorizedHandler(() => setAuthed(false));
    return () => setUnauthorizedHandler(null);
  }, []);

  if (!authed) {
    return <TokenGate onSave={() => setAuthed(true)} />;
  }

  return (
    <ToastProvider>
      <Shell />
    </ToastProvider>
  );
}
