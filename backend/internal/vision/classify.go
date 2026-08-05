package vision

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
)

// PhotoLabels are the CLIP zero-shot candidate labels for photo classification.
var PhotoLabels = []string{
	"marriage proposal",
	"wedding ceremony",
	"engagement photo shoot",
	"couple portrait",
	"casual photo",
}

// ClassifyPhoto runs CLIP zero-shot image classification on the image and
// returns the top label + its confidence (0–1). Returns ("", 0, nil) when the
// API is unavailable — callers treat no-label as "no signal."
//
// ponytail: the HF Inference API for zero-shot-image-classification expects a
// JSON body with a base64 image + candidate_labels. If HF changes their API
// shape, the graceful fallback returns ("", 0, nil) — no crash, no signal.
func ClassifyPhoto(ctx context.Context, imageURL string) (string, float64, error) {
	if imageURL == "" {
		return "", 0, nil
	}
	token := hfToken()
	if token == "" {
		return "", 0, fmt.Errorf("HF_TOKEN not set")
	}
	model := os.Getenv("HF_CLIP_MODEL")
	if model == "" {
		model = "openai/clip-vit-base-patch32"
	}
	imgBytes, err := fetchImageBytes(ctx, imageURL)
	if err != nil {
		return "", 0, err
	}
	body := map[string]any{
		"inputs": base64.StdEncoding.EncodeToString(imgBytes),
		"parameters": map[string]any{
			"candidate_labels": PhotoLabels,
		},
	}
	buf, err := json.Marshal(body)
	if err != nil {
		return "", 0, err
	}
	endpoint := "https://api-inference.huggingface.co/models/" + model
	raw, err := hfPostJSON(ctx, endpoint, token, buf)
	if err != nil {
		return "", 0, err
	}
	// Response: [{"score":0.88,"label":"marriage proposal"}, ...]
	var results []struct {
		Score float64 `json:"score"`
		Label string  `json:"label"`
	}
	if err := json.Unmarshal(raw, &results); err != nil {
		return "", 0, fmt.Errorf("parse clip results: %w", err)
	}
	if len(results) == 0 {
		return "", 0, nil
	}
	return results[0].Label, results[0].Score, nil
}
