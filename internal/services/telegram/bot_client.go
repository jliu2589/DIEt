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

const (
	telegramAPIBaseURL      = "https://api.telegram.org"
	telegramRequestTimeout  = 15 * time.Second
	maxTelegramErrorPreview = 512
)

type BotClient struct {
	botToken   string
	httpClient *http.Client
}

type sendMessageRequest struct {
	ChatID int64  `json:"chat_id"`
	Text   string `json:"text"`
}

type telegramAPIResponse struct {
	OK          bool   `json:"ok"`
	ErrorCode   int    `json:"error_code,omitempty"`
	Description string `json:"description,omitempty"`
}

func NewBotClient(botToken string) *BotClient {
	return &BotClient{
		botToken: strings.TrimSpace(botToken),
		httpClient: &http.Client{
			Timeout: telegramRequestTimeout,
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

	trimmedText := strings.TrimSpace(text)
	if trimmedText == "" {
		return fmt.Errorf("text is required")
	}

	body, err := json.Marshal(sendMessageRequest{
		ChatID: chatID,
		Text:   trimmedText,
	})
	if err != nil {
		return fmt.Errorf("marshal telegram sendMessage payload: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), telegramRequestTimeout)
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

	var apiResp telegramAPIResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		preview := strings.TrimSpace(string(respBody))
		if len(preview) > maxTelegramErrorPreview {
			preview = preview[:maxTelegramErrorPreview]
		}
		return fmt.Errorf("decode telegram sendMessage response: %w (body=%q)", err, preview)
	}

	if resp.StatusCode >= http.StatusBadRequest || !apiResp.OK {
		description := strings.TrimSpace(apiResp.Description)
		if description == "" {
			description = "unknown telegram error"
		}
		if apiResp.ErrorCode != 0 {
			return fmt.Errorf("telegram sendMessage failed (%d/%d): %s", resp.StatusCode, apiResp.ErrorCode, description)
		}
		return fmt.Errorf("telegram sendMessage failed (%d): %s", resp.StatusCode, description)
	}

	return nil
}
