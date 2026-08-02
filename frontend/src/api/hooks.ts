import { useEffect, useRef, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { api, getToken } from "./client";
import type {
  AuditEvent,
  ConfidenceBreakdown,
  CountyGovernmentView,
  CoupleGraph,
  CoupleSummary,
  CRMLead,
  DioceseView,
  Evidence,
  IngestStatus,
  NeptuneCase,
  OverviewCityView,
  ProspectBoard,
  ProspectPin,
  RecommendedAction,
  RelationshipResponse,
  Signal,
  SocialMarketView,
  OpsSummary,
  SourcePost,
  CongratulateKit,
  CoupleDossier,
  FunnelEvent,
  FunnelStats,
  AutopsyReport,
  StateCoverage,
  WatchedSource,
  SearchResult,
  DLQItem,
  UserSummary,
  LiveEvent,
} from "./types";

const keys = {
  signals: (monitor?: string) => ["signals", monitor] as const,
  coupleGraph: (coupleId?: string) => ["couple", coupleId, "graph"] as const,
  relationship: (coupleId?: string) => ["couple", coupleId, "relationship"] as const,
  dossier: (coupleId?: string) => ["couple", coupleId, "dossier"] as const,
  evidence: (hypId?: string) => ["hypothesis", hypId, "evidence"] as const,
  confidence: (hypId?: string) => ["hypothesis", hypId, "confidence"] as const,
  cases: ["cases"] as const,
  leads: ["leads"] as const,
  actions: (status?: string) => ["actions", status] as const,
  audit: (monitor?: string) => ["audit", monitor] as const,
  sources: ["sources"] as const,
  sourcePosts: (handle: string) => ["sources", handle, "posts"] as const,
  kits: ["kits"] as const,
  kit: (id: string) => ["kits", id] as const,
  coupleKit: (coupleId: string) => ["couples", coupleId, "kit"] as const,
  prospectBoard: ["prospects", "board"] as const,
  prospectPins: ["map", "prospects"] as const,
  ingestStatus: ["ingest", "status"] as const,
  mapCoverage: ["map", "coverage"] as const,
  stateOverview: (st: string) => ["map", "states", st, "overview"] as const,
  stateGovernment: (st: string) => ["map", "states", st, "government"] as const,
  stateChurches: (st: string) => ["map", "states", st, "churches"] as const,
  stateSocial: (st: string) => ["map", "states", st, "social"] as const,
  // legacy aliases
  ohioOverview: ["map", "states", "OH", "overview"] as const,
  ohioGovernment: ["map", "states", "OH", "government"] as const,
  ohioChurches: ["map", "states", "OH", "churches"] as const,
  ohioSocial: ["map", "states", "OH", "social"] as const,
  search: (q: string, type: string) => ["search", q, type] as const,
  dlq: (status: string, limit: number) => ["dlq", status, limit] as const,
  scanJobs: (limit: number, status: string) => ["scan-jobs", limit, status] as const,
  users: ["users"] as const,
};

function useInvalidateAll() {
  const qc = useQueryClient();
  return () => {
    qc.invalidateQueries({ queryKey: ["signals"] });
    qc.invalidateQueries({ queryKey: ["couples"] });
    qc.invalidateQueries({ queryKey: ["couple"] });
    qc.invalidateQueries({ queryKey: ["hypothesis"] });
    qc.invalidateQueries({ queryKey: ["cases"] });
    qc.invalidateQueries({ queryKey: ["leads"] });
    qc.invalidateQueries({ queryKey: ["actions"] });
    qc.invalidateQueries({ queryKey: ["audit"] });
    qc.invalidateQueries({ queryKey: ["sources"] });
    qc.invalidateQueries({ queryKey: ["ingest"] });
    qc.invalidateQueries({ queryKey: ["map"] });
    qc.invalidateQueries({ queryKey: ["ops"] });
    qc.invalidateQueries({ queryKey: ["prospects"] });
  };
}

// Scoped invalidation: each mutation refetches only the queries its side
// effects can actually change. The blanket invalidate-all above made one
// Approve click fire 13 HTTP refetches, including the Ohio map queries.
function useInvalidateKeys(...prefixes: string[]) {
  const qc = useQueryClient();
  return () => {
    for (const p of prefixes) {
      qc.invalidateQueries({ queryKey: [p] });
    }
  };
}

// When the global watch loop is paused, stop auto-refetch on live views so the
// console freezes with the radar. Status itself keeps polling so play works.
function useLiveRefetch(ms: number): number | false {
  const { data } = useIngestStatus();
  if (data?.paused) return false;
  return ms;
}

// The live feed polls — the watch loop ingests continuously, so the console
// refreshes on an interval instead of being stepped manually.
export function useSignals(monitor?: string) {
  const interval = useLiveRefetch(10_000);
  return useQuery({
    queryKey: keys.signals(monitor),
    queryFn: () => api.get<Signal[]>(`/api/signals${monitor ? `?monitor=${encodeURIComponent(monitor)}` : ""}`),
    refetchInterval: interval,
  });
}

export function useCouples() {
  return useQuery({ queryKey: ["couples"], queryFn: () => api.get<CoupleSummary[]>("/api/couples") });
}

export function useCoupleGraph(coupleId?: string) {
  return useQuery({
    queryKey: keys.coupleGraph(coupleId),
    queryFn: () => api.get<CoupleGraph>(`/api/couples/${coupleId}/graph`),
    enabled: !!coupleId,
  });
}

export function useRelationship(coupleId?: string) {
  return useQuery({
    queryKey: keys.relationship(coupleId),
    queryFn: () => api.get<RelationshipResponse>(`/api/couples/${coupleId}/relationship`),
    enabled: !!coupleId,
  });
}

export function usePauseCouple() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (coupleId: string) => api.post(`/api/couples/${coupleId}/pause`),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["couple"] });
      qc.invalidateQueries({ queryKey: ["actions"] });
    },
  });
}

export function useResumeCouple() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (coupleId: string) => api.post(`/api/couples/${coupleId}/resume`),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["couple"] });
      qc.invalidateQueries({ queryKey: ["actions"] });
    },
  });
}

export function useEvidence(hypId?: string) {
  return useQuery({
    queryKey: keys.evidence(hypId),
    queryFn: () => api.get<Evidence[]>(`/api/hypotheses/${hypId}/evidence`),
    enabled: !!hypId,
  });
}

export function useConfidence(hypId?: string) {
  return useQuery({
    queryKey: keys.confidence(hypId),
    queryFn: () => api.get<ConfidenceBreakdown>(`/api/hypotheses/${hypId}/confidence`),
    enabled: !!hypId,
  });
}

export function useCases() {
  return useQuery({ queryKey: keys.cases, queryFn: () => api.get<NeptuneCase[]>("/api/cases") });
}

export function useLeads() {
  return useQuery({ queryKey: keys.leads, queryFn: () => api.get<CRMLead[]>("/api/leads") });
}

export function useActions(status?: string) {
  const interval = useLiveRefetch(10_000);
  return useQuery({
    queryKey: keys.actions(status),
    queryFn: () => api.get<RecommendedAction[]>(`/api/actions${status ? `?status=${status}` : ""}`),
    refetchInterval: interval,
  });
}

// Approve/ignore create leads + cases, confirm/reject the hypothesis, and
// can flip a couple's stage — but they never touch signals, sources, ingest
// status, or the Ohio map registry queries.
export function useApproveAction() {
  const invalidate = useInvalidateKeys("actions", "cases", "leads", "hypothesis", "couples", "couple", "prospects", "ops");
  return useMutation({
    mutationFn: (actionId: string) => api.post(`/api/actions/${actionId}/approve`),
    onSuccess: invalidate,
  });
}

export function useIgnoreAction() {
  const invalidate = useInvalidateKeys("actions", "cases", "leads", "hypothesis", "couples", "couple", "prospects", "ops");
  return useMutation({
    mutationFn: (actionId: string) => api.post(`/api/actions/${actionId}/ignore`),
    onSuccess: invalidate,
  });
}

export function useAudit(monitor?: string) {
  const interval = useLiveRefetch(15_000);
  return useQuery({
    queryKey: keys.audit(monitor),
    queryFn: () => api.get<AuditEvent[]>(`/api/audit${monitor ? `?monitor=${encodeURIComponent(monitor)}` : ""}`),
    refetchInterval: interval,
  });
}

export function useSources(activeOnly = true) {
  return useQuery({
    queryKey: ["sources", activeOnly ? "active" : "all"],
    queryFn: () => api.get<WatchedSource[]>(`/api/sources${activeOnly ? "" : "?active=false"}`),
  });
}

export function useSourcePosts(handle?: string) {
  return useQuery({
    queryKey: keys.sourcePosts(handle ?? ""),
    queryFn: () => api.get<SourcePost[]>(`/api/sources/${encodeURIComponent(handle!)}/posts`),
    enabled: !!handle,
  });
}

export function useProspectBoard() {
  const interval = useLiveRefetch(15_000);
  return useQuery({
    queryKey: keys.prospectBoard,
    queryFn: () => api.get<ProspectBoard>("/api/prospects/board"),
    refetchInterval: interval,
  });
}

export function useProspectPins(enabled = true) {
  return useQuery({
    queryKey: keys.prospectPins,
    queryFn: () => api.get<ProspectPin[]>("/api/map/prospects"),
    enabled,
    staleTime: 30_000,
  });
}

export function useOpsSummary() {
  return useQuery({
    queryKey: ["ops", "summary"],
    queryFn: () => api.get<OpsSummary>("/api/ops/summary"),
    refetchInterval: 12_000,
  });
}

export function useSuppressCouple() {
  const invalidate = useInvalidateAll();
  return useMutation({
    mutationFn: ({ id, reason }: { id: string; reason?: string }) =>
      api.post(`/api/couples/${id}/suppress`, { reason: reason ?? "not_a_couple" }),
    onSuccess: invalidate,
  });
}

export function useEnrichMissing() {
  const invalidate = useInvalidateAll();
  return useMutation({
    mutationFn: (limit: number) =>
      api.post<{ attempted: number; succeeded: number }>(`/api/prospects/enrich-missing?limit=${limit}`),
    onSuccess: invalidate,
  });
}

export function useBackfillLocations() {
  const invalidate = useInvalidateAll();
  return useMutation({
    mutationFn: (limit: number) =>
      api.post<{ checked: number; updated: number }>(`/api/prospects/backfill-locations?limit=${limit}`),
    onSuccess: invalidate,
  });
}

export function useIngestStatus() {
  return useQuery({
    queryKey: keys.ingestStatus,
    queryFn: () => api.get<IngestStatus>("/api/ingest/status"),
    // Always poll status so Pause/Play stays accurate even while paused.
    refetchInterval: 5_000,
  });
}

export function usePauseIngest() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () => api.post<{ paused: boolean; running: boolean }>("/api/ingest/pause"),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: keys.ingestStatus });
      qc.invalidateQueries({ queryKey: ["audit"] });
    },
  });
}

export function useResumeIngest() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () => api.post<{ paused: boolean; running: boolean }>("/api/ingest/resume"),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: keys.ingestStatus });
      qc.invalidateQueries({ queryKey: ["audit"] });
      qc.invalidateQueries({ queryKey: ["signals"] });
      qc.invalidateQueries({ queryKey: ["actions"] });
    },
  });
}

export function useAddSource() {
  const invalidate = useInvalidateAll();
  return useMutation({
    mutationFn: (input: { handle: string; source_class: string; city?: string; state?: string }) =>
      api.post<WatchedSource>("/api/sources", input),
    onSuccess: invalidate,
  });
}

export function useStartScanJob() {
  return useMutation({
    mutationFn: ({ handle, limit }: { handle: string; limit?: number }) =>
      api.post<{ job_id: string; status: string }>(
        `/api/sources/${encodeURIComponent(handle)}/scan${limit ? `?limit=${limit}` : ""}`,
      ),
  });
}

export function useScanJob(jobId?: string) {
  return useQuery({
    queryKey: ["scan-job", jobId],
    queryFn: () => api.get<import("./types").ScanJob>(`/api/scan-jobs/${jobId}`),
    enabled: !!jobId,
    refetchInterval: (q) => {
      const st = q.state.data?.status;
      if (st === "done" || st === "failed") return false;
      return 1200;
    },
  });
}

export function useBulkScan() {
  const invalidate = useInvalidateAll();
  return useMutation({
    mutationFn: (body: { stale_only?: boolean; classes?: string[]; limit?: number; posts_per_source?: number }) =>
      api.post<{ job_id: string; handles?: string[]; count?: number; message?: string }>(
        "/api/sources/scan-bulk",
        body,
      ),
    onSuccess: invalidate,
  });
}

export function useSuppressVendorPairs() {
  const invalidate = useInvalidateAll();
  return useMutation({
    mutationFn: () => api.post<{ suppressed: number }>("/api/prospects/suppress-vendor-pairs"),
    onSuccess: invalidate,
  });
}

export function useEnrichSource() {
  const invalidate = useInvalidateAll();
  return useMutation({
    mutationFn: (handle: string) =>
      api.post<WatchedSource>(`/api/sources/${encodeURIComponent(handle)}/enrich`),
    onSuccess: invalidate,
  });
}

export function usePatchSourceLocation() {
  const invalidate = useInvalidateAll();
  return useMutation({
    mutationFn: (input: { handle: string; city: string; state: string }) =>
      api.patch<WatchedSource>(`/api/sources/${encodeURIComponent(input.handle)}/location`, {
        city: input.city,
        state: input.state,
      }),
    onSuccess: invalidate,
  });
}

export function useRemoveSource() {
  const invalidate = useInvalidateAll();
  return useMutation({
    mutationFn: (handle: string) => api.del(`/api/sources/${encodeURIComponent(handle)}`),
    onSuccess: invalidate,
  });
}

// useReactivateSource re-activates a paused source. The backend's addSource
// handler already does ON CONFLICT DO UPDATE SET active=TRUE, so re-adding
// the same handle + class resumes it. We need the source_class to send, so
// the mutation takes the full WatchedSource.
export function useReactivateSource() {
  const invalidate = useInvalidateAll();
  return useMutation({
    mutationFn: (source: WatchedSource) =>
      api.post<WatchedSource>("/api/sources", { handle: source.handle, source_class: source.source_class }),
    onSuccess: invalidate,
  });
}

// --- Congratulate kits (postcard + address research) -----------------------

export function useKits(status?: string) {
  return useQuery({
    queryKey: [...keys.kits, status ?? "all"],
    queryFn: () =>
      api.get<CongratulateKit[]>(`/api/kits${status ? `?status=${encodeURIComponent(status)}` : ""}`),
  });
}

export function useCoupleKit(coupleId?: string) {
  return useQuery({
    queryKey: keys.coupleKit(coupleId ?? ""),
    queryFn: () => api.get<CongratulateKit>(`/api/couples/${coupleId}/kit`),
    enabled: !!coupleId,
    retry: false,
  });
}

export function useBuildCongratulateKit() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (coupleId: string) =>
      api.post<CongratulateKit>(`/api/couples/${encodeURIComponent(coupleId)}/congratulate`),
    onSuccess: (kit) => {
      qc.invalidateQueries({ queryKey: keys.kits });
      qc.invalidateQueries({ queryKey: keys.coupleKit(kit.couple_id) });
      qc.invalidateQueries({ queryKey: keys.kit(kit.id) });
    },
  });
}

export function usePatchKit() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({
      id,
      ...body
    }: {
      id: string;
      address_line1?: string;
      address_line2?: string;
      address_city?: string;
      address_region?: string;
      address_postal?: string;
      address_country?: string;
      headline?: string;
      body_message?: string;
      first_name_a?: string;
      last_name_a?: string;
      first_name_b?: string;
      last_name_b?: string;
      verify?: boolean;
      verified_by?: string;
    }) => api.patch<CongratulateKit>(`/api/kits/${encodeURIComponent(id)}`, body),
    onSuccess: (kit) => {
      qc.invalidateQueries({ queryKey: keys.kits });
      qc.invalidateQueries({ queryKey: keys.coupleKit(kit.couple_id) });
      qc.invalidateQueries({ queryKey: keys.kit(kit.id) });
    },
  });
}

export function useKitReadyToMail() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.post<CongratulateKit>(`/api/kits/${encodeURIComponent(id)}/ready-to-mail`),
    onSuccess: () => qc.invalidateQueries({ queryKey: keys.kits }),
  });
}

export function useKitMarkMailed() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.post<CongratulateKit>(`/api/kits/${encodeURIComponent(id)}/mailed`),
    onSuccess: () => qc.invalidateQueries({ queryKey: keys.kits }),
  });
}

export function useRunDetective() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.post<CongratulateKit>(`/api/kits/${encodeURIComponent(id)}/run-detective`),
    onSuccess: (kit) => {
      qc.invalidateQueries({ queryKey: keys.kits });
      qc.invalidateQueries({ queryKey: keys.kit(kit.id) });
      qc.invalidateQueries({ queryKey: keys.coupleKit(kit.couple_id) });
    },
  });
}

export function useApplyCandidate() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, index }: { id: string; index: number }) =>
      api.post<CongratulateKit>(`/api/kits/${encodeURIComponent(id)}/apply-candidate`, { index }),
    onSuccess: (kit) => {
      qc.invalidateQueries({ queryKey: keys.kits });
      qc.invalidateQueries({ queryKey: keys.coupleKit(kit.couple_id) });
    },
  });
}

export function useVerifyKitAddress() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) =>
      api.post<CongratulateKit>(`/api/kits/${encodeURIComponent(id)}/verify-address`, {
        verified_by: "operator",
      }),
    onSuccess: (kit) => {
      qc.invalidateQueries({ queryKey: keys.kits });
      qc.invalidateQueries({ queryKey: keys.coupleKit(kit.couple_id) });
    },
  });
}

export function useSendPostcard() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.post<CongratulateKit>(`/api/kits/${encodeURIComponent(id)}/send-postcard`),
    onSuccess: (kit) => {
      qc.invalidateQueries({ queryKey: keys.kits });
      qc.invalidateQueries({ queryKey: keys.coupleKit(kit.couple_id) });
    },
  });
}

export function useRunJanitor() {
  const invalidate = useInvalidateAll();
  return useMutation({
    mutationFn: () =>
      api.post<{ vendor_pairs_suppressed: number; observation_facts_backfilled: number; errors?: string[] }>(
        "/api/ops/janitor",
      ),
    onSuccess: invalidate,
  });
}

export function useCoupleDossier(coupleId?: string) {
  return useQuery({
    queryKey: keys.dossier(coupleId),
    queryFn: () => api.get<CoupleDossier>(`/api/couples/${encodeURIComponent(coupleId!)}/dossier`),
    enabled: !!coupleId,
  });
}

export function useCreateHandoff() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (coupleId: string) =>
      api.post<{ couple_id: string; handoff_code: string; handoff_url: string; handoff_utm: string; journey_stage: string }>(
        `/api/couples/${encodeURIComponent(coupleId)}/handoff`,
      ),
    onSuccess: (_d, coupleId) => {
      qc.invalidateQueries({ queryKey: keys.dossier(coupleId) });
      qc.invalidateQueries({ queryKey: keys.prospectBoard });
      qc.invalidateQueries({ queryKey: ["ops"] });
    },
  });
}

export function useSetJourneyStage() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ coupleId, stage }: { coupleId: string; stage: string }) =>
      api.post<{ couple_id: string; journey_stage: string }>(
        `/api/couples/${encodeURIComponent(coupleId)}/journey`,
        { stage },
      ),
    onSuccess: (_d, vars) => {
      qc.invalidateQueries({ queryKey: keys.dossier(vars.coupleId) });
      qc.invalidateQueries({ queryKey: keys.prospectBoard });
      qc.invalidateQueries({ queryKey: ["ops"] });
    },
  });
}

export function useFunnelStats() {
  return useQuery({
    queryKey: ["funnel", "stats"],
    queryFn: () => api.get<FunnelStats>("/api/funnel/stats"),
    staleTime: 30_000,
  });
}

export function useFunnelEvents(coupleId?: string) {
  return useQuery({
    queryKey: ["funnel", "events", coupleId ?? "all"],
    queryFn: () =>
      api.get<FunnelEvent[]>(
        `/api/funnel/events${coupleId ? `?couple_id=${encodeURIComponent(coupleId)}` : ""}`,
      ),
    staleTime: 15_000,
  });
}

export function useAutopsies() {
  return useQuery({
    queryKey: ["trust", "autopsies"],
    queryFn: () => api.get<AutopsyReport[]>("/api/trust/autopsies"),
    staleTime: 30_000,
  });
}

export function useGenerateAutopsy() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (args?: { days?: number }) =>
      api.post<AutopsyReport>("/api/trust/autopsy", {
        days: args?.days ?? 7,
        generated_by: "human:concierge",
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["trust"] });
      qc.invalidateQueries({ queryKey: ["audit"] });
    },
  });
}

// --- National source registry map (any USPS state) ------------------------
// Gated by `enabled` so nothing fetches until a state is selected.
// All four fetch together so layer tab switches are instant.

export function useNationalCoverage(enabled = true) {
  return useQuery({
    queryKey: keys.mapCoverage,
    queryFn: () => api.get<StateCoverage[]>("/api/map/coverage"),
    enabled,
    staleTime: 60_000,
  });
}

export function useStateOverview(state: string | undefined, enabled: boolean) {
  const st = (state || "").toUpperCase();
  return useQuery({
    queryKey: keys.stateOverview(st),
    queryFn: () => api.get<OverviewCityView[]>(`/api/map/states/${st}/overview`),
    enabled: enabled && st.length === 2,
    staleTime: 60_000,
  });
}

export function useStateGovernment(state: string | undefined, enabled: boolean) {
  const st = (state || "").toUpperCase();
  return useQuery({
    queryKey: keys.stateGovernment(st),
    queryFn: () => api.get<CountyGovernmentView[]>(`/api/map/states/${st}/government`),
    enabled: enabled && st.length === 2,
    staleTime: 60_000,
  });
}

export function useStateChurches(state: string | undefined, enabled: boolean) {
  const st = (state || "").toUpperCase();
  return useQuery({
    queryKey: keys.stateChurches(st),
    queryFn: () => api.get<DioceseView[]>(`/api/map/states/${st}/churches`),
    enabled: enabled && st.length === 2,
    staleTime: 60_000,
  });
}

export function useStateSocial(state: string | undefined, enabled: boolean) {
  const st = (state || "").toUpperCase();
  return useQuery({
    queryKey: keys.stateSocial(st),
    queryFn: async () => {
      const raw = await api.get<SocialMarketView | SocialMarketView[]>(`/api/map/states/${st}/social`);
      // OH may still return a single market object; normalize to array.
      if (Array.isArray(raw)) return raw;
      return raw ? [raw] : [];
    },
    enabled: enabled && st.length === 2,
    staleTime: 60_000,
  });
}

/** @deprecated use useStateOverview("OH", enabled) */
export function useOhioOverview(enabled: boolean) {
  return useStateOverview("OH", enabled);
}
/** @deprecated use useStateGovernment("OH", enabled) */
export function useOhioGovernment(enabled: boolean) {
  return useStateGovernment("OH", enabled);
}
/** @deprecated use useStateChurches("OH", enabled) */
export function useOhioChurches(enabled: boolean) {
  return useStateChurches("OH", enabled);
}
/** @deprecated use useStateSocial — returns array */
export function useOhioSocial(enabled: boolean) {
  const q = useStateSocial("OH", enabled);
  return {
    ...q,
    data: q.data?.[0] as SocialMarketView | undefined,
  };
}

// --- Universal search / DLQ / job queue / admin ---------------------------

export function useSearch(query: string, type = "all") {
  return useQuery({
    queryKey: keys.search(query, type),
    queryFn: () =>
      api.get<SearchResult>(
        `/api/search?q=${encodeURIComponent(query)}&type=${encodeURIComponent(type)}`,
      ),
    enabled: query.trim().length > 0,
    staleTime: 15_000,
  });
}

export function useDLQ(status = "pending", limit = 50) {
  return useQuery({
    queryKey: keys.dlq(status, limit),
    queryFn: () =>
      api.get<DLQItem[]>(`/api/dlq?status=${encodeURIComponent(status)}&limit=${limit}`),
    refetchInterval: 30_000,
  });
}

export function useReplayDLQ() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.post(`/api/dlq/${encodeURIComponent(id)}/replay`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["dlq"] }),
  });
}

export function useScanJobs(limit = 50, status = "all") {
  return useQuery({
    queryKey: keys.scanJobs(limit, status),
    queryFn: () =>
      api.get<import("./types").ScanJob[]>(
        `/api/scan-jobs?limit=${limit}&status=${encodeURIComponent(status)}`,
      ),
    // ponytail: poll every 10s — running jobs finish in seconds-to-minutes,
    // so 10s is responsive without hammering. Upgrade: websocket when added.
    refetchInterval: 10_000,
  });
}

export function useUsers() {
  return useQuery({
    queryKey: keys.users,
    queryFn: () => api.get<UserSummary[]>("/api/users"),
    staleTime: 30_000,
  });
}

export function useRotateAPIKey() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.post(`/api/users/${encodeURIComponent(id)}/rotate-key`),
    onSuccess: () => qc.invalidateQueries({ queryKey: keys.users }),
  });
}

export function useDisableUser() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.post(`/api/users/${encodeURIComponent(id)}/disable`),
    onSuccess: () => qc.invalidateQueries({ queryKey: keys.users }),
  });
}

export function useEnableUser() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.post(`/api/users/${encodeURIComponent(id)}/enable`),
    onSuccess: () => qc.invalidateQueries({ queryKey: keys.users }),
  });
}

// SSE live feed. EventSource can't set headers, so the bearer token rides as
// a query param. Reconnects with exponential backoff (1s → 2s → 4s … capped 30s).
const BASE_URL = import.meta.env.VITE_API_URL ?? "";

export function useLiveEvents() {
  const [events, setEvents] = useState<LiveEvent[]>([]);
  const [connected, setConnected] = useState(false);
  const esRef = useRef<EventSource | null>(null);
  const backoff = useRef(1_000);

  useEffect(() => {
    let stopped = false;

    const connect = () => {
      if (stopped) return;
      const token = getToken();
      const url = `${BASE_URL}/api/events/stream${token ? `?token=${encodeURIComponent(token)}` : ""}`;
      const es = new EventSource(url);
      esRef.current = es;

      es.onopen = () => {
        setConnected(true);
        backoff.current = 1_000;
      };

      es.onmessage = (e) => {
        try {
          const evt = JSON.parse(e.data) as LiveEvent;
          setEvents((prev) => [...prev.slice(-49), evt]);
        } catch {
          /* malformed frame — drop */
        }
      };

      es.onerror = () => {
        setConnected(false);
        es.close();
        esRef.current = null;
        // ponytail: capped exponential backoff; EventSource auto-reconnects,
        // but we manage it ourselves so a bad token doesn't spin a tight loop.
        const delay = Math.min(backoff.current, 30_000);
        backoff.current = Math.min(backoff.current * 2, 30_000);
        window.setTimeout(connect, delay);
      };
    };

    connect();
    return () => {
      stopped = true;
      esRef.current?.close();
      esRef.current = null;
    };
  }, []);

  return { events, connected };
}
