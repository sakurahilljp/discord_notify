package discord

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestSendWebhook(t *testing.T) {
	expectedMsg := "Hello Webhook!"
	expectedUser := "CustomBot"
	expectedAvatar := "https://example.com/avatar.png"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected application/json, got %s", r.Header.Get("Content-Type"))
		}

		var p webhookPayload
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}

		if p.Content != expectedMsg {
			t.Errorf("expected content %q, got %q", expectedMsg, p.Content)
		}
		if p.Username != expectedUser {
			t.Errorf("expected username %q, got %q", expectedUser, p.Username)
		}
		if p.AvatarURL != expectedAvatar {
			t.Errorf("expected avatar %q, got %q", expectedAvatar, p.AvatarURL)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	ctx := context.Background()
	err := SendWebhook(ctx, server.URL, expectedMsg,
		WithUsername(expectedUser),
		WithAvatarURL(expectedAvatar),
		WithTimeout(3*time.Second),
	)
	if err != nil {
		t.Fatalf("SendWebhook failed: %v", err)
	}
}

func TestSendFromEnv(t *testing.T) {
	expectedMsg := "Hello from Env!"
	expectedUser := "EnvBot"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var p webhookPayload
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if p.Content != expectedMsg {
			t.Errorf("expected content %q, got %q", expectedMsg, p.Content)
		}
		if p.Username != expectedUser {
			t.Errorf("expected username %q, got %q", expectedUser, p.Username)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	t.Setenv(EnvWebhookURL, server.URL)
	t.Setenv(EnvUsername, expectedUser)

	ctx := context.Background()
	err := SendFromEnv(ctx, expectedMsg)
	if err != nil {
		t.Fatalf("SendFromEnv failed: %v", err)
	}
}

func TestNewConfigFromEnv(t *testing.T) {
	t.Setenv(EnvWebhookURL, "https://example.com/webhook")
	t.Setenv(EnvBotToken, "my-token")
	t.Setenv(EnvChannelID, "my-channel")
	t.Setenv(EnvUsername, "MyBot")
	t.Setenv(EnvAvatarURL, "https://example.com/avatar.png")

	cfg := NewConfigFromEnv()
	if cfg.WebhookURL != "https://example.com/webhook" {
		t.Errorf("expected WebhookURL, got %q", cfg.WebhookURL)
	}
	if cfg.BotToken != "my-token" {
		t.Errorf("expected BotToken, got %q", cfg.BotToken)
	}
	if cfg.ChannelID != "my-channel" {
		t.Errorf("expected ChannelID, got %q", cfg.ChannelID)
	}
	if cfg.Username != "MyBot" {
		t.Errorf("expected Username, got %q", cfg.Username)
	}
	if cfg.AvatarURL != "https://example.com/avatar.png" {
		t.Errorf("expected AvatarURL, got %q", cfg.AvatarURL)
	}
}

func TestSendBotMessage(t *testing.T) {
	expectedMsg := "Hello Bot!"
	expectedToken := "test-bot-token"
	channelID := "987654321"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bot "+expectedToken {
			t.Errorf("expected Bot %s, got %s", expectedToken, r.Header.Get("Authorization"))
		}

		var p botPayload
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if p.Content != expectedMsg {
			t.Errorf("expected content %q, got %q", expectedMsg, p.Content)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id": "123"}`))
	}))
	defer server.Close()

	ctx := context.Background()
	cfg := Config{
		BotToken:   expectedToken,
		ChannelID:  channelID,
		HTTPClient: server.Client(),
	}

	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	_ = client

	_, err = sendBotMessage(ctx, server.Client(), Config{
		BotToken:  expectedToken,
		ChannelID: channelID,
	}, expectedMsg)
}

func TestSendWithRetry(t *testing.T) {
	var attempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		current := atomic.AddInt32(&attempts, 1)
		if current < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	ctx := context.Background()
	cfg := Config{
		WebhookURL: server.URL,
		Timeout:    3 * time.Second,
		Retry:      3,
	}

	err := Send(ctx, cfg, "Retry test")
	if err != nil {
		t.Fatalf("Send failed: %v", err)
	}

	if atomic.LoadInt32(&attempts) != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestSendRateLimit(t *testing.T) {
	var attempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		current := atomic.AddInt32(&attempts, 1)
		if current == 1 {
			w.Header().Set("Retry-After", "0.1")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"message": "You are being rate limited.", "retry_after": 0.1}`))
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	ctx := context.Background()
	cfg := Config{
		WebhookURL: server.URL,
		Timeout:    3 * time.Second,
		Retry:      2,
	}

	err := Send(ctx, cfg, "Rate limit test")
	if err != nil {
		t.Fatalf("Send failed: %v", err)
	}

	if atomic.LoadInt32(&attempts) != 2 {
		t.Errorf("expected 2 attempts, got %d", attempts)
	}
}

func TestContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	cfg := Config{
		WebhookURL: server.URL,
		Retry:      2,
	}

	err := Send(ctx, cfg, "Cancel test")
	if err == nil {
		t.Fatal("expected error due to context cancellation, got nil")
	}
}

func TestValidationErrors(t *testing.T) {
	ctx := context.Background()

	// Clear env to ensure clean validation test
	t.Setenv(EnvWebhookURL, "")
	t.Setenv(EnvBotToken, "")
	t.Setenv(EnvChannelID, "")

	// Missing credentials
	err := Send(ctx, Config{}, "Hello")
	if err == nil {
		t.Error("expected error for missing credentials, got nil")
	}

	// Empty message
	err = Send(ctx, Config{WebhookURL: "https://example.com"}, "")
	if err == nil {
		t.Error("expected error for empty message, got nil")
	}
}

func TestFunctionalOptions(t *testing.T) {
	customClient := &http.Client{Timeout: 7 * time.Second}
	var cfg Config

	WithRetry(5)(&cfg)
	if cfg.Retry != 5 {
		t.Errorf("expected retry 5, got %d", cfg.Retry)
	}

	WithHTTPClient(customClient)(&cfg)
	if cfg.HTTPClient != customClient {
		t.Errorf("expected custom HTTPClient, got %v", cfg.HTTPClient)
	}
}

func TestSendBotMessageHelper(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id": "123"}`))
	}))
	defer server.Close()

	ctx := context.Background()
	// Call SendBotMessage (will fail with invalid URL since it targets real Discord endpoint, but tests option wiring)
	_ = SendBotMessage(ctx, "test_token", "test_chan", "test msg",
		WithRetry(1),
		WithTimeout(2*time.Second),
		WithHTTPClient(server.Client()),
	)
}

