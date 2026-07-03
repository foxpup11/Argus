// Package llm provides a lightweight LLM HTTP client supporting both
// OpenAI-compatible and Anthropic Messages API formats.
package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// APIType 表示 LLM 提供商的 API 协议类型。
type APIType string

const (
	APITypeOpenAI    APIType = "openai"
	APITypeAnthropic APIType = "anthropic"
)

// ProviderConfig holds the configuration for an LLM provider.
type ProviderConfig struct {
	Name    string  `json:"name"`    // display label, e.g. "MiMo V2.5"
	APIKey  string  `json:"apiKey"`  // user-supplied API key
	BaseURL string  `json:"baseUrl"` // API base URL, e.g. "https://api.xiaomimimo.com/anthropic"
	Model   string  `json:"model"`   // model identifier, e.g. "MiniMax-M2.5"
	APIType APIType `json:"apiType"` // API protocol: "openai" or "anthropic"
	Enabled bool    `json:"enabled"` // whether LLM enhancement is active
}

// PresetProviders returns the built-in provider presets.
func PresetProviders() map[string]ProviderConfig {
	return map[string]ProviderConfig{
		"mimo": {
			Name:    "MiMo V2.5",
			BaseURL: "https://api.xiaomimimo.com/anthropic",
			Model:   "MiniMax-M2.5",
			APIType: APITypeAnthropic,
		},
		"deepseek": {
			Name:    "DeepSeek",
			BaseURL: "https://api.deepseek.com/v1",
			Model:   "deepseek-chat",
			APIType: APITypeOpenAI,
		},
	}
}

// ResolveAPIType 根据 provider key 推导 API 类型。
func ResolveAPIType(provider string) APIType {
	presets := PresetProviders()
	if cfg, ok := presets[provider]; ok {
		return cfg.APIType
	}
	return APITypeOpenAI // custom 默认 OpenAI 兼容
}

// ChatMessage represents a single message in a chat completion request.
type ChatMessage struct {
	Role    string `json:"role"`    // "system", "user", "assistant"
	Content string `json:"content"` // message body
}

// ChatResponse is the parsed result of a chat completion call.
type ChatResponse struct {
	Content string `json:"content"` // the assistant's text response
	Model   string `json:"model"`   // model that generated the response
	Usage   Usage  `json:"usage"`   // token usage info
}

// Usage records token consumption for a single API call.
type Usage struct {
	InputTokens  int `json:"inputTokens"`
	OutputTokens int `json:"outputTokens"`
}

// Client is a minimal HTTP client for LLM chat completions.
type Client struct {
	httpClient *http.Client
	cfg        ProviderConfig
}

// NewClient creates a new LLM client with the given provider configuration.
func NewClient(cfg ProviderConfig) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 60 * time.Second},
		cfg:        cfg,
	}
}

// ChatCompletion sends a chat completion request and returns the response.
// It routes to the appropriate API implementation based on cfg.APIType.
func (c *Client) ChatCompletion(ctx context.Context, messages []ChatMessage) (*ChatResponse, error) {
	switch c.cfg.APIType {
	case APITypeAnthropic:
		return c.anthropicChatCompletion(ctx, messages)
	default:
		return c.openaiChatCompletion(ctx, messages)
	}
}

// ============================================
// OpenAI-compatible API (/v1/chat/completions)
// ============================================

type openaiChatRequest struct {
	Model          string        `json:"model"`
	Messages       []ChatMessage `json:"messages"`
	Temperature    float64       `json:"temperature,omitempty"`
	ResponseFormat *responseFmt  `json:"response_format,omitempty"`
}

type responseFmt struct {
	Type string `json:"type"` // "json_object" for structured output
}

type openaiChatResponse struct {
	Model   string `json:"model"`
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		InputTokens  int `json:"prompt_tokens"`
		OutputTokens int `json:"completion_tokens"`
	} `json:"usage"`
}

func (c *Client) openaiChatCompletion(ctx context.Context, messages []ChatMessage) (*ChatResponse, error) {
	endpoint := strings.TrimRight(c.cfg.BaseURL, "/") + "/chat/completions"

	reqBody := openaiChatRequest{
		Model:          c.cfg.Model,
		Messages:       messages,
		Temperature:    0.3,
		ResponseFormat: &responseFmt{Type: "json_object"},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("LLM API 请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("LLM API 返回错误状态 %d: %s", resp.StatusCode, string(respBytes))
	}

	return parseOpenAIResponse(respBytes, c.cfg.Model)
}

func parseOpenAIResponse(body []byte, fallbackModel string) (*ChatResponse, error) {
	var raw openaiChatResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("解析LLM响应JSON失败: %w", err)
	}

	if len(raw.Choices) == 0 {
		return nil, fmt.Errorf("LLM响应中没有choices")
	}

	model := raw.Model
	if model == "" {
		model = fallbackModel
	}

	return &ChatResponse{
		Content: raw.Choices[0].Message.Content,
		Model:   model,
		Usage: Usage{
			InputTokens:  raw.Usage.InputTokens,
			OutputTokens: raw.Usage.OutputTokens,
		},
	}, nil
}

// ============================================
// Anthropic Messages API (/v1/messages)
// ============================================

type anthropicRequest struct {
	Model     string           `json:"model"`
	MaxTokens int              `json:"max_tokens"`
	System    string           `json:"system,omitempty"`   // top-level system prompt
	Messages  []anthropicMsg   `json:"messages"`
}

type anthropicMsg struct {
	Role    string `json:"role"`    // "user" or "assistant"
	Content string `json:"content"` // text content
}

type anthropicResponse struct {
	ID      string `json:"id"`
	Model   string `json:"model"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

func (c *Client) anthropicChatCompletion(ctx context.Context, messages []ChatMessage) (*ChatResponse, error) {
	endpoint := strings.TrimRight(c.cfg.BaseURL, "/") + "/v1/messages"

	// 分离 system prompt 和对话消息
	var systemPrompt string
	var chatMessages []anthropicMsg
	for _, msg := range messages {
		if msg.Role == "system" {
			// Anthropic 只支持单个 system prompt，合并多个
			if systemPrompt != "" {
				systemPrompt += "\n\n"
			}
			systemPrompt += msg.Content
		} else {
			chatMessages = append(chatMessages, anthropicMsg{
				Role:    msg.Role,
				Content: msg.Content,
			})
		}
	}

	reqBody := anthropicRequest{
		Model:     c.cfg.Model,
		MaxTokens: 4096,
		System:    systemPrompt,
		Messages:  chatMessages,
	}
	// 如果只有 system 消息（如 TestLLMConnection），添加一个 user 占位消息
	if len(chatMessages) == 0 {
		reqBody.Messages = []anthropicMsg{
			{Role: "user", Content: "Reply with exactly the word OK and nothing else."},
		}
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.cfg.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("LLM API 请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("LLM API 返回错误状态 %d: %s", resp.StatusCode, string(respBytes))
	}

	return parseAnthropicResponse(respBytes, c.cfg.Model)
}

func parseAnthropicResponse(body []byte, fallbackModel string) (*ChatResponse, error) {
	var raw anthropicResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("解析LLM响应JSON失败: %w", err)
	}

	model := raw.Model
	if model == "" {
		model = fallbackModel
	}

	// 合并所有 text content blocks
	var textParts []string
	for _, block := range raw.Content {
		if block.Type == "text" && block.Text != "" {
			textParts = append(textParts, block.Text)
		}
	}
	content := strings.Join(textParts, "\n")

	return &ChatResponse{
		Content: content,
		Model:   model,
		Usage: Usage{
			InputTokens:  raw.Usage.InputTokens,
			OutputTokens: raw.Usage.OutputTokens,
		},
	}, nil
}

// maskKey returns a masked version of the API key for safe logging.
func maskKey(key string) string {
	if len(key) <= 8 {
		return "***"
	}
	return key[:4] + "***" + key[len(key)-4:]
}

// Config returns a copy of the provider configuration.
func (c *Client) Config() ProviderConfig {
	return c.cfg
}
