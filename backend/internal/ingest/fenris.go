package ingest

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"neptune-social-radar/backend/internal/ontology"
	"neptune-social-radar/backend/internal/store"
)

// fenrisAPIBase is the Fenris Digital life events search endpoint.
const fenrisAPIBase = "https://api.fenrisd.com/services/lifeevents/v1/events/search"

// fenrisEventTypes are the life events we care about for couple discovery.
var fenrisEventTypes = []string{"Newly Engaged", "Newly Married"}

// fenrisPollInterval is how often the poller checks for new events.
const fenrisPollInterval = 24 * time.Hour

// fenrisHTTPTimeout caps each API call.
const fenrisHTTPTimeout = 30 * time.Second

// fenrisResponse is the JSON envelope from the Fenris life events search API.
// ponytail: Fenris docs are behind a login; this follows the common pattern
// for their API (events array with standard life-event fields). Ceiling: if
// the real API uses a different envelope shape, adjust this struct + the
// field mappings in FetchLifeEvents. The test covers the parsing path.
type fenrisResponse struct {
	Events []fenrisEvent `json:"events"`
}

type fenrisEvent struct {
	EventType   string  `json:"event_type"`
	PersonName  string  `json:"person_name"`
	HouseholdID string  `json:"household_id"`
	Address     string  `json:"address"`
	City        string  `json:"city"`
	State       string  `json:"state"`
	Zip         string  `json:"zip"`
	EventDate   string  `json:"event_date"` // ISO 8601
	Confidence  float64 `json:"confidence"`
}

// FetchLifeEvents calls the Fenris Digital life events search endpoint and
// returns parsed LifeEvent structs. Auth is via the API key in the
// X-API-Key header (FENRIS_API_KEY env var).
func FetchLifeEvents(apiKey string, eventTypes []string, since time.Time) ([]ontology.LifeEvent, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("fenris: empty API key")
	}
	params := url.Values{}
	params.Set("api_key", apiKey) // ponytail: also pass as query param per Fenris pattern
	for _, et := range eventTypes {
		params.Add("event_types", et)
	}
	params.Set("since", since.Format("2006-01-02"))

	req, err := http.NewRequest(http.MethodGet, fenrisAPIBase+"?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("fenris: build request: %w", err)
	}
	req.Header.Set("X-API-Key", apiKey)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: fenrisHTTPTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fenris: request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("fenris: API returned %d: %s", resp.StatusCode, string(body))
	}

	var fr fenrisResponse
	if err := json.NewDecoder(resp.Body).Decode(&fr); err != nil {
		return nil, fmt.Errorf("fenris: decode response: %w", err)
	}

	events := make([]ontology.LifeEvent, 0, len(fr.Events))
	for _, fe := range fr.Events {
		date, err := time.Parse("2006-01-02", fe.EventDate)
		if err != nil {
			// Try full ISO 8601 as fallback
			date, err = time.Parse(time.RFC3339, fe.EventDate)
			if err != nil {
				log.Printf("[fenris] skipping event with unparseable date %q: %v", fe.EventDate, err)
				continue
			}
		}
		events = append(events, ontology.LifeEvent{
			EventType:   fe.EventType,
			PersonName:  fe.PersonName,
			HouseholdID: fe.HouseholdID,
			Address:     fe.Address,
			City:        fe.City,
			State:       fe.State,
			Zip:         fe.Zip,
			EventDate:   date,
			Confidence:  fe.Confidence,
		})
	}
	return events, nil
}

// RunFenrisPoller is the daily Fenris life events poll loop. If
// FENRIS_API_KEY is not set, it no-ops (logs once, then idles). Non-blocking
// errors (API down, parse failures) are logged and the loop continues — the
// social watch loop is unaffected.
func (w *Worker) RunFenrisPoller(ctx context.Context) {
	apiKey := os.Getenv("FENRIS_API_KEY")
	if apiKey == "" {
		log.Println("[fenris] FENRIS_API_KEY not set — life events poller idle")
		<-ctx.Done()
		return
	}
	log.Println("[fenris] life events poller started (interval=24h)")
	w.pollFenris(ctx, apiKey) // first poll immediately
	ticker := time.NewTicker(fenrisPollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Println("[fenris] life events poller stopped")
			return
		case <-ticker.C:
			w.pollFenris(ctx, apiKey)
		}
	}
}

// pollFenris fetches new life events and processes them: creates couples for
// new discoveries, cross-validates existing couples, and stores every event.
func (w *Worker) pollFenris(ctx context.Context, apiKey string) {
	if w.paused.Load() {
		return
	}
	since := time.Now().UTC().Add(-90 * 24 * time.Hour) // Fenris provides past 90 days
	events, err := FetchLifeEvents(apiKey, fenrisEventTypes, since)
	if err != nil {
		log.Printf("[fenris] fetch failed, continuing: %v", err)
		return
	}
	if len(events) == 0 {
		return
	}
	log.Printf("[fenris] fetched %d life events", len(events))

	for _, ev := range events {
		// Only process the event types we care about.
		if ev.EventType != "Newly Engaged" && ev.EventType != "Newly Married" {
			continue
		}
		externalID := fmt.Sprintf("fenris:%s:%s:%s", ev.HouseholdID, ev.EventType, ev.EventDate.Format("2006-01-02"))
		if ev.HouseholdID == "" {
			// ponytail: fall back to name+state+date for dedup when household_id absent.
			externalID = fmt.Sprintf("fenris:%s:%s:%s", strings.ToLower(ev.PersonName), ev.State, ev.EventDate.Format("2006-01-02"))
		}

		// Try to match an existing couple (discovered via Instagram) by name + state.
		couple, err := w.store.FindCoupleByNameState(ev.PersonName, ev.State)
		crossValidated := false
		coupleID := ""

		if err == nil {
			// Existing couple found — cross-validate with Fenris signal.
			coupleID = couple.ID
			crossValidated = true
			if !couple.FenrisValidated {
				if err := w.store.SetFenrisValidated(couple.ID); err != nil {
					log.Printf("[fenris] set fenris_validated for couple %s: %v", couple.ID, err)
				} else {
					log.Printf("[fenris] cross-validated couple %s with %s event for %s",
						couple.ID, ev.EventType, ev.PersonName)
				}
			}
		} else if err == store.ErrDuplicateObservation {
			continue // already processed
		} else {
			// No existing couple — create a new one from the Fenris event.
			coupleID = w.createCoupleFromFenris(ev)
		}

		// Store the event (deduped by external_id).
		if _, err := w.store.InsertFenrisEvent(ev, externalID, coupleID, crossValidated); err != nil {
			if err == store.ErrDuplicateFenrisEvent {
				continue // already ingested
			}
			log.Printf("[fenris] store event for %s: %v", ev.PersonName, err)
		}
	}
}

// createCoupleFromFenris creates a person + placeholder partner + couple from
// a Fenris life event. Returns the couple ID, or "" on failure.
// ponytail: we only have one name from Fenris; the partner is "Unknown Partner"
// until Instagram discovery or identity resolution fills them in. Ceiling: a
// second Fenris event for the partner would upgrade this; upgrade path =
// match on household_id to pair two events.
func (w *Worker) createCoupleFromFenris(ev ontology.LifeEvent) string {
	personA, err := w.store.CreatePerson(ontology.Person{
		DisplayName: ev.PersonName,
		CRMSource:   "fenris_life_event",
	})
	if err != nil {
		log.Printf("[fenris] create person %s: %v", ev.PersonName, err)
		return ""
	}
	personB, err := w.store.CreatePerson(ontology.Person{
		DisplayName: "Unknown Partner",
		CRMSource:   "fenris_life_event",
	})
	if err != nil {
		log.Printf("[fenris] create placeholder partner: %v", err)
		return ""
	}
	couple, err := w.store.EnsureCouple(personA.ID, personB.ID)
	if err != nil {
		log.Printf("[fenris] ensure couple: %v", err)
		return ""
	}
	// Tag the couple with Fenris source and location.
	_ = w.store.SetCoupleSource(couple.ID, "fenris_life_event")
	if ev.City != "" || ev.State != "" {
		lat, lng := cityCoords(ev.City, ev.State)
		_ = w.store.UpdateCoupleLocation(couple.ID, ev.City, ev.State, "fenris_life_event", lat, lng)
	}
	log.Printf("[fenris] created couple %s from %s event for %s (%s, %s)",
		couple.ID, ev.EventType, ev.PersonName, ev.City, ev.State)
	return couple.ID
}
