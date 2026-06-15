package ai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Message is a single turn in the conversation history.
type Message struct {
	Role    string `json:"role"`    // "system" | "user" | "assistant"
	Content string `json:"content"`
}

// GenerateRequest is the payload sent to Ollama /api/chat.
type GenerateRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
	Options  *Options  `json:"options,omitempty"`
}

// Options holds optional Ollama generation parameters.
type Options struct {
	Temperature float64 `json:"temperature,omitempty"`
	TopP        float64 `json:"top_p,omitempty"`
	NumCtx      int     `json:"num_ctx,omitempty"`
}

// StreamChunk is a single streamed response object from Ollama.
type StreamChunk struct {
	Model     string  `json:"model"`
	CreatedAt string  `json:"created_at"`
	Message   Message `json:"message"`
	Done      bool    `json:"done"`
}

// ModelInfo is a single model entry returned by /api/tags.
type ModelInfo struct {
	Name       string    `json:"name"`
	ModifiedAt time.Time `json:"modified_at"`
	Size       int64     `json:"size"`
}

// Client talks to Ollama.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// New creates an Ollama client pointed at baseURL (e.g. "http://localhost:11434").
func New(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 0, // no global timeout — streaming can be long
		},
	}
}

// ListModels fetches all locally available models from Ollama.
func (c *Client) ListModels(ctx context.Context) ([]ModelInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/tags", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ollama unreachable: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Models []ModelInfo `json:"models"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("ollama /api/tags decode: %w", err)
	}
	return result.Models, nil
}

// StreamChat sends messages to Ollama and calls onChunk for every streamed
// token. The function returns when the stream is complete or ctx is cancelled.
//
// onChunk receives the partial assistant text. Return a non-nil error from
// onChunk to abort early.
func (c *Client) StreamChat(
	ctx context.Context,
	model string,
	messages []Message,
	opts *Options,
	onChunk func(token string) error,
) error {
	payload := GenerateRequest{
		Model:    model,
		Messages: messages,
		Stream:   true,
		Options:  opts,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("ollama /api/chat: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ollama error %d: %s", resp.StatusCode, raw)
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var chunk StreamChunk
		if err := json.Unmarshal([]byte(line), &chunk); err != nil {
			continue // skip malformed lines
		}

		if chunk.Message.Content != "" {
			if err := onChunk(chunk.Message.Content); err != nil {
				return err
			}
		}

		if chunk.Done {
			break
		}
	}

	return scanner.Err()
}

// Chat is a convenience non-streaming call that returns the full response.
func (c *Client) Chat(
	ctx context.Context,
	model string,
	messages []Message,
	opts *Options,
) (string, error) {
	var sb strings.Builder
	err := c.StreamChat(ctx, model, messages, opts, func(token string) error {
		sb.WriteString(token)
		return nil
	})
	return sb.String(), err
}
