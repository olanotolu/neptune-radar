package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const defaultModel = "claude-sonnet-5"
const anthropicVersion = "2023-06-01"
const anthropicEndpoint = "https://api.anthropic.com/v1/messages"

// ClaudeInterpreter calls the real Anthropic Messages API. It is the
// "model proposes" half of the system: everything it returns is a
// suggestion that pipeline/policy independently validates and clamps.
type ClaudeInterpreter struct {
	apiKey string
	model  string
	client *http.Client
}

func NewClaudeInterpreter() *ClaudeInterpreter {
	model := os.Getenv("NEPTUNE_LLM_MODEL")
	if model == "" {
		model = defaultModel
	}
	return &ClaudeInterpreter{
		apiKey: os.Getenv("ANTHROPIC_API_KEY"),
		model:  model,
		client: &http.Client{Timeout: 20 * time.Second},
	}
}

func (c *ClaudeInterpreter) Available() bool { return c.apiKey != "" }

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	System    string             `json:"system"`
	Messages  []anthropicMessage `json:"messages"`
}

type anthropicResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (c *ClaudeInterpreter) complete(ctx context.Context, system, prompt string) (string, error) {
	if c.apiKey == "" {
		return "", fmt.Errorf("ANTHROPIC_API_KEY not set")
	}
	reqBody := anthropicRequest{
		Model:     c.model,
		MaxTokens: 512,
		System:    system,
		Messages:  []anthropicMessage{{Role: "user", Content: prompt}},
	}
	buf, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, anthropicEndpoint, bytes.NewReader(buf))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("content-type", "application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", anthropicVersion)

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("anthropic request: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var ar anthropicResponse
	if err := json.Unmarshal(body, &ar); err != nil {
		return "", fmt.Errorf("decode anthropic response: %w", err)
	}
	if ar.Error != nil {
		return "", fmt.Errorf("anthropic error: %s", ar.Error.Message)
	}
	if len(ar.Content) == 0 {
		return "", fmt.Errorf("anthropic response had no content")
	}
	return ar.Content[0].Text, nil
}

func extractJSON(s string) string {
	start := strings.IndexByte(s, '{')
	end := strings.LastIndexByte(s, '}')
	if start == -1 || end == -1 || end < start {
		return s
	}
	return s[start : end+1]
}

const signalSystemPrompt = `You are the Relationship Analyst inside Neptune's social-signal pipeline.
You receive one candidate life-event signal that deterministic code has already
flagged as worth interpreting. Judge how strongly the actual language supports
the hypothesis. Captions matter more than hashtags: understand semantic
variations ("he asked and I said forever") instead of relying on exact phrase
matching — deterministic matching has already had its say; your added value is
judging the variations and the context around them. Be conservative: precision
matters more than recall, and you must never assert a breakup happened — only
that context changed. Respond with
ONLY a JSON object: {"confidence": 0.0-1.0, "proposed_stage": "engaged"|"status_uncertain"|"unknown", "rationale": "one sentence"}`

func formatSignalPrompt(req SignalRequest) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Candidate event type: %s\nObservation type: %s\n", req.CandidateEventType, req.ObservationType))
	if text := fence("caption_or_bio", req.Text); text != "" {
		sb.WriteString(text + "\n")
	}
	sb.WriteString(fmt.Sprintf("Handle: %s\nPartner handle: %s\nPrior relationship stage: %s\n", req.Handle, req.PartnerHandle, req.PriorStage))
	if len(req.ExistingEvidence) > 0 {
		sanitized := make([]string, len(req.ExistingEvidence))
		for i, e := range req.ExistingEvidence {
			sanitized[i] = sanitizeLLMInput(e)
		}
		sb.WriteString(fmt.Sprintf("Existing evidence: %v\n", sanitized))
	}
	if req.SignalContext != "" {
		sb.WriteString("Deterministic signal-vocabulary matches: " + sanitizeLLMInput(req.SignalContext))
	}
	return sb.String()
}

func (c *ClaudeInterpreter) InterpretSignal(ctx context.Context, req SignalRequest) (Interpretation, error) {
	prompt := formatSignalPrompt(req)
	raw, err := c.complete(ctx, signalSystemPrompt, prompt)
	if err != nil {
		return Interpretation{}, err
	}
	var out Interpretation
	if err := json.Unmarshal([]byte(extractJSON(raw)), &out); err != nil {
		return Interpretation{}, fmt.Errorf("parse claude interpretation: %w", err)
	}
	out.Source = "claude:" + c.model
	return out, nil
}

const copySystemPrompt = `You are the Conversation Agent inside Neptune, a couples' operating system for
prenups/estate/tax/finance. Write two pieces of copy for one recommended action:
1. internal_note: for Neptune's internal team only — can be funny/blunt (this is an internal joke: "PagerDuty for couples").
   If action type is "review" (a high-confidence engagement prospect), format it as a "NEW NEPTUNE PROSPECT" card: People,
   Event, Detected/Location if given, Evidence bullets, then both "Engagement confidence: X%" and
   "Partner-match confidence: Y%" AS SEPARATE numbers (never average them into one), and
   "Recommended action: Human review". A confident engagement caption tagging the wrong second person is a
   real failure mode — the two scores exist so the reader can see that distinction, not just a friendly write-up.
   If action type is "investigate", the prospect scored below the create-prospect bar but above the discard
   bar — same card format, but titled as needing investigation, and "Recommended action: Human investigation"
   (verify the couple and the event are real before any outreach is considered).
2. customer_facing: sent to the actual customer if a human approves it. This MUST be calm, neutral,
respectful, brief, and must NEVER claim or imply a breakup or relationship failure — if the action type
is "concierge_review", ask an open, no-pressure question about whether to continue/pause/close their
Neptune process, with no explanation required from them.
Respond with ONLY a JSON object: {"internal_note": "...", "customer_facing": "..."}`

func (c *ClaudeInterpreter) DraftCopy(ctx context.Context, req CopyRequest) (Copy, error) {
	prompt := formatCopyPrompt(req)
	raw, err := c.complete(ctx, copySystemPrompt, prompt)
	if err != nil {
		return Copy{}, err
	}
	var out Copy
	if err := json.Unmarshal([]byte(extractJSON(raw)), &out); err != nil {
		return Copy{}, fmt.Errorf("parse claude copy: %w", err)
	}
	return out, nil
}
