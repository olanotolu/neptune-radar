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

const openaiEndpoint = "https://api.openai.com/v1/chat/completions"

// OpenAIInterpreter calls OpenAI Chat Completions (or an OpenAI-compatible gateway).
type OpenAIInterpreter struct {
	apiKey  string
	model   string
	baseURL string
	client  *http.Client
}

func NewOpenAIInterpreter() *OpenAIInterpreter {
	model := os.Getenv("OPENAI_MODEL")
	if model == "" {
		model = "gpt-4.1-mini"
	}
	base := os.Getenv("OPENAI_BASE_URL")
	if base == "" {
		base = openaiEndpoint
	}
	return &OpenAIInterpreter{
		apiKey:  os.Getenv("OPENAI_API_KEY"),
		model:   model,
		baseURL: base,
		client:  &http.Client{Timeout: 90 * time.Second},
	}
}

func (o *OpenAIInterpreter) Available() bool {
	return strings.TrimSpace(o.apiKey) != "" && strings.TrimSpace(o.model) != ""
}

func (o *OpenAIInterpreter) complete(ctx context.Context, system, prompt string) (string, LLMUsage, error) {
	if !o.Available() {
		return "", LLMUsage{}, fmt.Errorf("OPENAI_API_KEY not set")
	}
	body := map[string]any{
		"model": o.model,
		"messages": []map[string]string{
			{"role": "system", "content": system},
			{"role": "user", "content": prompt},
		},
		"max_tokens":  512,
		"temperature": 0.2,
	}
	buf, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, o.baseURL, bytes.NewReader(buf))
	if err != nil {
		return "", LLMUsage{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.apiKey)
	resp, err := o.client.Do(req)
	if err != nil {
		return "", LLMUsage{}, fmt.Errorf("openai request: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if resp.StatusCode >= 400 {
		return "", LLMUsage{}, fmt.Errorf("openai http %d: %s", resp.StatusCode, truncateLLM(string(raw), 300))
	}
	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage *struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
		} `json:"usage"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", LLMUsage{}, fmt.Errorf("openai decode: %w", err)
	}
	if parsed.Error != nil && parsed.Error.Message != "" {
		return "", LLMUsage{}, fmt.Errorf("openai error: %s", parsed.Error.Message)
	}
	if len(parsed.Choices) == 0 || parsed.Choices[0].Message.Content == "" {
		return "", LLMUsage{}, fmt.Errorf("openai empty response")
	}
	var u LLMUsage
	if parsed.Usage != nil {
		u.PromptTokens = parsed.Usage.PromptTokens
		u.CompletionTokens = parsed.Usage.CompletionTokens
	}
	return parsed.Choices[0].Message.Content, u, nil
}

func (o *OpenAIInterpreter) InterpretSignal(ctx context.Context, req SignalRequest) (Interpretation, error) {
	raw, usage, err := o.complete(ctx, signalSystemPrompt, formatSignalPrompt(req))
	if err != nil {
		return Interpretation{}, err
	}
	var out Interpretation
	if err := json.Unmarshal([]byte(extractJSON(raw)), &out); err != nil {
		return Interpretation{}, fmt.Errorf("parse openai interpretation: %w", err)
	}
	out.Source = "openai:" + o.model
	out.PromptTokens, out.CompletionTokens = usage.PromptTokens, usage.CompletionTokens
	return out, nil
}

func (o *OpenAIInterpreter) DraftCopy(ctx context.Context, req CopyRequest) (Copy, error) {
	raw, _, err := o.complete(ctx, copySystemPrompt, formatCopyPrompt(req))
	if err != nil {
		return Copy{}, err
	}
	var out Copy
	if err := json.Unmarshal([]byte(extractJSON(raw)), &out); err != nil {
		return Copy{}, fmt.Errorf("parse openai copy: %w", err)
	}
	return out, nil
}

func truncateLLM(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
