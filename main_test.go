package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParseArgs(t *testing.T) {
	tests := []struct {
		name          string
		argv          []string
		expectedMsg   string
		expectedURL   string
		expectedTok   string
		expectedChan  string
		expectedError bool
	}{
		{
			name:        "Webhook option with position message",
			argv:        []string{"-w", "https://discord.com/api/webhooks/test", "Hello World"},
			expectedMsg: "Hello World",
			expectedURL: "https://discord.com/api/webhooks/test",
		},
		{
			name:         "Bot token and channel option with -m flag",
			argv:         []string{"-t", "mytoken", "-c", "12345", "-m", "Bot message"},
			expectedMsg:  "Bot message",
			expectedTok:  "mytoken",
			expectedChan: "12345",
		},
		{
			name:          "Missing credentials error",
			argv:          []string{"Hello World"},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := parseArgs(tt.argv)
			if (err != nil) != tt.expectedError {
				t.Fatalf("parseArgs() error = %v, expectedError %v", err, tt.expectedError)
			}
			if !tt.expectedError {
				if cfg.Message != tt.expectedMsg {
					t.Errorf("expected message %q, got %q", tt.expectedMsg, cfg.Message)
				}
				if tt.expectedURL != "" && cfg.WebhookURL != tt.expectedURL {
					t.Errorf("expected webhook URL %q, got %q", tt.expectedURL, cfg.WebhookURL)
				}
				if tt.expectedTok != "" && cfg.BotToken != tt.expectedTok {
					t.Errorf("expected bot token %q, got %q", tt.expectedTok, cfg.BotToken)
				}
				if tt.expectedChan != "" && cfg.ChannelID != tt.expectedChan {
					t.Errorf("expected channel ID %q, got %q", tt.expectedChan, cfg.ChannelID)
				}
			}
		})
	}
}

func TestSendWebhookMessage(t *testing.T) {
	expectedMessage := "Hello Webhook!"
	expectedUsername := "TestBot"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected method POST, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		var payload WebhookPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("Failed to decode payload: %v", err)
		}

		if payload.Content != expectedMessage {
			t.Errorf("Expected content %q, got %q", expectedMessage, payload.Content)
		}
		if payload.Username != expectedUsername {
			t.Errorf("Expected username %q, got %q", expectedUsername, payload.Username)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	cfg := &Config{
		WebhookURL: server.URL,
		Message:    expectedMessage,
		Username:   expectedUsername,
	}

	err := sendMessage(cfg)
	if err != nil {
		t.Fatalf("sendMessage returned error: %v", err)
	}
}

func TestSendBotMessage(t *testing.T) {
	expectedToken := "test-bot-token"
	expectedMessage := "Hello Bot!"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected method POST, got %s", r.Method)
		}
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bot "+expectedToken {
			t.Errorf("Expected Authorization 'Bot %s', got %q", expectedToken, authHeader)
		}

		var payload BotPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("Failed to decode payload: %v", err)
		}

		if payload.Content != expectedMessage {
			t.Errorf("Expected content %q, got %q", expectedMessage, payload.Content)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id": "123456789"}`))
	}))
	defer server.Close()

	client := server.Client()
	cfg := &Config{
		BotToken:  expectedToken,
		ChannelID: "123456",
		Message:   expectedMessage,
	}

	payload := BotPayload{Content: cfg.Message}
	jsonBytes, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, server.URL, io.NopCloser(bytes.NewReader(jsonBytes)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bot "+cfg.BotToken)

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}
