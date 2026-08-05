// Package ontology defines the plain data types that mirror the SQLite schema.
// These are the nouns of the living relationship graph: Person, SocialAccount,
// Couple, Relationship, SocialObservation, Evidence, LifeEventHypothesis,
// CRMLead, NeptuneCase, ConsentPolicy, RecommendedAction, ExecutedAction, AuditEvent.
package ontology

import "time"

// VisibilityScope controls who/what may see a piece of derived state.
// unconfirmed_inference is the default for anything the system proposed but
// no human or corroborating signal has confirmed.
type VisibilityScope string

const (
	ScopePrivatePersonA   VisibilityScope = "private_person_a"
	ScopePrivatePersonB   VisibilityScope = "private_person_b"
	ScopeSharedCouple     VisibilityScope = "shared_couple"
	ScopeNeptuneInternal  VisibilityScope = "neptune_internal"
	ScopeAttorneyOnly     VisibilityScope = "attorney_only"
	ScopeUnconfirmedInfer VisibilityScope = "unconfirmed_inference"
)

type RelationshipStage string

const (
	StageUnknown         RelationshipStage = "unknown"
	StageDatingSuspected RelationshipStage = "dating_suspected"
	StageEngaged         RelationshipStage = "engaged"
	StageMarried         RelationshipStage = "married"
	StageStatusUncertain RelationshipStage = "status_uncertain"
	StageEndedSuspected  RelationshipStage = "ended_suspected"
)

type HypothesisStatus string

const (
	HypothesisUnconfirmed   HypothesisStatus = "unconfirmed"
	HypothesisCorroborating HypothesisStatus = "corroborating"
	HypothesisConfirmed     HypothesisStatus = "confirmed"
	HypothesisRejected      HypothesisStatus = "rejected"
	HypothesisExpired       HypothesisStatus = "expired"
)

// EventType values for LifeEventHypothesis.EventType. Defined here (rather
// than in pipeline/analyst) so that internal/pipeline/policy can reference
// them without importing analyst — analyst imports internal/llm, and policy
// must never transitively depend on it.
const (
	EventTypeEngagement         = "engagement"
	EventTypeRelationshipChange = "relationship_state_change"
)

// EvidenceKind values for Evidence.Kind — the named signals the scorer writes.
// Defined here so both the scorer and the dashboard can reference them without
// a circular import on pipeline/scorer.
const (
	EvidenceRingDetected    = "ring_detected"    // YOLOv8 ring detection (confidence ≥ 0.5)
	EvidenceProposalPhoto   = "proposal_photo"   // CLIP zero-shot classified as proposal/engagement shoot
	EvidenceDispersionScore = "dispersion_score" // FAIR dispersion metric for relationship scoring
)

type ActionType string

const (
	ActionReview          ActionType = "review"
	ActionIgnore          ActionType = "ignore"
	ActionDraftOutreach   ActionType = "draft_outreach"
	ActionPauseAutomation ActionType = "pause_automation"
	ActionCreateCase      ActionType = "create_case"
	ActionConciergeReview ActionType = "concierge_review"
	// ActionInvestigate is the human investigation queue: an engagement
	// prospect that cleared the discard bar but not the create-prospect bar.
	// A human verifies the couple and the event are real; approving promotes
	// it to a lead, ignoring rejects the hypothesis.
	ActionInvestigate ActionType = "investigate"
	ActionNoAction    ActionType = "no_action"
)

type ActionStatus string

const (
	ActionPending  ActionStatus = "pending"
	ActionApproved ActionStatus = "approved"
	ActionIgnored  ActionStatus = "ignored"
	ActionExecuted ActionStatus = "executed"
	ActionFailed   ActionStatus = "failed"
)

type Person struct {
	ID          string    `json:"id"`
	DisplayName string    `json:"display_name"`
	Email       string    `json:"email,omitempty"`
	CRMSource   string    `json:"crm_source,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type SocialAccount struct {
	ID               string     `json:"id"`
	PersonID         string     `json:"person_id,omitempty"` // empty until identity resolved
	Platform         string     `json:"platform"`
	Handle           string     `json:"handle"`
	DisplayName      string     `json:"display_name,omitempty"`
	BioText          string     `json:"bio_text,omitempty"`
	IsPrivate        bool       `json:"is_private"`
	IsDisabled       bool       `json:"is_disabled"`
	LastSeenAt       time.Time  `json:"last_seen_at,omitempty"`
	ProfilePicURL    string     `json:"profile_pic_url,omitempty"`
	FollowerCount    *int       `json:"follower_count,omitempty"`
	FollowingCount   *int       `json:"following_count,omitempty"`
	ProfileCheckedAt *time.Time `json:"profile_checked_at,omitempty"`
	InferredCity     string     `json:"inferred_city,omitempty"`
	InferredRegion   string     `json:"inferred_region,omitempty"`
	LocationSource   string     `json:"location_source,omitempty"`
}

type Couple struct {
	ID               string     `json:"id"`
	PersonAID        string     `json:"person_a_id"`
	PersonBID        string     `json:"person_b_id"`
	CreatedAt        time.Time  `json:"created_at"`
	InferredCity     string     `json:"inferred_city,omitempty"`
	InferredRegion   string     `json:"inferred_region,omitempty"`
	InferredLat      *float64   `json:"inferred_lat,omitempty"`
	InferredLng      *float64   `json:"inferred_lng,omitempty"`
	LocationSource   string     `json:"location_source,omitempty"`
	SuppressedAt     *time.Time `json:"suppressed_at,omitempty"`
	SuppressedReason string     `json:"suppressed_reason,omitempty"`
	Mistaken         bool       `json:"mistaken,omitempty"`
	MistakenReason   string     `json:"mistaken_reason,omitempty"`
	MistakenBy       string     `json:"mistaken_by,omitempty"`
	MistakenAt       *time.Time `json:"mistaken_at,omitempty"`
	// Marriage-license monitoring fields. Source is "social" (default, the
	// Instagram watch loop), "marriage_license" (public filings feed), or
	// "fenris_life_event" (Fenris Digital life-events API). The license fields
	// are set only when source = "marriage_license".
	Source               string     `json:"source,omitempty"`
	LicenseCounty        string     `json:"license_county,omitempty"`
	LicenseFilingDate    *time.Time `json:"license_filing_date,omitempty"`
	PredictedWeddingDate *time.Time `json:"predicted_wedding_date,omitempty"`
	WeddingDate          *time.Time `json:"wedding_date,omitempty"`
	// FenrisValidated is true when a Fenris Digital life event (Newly Engaged
	// or Newly Married) independently cross-validates this couple — two
	// independent signals = higher confidence.
	FenrisValidated bool `json:"fenris_validated,omitempty"`
	// PrenupIntentScore predicts how likely a couple is to need/want a prenup
	// (0-1). Set once by the prep gate via LLM inference, stored — not recomputed.
	PrenupIntentScore   float64  `json:"prenup_intent_score,omitempty"`
	PrenupIntentReason  string   `json:"prenup_intent_reason,omitempty"`
	PrenupIntentSignals []string `json:"prenup_intent_signals,omitempty"`
	// SocialWeddingPrediction is the LLM's rationale for a wedding date
	// predicted from Instagram signals (captions, bios, venue/vendor tags) —
	// distinct from the more authoritative marriage-license prediction. Empty
	// when no social signal was found or a license prediction already exists.
	SocialWeddingPrediction string  `json:"social_wedding_prediction,omitempty"`
	SocialWeddingConfidence float64 `json:"social_wedding_confidence,omitempty"`
}

// LifeEvent is a Fenris Digital life-events trigger: a licensed data-broker
// signal (Newly Engaged, Newly Married, etc.) that cross-validates our
// Instagram discovery. Two independent signals per couple = high confidence.
type LifeEvent struct {
	EventType   string    `json:"event_type"`   // "Newly Engaged", "Newly Married"
	PersonName  string    `json:"person_name"`  // full name of the person
	HouseholdID string    `json:"household_id"` // Fenris household identifier
	Address     string    `json:"address"`      // street address (when available)
	City        string    `json:"city"`
	State       string    `json:"state"`
	Zip         string    `json:"zip"`
	EventDate   time.Time `json:"event_date"`
	Confidence  float64   `json:"confidence"` // 0–1, Fenris-provided match confidence
}

type Relationship struct {
	ID               string            `json:"id"`
	CoupleID         string            `json:"couple_id"`
	Stage            RelationshipStage `json:"stage"`
	Confidence       float64           `json:"confidence"`
	EffectiveFrom    time.Time         `json:"effective_from"`
	EffectiveTo      *time.Time        `json:"effective_to,omitempty"`
	SupersededBy     string            `json:"superseded_by,omitempty"`
	AutomationPaused bool              `json:"automation_paused"`
	VisibilityScope  VisibilityScope   `json:"visibility_scope"`
}

// EdgeKind covers low-attribute social edges. Richer edges (partner_of,
// supersedes, supported_by_evidence, associated_with_lead, enrolled_in_case)
// are foreign keys on their owning entity instead of rows here.
type EdgeKind string

const (
	EdgeFollows     EdgeKind = "follows"
	EdgeTaggedWith  EdgeKind = "tagged_with"
	EdgeMentionedBy EdgeKind = "mentioned_by"
)

type Edge struct {
	ID                  string    `json:"id"`
	Kind                EdgeKind  `json:"kind"`
	FromAccountID       string    `json:"from_account_id"`
	ToAccountID         string    `json:"to_account_id"`
	Active              bool      `json:"active"`
	FirstObservedAt     time.Time `json:"first_observed_at"`
	LastObservedAt      time.Time `json:"last_observed_at"`
	SourceObservationID string    `json:"source_observation_id,omitempty"`
}

type SocialObservation struct {
	ID               string          `json:"id"`
	Monitor          string          `json:"monitor"`
	ExternalEventID  string          `json:"external_event_id"`
	AccountID        string          `json:"account_id,omitempty"`
	ObservationType  string          `json:"observation_type"`
	RawPayload       string          `json:"raw_payload"` // JSON
	ObservedAt       time.Time       `json:"observed_at"`
	IngestedAt       time.Time       `json:"ingested_at"`
	Source           string          `json:"source"`
	FreshnessSeconds int             `json:"freshness_seconds"`
	ConsentScope     VisibilityScope `json:"consent_scope"`
}

type Evidence struct {
	ID            string    `json:"id"`
	HypothesisID  string    `json:"hypothesis_id"`
	ObservationID string    `json:"observation_id,omitempty"`
	Kind          string    `json:"kind"`
	Description   string    `json:"description"`
	Weight        float64   `json:"weight"`
	Confirmed     bool      `json:"confirmed"`
	CreatedAt     time.Time `json:"created_at"`
}

type LifeEventHypothesis struct {
	ID            string            `json:"id"`
	CoupleID      string            `json:"couple_id,omitempty"`
	PersonID      string            `json:"person_id,omitempty"`
	EventType     string            `json:"event_type"`
	ProposedStage RelationshipStage `json:"proposed_stage,omitempty"`
	Confidence    float64           `json:"confidence"`
	// EngagementConfidence and PartnerConfidence are the two independently-
	// gated scores event-first discovery requires: "did an engagement-shaped
	// event happen" and "did we identify the right two people." Both nil for
	// non-engagement hypotheses (e.g. relationship_state_change).
	EngagementConfidence *float64         `json:"engagement_confidence,omitempty"`
	PartnerConfidence    *float64         `json:"partner_confidence,omitempty"`
	ModelOrRule          string           `json:"model_or_rule"` // e.g. "claude-sonnet-5" or "template:bio_regex_v1" or "policy:unfollow_check_v1"
	Status               HypothesisStatus `json:"status"`
	VisibilityScope      VisibilityScope  `json:"visibility_scope"`
	ConsentScope         VisibilityScope  `json:"consent_scope"`
	ExpiresAt            *time.Time       `json:"expires_at,omitempty"`
	CreatedAt            time.Time        `json:"created_at"`
	UpdatedAt            time.Time        `json:"updated_at"`
}

type ConsentPolicy struct {
	ID             string          `json:"id"`
	PersonID       string          `json:"person_id"`
	Scope          VisibilityScope `json:"scope"`
	AllowedActions []string        `json:"allowed_actions"`
	GrantedAt      time.Time       `json:"granted_at"`
	RevokedAt      *time.Time      `json:"revoked_at,omitempty"`
}

type CRMLead struct {
	ID           string    `json:"id"`
	PersonID     string    `json:"person_id"`
	HypothesisID string    `json:"hypothesis_id,omitempty"`
	LeadType     string    `json:"lead_type"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

type NeptuneCase struct {
	ID        string    `json:"id"`
	CoupleID  string    `json:"couple_id,omitempty"`
	LeadID    string    `json:"lead_id,omitempty"`
	CaseType  string    `json:"case_type"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type RecommendedAction struct {
	ID              string       `json:"id"`
	HypothesisID    string       `json:"hypothesis_id,omitempty"`
	CaseID          string       `json:"case_id,omitempty"`
	ActionType      ActionType   `json:"action_type"`
	ProposedPayload string       `json:"proposed_payload"` // JSON
	Status          ActionStatus `json:"status"`
	CreatedAt       time.Time    `json:"created_at"`
	DecidedAt       *time.Time   `json:"decided_at,omitempty"`
	DecidedBy       string       `json:"decided_by,omitempty"`
	Priority        int          `json:"priority"`
	Owner           string       `json:"owner,omitempty"`
	SnoozeUntil     *time.Time   `json:"snooze_until,omitempty"`
	Reason          string       `json:"reason,omitempty"`
}

type ExecutedAction struct {
	ID                  string    `json:"id"`
	RecommendedActionID string    `json:"recommended_action_id"`
	Result              string    `json:"result"`
	Detail              string    `json:"detail,omitempty"`
	Verified            bool      `json:"verified"`
	ExecutedAt          time.Time `json:"executed_at"`
}

type AuditEvent struct {
	ID         string    `json:"id"`
	EntityType string    `json:"entity_type"`
	EntityID   string    `json:"entity_id"`
	Event      string    `json:"event"`
	Detail     string    `json:"detail,omitempty"` // JSON
	Monitor    string    `json:"monitor,omitempty"`
	StepIndex  int       `json:"step_index"`
	CreatedAt  time.Time `json:"created_at"`
}

// PipelineRun is the summary index row for one orchestrator ProcessEvent
// execution. Per-stage detail lives in audit_events and pipeline_timings
// (both keyed by observation_id); this row is the cross-cutting summary that
// makes "show me this run" one query instead of a four-table join.
type PipelineRun struct {
	ID               string     `json:"id"`
	ObservationID    string     `json:"observation_id"`
	AgentName        string     `json:"agent_name"`
	Model            string     `json:"model,omitempty"`
	PromptTokens     int        `json:"prompt_tokens"`
	CompletionTokens int        `json:"completion_tokens"`
	CostUSD          float64    `json:"cost_usd"`
	Confidence       *float64   `json:"confidence,omitempty"`
	StopReason       string     `json:"stop_reason"`
	HypothesisID     string     `json:"hypothesis_id,omitempty"`
	ActionID         string     `json:"action_id,omitempty"`
	CoupleID         string     `json:"couple_id,omitempty"`
	Monitor          string     `json:"monitor,omitempty"`
	StartedAt        time.Time  `json:"started_at"`
	EndedAt          *time.Time `json:"ended_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
}

// WatchedSource is a curated public account the radar monitors (an engagement
// photographer, proposal planner, venue, jeweler, publication, registry
// provider, or boutique — the classes in signals.WatchedSourceClasses).
// State/City are optional geographic tags — set when this account is also
// registered in the source registry's SocialSource (see registry.go), so the
// coverage map can filter to a state/city without a parallel accounts table.
// The profile-stat fields (FollowerCount..ProfileCheckedAt) are only ever
// populated by a real successful Apify profile fetch — nil until then.
type WatchedSource struct {
	ID               string     `json:"id"`
	Handle           string     `json:"handle"`
	SourceClass      string     `json:"source_class"`
	Active           bool       `json:"active"`
	State            string     `json:"state,omitempty"`
	City             string     `json:"city,omitempty"`
	FollowerCount    *int       `json:"follower_count,omitempty"`
	FollowingCount   *int       `json:"following_count,omitempty"`
	PostCount        *int       `json:"post_count,omitempty"`
	FullName         string     `json:"full_name,omitempty"`
	ProfilePicURL    string     `json:"profile_pic_url,omitempty"`
	Verified         *bool      `json:"verified,omitempty"`
	ProfileCheckedAt *time.Time `json:"profile_checked_at,omitempty"`
	LastScannedAt    *time.Time `json:"last_scanned_at,omitempty"`
	LastScanCouples  *int       `json:"last_scan_couples,omitempty"`
	LastScanActions  *int       `json:"last_scan_actions,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
}

// IngestCursor is per-monitor read progress: how far a watch source has been
// consumed, so restarts and overlapping runs neither re-ingest nor miss.
type IngestCursor struct {
	Monitor    string     `json:"monitor"`
	LastSeenAt *time.Time `json:"last_seen_at,omitempty"`
	LastRunAt  *time.Time `json:"last_run_at,omitempty"`
	UpdatedAt  time.Time  `json:"updated_at"`
}
