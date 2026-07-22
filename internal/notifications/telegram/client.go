package telegram

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

type HTTPBotClient struct {
	token   string
	baseURL string
	client  *http.Client
}

type apiResponse[T any] struct {
	OK          bool   `json:"ok"`
	Result      T      `json:"result"`
	Description string `json:"description"`
}

type webhookRequest struct {
	URL         string `json:"url"`
	SecretToken string `json:"secret_token,omitempty"`
}

func NewHTTPBotClient(token, baseURL string) *HTTPBotClient {
	if baseURL == "" {
		baseURL = "https://api.telegram.org"
	}
	return &HTTPBotClient{
		token:   token,
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  &http.Client{Timeout: 45 * time.Second},
	}
}

func (c *HTTPBotClient) SendMessage(ctx context.Context, req SendMessageRequest) error {
	return c.postJSON(ctx, "sendMessage", req, nil)
}

func (c *HTTPBotClient) SetWebhook(ctx context.Context, webhookURL, secret string) error {
	return c.postJSON(ctx, "setWebhook", webhookRequest{URL: webhookURL, SecretToken: secret}, nil)
}

func (c *HTTPBotClient) DeleteWebhook(ctx context.Context) error {
	return c.postJSON(ctx, "deleteWebhook", map[string]any{}, nil)
}

func (c *HTTPBotClient) GetUpdates(ctx context.Context, req GetUpdatesRequest) ([]Update, error) {
	var result []Update
	if err := c.postJSON(ctx, "getUpdates", req, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *HTTPBotClient) GetMe(ctx context.Context) (BotInfo, error) {
	var result BotInfo
	if err := c.postJSON(ctx, "getMe", map[string]any{}, &result); err != nil {
		return BotInfo{}, err
	}
	return result, nil
}

func (c *HTTPBotClient) postJSON(ctx context.Context, method string, payload any, out any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal telegram %s request: %w", method, err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/bot"+c.token+"/"+method, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build telegram %s request: %w", method, err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("execute telegram %s request: %w", method, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("telegram %s returned status %d", method, resp.StatusCode)
	}

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read telegram %s response: %w", method, err)
	}

	var envelope apiResponse[json.RawMessage]
	if err := json.Unmarshal(responseBody, &envelope); err != nil {
		return fmt.Errorf("decode telegram %s response: %w", method, err)
	}
	if !envelope.OK {
		return fmt.Errorf("telegram %s failed: %s", method, envelope.Description)
	}

	if out == nil {
		return nil
	}

	if err := json.Unmarshal(envelope.Result, out); err != nil {
		return fmt.Errorf("decode telegram %s result: %w", method, err)
	}
	return nil
}
