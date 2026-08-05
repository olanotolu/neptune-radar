package ops

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"neptune-social-radar/backend/internal/mail"
	"neptune-social-radar/backend/internal/outreach"
	"neptune-social-radar/backend/internal/store"
)

// FollowUpResult is the outcome of sending one follow-up postcard.
type FollowUpResult struct {
	KitID      string `json:"kit_id"`
	ExternalID string `json:"external_id,omitempty"`
	Error      string `json:"error,omitempty"`
	Sent       bool   `json:"sent"`
}

// ProcessFollowUpQueueResult summarizes a full queue sweep.
type ProcessFollowUpQueueResult struct {
	Sent    int               `json:"sent"`
	Skipped int               `json:"skipped"`
	Failed  int               `json:"failed"`
	Results []FollowUpResult  `json:"results,omitempty"`
}

// MinAddressConfidence is the lowest address quality we'll auto-send to.
const MinAddressConfidence = 0.50

// MaxFollowUps caps how many follow-up cards a single kit receives.
const MaxFollowUps = 2

// IsDueForFollowUp reports whether a kit meets all conditions for an
// automatic follow-up: FollowUpAt set and in the past, not yet sent,
// under the max count, and address confidence at or above the floor.
// This is the single source of truth for the due-check — the API queue
// endpoint, the cron, and the test all call it.
func IsDueForFollowUp(k store.CongratulateKit, now time.Time) bool {
	if k.FollowUpAt == nil || k.FollowUpAt.After(now) {
		return false
	}
	if k.FollowUpSentAt != nil {
		return false
	}
	if k.FollowUpCount >= MaxFollowUps {
		return false
	}
	if k.AddressConfidence < MinAddressConfidence {
		return false
	}
	return true
}

// SendFollowUp sends one follow-up postcard for a kit via Lob and records
// the result (FollowUpSentAt + FollowUpCount increment + audit trail).
// This is the shared sending path — both the manual API endpoint and the
// automatic cron call it so there's no duplicated mail logic.
func SendFollowUp(ctx context.Context, s *store.Store, mailer *mail.Client, kitID string) (FollowUpResult, error) {
	k, err := s.GetCongratulateKit(kitID)
	if err != nil {
		return FollowUpResult{KitID: kitID, Error: err.Error()}, err
	}
	if k.Status != "mailed" {
		return FollowUpResult{KitID: kitID, Error: "kit must be mailed before follow-up"},
			fmt.Errorf("kit must be mailed before follow-up")
	}
	if k.FollowUpCount >= MaxFollowUps {
		return FollowUpResult{KitID: kitID, Error: "max follow-ups reached"},
			fmt.Errorf("max %d follow-ups per kit", MaxFollowUps)
	}
	if k.AddressLine1 == "" || k.AddressPostal == "" {
		return FollowUpResult{KitID: kitID, Error: "complete address required"},
			fmt.Errorf("complete address required for follow-up")
	}
	if mailer == nil || !mailer.Available() {
		return FollowUpResult{KitID: kitID, Error: "LOB_API_KEY not configured"},
			fmt.Errorf("lob unavailable")
	}

	// Use the follow-up template (different from first card).
	tpls := outreach.TemplateLibrary()
	tplID := k.FollowUpTemplate
	if tplID == "" {
		tplID = "bright_casual"
	}
	var tpl *outreach.GreetingTemplate
	for i := range tpls {
		if tpls[i].ID == tplID {
			tpl = &tpls[i]
			break
		}
	}
	if tpl == nil {
		tpl = &tpls[0]
	}

	data := outreach.TemplateData{
		NameA:    k.FirstNameA,
		NameB:    k.FirstNameB,
		Location: k.MarketCity,
	}
	followUpBody := outreach.RenderTemplate(*tpl, data)

	front := outreach.RenderPostcardHTML(k)
	back := fmt.Sprintf(`<html><body style="margin:0;font-family:Georgia,serif;padding:24px;font-size:13px;line-height:1.5;color:#1c1917">%s<p style="margin-top:20px;font-size:12px;color:#57534e">Neptune · with care (follow-up)</p></body></html>`,
		strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(followUpBody, "&", "&amp;"), "<", "&lt;"), ">", "&gt;"))
	to := mail.Address{
		Name:           strings.TrimSpace(k.PersonAName + " & " + k.PersonBName),
		AddressLine1:   k.AddressLine1,
		AddressLine2:   k.AddressLine2,
		AddressCity:    k.AddressCity,
		AddressState:   k.AddressRegion,
		AddressZip:     k.AddressPostal,
		AddressCountry: firstNonEmpty(k.AddressCountry, "US"),
	}
	res, err := mailer.SendPostcard(ctx, to, front, back,
		fmt.Sprintf("Neptune follow-up %s & %s (#%d)", k.PersonAName, k.PersonBName, k.FollowUpCount+1))
	if err != nil {
		return FollowUpResult{KitID: kitID, Error: err.Error()}, err
	}

	// Record follow-up.
	k.FollowUpCount++
	now := time.Now().UTC()
	k.FollowUpSentAt = &now
	k.InternalNote = strings.TrimSpace(k.InternalNote + fmt.Sprintf(
		"\n\nFollow-up #%d sent via Lob: %s", k.FollowUpCount, res.ExternalID))
	if _, err := s.UpsertCongratulateKit(k); err != nil {
		return FollowUpResult{KitID: kitID, Error: err.Error()}, err
	}
	// Audit trail.
	_, _ = s.Audit("kit", k.ID, "follow_up_sent",
		map[string]any{"couple_id": k.CoupleID, "external_id": res.ExternalID,
			"follow_up_count": k.FollowUpCount, "source": "auto"}, "", 0)

	return FollowUpResult{KitID: kitID, ExternalID: res.ExternalID, Sent: true}, nil
}

// ProcessFollowUpQueue scans mailed kits and auto-sends follow-up postcards
// for every kit that is due. If sending fails for one kit it logs and
// continues — one bad address never blocks the rest of the queue.
func ProcessFollowUpQueue(s *store.Store, mailer *mail.Client) ProcessFollowUpQueueResult {
	return ProcessFollowUpQueueCtx(context.Background(), s, mailer)
}

// ProcessFollowUpQueueCtx is the testable variant that accepts a context.
func ProcessFollowUpQueueCtx(ctx context.Context, s *store.Store, mailer *mail.Client) ProcessFollowUpQueueResult {
	var res ProcessFollowUpQueueResult
	kits, err := s.ListCongratulateKits("mailed", 200)
	if err != nil {
		log.Printf("[followup] list mailed kits: %v", err)
		res.Failed = -1 // ponytail: sentinel — caller checks < 0 for hard error
		return res
	}
	now := time.Now().UTC()
	for _, k := range kits {
		if !IsDueForFollowUp(k, now) {
			res.Skipped++
			continue
		}
		fr, err := SendFollowUp(ctx, s, mailer, k.ID)
		if err != nil {
			log.Printf("[followup] send failed for kit %s: %v", k.ID, err)
			res.Failed++
			res.Results = append(res.Results, fr)
			continue // don't block the whole queue
		}
		res.Sent++
		res.Results = append(res.Results, fr)
		log.Printf("[followup] sent kit=%s couple=%s count=%d ext=%s",
			k.ID, k.CoupleID, k.FollowUpCount, fr.ExternalID)
	}
	if res.Sent > 0 {
		log.Printf("[followup] queue sweep: sent=%d skipped=%d failed=%d",
			res.Sent, res.Skipped, res.Failed)
	}
	return res
}

// RunFollowUpCron is the leader-elected hourly loop. Only the leader replica
// processes follow-ups — same advisory-lock pattern as the ingest worker.
// Non-leaders idle until context cancel (the API still serves on every replica).
func RunFollowUpCron(ctx context.Context, s *store.Store, mailer *mail.Client) {
	leader, err := s.TryAcquireLeaderLock()
	if err != nil {
		log.Printf("[followup] leader lock acquire failed: %v", err)
		<-ctx.Done()
		return
	}
	if !leader {
		log.Println("[followup] another replica holds the leader lock — cron idle")
		<-ctx.Done()
		return
	}
	log.Println("[followup] acquired leader lock — hourly follow-up cron active")
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Println("[followup] cron stopped")
			return
		case <-ticker.C:
			ProcessFollowUpQueue(s, mailer)
		}
	}
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
