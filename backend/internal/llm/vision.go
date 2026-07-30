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

// VisionClassifier is the visual-signal producer: given a post image it
// returns which of the known visual engagement signals (ring, ring box,
// "marry me" sign, on-screen engagement text, ...) are visible. It is a
// signal PRODUCER, not a decision maker — its output lands in the post
// payload and is scored deterministically like everything else. It never
// identifies people, and it is only ever called on posts that already passed
// the cheap language/vendor pre-filter (see the watchtower worker).
type VisionClassifier interface {
	ClassifyVisualSignals(ctx context.Context, imageURL string) ([]string, error)
}

// NoopVision is the default when no vision model is configured: visual
// evidence simply never fires, and the points table degrades gracefully.
type NoopVision struct{}

func (NoopVision) ClassifyVisualSignals(ctx context.Context, imageURL string) ([]string, error) {
	return nil, nil
}

// BasetenVision classifies post images with a multimodal model hosted on
// Baseten's OpenAI-compatible chat completions endpoint.
type BasetenVision struct {
	apiKey string
	model  string
}

func NewBasetenVision() *BasetenVision {
	model := os.Getenv("BASETEN_VISION_MODEL")
	if model == "" {
		model = os.Getenv("BASETEN_MODEL")
	}
	return &BasetenVision{apiKey: os.Getenv("BASETEN_API_KEY"), model: model}
}

func (b *BasetenVision) Available() bool { return b.apiKey != "" && b.model != "" }

const visionSystemPrompt = `You are a visual signal classifier inside a social-listening pipeline for engagement photography.
Look at the image and report which of these signals are clearly visible:
ring (an engagement/wedding ring on a hand), ring_box, proposal_scene (one person proposing to another),
marry_me_sign, engagement_party_signage, champagne_celebration, on_screen_text_engaged
(on-screen text announcing an engagement, e.g. "POV: you just got engaged"),
countdown_screenshot (wedding countdown), save_the_date_card,
couple_portrait (two people clearly shown together as a couple — faces/bodies, not just hands),
people_present (one or more people clearly visible in frame),
venue_only (interior/exterior of a venue, chandelier, tables, architecture — NO clear people as subject),
product_only (rings/flowers/decor as product still life without people as the subject).
Be conservative. Prefer couple_portrait when two people are the clear subject.
Respond with ONLY a JSON array of matching labels, e.g. ["couple_portrait","ring"], or [] if none.
Never describe or identify people by name.`

func (b *BasetenVision) ClassifyVisualSignals(ctx context.Context, imageURL string) ([]string, error) {
	if !b.Available() {
		return nil, fmt.Errorf("BASETEN_API_KEY/BASETEN_VISION_MODEL not set")
	}
	type contentPart struct {
		Type     string `json:"type"`
		Text     string `json:"text,omitempty"`
		ImageURL *struct {
			URL string `json:"url"`
		} `json:"image_url,omitempty"`
	}
	type message struct {
		Role    string        `json:"role"`
		Content []contentPart `json:"content"`
	}
	reqBody := map[string]any{
		"model":      b.model,
		"max_tokens": 128,
		"messages": []message{
			{Role: "system", Content: []contentPart{{Type: "text", Text: visionSystemPrompt}}},
			{Role: "user", Content: []contentPart{
				{Type: "text", Text: "Which signals are visible in this post image?"},
				{Type: "image_url", ImageURL: &struct {
					URL string `json:"url"`
				}{URL: imageURL}},
			}},
		},
	}
	raw, err := b.complete(ctx, reqBody)
	if err != nil {
		return nil, err
	}
	var labels []string
	if err := json.Unmarshal([]byte(extractJSONArray(raw)), &labels); err != nil {
		return nil, fmt.Errorf("parse vision labels: %w", err)
	}
	return labels, nil
}

func (b *BasetenVision) complete(ctx context.Context, reqBody map[string]any) (string, error) {
	buf, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, basetenEndpoint, bytes.NewReader(buf))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("content-type", "application/json")
	httpReq.Header.Set("authorization", "Bearer "+b.apiKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("baseten vision request: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var br basetenResponse
	if err := json.Unmarshal(body, &br); err != nil {
		return "", fmt.Errorf("decode baseten vision response: %w", err)
	}
	if br.Error != nil {
		return "", fmt.Errorf("baseten vision error: %s", br.Error.Message)
	}
	if len(br.Choices) == 0 || br.Choices[0].Message.Content == "" {
		return "", fmt.Errorf("baseten vision response had no content")
	}
	return br.Choices[0].Message.Content, nil
}

// extractJSONArray mirrors extractJSON for array-shaped model output.
func extractJSONArray(s string) string {
	start := strings.IndexByte(s, '[')
	end := strings.LastIndexByte(s, ']')
	if start == -1 || end == -1 || end < start {
		return s
	}
	return s[start : end+1]
}

// NewVisionClassifier picks the configured vision classifier, defaulting to
// the no-op so the system runs (degraded) without any vision credentials.
func NewVisionClassifier() VisionClassifier {
	if v := NewBasetenVision(); v.Available() {
		return v
	}
	return NoopVision{}
}
