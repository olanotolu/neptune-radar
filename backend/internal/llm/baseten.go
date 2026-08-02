package llm

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

const basetenEndpoint = "https://api.baseten.co/v1/chat/completions"

// BasetenInterpreter calls a model deployed on Baseten via their
// OpenAI-compatible chat completions endpoint.
type BasetenInterpreter struct {
	apiKey string
	model  string
	client *http.Client
}

func NewBasetenInterpreter() *BasetenInterpreter {
	model := os.Getenv("BASETEN_MODEL")
	if model == "" {
		model = os.Getenv("NEPTUNE_LLM_MODEL")
	}
	return &BasetenInterpreter{
		apiKey: os.Getenv("BASETEN_API_KEY"),
		model:  model,
		client: &http.Client{Timeout: 60 * time.Second},
	}
}

func (b *BasetenInterpreter) Available() bool { return b.apiKey != "" && b.model != "" }

type basetenMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type basetenRequest struct {
	Model     string           `json:"model"`
	MaxTokens int              `json:"max_tokens"`
	Messages  []basetenMessage `json:"messages"`
}

type basetenResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage *struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage,omitempty"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

func (b *BasetenInterpreter) complete(ctx context.Context, system, prompt string) (string, LLMUsage, error) {
	if b.apiKey == "" {
		return "", LLMUsage{}, fmt.Errorf("BASETEN_API_KEY not set")
	}
	if b.model == "" {
		return "", LLMUsage{}, fmt.Errorf("BASETEN_MODEL not set")
	}

	reqBody := basetenRequest{
		Model:     b.model,
		MaxTokens: 512,
		Messages: []basetenMessage{
			{Role: "system", Content: system},
			{Role: "user", Content: prompt},
		},
	}
	buf, err := json.Marshal(reqBody)
	if err != nil {
		return "", LLMUsage{}, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, basetenEndpoint, bytes.NewReader(buf))
	if err != nil {
		return "", LLMUsage{}, err
	}
	httpReq.Header.Set("content-type", "application/json")
	httpReq.Header.Set("authorization", "Bearer "+b.apiKey)

	resp, err := b.client.Do(httpReq)
	if err != nil {
		return "", LLMUsage{}, fmt.Errorf("baseten request: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", LLMUsage{}, err
	}
	if resp.StatusCode >= 400 {
		return "", LLMUsage{}, fmt.Errorf("baseten http %d: %s", resp.StatusCode, string(body))
	}
	var br basetenResponse
	if err := json.Unmarshal(body, &br); err != nil {
		return "", LLMUsage{}, fmt.Errorf("decode baseten response: %w", err)
	}
	if br.Error != nil {
		return "", LLMUsage{}, fmt.Errorf("baseten error: %s", br.Error.Message)
	}
	if len(br.Choices) == 0 || br.Choices[0].Message.Content == "" {
		return "", LLMUsage{}, fmt.Errorf("baseten response had no content")
	}
	var u LLMUsage
	if br.Usage != nil {
		u.PromptTokens = br.Usage.PromptTokens
		u.CompletionTokens = br.Usage.CompletionTokens
	}
	return br.Choices[0].Message.Content, u, nil
}

func (b *BasetenInterpreter) InterpretSignal(ctx context.Context, req SignalRequest) (Interpretation, error) {
	raw, usage, err := b.complete(ctx, signalSystemPrompt, formatSignalPrompt(req))
	if err != nil {
		return Interpretation{}, err
	}
	var out Interpretation
	if err := json.Unmarshal([]byte(extractJSON(raw)), &out); err != nil {
		return Interpretation{}, fmt.Errorf("parse baseten interpretation: %w", err)
	}
	out.Source = "baseten:" + b.model
	out.PromptTokens, out.CompletionTokens = usage.PromptTokens, usage.CompletionTokens
	return out, nil
}

func (b *BasetenInterpreter) DraftCopy(ctx context.Context, req CopyRequest) (Copy, error) {
	prompt := formatCopyPrompt(req)
	raw, _, err := b.complete(ctx, copySystemPrompt, prompt)
	if err != nil {
		return Copy{}, err
	}
	var out Copy
	if err := json.Unmarshal([]byte(extractJSON(raw)), &out); err != nil {
		return Copy{}, fmt.Errorf("parse baseten copy: %w", err)
	}
	return out, nil
}
