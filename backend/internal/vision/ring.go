// Package vision provides specialized visual analysis for engagement
// detection: YOLOv8 ring detection and CLIP zero-shot photo classification
// via the HuggingFace Inference API, plus a pure-Go dispersion metric for
// relationship scoring. All hosted-API calls fall back to zero confidence on
// error so ingest never crashes when the API is down or unconfigured.
package vision

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

var hfClient = &http.Client{Timeout: 30 * time.Second}

// hfToken returns the HuggingFace API token from the environment, or "".
func hfToken() string { return os.Getenv("HF_TOKEN") }

// fetchImageBytes downloads the image at imageURL. HF vision endpoints take
// raw image bytes, not a URL.
func fetchImageBytes(ctx context.Context, imageURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, imageURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := hfClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch image: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch image: status %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

// hfPostRaw posts raw bytes (image data) to a HF Inference API model endpoint.
func hfPostRaw(ctx context.Context, endpoint, token string, body []byte) (json.RawMessage, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "image/jpeg") // ponytail: Instagram serves JPEG; HF infers from bytes anyway
	return hfDo(ctx, req)
}

// hfPostJSON posts a JSON body to a HF Inference API model endpoint.
func hfPostJSON(ctx context.Context, endpoint, token string, body []byte) (json.RawMessage, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	return hfDo(ctx, req)
}

func hfDo(ctx context.Context, req *http.Request) (json.RawMessage, error) {
	resp, err := hfClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("hf inference: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("hf inference: status %d: %s", resp.StatusCode, string(b))
	}
	return io.ReadAll(resp.Body)
}

// hfInference posts image bytes to a HuggingFace Inference API model and
// returns the raw JSON response body. Returns an error if the token is
// missing or the request fails — callers fall back gracefully.
func hfInference(ctx context.Context, model, imageURL string) (json.RawMessage, error) {
	token := hfToken()
	if token == "" {
		return nil, fmt.Errorf("HF_TOKEN not set")
	}
	imgBytes, err := fetchImageBytes(ctx, imageURL)
	if err != nil {
		return nil, err
	}
	endpoint := "https://api-inference.huggingface.co/models/" + model
	return hfPostRaw(ctx, endpoint, token, imgBytes)
}

// DetectRing calls a YOLOv8 ring-detection model on the image and returns a
// confidence score 0–1 that an engagement ring is visible. Returns 0 (not an
// error) when no ring is detected or the API is unavailable — callers should
// treat 0 as "no signal," not a failure.
func DetectRing(ctx context.Context, imageURL string) (float64, error) {
	if imageURL == "" {
		return 0, nil
	}
	model := os.Getenv("HF_RING_MODEL")
	if model == "" {
		model = "humaisj/ring-detector-yolov8"
	}
	raw, err := hfInference(ctx, model, imageURL)
	if err != nil {
		return 0, err // caller logs, scores 0
	}
	// YOLOv8 object-detection response: [{"score":0.95,"label":"ring","box":{...}}, ...]
	var detections []struct {
		Score float64 `json:"score"`
		Label string  `json:"label"`
	}
	if err := json.Unmarshal(raw, &detections); err != nil {
		return 0, fmt.Errorf("parse ring detections: %w", err)
	}
	var best float64
	for _, d := range detections {
		if d.Score > best {
			best = d.Score // highest-confidence detection of any label
		}
	}
	return best, nil
}
