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

const telegramAPIBaseURL = "https://api.telegram.org"

type BotClient struct {
	botToken   string
	httpClient *http.Client
}

func NewBotClient(botToken string) *BotClient {
	return &BotClient{
		botToken: strings.TrimSpace(botToken),
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (c *BotClient) SendMessage(chatID int64, text string) error {
	if c.botToken == "" {
		return fmt.Errorf("telegram bot token is required")
	}
	if chatID == 0 {
		return fmt.Errorf("chatID is required")
	}
	if strings.TrimSpace(text) == "" {
		return fmt.Errorf("text is required")
	}

	payload := map[string]any{
		"chat_id": chatID,
		"text":    text,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal telegram sendMessage payload: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	url := fmt.Sprintf("%s/bot%s/sendMessage", telegramAPIBaseURL, c.botToken)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build telegram sendMessage request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send telegram sendMessage request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read telegram sendMessage response: %w", err)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("telegram sendMessage failed (%d): %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var apiResp struct {
		OK bool `json:"ok"`
	}
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return fmt.Errorf("decode telegram sendMessage response: %w", err)
	}
	if !apiResp.OK {
		return fmt.Errorf("telegram sendMessage failed: api returned ok=false")
	}

	return nil
}
