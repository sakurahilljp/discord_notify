package discord

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"
)

type discordErrorResponse struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

type discordRateLimitResponse struct {
	Message    string  `json:"message"`
	RetryAfter float64 `json:"retry_after"`
	Global     bool    `json:"global"`
}

func sendWithRetry(ctx context.Context, client *http.Client, cfg Config, message string) error {
	maxAttempts := cfg.Retry + 1
	var lastErr error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		// Check context before trying
		if err := ctx.Err(); err != nil {
			return err
		}

		var retryAfter time.Duration
		var err error

		if cfg.WebhookURL != "" {
			retryAfter, err = sendWebhook(ctx, client, cfg, message)
		} else {
			retryAfter, err = sendBotMessage(ctx, client, cfg, message)
		}

		if err == nil {
			return nil
		}

		lastErr = err

		if attempt < maxAttempts {
			// Exponential backoff by default (1s, 2s, 4s...)
			sleepDuration := time.Duration(1<<uint(attempt-1)) * time.Second
			if retryAfter > 0 {
				sleepDuration = retryAfter
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(sleepDuration):
			}
		}
	}

	return lastErr
}

func parseRetryAfter(resp *http.Response) time.Duration {
	if headerVal := resp.Header.Get("Retry-After"); headerVal != "" {
		if sec, err := strconv.ParseFloat(headerVal, 64); err == nil && sec > 0 {
			return time.Duration(sec * float64(time.Second))
		}
	}

	body, err := io.ReadAll(resp.Body)
	if err == nil {
		// Restore body for any subsequent readers
		resp.Body = io.NopCloser(bytes.NewBuffer(body))

		var rl discordRateLimitResponse
		if json.Unmarshal(body, &rl) == nil && rl.RetryAfter > 0 {
			return time.Duration(rl.RetryAfter * float64(time.Second))
		}
	}

	return 0
}
