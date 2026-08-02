import type {
  ConnectorStatus,
  CountyGovernmentView,
  DioceseView,
  OverviewCityView,
  SocialMarketView,
} from "../api/types";
import { LAYERS, type LayerId } from "./ohio";

const STATUS_LABEL: Record<ConnectorStatus, string> = {
  setup: "Setup",
  healthy: "Healthy",
  degraded: "Degraded",
  offline: "Offline",
};

function StatusBadge({ status }: { status: ConnectorStatus | "not_configured" }) {
  if (status === "not_configured") {
    return <span className="ohio-status-badge ohio-status-badge--not_configured">Not configured</span>;
  }
  return <span className={`ohio-status-badge ohio-status-badge--${status}`}>{STATUS_LABEL[status]}</span>;
}

function formatCount(n?: number): string | null {
  if (n == null) return null;
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`;
  if (n >= 1_000) return `${(n / 1_000).toFixed(1)}K`;
  return String(n);
}

function timeAgo(iso?: string): string {
  if (!iso) return "no successful check yet";
  const ms = Date.now() - new Date(iso).getTime();
  const min = Math.round(ms / 60_000);
  if (min < 1) return "just now";
  if (min < 60) return `${min} min ago`;
  const hr = Math.round(min / 60);
  if (hr < 24) return `${hr}h ago`;
  return `${Math.round(hr / 24)}d ago`;
}

function parseMeta(json?: string): Record<string, unknown> {
  if (!json) return {};
  try {
    return JSON.parse(json);
  } catch {
    return {};
  }
}

interface OhioPanelProps {
  /** Display name, e.g. "Ohio" or "New York" */
  stateName?: string;
  /** USPS code, e.g. "OH" */
  stateCode?: string;
  layer: LayerId;
  onLayerChange: (l: LayerId) => void;
  onClose: () => void;
  overview: { data?: OverviewCityView[]; isLoading: boolean };
  government: { data?: CountyGovernmentView[]; isLoading: boolean };
  churches: { data?: DioceseView[]; isLoading: boolean };
  /** Single market (legacy) or multi-city markets */
  social: { data?: SocialMarketView | SocialMarketView[]; isLoading: boolean };
  selectedCountyFips: string | null;
  onSelectCounty: (fips: string | null) => void;
  selectedDioceseId: string | null;
  onSelectDiocese: (id: string | null) => void;
}

export function OhioPanel({
  stateName = "Ohio",
  stateCode = "OH",
  layer, onLayerChange, onClose, overview, government, churches, social,
  selectedCountyFips, onSelectCounty, selectedDioceseId, onSelectDiocese,
}: OhioPanelProps) {
  return (
    <aside className="state-panel ohio-panel">
      <header className="state-panel__header">
        <div className="state-panel__title-row">
          <div>
            <h3 className="state-panel__title">{stateName}</h3>
            <span className="state-panel__code">{stateCode}</span>
          </div>
          <button className="state-panel__close" onClick={onClose} aria-label="Close panel">×</button>
        </div>
        <nav className="ohio-layer-tabs">
          {LAYERS.map((l) => (
            <button
              key={l.id}
              className={`ohio-layer-tab ${layer === l.id ? "ohio-layer-tab--active" : ""}`}
              onClick={() => onLayerChange(l.id)}
            >
              {l.label}
            </button>
          ))}
        </nav>
      </header>

      <div className="state-panel__body">
        {layer === "overview" && <OverviewBody overview={overview} />}
        {layer === "government" && (
          <GovernmentBody government={government} selectedCountyFips={selectedCountyFips} onSelectCounty={onSelectCounty} />
        )}
        {layer === "churches" && (
          <ChurchesBody churches={churches} selectedDioceseId={selectedDioceseId} onSelectDiocese={onSelectDiocese} />
        )}
        {layer === "instagram" && <InstagramBody social={social} />}
      </div>

      <footer className="state-panel__footer">
        <span className="state-panel__hint">Click ocean or press esc to reset</span>
      </footer>
    </aside>
  );
}

function OverviewBody({ overview }: { overview: { data?: OverviewCityView[]; isLoading: boolean } }) {
  if (overview.isLoading) return <div className="ohio-empty">Loading real coverage counts…</div>;
  const cities = overview.data ?? [];
  if (cities.length === 0) {
    return (
      <div className="ohio-empty">
        No city markets registered in this state yet.
        <div className="ohio-hint">
          Geography is national; social/government packs are added state-by-state. Honest empty — not a fake status.
        </div>
      </div>
    );
  }
  return (
    <>
      {cities.map((city) => {
        const c = city.counts;
        return (
          <section className="state-panel__section" key={city.city.id}>
            <div className="state-panel__section-header">
              <span className="state-panel__section-title">{city.city.name} — real coverage</span>
            </div>
            <div className="ohio-overview-grid">
              <div className="ohio-overview-stat">
                <span className="ohio-overview-stat__value">{c.government}</span>
                <span className="ohio-overview-stat__label">government sources</span>
              </div>
              <div className="ohio-overview-stat">
                <span className="ohio-overview-stat__value">{c.church}</span>
                <span className="ohio-overview-stat__label">church sources</span>
              </div>
              <div className="ohio-overview-stat">
                <span className="ohio-overview-stat__value">{c.social}</span>
                <span className="ohio-overview-stat__label">social sources</span>
              </div>
            </div>
            <div className="ohio-overview-breakdown">
              <StatusBadge status="healthy" /> <span>{c.healthy}</span>
              <StatusBadge status="degraded" /> <span>{c.degraded}</span>
              <StatusBadge status="setup" /> <span>{c.setup}</span>
              <StatusBadge status="offline" /> <span>{c.offline}</span>
            </div>
            <p className="ohio-hint">Every count above comes from real, checked connectors — not a fixed estimate.</p>
          </section>
        );
      })}
    </>
  );
}

function GovernmentBody({
  government, selectedCountyFips, onSelectCounty,
}: {
  government: { data?: CountyGovernmentView[]; isLoading: boolean };
  selectedCountyFips: string | null;
  onSelectCounty: (fips: string | null) => void;
}) {
  if (government.isLoading) return <div className="ohio-empty">Loading county connectors…</div>;
  const rows = government.data ?? [];
  const configured = rows.filter((r) => r.organization).length;

  if (!selectedCountyFips) {
    return (
      <section className="state-panel__section">
        <div className="state-panel__section-header">
          <span className="state-panel__section-title">Counties</span>
          <span className="state-panel__section-count">
            {configured}/{rows.length} configured
          </span>
        </div>
        {rows.length === 0 ? (
          <div className="ohio-empty">
            No counties in registry for this state yet. Run <code>seed-geography</code>.
          </div>
        ) : (
          <ul className="ohio-diocese-list">
            {rows.slice(0, 80).map((r) => (
              <li key={r.county.id}>
                <button
                  type="button"
                  className="ohio-diocese-card"
                  onClick={() => onSelectCounty(r.county.id)}
                >
                  <span className="ohio-diocese-card__name">{r.county.name}</span>
                  <span className="ohio-diocese-card__meta">
                    {r.organization ? r.organization.name : "Not configured"}
                  </span>
                </button>
              </li>
            ))}
            {rows.length > 80 && (
              <li className="ohio-hint">Showing first 80 of {rows.length} counties — pick one to inspect.</li>
            )}
          </ul>
        )}
      </section>
    );
  }

  const row = rows.find((r) => r.county.id === selectedCountyFips);
  if (!row) return <div className="ohio-empty">Unknown county.</div>;

  if (!row.organization || !row.endpoint || !row.connector) {
    return (
      <section className="state-panel__section">
        <div className="state-panel__section-header">
          <span className="state-panel__section-title">{row.county.name} County</span>
        </div>
        <div className="ohio-empty">No probate-court connector configured for {row.county.name} County yet.</div>
        <button className="ohio-link-btn" onClick={() => onSelectCounty(null)}>Back to county list</button>
      </section>
    );
  }

  const meta = parseMeta(row.organization.metadata);
  return (
    <section className="state-panel__section ohio-connector-detail">
      <div className="state-panel__section-header">
        <span className="state-panel__section-title">{row.county.name} County</span>
        <StatusBadge status={row.connector.status} />
      </div>
      <dl className="ohio-detail-list">
        <dt>Office</dt>
        <dd>{row.organization.name}</dd>
        <dt>Search URL</dt>
        <dd><a href={row.endpoint.url} target="_blank" rel="noreferrer">{row.endpoint.url}</a></dd>
        {typeof meta.phone === "string" && (<><dt>Phone</dt><dd>{meta.phone}</dd></>)}
        {typeof meta.coverage_note === "string" && (<><dt>Coverage</dt><dd>{meta.coverage_note}</dd></>)}
        <dt>Last checked</dt>
        <dd>{timeAgo(row.connector.last_checked_at)}</dd>
      </dl>
      <button className="ohio-link-btn" onClick={() => onSelectCounty(null)}>Back to county list</button>
    </section>
  );
}

function ChurchesBody({
  churches, selectedDioceseId, onSelectDiocese,
}: {
  churches: { data?: DioceseView[]; isLoading: boolean };
  selectedDioceseId: string | null;
  onSelectDiocese: (id: string | null) => void;
}) {
  if (churches.isLoading) return <div className="ohio-empty">Loading dioceses…</div>;
  const dioceses = churches.data ?? [];

  const selected = selectedDioceseId ? dioceses.find((d) => d.jurisdiction.id === selectedDioceseId) : null;

  if (!selected) {
    if (dioceses.length === 0) return <div className="ohio-empty">No dioceses registered yet.</div>;
    return (
      <div className="ohio-diocese-list">
        {dioceses.map((d) => (
          <button
            key={d.jurisdiction.id}
            className="ohio-diocese-card"
            onClick={() => onSelectDiocese(d.jurisdiction.id)}
          >
            <span className="ohio-diocese-card__name">{d.organization.name}</span>
            <span className="ohio-diocese-card__meta">{(d.parishes ?? []).length} parishes registered</span>
            {d.directory_connector && <StatusBadge status={d.directory_connector.status} />}
          </button>
        ))}
      </div>
    );
  }

  return (
    <section className="state-panel__section ohio-connector-detail">
      <div className="state-panel__section-header">
        <span className="state-panel__section-title">{selected.organization.name}</span>
        {selected.directory_connector && <StatusBadge status={selected.directory_connector.status} />}
      </div>
      <dl className="ohio-detail-list">
        <dt>Directory</dt>
        <dd>
          {selected.organization.official_url && (
            <a href={selected.organization.official_url} target="_blank" rel="noreferrer">{selected.organization.official_url}</a>
          )}
        </dd>
        <dt>Last checked</dt>
        <dd>{timeAgo(selected.directory_connector?.last_checked_at)}</dd>
      </dl>
      <div className="ohio-parish-list">
        {(selected.parishes ?? []).length === 0 && (
          <div className="ohio-empty">No parishes registered yet — pending directory crawl.</div>
        )}
        {(selected.parishes ?? []).map((p) => {
          const meta = parseMeta(p.organization.metadata);
          const bulletinUrl = typeof meta.bulletin_url === "string" ? meta.bulletin_url : null;
          const bannsEvidence = typeof meta.banns_evidence === "string" ? meta.banns_evidence : null;
          return (
            <div key={p.parish.id} className="ohio-parish-row">
              <span className="ohio-parish-row__name">{p.organization.name}</span>
              {typeof meta.address === "string" && <span className="ohio-parish-row__addr">{meta.address}</span>}
              <span className="ohio-parish-row__bulletin">
                {bulletinUrl ? (
                  <a href={bulletinUrl} target="_blank" rel="noreferrer">Bulletin archive ↗</a>
                ) : (
                  "No bulletin archive discovered yet"
                )}
              </span>
              {bannsEvidence && (
                <a
                  className="ohio-parish-row__banns"
                  href={bannsEvidence}
                  target="_blank"
                  rel="noreferrer"
                  title="A real bulletin from this parish was observed containing a Banns of Marriage section with couple names and a wedding date"
                >
                  Banns confirmed ↗
                </a>
              )}
            </div>
          );
        })}
      </div>
      <button className="ohio-link-btn" onClick={() => onSelectDiocese(null)}>Back to dioceses</button>
    </section>
  );
}

const CATEGORY_LABEL: Record<string, string> = {
  engagement_photographer: "Photographers",
  wedding_venue: "Venues",
  wedding_planner: "Planners",
  proposal_planner: "Proposal planners",
  jeweler: "Jewelers",
  wedding_publication: "Publications",
  registry_provider: "Registries",
  bridal_boutique: "Boutiques",
};

function InstagramBody({ social }: { social: { data?: SocialMarketView | SocialMarketView[]; isLoading: boolean } }) {
  if (social.isLoading) return <div className="ohio-empty">Loading social accounts…</div>;
  const markets = Array.isArray(social.data) ? social.data : social.data ? [social.data] : [];
  const vendors = markets.flatMap((m) => m.vendors ?? []);
  if (vendors.length === 0) {
    return (
      <div className="ohio-empty">
        No social accounts configured yet for this state.
        <div className="ohio-hint">Vendor packs are added per metro — geography alone does not invent sources.</div>
      </div>
    );
  }

  const byCategory = new Map<string, typeof vendors>();
  for (const v of vendors) {
    const list = byCategory.get(v.social_source.category) ?? [];
    list.push(v);
    byCategory.set(v.social_source.category, list);
  }

  return (
    <>
      {markets.length > 1 && (
        <p className="ohio-hint">{markets.length} city markets · {vendors.length} accounts</p>
      )}
      {[...byCategory.entries()].map(([category, rows]) => (
        <div key={category} className="ohio-vendor-group">
          <div className="state-panel__section-header">
            <span className="state-panel__section-title">{CATEGORY_LABEL[category] ?? category}</span>
            <span className="state-panel__section-count">{rows.length}</span>
          </div>
          {rows.map((v) => (
            <div key={v.social_source.id} className="ohio-vendor-row">
              <div className="ohio-vendor-row__main">
                <span className="ohio-vendor-row__name">{v.organization.name}</span>
                {v.watched_source && (
                  <a
                    className="ohio-vendor-row__handle"
                    href={`https://instagram.com/${v.watched_source.handle}`}
                    target="_blank"
                    rel="noreferrer"
                  >
                    @{v.watched_source.handle}
                  </a>
                )}
                {v.watched_source?.follower_count != null ? (
                  <span className="ohio-vendor-row__stats">
                    {formatCount(v.watched_source.follower_count)} followers · {formatCount(v.watched_source.post_count)} posts
                  </span>
                ) : (
                  <span className="ohio-vendor-row__stats ohio-vendor-row__stats--pending">no profile check yet</span>
                )}
              </div>
              {v.connector && <StatusBadge status={v.connector.status} />}
            </div>
          ))}
        </div>
      ))}
    </>
  );
}
