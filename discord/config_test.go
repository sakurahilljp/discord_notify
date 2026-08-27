package discord

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadDotEnv(t *testing.T) {
	// Create temporary .env file
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env")

	content := `
# This is a comment
DISCORD_TEST_KEY1=value1
DISCORD_TEST_KEY2="value2 with spaces"
DISCORD_TEST_KEY3='value3 single quoted'
export DISCORD_TEST_KEY4=value4 # inline comment
DISCORD_TEST_EXISTING=new_value
`
	if err := os.WriteFile(envPath, []byte(content), 0600); err != nil {
		t.Fatalf("failed to write temp env file: %v", err)
	}

	// Set an existing env var
	_ = os.Setenv("DISCORD_TEST_EXISTING", "original_value")
	defer func() {
		os.Unsetenv("DISCORD_TEST_KEY1")
		os.Unsetenv("DISCORD_TEST_KEY2")
		os.Unsetenv("DISCORD_TEST_KEY3")
		os.Unsetenv("DISCORD_TEST_KEY4")
		os.Unsetenv("DISCORD_TEST_EXISTING")
	}()

	if err := LoadDotEnv(envPath); err != nil {
		t.Fatalf("LoadDotEnv failed: %v", err)
	}

	if got := os.Getenv("DISCORD_TEST_KEY1"); got != "value1" {
		t.Errorf("expected value1, got %q", got)
	}
	if got := os.Getenv("DISCORD_TEST_KEY2"); got != "value2 with spaces" {
		t.Errorf("expected 'value2 with spaces', got %q", got)
	}
	if got := os.Getenv("DISCORD_TEST_KEY3"); got != "value3 single quoted" {
		t.Errorf("expected 'value3 single quoted', got %q", got)
	}
	if got := os.Getenv("DISCORD_TEST_KEY4"); got != "value4" {
		t.Errorf("expected value4, got %q", got)
	}
	if got := os.Getenv("DISCORD_TEST_EXISTING"); got != "original_value" {
		t.Errorf("expected existing variable to not be overwritten, got %q", got)
	}
}

func TestLoadDotEnvIfExists(t *testing.T) {
	err := LoadDotEnvIfExists("non_existent_file_path_12345.env")
	if err != nil {
		t.Fatalf("LoadDotEnvIfExists should return nil for non-existent file, got: %v", err)
	}
}

func TestLoadYAMLProfile(t *testing.T) {
	tmpDir := t.TempDir()
	yamlPath := filepath.Join(tmpDir, "config.yaml")

	yamlContent := `
default_profile: dev

profiles:
  dev:
    webhook_url: "https://discord.com/api/webhooks/dev"
    username: "Dev Bot"
    avatar_url: "https://example.com/avatar.png"
    timeout: "5s"
    retry: 2

  prod:
    webhook_url: "https://discord.com/api/webhooks/prod"
    username: "Prod Bot"
    retry: 0

  bot:
    bot_token: "test_token"
    channel_id: "test_channel"
    timeout: "15s"

  invalid_timeout:
    webhook_url: "https://discord.com/api/webhooks/test"
    timeout: "invalid-duration"

  invalid_retry:
    webhook_url: "https://discord.com/api/webhooks/test"
    retry: -1
`
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0600); err != nil {
		t.Fatalf("failed to write temp yaml file: %v", err)
	}

	t.Run("load default profile", func(t *testing.T) {
		cfg, err := LoadYAMLProfile(yamlPath, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.WebhookURL != "https://discord.com/api/webhooks/dev" {
			t.Errorf("expected dev webhook URL, got %q", cfg.WebhookURL)
		}
		if cfg.Username != "Dev Bot" {
			t.Errorf("expected 'Dev Bot', got %q", cfg.Username)
		}
		if cfg.AvatarURL != "https://example.com/avatar.png" {
			t.Errorf("expected avatar URL, got %q", cfg.AvatarURL)
		}
		if cfg.Timeout != 5*time.Second {
			t.Errorf("expected 5s timeout, got %v", cfg.Timeout)
		}
		if cfg.Retry != 2 {
			t.Errorf("expected retry 2, got %d", cfg.Retry)
		}
	})

	t.Run("load specific prod profile", func(t *testing.T) {
		cfg, err := LoadYAMLProfile(yamlPath, "prod")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.WebhookURL != "https://discord.com/api/webhooks/prod" {
			t.Errorf("expected prod webhook URL, got %q", cfg.WebhookURL)
		}
		if cfg.Username != "Prod Bot" {
			t.Errorf("expected 'Prod Bot', got %q", cfg.Username)
		}
		if cfg.Retry != 0 {
			t.Errorf("expected retry 0, got %d", cfg.Retry)
		}
	})

	t.Run("load bot profile", func(t *testing.T) {
		cfg, err := LoadYAMLProfile(yamlPath, "bot")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.BotToken != "test_token" {
			t.Errorf("expected bot token, got %q", cfg.BotToken)
		}
		if cfg.ChannelID != "test_channel" {
			t.Errorf("expected channel id, got %q", cfg.ChannelID)
		}
		if cfg.Timeout != 15*time.Second {
			t.Errorf("expected 15s timeout, got %v", cfg.Timeout)
		}
	})

	t.Run("non existent profile", func(t *testing.T) {
		_, err := LoadYAMLProfile(yamlPath, "unknown")
		if err == nil {
			t.Fatal("expected error for non-existent profile, got nil")
		}
	})

	t.Run("invalid timeout", func(t *testing.T) {
		_, err := LoadYAMLProfile(yamlPath, "invalid_timeout")
		if err == nil {
			t.Fatal("expected error for invalid timeout, got nil")
		}
	})

	t.Run("invalid retry", func(t *testing.T) {
		_, err := LoadYAMLProfile(yamlPath, "invalid_retry")
		if err == nil {
			t.Fatal("expected error for invalid retry, got nil")
		}
	})

	t.Run("non existent file", func(t *testing.T) {
		_, err := LoadYAMLProfile("non_existent_file.yaml", "")
		if err == nil {
			t.Fatal("expected error for non-existent file, got nil")
		}
	})

	t.Run("empty default_profile and fallback to default key", func(t *testing.T) {
		emptyDefaultPath := filepath.Join(tmpDir, "no_default.yaml")
		content := `
profiles:
  default:
    webhook_url: "https://discord.com/api/webhooks/fallback_default"
`
		if err := os.WriteFile(emptyDefaultPath, []byte(content), 0600); err != nil {
			t.Fatalf("failed to write yaml: %v", err)
		}

		cfg, err := LoadYAMLProfile(emptyDefaultPath, "")
		if err != nil {
			t.Fatalf("expected success with default fallback: %v", err)
		}
		if cfg.WebhookURL != "https://discord.com/api/webhooks/fallback_default" {
			t.Errorf("expected fallback_default URL, got %q", cfg.WebhookURL)
		}
	})
}

func TestFindDefaultConfigFile(t *testing.T) {
	// When none of the default files exist in current dir
	_ = FindDefaultConfigFile() // Should not panic and return string

	// Create temporary directory and change working directory to test local file discovery
	tmpDir := t.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get wd: %v", err)
	}
	defer func() {
		_ = os.Chdir(originalWd)
	}()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to chdir to tmpDir: %v", err)
	}

	// 1. None exists
	if got := FindDefaultConfigFile(); got != "" {
		t.Logf("FindDefaultConfigFile returned %q (possibly from user config dir)", got)
	}

	// 2. Create .discord_notify.yaml
	dotYamlPath := filepath.Join(tmpDir, ".discord_notify.yaml")
	if err := os.WriteFile(dotYamlPath, []byte("profiles: {}"), 0600); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	if got := FindDefaultConfigFile(); got != ".discord_notify.yaml" {
		t.Errorf("expected .discord_notify.yaml, got %q", got)
	}
}
