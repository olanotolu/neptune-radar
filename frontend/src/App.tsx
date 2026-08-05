import { Suspense, lazy, useCallback, useEffect, useMemo, useRef, useState } from "react";
import "./App.css";
import { useIngestStatus, useLiveEvents, useOpsSummary, usePauseIngest, useResumeIngest } from "./api/hooks";
import { KeyboardHelp } from "./components/KeyboardHelp";
import { NotificationCenter } from "./components/NotificationCenter";
import { OnboardingTour, isOnboarded } from "./components/OnboardingTour";
import { ToastProvider } from "./components/Toast";
import { useDarkMode } from "./hooks/useDarkMode";
import { useKeyboard } from "./hooks/useKeyboard";
// Only Today + shell stay eager. Everything else is route-lazy.
import { TodayView } from "./views/TodayView";

const WorkView = lazy(() => import("./views/WorkView").then((m) => ({ default: m.WorkView })));
const SourcesView = lazy(() => import("./views/SourcesView").then((m) => ({ default: m.SourcesView })));
const MapView = lazy(() => import("./views/MapView").then((m) => ({ default: m.MapView })));
const CongratulateView = lazy(() =>
  import("./views/CongratulateView").then((m) => ({ default: m.CongratulateView })),
);
const FeedView = lazy(() => import("./views/FeedView").then((m) => ({ default: m.FeedView })));
const CoupleGraphView = lazy(() =>
  import("./views/CoupleGraphView").then((m) => ({ default: m.CoupleGraphView })),
);
const CaseDetailView = lazy(() =>
  import("./views/CaseDetailView").then((m) => ({ default: m.CaseDetailView })),
);
const AuditTrailView = lazy(() =>
  import("./views/AuditTrailView").then((m) => ({ default: m.AuditTrailView })),
);
const AgentRunsView = lazy(() =>
  import("./views/AgentRunsView").then((m) => ({ default: m.AgentRunsView })),
);
const DossierView = lazy(() => import("./views/DossierView").then((m) => ({ default: m.DossierView })));
const SearchView = lazy(() => import("./views/SearchView").then((m) => ({ default: m.SearchView })));
const DLQView = lazy(() => import("./views/DLQView").then((m) => ({ default: m.DLQView })));
const JobsView = lazy(() => import("./views/JobsView").then((m) => ({ default: m.JobsView })));
const SettingsView = lazy(() =>
  import("./views/SettingsView").then((m) => ({ default: m.SettingsView })),
);
const FunnelView = lazy(() => import("./views/FunnelView").then((m) => ({ default: m.FunnelView })));
const CostView = lazy(() => import("./views/CostView").then((m) => ({ default: m.CostView })));
const OpsView = lazy(() => import("./views/OpsView").then((m) => ({ default: m.OpsView })));
const InterviewView = lazy(() =>
  import("./views/InterviewView").then((m) => ({ default: m.InterviewView })),
);
const OrganismView = lazy(() =>
  import("./views/OrganismView").then((m) => ({ default: m.OrganismView })),
);
const ProviderAccuracyView = lazy(() =>
  import("./views/ProviderAccuracyView").then((m) => ({ default: m.ProviderAccuracyView })),
);
const VisionView = lazy(() =>
  import("./views/VisionView").then((m) => ({ default: m.VisionView })),
);
const LifeEventsView = lazy(() =>
  import("./views/LifeEventsView").then((m) => ({ default: m.LifeEventsView })),
);
const MarriageLicensesView = lazy(() =>
  import("./views/MarriageLicensesView").then((m) => ({ default: m.MarriageLicensesView })),
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

type NavItem = { id: string; label: string; path: string };

/** Daily operator destinations — keep this short. */
const NAV_PRIMARY: NavItem[] = [
  { id: "today", label: "Today", path: "/today" },
  { id: "work", label: "Work", path: "/work?filter=action" },
  { id: "licenses", label: "Licenses", path: "/licenses" },
  { id: "congratulate", label: "Congratulate", path: "/congratulate" },
  { id: "sources", label: "Sources", path: "/sources" },
  { id: "map", label: "Map", path: "/map" },
];

/** Everything else lives under More, grouped by job. */
const NAV_MORE: { group: string; items: NavItem[] }[] = [
  {
    group: "Explore",
    items: [
      { id: "feed", label: "Feed", path: "/feed" },
      { id: "life-events", label: "Life Events", path: "/life-events" },
      { id: "graph", label: "Graph", path: "/graph" },
      { id: "case", label: "Cases", path: "/case" },
      { id: "interview", label: "Interview", path: "/interview" },
      { id: "search", label: "Search", path: "/search" },
    ],
  },
  {
    group: "Insights",
    items: [
      { id: "organism", label: "Organism", path: "/organism" },
      { id: "vision", label: "Vision", path: "/vision" },
      { id: "funnel", label: "Funnel", path: "/funnel" },
      { id: "cost", label: "Budget", path: "/cost" },
      { id: "provider-accuracy", label: "Provider Accuracy", path: "/provider-accuracy" },
    ],
  },
  {
    group: "System",
    items: [
      { id: "ops", label: "Ops", path: "/ops" },
      { id: "runs", label: "Runs", path: "/runs" },
      { id: "audit", label: "Audit", path: "/audit" },
      { id: "dlq", label: "DLQ", path: "/dlq" },
      { id: "jobs", label: "Jobs", path: "/jobs" },
      { id: "settings", label: "Settings", path: "/settings" },
    ],
  },
];

const NAV_MORE_FLAT = NAV_MORE.flatMap((g) => g.items);
const WORK_ALIASES = new Set(["work", "prospects", "queue"]);

function isNavActive(tab: string, id: string): boolean {
  if (id === "work") return WORK_ALIASES.has(tab);
  if (id === "congratulate") return tab === "congratulate" || tab === "kits";
  return tab === id;
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
  const pending = ops?.pending_actions ?? 0;

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

  const budgetTitle = budget
    ? `${used}/${budget} results today (${pct}%)`
    : `${used} results today`;

  return (
    <div className="watch-transport" title={`Radar control — ${budgetTitle}`}>
      <span className={`watch-transport__pill watch-transport__pill--${tone}`}>
        <span className="watch-transport__dot" aria-hidden />
        <span className="watch-transport__label">{label}</span>
        <span className="watch-transport__meta" aria-label={budgetTitle}>
          {pct}%
        </span>
      </span>
      {pending > 0 && (
        <button
          type="button"
          className="watch-transport__queue"
          onClick={() => setHash("/work?filter=action")}
          title={`${pending} pending approvals`}
        >
          {pending}
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
          <span className="watch-transport__icon" aria-hidden>
            <svg viewBox="0 0 10 10" fill="currentColor"><path d="M2 1.5v7l6.5-3.5L2 1.5z" /></svg>
          </span>
        ) : (
          <span className="watch-transport__icon" aria-hidden>
            <svg viewBox="0 0 10 10" fill="currentColor"><rect x="2" y="1.5" width="2.2" height="7" rx="0.4" /><rect x="5.8" y="1.5" width="2.2" height="7" rx="0.4" /></svg>
          </span>
        )}
        <span className="watch-transport__btn-text">{paused ? "Play" : "Pause"}</span>
      </button>
    </div>
  );
}

function MoreNav({
  activeTab,
  onNavigate,
}: {
  activeTab: string;
  onNavigate: (path: string) => void;
}) {
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement | null>(null);
  const moreActive = NAV_MORE_FLAT.some((t) => isNavActive(activeTab, t.id));
  const activeLabel = NAV_MORE_FLAT.find((t) => isNavActive(activeTab, t.id))?.label;

  useEffect(() => {
    if (!open) return;
    const onDown = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false);
    };
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") setOpen(false);
    };
    document.addEventListener("mousedown", onDown);
    document.addEventListener("keydown", onKey);
    return () => {
      document.removeEventListener("mousedown", onDown);
      document.removeEventListener("keydown", onKey);
    };
  }, [open]);

  return (
    <div className={`app-more ${open ? "app-more--open" : ""} ${moreActive ? "app-more--active" : ""}`} ref={ref}>
      <button
        type="button"
        className={`app-nav__tab app-more__trigger ${moreActive ? "app-nav__tab--active" : ""}`}
        aria-expanded={open}
        aria-haspopup="menu"
        data-testid="nav-more"
        onClick={() => setOpen((v) => !v)}
      >
        {moreActive && activeLabel ? activeLabel : "More"}
        <svg className="app-more__chevron" viewBox="0 0 12 12" width="12" height="12" aria-hidden>
          <path d="M3 4.5 L6 7.5 L9 4.5" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
        </svg>
      </button>
      {open && (
        <div className="app-more__menu" role="menu">
          {NAV_MORE.map((group) => (
            <div key={group.group} className="app-more__group">
              <div className="app-more__group-label">{group.group}</div>
              {group.items.map((item) => {
                const active = isNavActive(activeTab, item.id);
                return (
                  <button
                    key={item.id}
                    type="button"
                    role="menuitem"
                    className={`app-more__item ${active ? "app-more__item--active" : ""}`}
                    data-testid={`nav-${item.id}`}
                    aria-current={active ? "page" : undefined}
                    onClick={() => {
                      setOpen(false);
                      onNavigate(item.path);
                    }}
                  >
                    {item.label}
                  </button>
                );
              })}
            </div>
          ))}
        </div>
      )}
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
      case "licenses":
        return <MarriageLicensesView />;
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
      case "life-events":
        return <LifeEventsView />;
      case "graph":
        return <CoupleGraphView />;
      case "case":
        return <CaseDetailView />;
      case "organism":
        return <OrganismView onNavigate={navigate} />;
      case "funnel":
        return <FunnelView />;
      case "cost":
        return <CostView />;
      case "provider-accuracy":
        return <ProviderAccuracyView />;
      case "vision":
        return <VisionView />;
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
            <div className="app-header__title">Neptune</div>
          </div>
          <div className="app-header__center">
            <SearchBar inputRef={searchRef} />
          </div>
          <div className="app-header__right">
            <WatchTransport />
            <span className="app-header__divider" aria-hidden />
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
              {dark ? (
                <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" aria-hidden>
                  <circle cx="8" cy="8" r="3.2" />
                  <path strokeLinecap="round" d="M8 1.5v1.6M8 12.9v1.6M1.5 8h1.6M12.9 8h1.6M3.4 3.4l1.1 1.1M11.5 11.5l1.1 1.1M12.6 3.4l-1.1 1.1M4.5 11.5l-1.1 1.1" />
                </svg>
              ) : (
                <svg viewBox="0 0 16 16" fill="currentColor" aria-hidden>
                  <path d="M12.8 10.2A5.6 5.6 0 0 1 5.8 3.2 5.8 5.8 0 1 0 12.8 10.2Z" />
                </svg>
              )}
            </button>
          </div>
        </div>
        <nav className="app-nav" aria-label="Main navigation" data-testid="app-nav">
          <div className="app-nav__primary">
            {NAV_PRIMARY.map((t) => {
              const isActive = isNavActive(route.tab, t.id);
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
          </div>
          <MoreNav activeTab={route.tab} onNavigate={navigate} />
        </nav>
      </header>
      <main
        id="app-main"
        className={`app-main ${
          route.tab === "work" ||
          route.tab === "sources" ||
          route.tab === "map" ||
          route.tab === "congratulate" ||
          route.tab === "kits"
            ? "app-main--wide"
            : ""
        }`}
        role="main"
        data-testid="app-main"
      >
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
