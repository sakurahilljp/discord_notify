package main

import (
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
				if cfg.message != tt.expectedMsg {
					t.Errorf("expected message %q, got %q", tt.expectedMsg, cfg.message)
				}
				if tt.expectedURL != "" && cfg.discordConfig.WebhookURL != tt.expectedURL {
					t.Errorf("expected webhook URL %q, got %q", tt.expectedURL, cfg.discordConfig.WebhookURL)
				}
				if cfg.ignoreErrors != tt.expectedIgnoreErrors {
					t.Errorf("expected IgnoreErrors %v, got %v", tt.expectedIgnoreErrors, cfg.ignoreErrors)
				}
				if cfg.discordConfig.Timeout != tt.expectedTimeout {
					t.Errorf("expected Timeout %v, got %v", tt.expectedTimeout, cfg.discordConfig.Timeout)
				}
				if cfg.discordConfig.Retry != tt.expectedRetry {
					t.Errorf("expected Retry %d, got %d", tt.expectedRetry, cfg.discordConfig.Retry)
				}
			}
		})
	}
}
