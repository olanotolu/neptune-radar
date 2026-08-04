import { Suspense, lazy, useCallback, useEffect, useMemo, useRef, useState } from "react";
import "./App.css";
import { useIngestStatus, useLiveEvents, useOpsSummary, usePauseIngest, useResumeIngest } from "./api/hooks";
import { KeyboardHelp } from "./components/KeyboardHelp";
import { NotificationCenter } from "./components/NotificationCenter";
import { OnboardingTour, isOnboarded } from "./components/OnboardingTour";
import { ToastProvider } from "./components/Toast";
import { useDarkMode } from "./hooks/useDarkMode";
import { useKeyboard } from "./hooks/useKeyboard";
import { FeedView } from "./views/FeedView";
import { CoupleGraphView } from "./views/CoupleGraphView";
import { CaseDetailView } from "./views/CaseDetailView";
import { AuditTrailView } from "./views/AuditTrailView";
import { AgentRunsView } from "./views/AgentRunsView";
import { SourcesView } from "./views/SourcesView";
import { WorkView } from "./views/WorkView";
import { TodayView } from "./views/TodayView";
import { DossierView } from "./views/DossierView";
import { SearchView } from "./views/SearchView";
import { DLQView } from "./views/DLQView";
import { JobsView } from "./views/JobsView";
import { SettingsView } from "./views/SettingsView";
import { FunnelView } from "./views/FunnelView";
import { CostView } from "./views/CostView";
import { OpsView } from "./views/OpsView";
import { InterviewView } from "./views/InterviewView";

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
  if (tab === "work" || tab === "dossier") {
    const f = qs.get("filter");
    if (f === "action" || f === "pics" || f === "loc" || f === "all") route.filter = f;
    const c = qs.get("couple") || parts[1];
    if (c) route.coupleId = decodeURIComponent(c);
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
  { id: "interview", label: "Interview", path: "/interview" },
  { id: "sources", label: "Sources", path: "/sources" },
  { id: "map", label: "Map", path: "/map" },
  { id: "feed", label: "Feed", path: "/feed" },
  { id: "graph", label: "Graph", path: "/graph" },
  { id: "case", label: "Cases", path: "/case" },
  { id: "funnel", label: "Funnel", path: "/funnel" },
  { id: "cost", label: "Budget", path: "/cost" },
  { id: "ops", label: "Ops", path: "/ops" },
  { id: "runs", label: "Runs", path: "/runs" },
  { id: "audit", label: "System", path: "/audit" },
  { id: "search", label: "Search", path: "/search" },
  { id: "dlq", label: "DLQ", path: "/dlq" },
  { id: "jobs", label: "Jobs", path: "/jobs" },
  { id: "settings", label: "Settings", path: "/settings" },
];

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

function SearchBar({ inputRef }: { inputRef: React.RefObject<HTMLInputElement | null> }) {
  const [q, setQ] = useState("");
  return (
    <form
      className="app-search"
      role="search"
      onSubmit={(e) => {
        e.preventDefault();
        const v = q.trim();
        if (v) setHash(`/search?q=${encodeURIComponent(v)}`);
      }}
    >
      <label htmlFor="app-search-input" className="sr-only">Search</label>
      <input
        id="app-search-input"
        ref={inputRef}
        type="search"
        className="app-search__input"
        data-testid="app-search-input"
        placeholder="Search couples, leads, cases…"
        value={q}
        onChange={(e) => setQ(e.target.value)}
      />
    </form>
  );
}

function LiveIndicator({ connected }: { connected: boolean }) {
  return (
    <span className={`live-indicator ${connected ? "live-indicator--on" : "live-indicator--off"}`} title={connected ? "Live feed connected" : "Live feed disconnected"}>
      <span className="live-indicator__dot" aria-hidden />
      LIVE
    </span>
  );
}

function Shell() {
  const [route, setRoute] = useState<Route>(() => parseHash());
  const [showHelp, setShowHelp] = useState(false);
  const [showTour, setShowTour] = useState(() => !isOnboarded());
  const [dark, toggleDark] = useDarkMode();
  const searchRef = useRef<HTMLInputElement | null>(null);
  const gPressed = useRef(false);
  const { connected } = useLiveEvents();

  useEffect(() => {
    const onHash = () => setRoute(parseHash());
    window.addEventListener("hashchange", onHash);
    if (!window.location.hash) setHash("/today");
    return () => window.removeEventListener("hashchange", onHash);
  }, []);

  const navigate = useCallback((path: string) => setHash(path), []);

  useKeyboard((e) => {
    // `g` is a prefix key — set a flag and wait for the next key.
    if (gPressed.current) {
      gPressed.current = false;
      if (e.key === "t") return navigate("/today");
      if (e.key === "w") return navigate("/work?filter=action");
      if (e.key === "m") return navigate("/map");
      if (e.key === "s") return navigate("/sources");
      if (e.key === "f") return navigate("/funnel");
      return;
    }
    if (e.key === "g") {
      gPressed.current = true;
      // Reset the prefix if no second key arrives within 700ms.
      window.setTimeout(() => {
        gPressed.current = false;
      }, 700);
      return;
    }
    if (e.key === "/") {
      e.preventDefault();
      searchRef.current?.focus();
      return;
    }
    if (e.key === "?") {
      e.preventDefault();
      setShowHelp(true);
    }
    if (e.key === "Escape") {
      setShowHelp(false);
    }
  });

  const body = useMemo(() => {
    switch (route.tab) {
      case "today":
        return <TodayView onNavigate={navigate} />;
      case "dossier":
        if (!route.coupleId) {
          return <TodayView onNavigate={navigate} />;
        }
        return (
          <DossierView
            coupleId={route.coupleId}
            onClose={() => navigate("/work?filter=action")}
            onCongratulate={(id) => navigate(`/congratulate?couple=${encodeURIComponent(id)}`)}
          />
        );
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
      case "interview":
        return <InterviewView />;
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
      case "funnel":
        return <FunnelView />;
      case "cost":
        return <CostView />;
      case "ops":
        return <OpsView />;
      case "audit":
        return <AuditTrailView />;
      case "runs":
        return <AgentRunsView />;
      case "search":
        return <SearchView />;
      case "dlq":
        return <DLQView />;
      case "jobs":
        return <JobsView />;
      case "settings":
        return <SettingsView />;
      default:
        return <TodayView onNavigate={navigate} />;
    }
  }, [route, navigate]);

  return (
    <div className="app-shell">
      <a href="#app-main" className="skip-link">Skip to content</a>
      <header className="app-header" role="banner">
        <div className="app-header__top">
          <div className="app-header__brand">
            <span className="app-header__logo" aria-hidden="true">N</span>
            <div>
              <div className="app-header__title">Neptune Radar</div>
              <div className="app-header__subtitle">Neptune Growth OS · celebrate first · human in the loop</div>
            </div>
          </div>
          <div className="app-header__center">
            <WatchTransport />
            <SearchBar inputRef={searchRef} />
          </div>
          <div className="app-header__right">
            <LiveIndicator connected={connected} />
            <NotificationCenter onNavigate={navigate} />
            <button
              type="button"
              className="dark-mode-toggle"
              data-testid="dark-mode-toggle"
              onClick={toggleDark}
              aria-label={dark ? "Switch to light mode" : "Switch to dark mode"}
              title={dark ? "Switch to light mode" : "Switch to dark mode"}
            >
              {dark ? "☀" : "☾"}
            </button>
          </div>
        </div>
        <nav className="app-nav" aria-label="Main navigation" data-testid="app-nav">
          {NAV.map((t) => {
            const isActive = route.tab === t.id || (t.id === "work" && (route.tab === "prospects" || route.tab === "queue"));
            return (
              <button
                key={t.id}
                type="button"
                className={`app-nav__tab ${isActive ? "app-nav__tab--active" : ""}`}
                data-testid={`nav-${t.id}`}
                aria-current={isActive ? "page" : undefined}
                onClick={() => navigate(t.path)}
              >
                {t.label}
              </button>
            );
          })}
        </nav>
      </header>
      <main id="app-main" className={`app-main ${route.tab === "work" || route.tab === "sources" || route.tab === "map" ? "app-main--wide" : ""}`} role="main" data-testid="app-main">
        <Suspense fallback={<div className="empty-state">Loading…</div>}>
          {body}
        </Suspense>
      </main>
      {showHelp && <KeyboardHelp onClose={() => setShowHelp(false)} data-testid="keyboard-help" />}
      {showTour && <OnboardingTour onClose={() => setShowTour(false)} />}
    </div>
  );
}

export default function App() {
  return (
    <ToastProvider>
      <Shell />
    </ToastProvider>
  );
}
