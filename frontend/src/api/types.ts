export type LiveEventType = "couple_detected" | "stage_transition" | "action_created" | "alert";

export interface LiveEvent {
  type: LiveEventType;
  data: Record<string, unknown>;
  time: string;
}

export interface Signal {
  id: string;
  observation_type: string;
  monitor: string;
  handle: string;
  summary: string;
  observed_at: string;
  consent_scope: string;
}

export interface CoupleSummary {
  id: string;
  person_a_label: string;
  person_b_label: string;
}

export interface GraphNode {
  id: string;
  type: "person" | "account";
  label: string;
}

export interface GraphEdge {
  from: string;
  to: string;
  kind: string;
  active: boolean;
}

export interface CoupleGraph {
  nodes: GraphNode[];
  edges: GraphEdge[];
}

export type RelationshipStage =
  | "unknown"
  | "dating_suspected"
  | "engaged"
  | "married"
  | "status_uncertain"
  | "ended_suspected";

export interface Relationship {
  id: string;
  couple_id: string;
  stage: RelationshipStage;
  confidence: number;
  effective_from: string;
  effective_to?: string;
  automation_paused: boolean;
  visibility_scope: string;
}

export interface RelationshipResponse {
  current: Relationship;
  history: Relationship[];
}

export interface Evidence {
  id: string;
  hypothesis_id: string;
  observation_id?: string;
  kind: string;
  description: string;
  weight: number;
  confirmed: boolean;
  created_at: string;
}

export interface ConfidenceComponent {
  kind: string;
  weight: number;
  description: string;
}

export interface ConfidenceBreakdown {
  final: number;
  components: ConfidenceComponent[];
}

export interface NeptuneCase {
  id: string;
  couple_id: string;
  lead_id: string;
  case_type: string;
  status: string;
  created_at: string;
  updated_at: string;
}

export interface CRMLead {
  id: string;
  person_id: string;
  hypothesis_id?: string;
  lead_type: string;
  status: string;
  created_at: string;
}

export type ActionType =
  | "review"
  | "ignore"
  | "draft_outreach"
  | "pause_automation"
  | "create_case"
  | "concierge_review"
  | "investigate"
  | "no_action";

export interface RecommendedAction {
  id: string;
  hypothesis_id: string;
  case_id?: string;
  action_type: ActionType;
  proposed_payload: string;
  status: "pending" | "approved" | "ignored" | "executed" | "failed";
  created_at: string;
  decided_at?: string;
  decided_by?: string;
}

export interface ActionPayload {
  internal_note: string;
  customer_facing: string;
  reasons: string[] | null;
}

export interface AuditEvent {
  id: string;
  entity_type: string;
  entity_id: string;
  event: string;
  detail?: string;
  monitor?: string;
  step_index: number;
  created_at: string;
}

export interface WatchedSource {
  id: string;
  handle: string;
  source_class: string;
  active: boolean;
  state?: string;
  city?: string;
  // Only ever populated by a real, successful Apify profile check — absent
  // until then, never a placeholder.
  follower_count?: number;
  following_count?: number;
  post_count?: number;
  full_name?: string;
  profile_pic_url?: string;
  verified?: boolean;
  profile_checked_at?: string;
  created_at: string;
  posts_stored?: number;
  last_post_at?: string;
  stale?: boolean;
  scan_mode?: "find_couples" | "monitor_only";
  last_scanned_at?: string;
  last_scan_couples?: number;
  last_scan_actions?: number;
}

export interface ScanJob {
  id: string;
  kind: "single" | "bulk";
  handle?: string;
  status: "queued" | "running" | "done" | "failed";
  step: string;
  progress: number;
  message?: string;
  result?: SourceScanResult;
  results?: SourceScanResult[];
  error?: string;
  created_at: string;
  updated_at: string;
}

export interface ScannedCouple {
  couple_id?: string;
  handle_a: string;
  handle_b: string;
  tags: string[];
  vendor_tags?: string[];
  post_url?: string;
  caption?: string;
  image_url?: string;
  action_id?: string;
  action_type?: string;
  confidence?: number;
  hypothesis_id?: string;
  quality?: number;
  quality_label?: "strong_couple" | "likely_couple" | "weak" | "vendor_noise";
  has_people_shot?: boolean;
}

export interface ScannedApproval {
  action_id: string;
  action_type?: string;
  couple_id?: string;
  handle_a?: string;
  handle_b?: string;
  confidence?: number;
}

export interface SourceScanResult {
  handle: string;
  posts_fetched: number;
  posts_processed: number;
  duplicates: number;
  tagged_posts: number;
  actions_created: number;
  city?: string;
  state?: string;
  full_name?: string;
  profile_pic_url?: string;
  follower_count?: number;
  couples: ScannedCouple[];
  pending_approvals: ScannedApproval[];
  errors?: string[];
  duration_ms: number;
}

export type KitStatus =
  | "draft"
  | "ready_review"
  | "address_verified"
  | "ready_to_mail"
  | "mailed"
  | "cancelled";

export interface ResearchStep {
  id: string;
  label: string;
  detail: string;
  status: "done" | "suggested" | "blocked";
  url?: string;
}

export interface AddressCandidate {
  line1?: string;
  line2?: string;
  city?: string;
  region?: string;
  postal?: string;
  country?: string;
  confidence: number;
  source: string;
  note?: string;
}

export interface CongratulateKit {
  id: string;
  couple_id: string;
  status: KitStatus;
  handle_a?: string;
  handle_b?: string;
  person_a_name?: string;
  person_b_name?: string;
  first_name_a?: string;
  last_name_a?: string;
  first_name_b?: string;
  last_name_b?: string;
  name_source_a?: string;
  name_source_b?: string;
  bio_a?: string;
  bio_b?: string;
  profile_pic_a?: string;
  profile_pic_b?: string;
  market_city?: string;
  market_region?: string;
  market_source?: string;
  source_handle?: string;
  source_class?: string;
  discovery_caption?: string;
  discovery_image_url?: string;
  discovery_post_url?: string;
  evidence?: string[];
  research_notes?: string;
  research_steps?: ResearchStep[];
  address_line1?: string;
  address_line2?: string;
  address_city?: string;
  address_region?: string;
  address_postal?: string;
  address_country?: string;
  address_confidence: number;
  address_source?: string;
  address_candidates?: AddressCandidate[];
  headline?: string;
  body_message?: string;
  internal_note?: string;
  postcard_html?: string;
  mail_payload?: Record<string, unknown>;
  verified_by?: string;
  verified_at?: string;
  mailed_at?: string;
  created_at: string;
  updated_at: string;
}

// --- Ohio source registry (real government/church/social connectors) -------

export type ConnectorStatus = "setup" | "healthy" | "degraded" | "offline";

export interface SourceOrganization {
  id: string;
  org_type: "government_office" | "diocese" | "parish" | "business";
  name: string;
  city_id?: string;
  county_id?: string;
  official_url?: string;
  provenance: string;
  data_mode: string;
  metadata?: string; // JSON string: org_type-specific real facts (phone, coverage dates, ...)
  created_at: string;
}

export interface SourceEndpoint {
  id: string;
  organization_id: string;
  endpoint_type: "marriage_record_search" | "bulletin_archive" | "parish_directory" | "social_profile";
  url: string;
  access_method: string;
  is_official: boolean;
  data_mode: string;
  created_at: string;
}

export interface Connector {
  id: string;
  source_endpoint_id: string;
  connector_type: string;
  provider: string;
  status: ConnectorStatus;
  schedule?: string;
  last_checked_at?: string;
  last_success_at?: string;
  last_failure_at?: string;
  error_message?: string;
  created_at: string;
}

export interface RegistryCounty {
  id: string;
  state_id: string;
  name: string;
}

export interface RegistryCity {
  id: string;
  state_id: string;
  primary_county_id?: string;
  name: string;
  lat?: number;
  lng?: number;
}

export interface CountyGovernmentView {
  county: RegistryCounty;
  organization?: SourceOrganization;
  endpoint?: SourceEndpoint;
  connector?: Connector;
}

export interface ChurchJurisdiction {
  id: string;
  source_organization_id: string;
  jurisdiction_type: string;
  hub_city_id?: string;
}

export interface Parish {
  id: string;
  source_organization_id: string;
  jurisdiction_id: string;
  bulletin_endpoint_id?: string;
}

export interface ParishView {
  organization: SourceOrganization;
  parish: Parish;
}

export interface DioceseView {
  organization: SourceOrganization;
  jurisdiction: ChurchJurisdiction;
  directory_endpoint?: SourceEndpoint;
  directory_connector?: Connector;
  parishes: ParishView[];
}

export interface SocialSource {
  id: string;
  source_organization_id: string;
  platform: string;
  category: string;
  city_market_id?: string;
  manually_verified: boolean;
  watched_source_id?: string;
}

export interface VendorView {
  organization: SourceOrganization;
  social_source: SocialSource;
  watched_source?: WatchedSource;
  connector?: Connector;
}

export interface SocialMarketView {
  city: RegistryCity;
  vendors: VendorView[];
}

export interface OverviewCounts {
  government: number;
  church: number;
  social: number;
  healthy: number;
  degraded: number;
  setup: number;
  offline: number;
}

export interface OverviewCityView {
  city: RegistryCity;
  counts: OverviewCounts;
}

export interface IngestCursor {
  monitor: string;
  last_seen_at?: string;
  last_run_at?: string;
  updated_at: string;
}

export interface IngestStatus {
  provider: string;
  provider_available: boolean;
  paused: boolean;
  running: boolean;
  poll_interval?: string;
  daily_budget?: number;
  results_used_today: number;
  cursors: IngestCursor[] | null;
}

export interface SourcePost {
  id: string;
  monitor: string;
  handle: string;
  caption?: string;
  url?: string;
  image_url?: string;
  tags?: string[];
  mentions?: string[];
  location?: string;
  observed_at: string;
}

export type ProspectColumnId =
  | "tagged_pair"
  | "investigating"
  | "engaged_signal"
  | "ready_outreach"
  | "approved_paused";

export interface ICPFit {
  score: number;
  tags: string[];
  labels: string[];
  employers?: string[];
  roles?: string[];
  market_priority: number;
  market_label: string;
}

export interface WeddingRunway {
  date?: string;
  raw?: string;
  source?: string;
  days_until?: number;
  band: "green" | "amber" | "red" | "past" | "unknown" | string;
  factor: number;
  suppress_outreach?: boolean;
  confidence?: number;
}

export type ConciergeQueue =
  | "congratulate"
  | "detective"
  | "runway_urgent"
  | "risk"
  | "action"
  | "other";

export interface ProspectCard {
  couple_id: string;
  column: ProspectColumnId;
  person_a_label: string;
  person_b_label: string;
  handle_a?: string;
  handle_b?: string;
  profile_pic_a?: string;
  profile_pic_b?: string;
  bio_a?: string;
  bio_b?: string;
  stage?: string;
  confidence?: number;
  hypothesis_score?: number;
  pending_action_id?: string;
  pending_action_type?: string;
  proposed_payload?: string;
  city?: string;
  region?: string;
  automation_paused: boolean;
  has_case: boolean;
  needs_pics?: boolean;
  needs_location?: boolean;
  needs_action?: boolean;
  created_at: string;
  icp?: ICPFit;
  runway?: WeddingRunway;
  runway_label?: string;
  neptune_rank?: number;
  journey_stage?: string;
  queue?: ConciergeQueue;
}

export interface OpsSummary {
  couples_total: number;
  couples_24h: number;
  pending_actions: number;
  needs_pics: number;
  needs_location: number;
  sources_total: number;
  sources_with_loc: number;
  sources_stale: number;
  map_pins: number;
  results_used_today: number;
  daily_budget?: number;
  paused?: boolean;
  provider_available?: boolean;
  running?: boolean;
  poll_interval?: string;
  queue_congratulate?: number;
  queue_detective?: number;
  queue_runway_urgent?: number;
  queue_risk?: number;
  kits_ready_to_mail?: number;
  kits_mailed?: number;
  funnel_chat_started_7d?: number;
  funnel_consult_booked_7d?: number;
  funnel_closed_won_7d?: number;
  funnel_handoffs_issued?: number;
  funnel_chat_rate?: number;
  funnel_book_rate?: number;
}

export interface FunnelEvent {
  id: string;
  couple_id?: string;
  event_type: string;
  external_id?: string;
  handoff_code?: string;
  source: string;
  payload_json?: string;
  journey_stage_before?: string;
  journey_stage_after?: string;
  matched_by?: string;
  occurred_at: string;
  created_at: string;
}

export interface FunnelStats {
  chat_started_7d: number;
  consult_booked_7d: number;
  closed_won_7d: number;
  closed_lost_7d: number;
  handoffs_issued: number;
  chat_rate: number;
  book_rate: number;
}

export interface AutopsyCase {
  couple_id?: string;
  kind: string;
  reason: string;
  handles?: string[];
  labels?: string[];
  score?: number;
  action_type?: string;
  lesson: string;
  occurred_at?: string;
  hypothesis_id?: string;
}

export interface AutopsySummary {
  period_start: string;
  period_end: string;
  suppressed_couples: number;
  ignored_actions: number;
  rejected_hypotheses: number;
  approved_actions: number;
  pending_actions: number;
  human_reject_rate: number;
  by_reason: Record<string, number>;
  by_action_type_ignored: Record<string, number>;
  top_failure_modes: string[];
  funnel: FunnelStats;
  notes: string[];
}

export interface AutopsyReport {
  id: string;
  period_start: string;
  period_end: string;
  summary: AutopsySummary;
  cases: AutopsyCase[];
  generated_by: string;
  created_at: string;
}

export interface DossierEvidence {
  id: string;
  kind: string;
  description: string;
  weight: number;
  points: number;
  confirmed: boolean;
  created_at?: string;
}

export interface DossierIdentity {
  kind: string;
  description: string;
  strength: string;
}

export interface JourneyStep {
  id: string;
  label: string;
  description: string;
  status: "done" | "current" | "upcoming" | "blocked" | string;
}

export interface BrandAction {
  id: string;
  title: string;
  body: string;
  tone: "celebrate" | "soft_invite" | "internal" | "risk" | string;
  allowed: boolean;
  block_reason?: string;
}

export interface DossierPost {
  id: string;
  caption?: string;
  image_url?: string;
  post_url?: string;
  location?: string;
  source_handle?: string;
  tags?: string[];
  observed_at?: string;
}

export interface CoupleDossier {
  couple_id: string;
  handle_a?: string;
  handle_b?: string;
  person_a_name?: string;
  person_b_name?: string;
  bio_a?: string;
  bio_b?: string;
  profile_pic_a?: string;
  profile_pic_b?: string;
  city?: string;
  region?: string;
  stage?: string;
  automation_paused?: boolean;
  has_case?: boolean;
  engagement_score: number;
  partner_score?: number;
  hypothesis_score: number;
  icp: ICPFit;
  runway: WeddingRunway;
  runway_label?: string;
  neptune_rank: number;
  deliverability: number;
  evidence: DossierEvidence[];
  identity: DossierIdentity[];
  pending_action_id?: string;
  pending_action_type?: string;
  proposed_payload?: string;
  hypothesis_id?: string;
  journey_stage: string;
  journey_steps: JourneyStep[];
  handoff_code?: string;
  handoff_url?: string;
  handoff_utm?: string;
  brand_actions: BrandAction[];
  celebrate_copy: string;
  soft_invite_copy: string;
  observations: DossierPost[];
  latest_kit_id?: string;
  latest_kit_status?: string;
  audit_trail: AuditEvent[];
  why_now: string[];
}

export interface ProspectBoard {
  columns: { id: ProspectColumnId; label: string }[];
  cards: Record<ProspectColumnId, ProspectCard[]>;
  total: number;
}

export interface ProspectPin {
  couple_id: string;
  person_a_label: string;
  person_b_label: string;
  handle_a?: string;
  handle_b?: string;
  profile_pic_a?: string;
  profile_pic_b?: string;
  city: string;
  region?: string;
  lat?: number;
  lng?: number;
  stage?: string;
  column?: ProspectColumnId;
}

export interface StateCoverage {
  state_id: string;
  name: string;
  county_count: number;
  counties_configured: number;
  cities: number;
  government_sources: number;
  church_sources: number;
  social_sources: number;
  watched_sources: number;
  alive_score: number;
}

// --- Universal search / DLQ / admin ----------------------------------------

// Reuses the existing CoupleSummary, CRMLead, and NeptuneCase types rather
// than inventing parallel *Summary aliases — the search endpoint returns the
// same shapes the list endpoints already produce.
export interface SearchResult {
  couples: CoupleSummary[];
  leads: CRMLead[];
  cases: NeptuneCase[];
}

export interface DLQItem {
  id: string;
  source: string;
  monitor?: string;
  external_id?: string;
  raw_payload?: string;
  error_message: string;
  retries: number;
  last_retry_at?: string;
  status: string;
  created_at: string;
}

export interface UserSummary {
  id: string;
  email: string;
  role: string;
  last_seen_at?: string;
  api_key_masked?: string;
  disabled?: boolean;
}

// Pipeline run ledger — one summary row per orchestrator ProcessEvent call.
// Per-stage detail (audit events, timings) is joinable by observation_id.
export interface PipelineRun {
  id: string;
  observation_id: string;
  agent_name: string;
  model?: string;
  prompt_tokens: number;
  completion_tokens: number;
  cost_usd: number;
  confidence?: number;
  stop_reason: string;
  hypothesis_id?: string;
  action_id?: string;
  couple_id?: string;
  monitor?: string;
  started_at: string;
  ended_at?: string;
  created_at: string;
}

export interface StageTiming {
  stage: string;
  duration_ms: number;
  event_id?: string;
  timestamp: string;
}

export interface PipelineRunDetail extends PipelineRun {
  timings: StageTiming[];
  events: AuditEvent[];
}
