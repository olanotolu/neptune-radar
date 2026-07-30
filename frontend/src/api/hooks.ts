import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { api } from "./client";
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
  WatchedSource,
} from "./types";

const keys = {
  signals: (monitor?: string) => ["signals", monitor] as const,
  coupleGraph: (coupleId?: string) => ["couple", coupleId, "graph"] as const,
  relationship: (coupleId?: string) => ["couple", coupleId, "relationship"] as const,
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
  ohioOverview: ["map", "ohio", "overview"] as const,
  ohioGovernment: ["map", "ohio", "government"] as const,
  ohioChurches: ["map", "ohio", "churches"] as const,
  ohioSocial: ["map", "ohio", "social"] as const,
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

// --- Ohio source registry: real government/church/social connectors --------
// Gated by `enabled` so nothing fetches until Ohio is actually selected on
// the map. All four fetch together once Ohio is selected (not per-tab) so
// switching layer tabs is instant, not a fresh spinner each time.

export function useOhioOverview(enabled: boolean) {
  return useQuery({
    queryKey: keys.ohioOverview,
    queryFn: () => api.get<OverviewCityView[]>("/api/map/states/OH/overview"),
    enabled,
    staleTime: 60_000,
  });
}

export function useOhioGovernment(enabled: boolean) {
  return useQuery({
    queryKey: keys.ohioGovernment,
    queryFn: () => api.get<CountyGovernmentView[]>("/api/map/states/OH/government"),
    enabled,
    staleTime: 60_000,
  });
}

export function useOhioChurches(enabled: boolean) {
  return useQuery({
    queryKey: keys.ohioChurches,
    queryFn: () => api.get<DioceseView[]>("/api/map/states/OH/churches"),
    enabled,
    staleTime: 60_000,
  });
}

export function useOhioSocial(enabled: boolean) {
  return useQuery({
    queryKey: keys.ohioSocial,
    queryFn: () => api.get<SocialMarketView>("/api/map/states/OH/social"),
    enabled,
    staleTime: 60_000,
  });
}
