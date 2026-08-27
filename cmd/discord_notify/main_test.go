package main

import (
	"os"
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
		expectedUsername     string
		expectedFilePath     string
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
			expectedTimeout: 0,
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
			name:             "File attachment with message",
			argv:             []string{"-w", "https://discord.com/api/webhooks/test", "-f", "./image.png", "Image caption"},
			expectedMsg:      "Image caption",
			expectedURL:      "https://discord.com/api/webhooks/test",
			expectedFilePath: "./image.png",
		},
		{
			name:             "File attachment without message",
			argv:             []string{"-w", "https://discord.com/api/webhooks/test", "-f", "./only_image.png"},
			expectedMsg:      "",
			expectedURL:      "https://discord.com/api/webhooks/test",
			expectedFilePath: "./only_image.png",
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
		{
			name:          "Missing both message and file error",
			argv:          []string{"-w", "https://discord.com/api/webhooks/test"},
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
				if tt.expectedFilePath != "" && cfg.discordConfig.FilePath != tt.expectedFilePath {
					t.Errorf("expected file path %q, got %q", tt.expectedFilePath, cfg.discordConfig.FilePath)
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

func TestParseArgsWithConfigAndEnv(t *testing.T) {
	tmpDir := t.TempDir()
	yamlPath := tmpDir + "/test_config.yaml"
	envPath := tmpDir + "/test.env"

	yamlData := `
default_profile: dev

profiles:
  dev:
    webhook_url: "https://discord.com/api/webhooks/yaml_dev"
    username: "YAML Dev"
    retry: 1
    timeout: "3s"

  prod:
    webhook_url: "https://discord.com/api/webhooks/yaml_prod"
    username: "YAML Prod"
    retry: 5
`
	if err := os.WriteFile(yamlPath, []byte(yamlData), 0600); err != nil {
		t.Fatalf("failed to write yaml file: %v", err)
	}

	envData := `
DISCORD_WEBHOOK_URL="https://discord.com/api/webhooks/from_env_file"
DISCORD_USERNAME="Env User"
`
	if err := os.WriteFile(envPath, []byte(envData), 0600); err != nil {
		t.Fatalf("failed to write env file: %v", err)
	}

	t.Run("load yaml default profile", func(t *testing.T) {
		cfg, err := parseArgs([]string{"--config", yamlPath, "Test Message"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.discordConfig.WebhookURL != "https://discord.com/api/webhooks/yaml_dev" {
			t.Errorf("expected yaml_dev webhook, got %q", cfg.discordConfig.WebhookURL)
		}
		if cfg.discordConfig.Username != "YAML Dev" {
			t.Errorf("expected 'YAML Dev', got %q", cfg.discordConfig.Username)
		}
		if cfg.discordConfig.Retry != 1 {
			t.Errorf("expected retry 1, got %d", cfg.discordConfig.Retry)
		}
		if cfg.discordConfig.Timeout != 3*time.Second {
			t.Errorf("expected timeout 3s, got %v", cfg.discordConfig.Timeout)
		}
	})

	t.Run("load yaml specific prod profile with flag override", func(t *testing.T) {
		cfg, err := parseArgs([]string{"--config", yamlPath, "-p", "prod", "-u", "Overridden User", "--retry", "2", "Prod Alert"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.discordConfig.WebhookURL != "https://discord.com/api/webhooks/yaml_prod" {
			t.Errorf("expected yaml_prod webhook, got %q", cfg.discordConfig.WebhookURL)
		}
		if cfg.discordConfig.Username != "Overridden User" {
			t.Errorf("expected 'Overridden User', got %q", cfg.discordConfig.Username)
		}
		if cfg.discordConfig.Retry != 2 {
			t.Errorf("expected retry 2, got %d", cfg.discordConfig.Retry)
		}
	})

	t.Run("load env-file flag", func(t *testing.T) {
		cfg, err := parseArgs([]string{"--env-file", envPath, "Env Message"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.discordConfig.WebhookURL != "https://discord.com/api/webhooks/from_env_file" {
			t.Errorf("expected env webhook, got %q", cfg.discordConfig.WebhookURL)
		}
		if cfg.discordConfig.Username != "Env User" {
			t.Errorf("expected 'Env User', got %q", cfg.discordConfig.Username)
		}
	})

	t.Run("invalid env-file error", func(t *testing.T) {
		_, err := parseArgs([]string{"--env-file", "non_existent.env", "Message"})
		if err == nil {
			t.Fatal("expected error for non-existent env-file, got nil")
		}
	})

	t.Run("profile without config file found in empty dir", func(t *testing.T) {
		originalWd, _ := os.Getwd()
		defer func() { _ = os.Chdir(originalWd) }()
		_ = os.Chdir(tmpDir)

		_, err := parseArgs([]string{"-p", "non_existent_profile", "Message"})
		if err == nil {
			t.Fatal("expected error when profile specified but no config found, got nil")
		}
	})

	t.Run("system environment variable overrides YAML config", func(t *testing.T) {
		t.Setenv("DISCORD_USERNAME", "System Env User")
		cfg, err := parseArgs([]string{"--config", yamlPath, "Test Message"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.discordConfig.Username != "System Env User" {
			t.Errorf("expected system env user to override YAML, got %q", cfg.discordConfig.Username)
		}
	})

	t.Run("positional message vs -m option", func(t *testing.T) {
		cfg, err := parseArgs([]string{"-w", "https://discord.com/api/webhooks/test", "-m", "Flag Message", "Positional Message"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.message != "Flag Message" {
			t.Errorf("expected -m flag to take priority, got %q", cfg.message)
		}
	})
}
