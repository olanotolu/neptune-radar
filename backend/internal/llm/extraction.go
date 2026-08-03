package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

// ConversationMessage is one utterance in a couple interview session.
type ConversationMessage struct {
	Speaker string `json:"speaker"` // "a1", "a2", "b1", "b2"
	Couple  string `json:"couple"`  // "A" or "B"
	Text    string `json:"text"`
}

// ExtractionAgent analyzes a conversation from one perspective.
type ExtractionAgent struct {
	Type   string // "relationship_stage", "wedding_timeline", "vendor_interest", "location", "budget"
	System string // system prompt for this agent
}

// ExtractionResult is what one agent returns.
type ExtractionResult struct {
	AgentType  string          `json:"agent_type"`
	Findings   json.RawMessage `json:"findings"`
	Confidence float64         `json:"confidence"`
	Summary    string          `json:"summary"`
}

// extractionAgents are the five parallel perspectives run over a conversation.
var extractionAgents = []ExtractionAgent{
	{
		Type: "relationship_stage",
		System: "You are a relationship intelligence analyst. Analyze this conversation between two couples " +
			"and extract signals about their relationship stages (dating, engaged, married, planning). " +
			"Identify each couple's likely stage, evidence quotes, and how confident you are.",
	},
	{
		Type: "wedding_timeline",
		System: "You are a wedding timeline analyst. Extract any wedding date, planning timeline, season " +
			"preferences, and how far along planning is. Note explicit dates and inferred timeframes.",
	},
	{
		Type: "vendor_interest",
		System: "You are a vendor interest analyst. Extract preferences and mentions for: photographer, " +
			"venue, florist, jeweler, planner, videographer, cake, bridal shop, officiant. " +
			"Capture which couple expressed interest and any specifics (style, budget hints, brands).",
	},
	{
		Type: "location",
		System: "You are a location analyst. Extract location and venue preferences: city, state, venue type " +
			"(barn, hotel, beach, church, etc.), travel willingness, and any geographic constraints.",
	},
	{
		Type: "budget",
		System: "You are a budget analyst. Extract budget signals, spending patterns, and price sensitivity. " +
			"Note explicit dollar figures, relative language (\"affordable\", \"splurge\"), and priorities.",
	},
}

// dollarRe matches budget patterns like "30k", "25 k", "$30,000", "30 grand".
var dollarRe = regexp.MustCompile(`\$\d[\d,]*|\b\d+\s*k\b|\b\d+\s*grand\b`)

// extractionResponse is the JSON shape each agent must return.
type extractionResponse struct {
	Findings   json.RawMessage `json:"findings"`
	Confidence float64         `json:"confidence"`
	Summary    string          `json:"summary"`
}

// RunExtractionAgents runs all 5 extraction agents over the conversation in
// parallel. Each agent calls the Baseten chat-completions endpoint directly.
// If the LLM is unavailable (no API key, billing issue, etc.), falls back to
// keyword-based extraction so the feature still works.
func RunExtractionAgents(ctx context.Context, messages []ConversationMessage) ([]ExtractionResult, error) {
	apiKey := os.Getenv("BASETEN_API_KEY")
	model := os.Getenv("BASETEN_MODEL")
	if model == "" {
		model = os.Getenv("NEPTUNE_LLM_MODEL")
	}
	if apiKey == "" || model == "" {
		return keywordExtraction(messages), nil
	}
	conversation := formatConversation(messages)

	results := make([]ExtractionResult, len(extractionAgents))
	errs := make([]error, len(extractionAgents))
	var wg sync.WaitGroup
	wg.Add(len(extractionAgents))
	for i, agent := range extractionAgents {
		go func(i int, agent ExtractionAgent) {
			defer wg.Done()
			res, err := runOneExtraction(ctx, apiKey, model, agent, conversation)
			results[i] = res
			errs[i] = err
		}(i, agent)
	}
	wg.Wait()

	// If any agent failed, fall back to keyword extraction for ALL agents.
	// ponytail: ceiling — all-or-nothing fallback. A partial LLM failure
	// (e.g. one agent times out) degrades to full keyword mode. Acceptable
	// for interview sessions where consistency matters more than coverage.
	for _, err := range errs {
		if err != nil {
			return keywordExtraction(messages), nil
		}
	}
	return results, nil
}

// runOneExtraction calls Baseten for a single agent and parses the JSON result.
func runOneExtraction(ctx context.Context, apiKey, model string, agent ExtractionAgent, conversation string) (ExtractionResult, error) {
	prompt := conversation + "\n\nReturn JSON only: {\"findings\": {...}, \"confidence\": 0.0-1.0, \"summary\": \"...\"}"
	reqBody := map[string]any{
		"model":      model,
		"max_tokens": 800,
		"messages": []map[string]string{
			{"role": "system", "content": agent.System},
			{"role": "user", "content": prompt},
		},
	}
	buf, err := json.Marshal(reqBody)
	if err != nil {
		return ExtractionResult{}, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, basetenEndpoint, bytes.NewReader(buf))
	if err != nil {
		return ExtractionResult{}, err
	}
	httpReq.Header.Set("content-type", "application/json")
	httpReq.Header.Set("authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 90 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return ExtractionResult{}, fmt.Errorf("extraction %s request: %w", agent.Type, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ExtractionResult{}, err
	}
	if resp.StatusCode >= 400 {
		return ExtractionResult{}, fmt.Errorf("extraction %s http %d: %s", agent.Type, resp.StatusCode, string(body))
	}
	var br basetenResponse
	if err := json.Unmarshal(body, &br); err != nil {
		return ExtractionResult{}, fmt.Errorf("extraction %s decode: %w", agent.Type, err)
	}
	if br.Error != nil {
		return ExtractionResult{}, fmt.Errorf("extraction %s error: %s", agent.Type, br.Error.Message)
	}
	if len(br.Choices) == 0 || br.Choices[0].Message.Content == "" {
		return ExtractionResult{}, fmt.Errorf("extraction %s: no content", agent.Type)
	}
	var out extractionResponse
	if err := json.Unmarshal([]byte(extractJSON(br.Choices[0].Message.Content)), &out); err != nil {
		return ExtractionResult{}, fmt.Errorf("extraction %s parse: %w", agent.Type, err)
	}
	if len(out.Findings) == 0 {
		out.Findings = json.RawMessage("{}")
	}
	return ExtractionResult{
		AgentType:  agent.Type,
		Findings:   out.Findings,
		Confidence: out.Confidence,
		Summary:    out.Summary,
	}, nil
}

// formatConversation renders the message list into a readable transcript.
func formatConversation(messages []ConversationMessage) string {
	var b []byte
	for _, m := range messages {
		label := m.Couple + " (" + m.Speaker + ")"
		b = append(b, []byte(fmt.Sprintf("%s: %s\n", label, sanitizeLLMInput(m.Text)))...)
	}
	return "Conversation transcript:\n" + string(b)
}

// keywordExtraction is the fallback when the LLM is unavailable. It scans the
// conversation for keywords relevant to each agent type and produces structured
// findings. Less nuanced than the LLM but always works.
// ponytail: ceiling — keyword matching can't infer sentiment or read between
// lines. Upgrade path: fix Baseten billing and the LLM path takes over.
func keywordExtraction(messages []ConversationMessage) []ExtractionResult {
	fullText := ""
	for _, m := range messages {
		fullText += m.Text + " "
	}
	lower := strings.ToLower(fullText)
	wordCount := len(strings.Fields(fullText))
	if wordCount == 0 {
		wordCount = 1
	}

	// --- relationship_stage ---
	stageFindings := map[string]any{}
	stageConf := 0.1
	stageSummary := "No strong relationship stage signals detected."
	if containsAny(lower, "engaged", "engagement", "proposed", "proposal", "ring") {
		stageFindings["stage"] = "engaged"
		stageFindings["evidence"] = "engagement-related keywords detected"
		stageConf = 0.7
		stageSummary = "Couple appears to be engaged — engagement keywords detected."
	} else if containsAny(lower, "married", "wedding", "got married", "our wedding") {
		stageFindings["stage"] = "married"
		stageFindings["evidence"] = "marriage keywords detected"
		stageConf = 0.6
		stageSummary = "Couple appears to be married — marriage keywords detected."
	} else if containsAny(lower, "dating", "boyfriend", "girlfriend", "relationship") {
		stageFindings["stage"] = "dating"
		stageFindings["evidence"] = "dating keywords detected"
		stageConf = 0.5
		stageSummary = "Couple appears to be dating."
	}

	// --- wedding_timeline ---
	timelineFindings := map[string]any{}
	timelineConf := 0.1
	timelineSummary := "No timeline signals detected."
	seasons := []string{"spring", "summer", "fall", "autumn", "winter"}
	for _, s := range seasons {
		if strings.Contains(lower, s) {
			timelineFindings["season"] = s
			timelineConf = 0.4
			timelineSummary = fmt.Sprintf("Mentions %s as a possible wedding season.", s)
		}
	}
	if containsAny(lower, "next year", "this year", "next month", "in 6 months", "2025", "2026", "2027") {
		timelineFindings["timeframe"] = "mentioned"
		timelineConf = max(timelineConf, 0.5)
		timelineSummary = "Wedding timeframe mentioned in conversation."
	}

	// --- vendor_interest ---
	vendorFindings := map[string]any{}
	vendorConf := 0.1
	vendorSummary := "No vendor preferences detected."
	vendorMap := map[string][]string{
		"photographer":  {"photographer", "photography", "photos", "photo"},
		"venue":         {"venue", "location", "place", "barn", "hotel", "beach"},
		"florist":       {"florist", "flowers", "floral", "bouquet"},
		"jeweler":       {"jeweler", "jewelry", "ring", "diamond"},
		"planner":       {"planner", "planning", "coordinator"},
		"videographer":  {"videographer", "video", "cinema", "film"},
		"cake":          {"cake", "bakery", "baker", "dessert"},
		"bridal_shop":   {"dress", "bridal", "gown", "attire"},
		"officiant":     {"officiant", "minister", "pastor", "priest", "rabbi"},
	}
	interestedVendors := []string{}
	for vendor, keywords := range vendorMap {
		if containsAny(lower, keywords...) {
			interestedVendors = append(interestedVendors, vendor)
		}
	}
	if len(interestedVendors) > 0 {
		vendorFindings["interested"] = interestedVendors
		vendorConf = float64(len(interestedVendors)) / 9.0
		if vendorConf > 0.8 {
			vendorConf = 0.8
		}
		vendorSummary = fmt.Sprintf("Interested in: %s.", strings.Join(interestedVendors, ", "))
	}

	// --- location ---
	locFindings := map[string]any{}
	locConf := 0.1
	locSummary := "No location preferences detected."
	venueTypes := []string{"beach", "barn", "hotel", "church", "garden", "outdoor", "indoor", "destination"}
	foundTypes := []string{}
	for _, vt := range venueTypes {
		if strings.Contains(lower, vt) {
			foundTypes = append(foundTypes, vt)
		}
	}
	if len(foundTypes) > 0 {
		locFindings["venue_type"] = foundTypes
		locConf = 0.4
		locSummary = fmt.Sprintf("Venue type preferences: %s.", strings.Join(foundTypes, ", "))
	}
	// Check for city mentions (capitalized words near "in" or "at")
	cities := extractCities(fullText)
	if len(cities) > 0 {
		locFindings["cities"] = cities
		locConf = max(locConf, 0.5)
		locSummary = fmt.Sprintf("Mentions cities: %s. %s", strings.Join(cities, ", "), locSummary)
	}

	// --- budget ---
	budgetFindings := map[string]any{}
	budgetConf := 0.1
	budgetSummary := "No budget signals detected."
	// Look for dollar amounts
	dollarAmounts := extractDollarAmounts(lower)
	if len(dollarAmounts) > 0 {
		budgetFindings["amounts"] = dollarAmounts
		budgetConf = 0.7
		budgetSummary = fmt.Sprintf("Budget amounts mentioned: %s.", strings.Join(dollarAmounts, ", "))
	}
	if containsAny(lower, "budget", "affordable", "expensive", "splurge", "cheap", "cost", "price", "k", "grand") {
		budgetFindings["price_sensitive"] = true
		budgetConf = max(budgetConf, 0.4)
		if budgetSummary == "No budget signals detected." {
			budgetSummary = "Budget-related language detected."
		}
	}

	return []ExtractionResult{
		{AgentType: "relationship_stage", Findings: toJSON(stageFindings), Confidence: stageConf, Summary: stageSummary},
		{AgentType: "wedding_timeline", Findings: toJSON(timelineFindings), Confidence: timelineConf, Summary: timelineSummary},
		{AgentType: "vendor_interest", Findings: toJSON(vendorFindings), Confidence: vendorConf, Summary: vendorSummary},
		{AgentType: "location", Findings: toJSON(locFindings), Confidence: locConf, Summary: locSummary},
		{AgentType: "budget", Findings: toJSON(budgetFindings), Confidence: budgetConf, Summary: budgetSummary},
	}
}

func containsAny(s string, words ...string) bool {
	for _, w := range words {
		if strings.Contains(s, w) {
			return true
		}
	}
	return false
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// extractCities finds capitalized words that might be city names (heuristic).
func extractCities(text string) []string {
	var cities []string
	words := strings.Fields(text)
	for i, w := range words {
		clean := strings.Trim(w, ",.;!?\"'()")
		if len(clean) < 3 {
			continue
		}
		// Check if it's capitalized and preceded by "in" or "at"
		if clean[0] >= 'A' && clean[0] <= 'Z' {
			if i > 0 {
				prev := strings.ToLower(strings.Trim(words[i-1], ",.;!?\"'()"))
				if prev == "in" || prev == "at" || prev == "near" || prev == "from" {
					cities = append(cities, clean)
				}
			}
		}
	}
	return cities
}

// extractDollarAmounts finds patterns like "30k", "25k", "$30,000", "30 grand".
func extractDollarAmounts(lower string) []string {
	var amounts []string
	// Match "Nk" or "N k" patterns (e.g. "30k", "25 k")
	for _, m := range dollarRe.FindAllString(lower, -1) {
		amounts = append(amounts, m)
	}
	return amounts
}

func toJSON(m map[string]any) json.RawMessage {
	b, _ := json.Marshal(m)
	if len(b) == 0 || string(b) == "null" {
		return json.RawMessage("{}")
	}
	return b
}
