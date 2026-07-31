package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestParseArgs(t *testing.T) {
	tests := []struct {
		name                 string
		argv                 []string
		expectedMsg          string
		expectedURL          string
		expectedTok          string
		expectedChan         string
		expectedIgnoreErrors bool
		expectedTimeout      time.Duration
		expectedRetry        int
		expectedError        bool
	}{
		{
			name:            "Webhook option with position message",
			argv:            []string{"-w", "https://discord.com/api/webhooks/test", "Hello World"},
			expectedMsg:     "Hello World",
			expectedURL:     "https://discord.com/api/webhooks/test",
			expectedTimeout: 10 * time.Second,
			expectedRetry:   0,
		},
		{
			name:                 "Ignore errors and timeout/retry flags",
			argv:                 []string{"-w", "https://discord.com/api/webhooks/test", "-i", "--timeout=5s", "--retry=3", "Hello"},
			expectedMsg:          "Hello",
			expectedURL:          "https://discord.com/api/webhooks/test",
			expectedIgnoreErrors: true,
			expectedTimeout:      5 * time.Second,
			expectedRetry:        3,
		},
		{
			name:          "Invalid timeout duration error",
			argv:          []string{"-w", "https://discord.com/api/webhooks/test", "--timeout=invalid", "Hello"},
			expectedError: true,
		},
		{
			name:          "Invalid retry count error",
			argv:          []string{"-w", "https://discord.com/api/webhooks/test", "--retry=-1", "Hello"},
			expectedError: true,
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
				if cfg.IgnoreErrors != tt.expectedIgnoreErrors {
					t.Errorf("expected IgnoreErrors %v, got %v", tt.expectedIgnoreErrors, cfg.IgnoreErrors)
				}
				if cfg.Timeout != tt.expectedTimeout {
					t.Errorf("expected Timeout %v, got %v", tt.expectedTimeout, cfg.Timeout)
				}
				if cfg.Retry != tt.expectedRetry {
					t.Errorf("expected Retry %d, got %d", tt.expectedRetry, cfg.Retry)
				}
			}
		})
	}
}

func TestSendMessageWithRetry(t *testing.T) {
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

	cfg := &Config{
		WebhookURL: server.URL,
		Message:    "Retry test",
		Timeout:    5 * time.Second,
		Retry:      3,
	}

	err := sendMessage(cfg)
	if err != nil {
		t.Fatalf("sendMessage returned unexpected error: %v", err)
	}

	if atomic.LoadInt32(&attempts) != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
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
		Timeout:    5 * time.Second,
	}

	err := sendMessage(cfg)
	if err != nil {
		t.Fatalf("sendMessage returned error: %v", err)
	}
}
