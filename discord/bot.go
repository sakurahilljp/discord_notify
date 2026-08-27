package discord

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type botPayload struct {
	Content string `json:"content"`
}

func sendBotMessage(ctx context.Context, client *http.Client, cfg Config, message string) (time.Duration, error) {
	apiURL := fmt.Sprintf("https://discord.com/api/v10/channels/%s/messages", cfg.ChannelID)

	payload := botPayload{
		Content: message,
	}

	req, err := buildHTTPRequest(ctx, apiURL, "Bot "+cfg.BotToken, payload, cfg.FilePath)
	if err != nil {
		return 0, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		retryAfter := parseRetryAfter(resp)
		return retryAfter, fmt.Errorf("rate limited by Discord (status 429)")
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		var apiErr discordErrorResponse
		if json.Unmarshal(body, &apiErr) == nil && apiErr.Message != "" {
			return 0, fmt.Errorf("Discord API error (status %d): %s (code: %d)", resp.StatusCode, apiErr.Message, apiErr.Code)
		}
		return 0, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return 0, nil
}
