package discord

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// YAMLConfig represents the top-level structure of a discord_notify YAML configuration file.
type YAMLConfig struct {
	DefaultProfile string                 `yaml:"default_profile"`
	Profiles       map[string]YAMLProfile `yaml:"profiles"`
}

// YAMLProfile represents an individual configuration profile in YAML.
type YAMLProfile struct {
	WebhookURL string `yaml:"webhook_url"`
	BotToken   string `yaml:"bot_token"`
	ChannelID  string `yaml:"channel_id"`
	Username   string `yaml:"username"`
	AvatarURL  string `yaml:"avatar_url"`
	FilePath   string `yaml:"file_path"`
	Timeout    string `yaml:"timeout"`
	Retry      *int   `yaml:"retry"`
}

// ToConfig converts a YAMLProfile into a Config struct.
func (p YAMLProfile) ToConfig() (Config, error) {
	cfg := Config{
		WebhookURL: p.WebhookURL,
		BotToken:   p.BotToken,
		ChannelID:  p.ChannelID,
		Username:   p.Username,
		AvatarURL:  p.AvatarURL,
		FilePath:   p.FilePath,
	}

	if p.Timeout != "" {
		d, err := time.ParseDuration(p.Timeout)
		if err != nil {
			return Config{}, fmt.Errorf("invalid timeout duration %q: %w", p.Timeout, err)
		}
		cfg.Timeout = d
	}

	if p.Retry != nil {
		if *p.Retry < 0 {
			return Config{}, fmt.Errorf("invalid retry count %d: must be non-negative", *p.Retry)
		}
		cfg.Retry = *p.Retry
	}

	return cfg, nil
}

// LoadYAMLConfig parses a YAML configuration file.
func LoadYAMLConfig(filename string) (*YAMLConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read YAML config file %q: %w", filename, err)
	}

	var yCfg YAMLConfig
	if err := yaml.Unmarshal(data, &yCfg); err != nil {
		return nil, fmt.Errorf("failed to parse YAML config file %q: %w", filename, err)
	}

	return &yCfg, nil
}

// LoadYAMLProfile reads the YAML config file and returns the Config for the specified profile.
// If profileName is empty, it uses DefaultProfile from the config, or falls back to "default".
func LoadYAMLProfile(filename string, profileName string) (Config, error) {
	yCfg, err := LoadYAMLConfig(filename)
	if err != nil {
		return Config{}, err
	}

	targetProfile := profileName
	if targetProfile == "" {
		if yCfg.DefaultProfile != "" {
			targetProfile = yCfg.DefaultProfile
		} else {
			targetProfile = "default"
		}
	}

	profile, exists := yCfg.Profiles[targetProfile]
	if !exists {
		return Config{}, fmt.Errorf("profile %q not found in config file %q", targetProfile, filename)
	}

	return profile.ToConfig()
}

// FindDefaultConfigFile searches for a configuration file in default locations:
// 1. ./.discord_notify.yaml
// 2. ./discord_notify.yaml
// 3. ~/.config/discord_notify/config.yaml
// Returns the path if found, or an empty string if none exists.
func FindDefaultConfigFile() string {
	candidates := []string{
		".discord_notify.yaml",
		"discord_notify.yaml",
	}

	if userConfigDir, err := os.UserConfigDir(); err == nil && userConfigDir != "" {
		candidates = append(candidates, filepath.Join(userConfigDir, "discord_notify", "config.yaml"))
	}

	for _, path := range candidates {
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return path
		}
	}

	return ""
}
