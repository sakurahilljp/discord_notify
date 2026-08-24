package discord

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"
)

// Environment variable names supported by the package.
const (
	EnvWebhookURL = "DISCORD_WEBHOOK_URL"
	EnvBotToken   = "DISCORD_BOT_TOKEN"
	EnvChannelID  = "DISCORD_CHANNEL_ID"
	EnvUsername   = "DISCORD_USERNAME"
	EnvAvatarURL  = "DISCORD_AVATAR_URL"
)

// DefaultTimeout is the default HTTP timeout used if not specified.
const DefaultTimeout = 10 * time.Second

// Config holds the configuration options for sending Discord messages.
type Config struct {
	WebhookURL string        // Discord Webhook URL
	BotToken   string        // Discord Bot Token
	ChannelID  string        // Discord Text Channel ID
	Username   string        // Custom sender username (Webhook only)
	AvatarURL  string        // Custom avatar image URL (Webhook only)
	Timeout    time.Duration // HTTP request timeout (defaults to DefaultTimeout)
	Retry      int           // Number of retry attempts on failure (defaults to 0)
	HTTPClient *http.Client  // Optional custom HTTP client
}

// NewConfigFromEnv creates a Config populated with values from environment variables.
func NewConfigFromEnv() Config {
	return Config{
		WebhookURL: os.Getenv(EnvWebhookURL),
		BotToken:   os.Getenv(EnvBotToken),
		ChannelID:  os.Getenv(EnvChannelID),
		Username:   os.Getenv(EnvUsername),
		AvatarURL:  os.Getenv(EnvAvatarURL),
		Timeout:    DefaultTimeout,
	}
}

func (c *Config) applyEnvDefaults() {
	if c.WebhookURL == "" {
		c.WebhookURL = os.Getenv(EnvWebhookURL)
	}
	if c.BotToken == "" {
		c.BotToken = os.Getenv(EnvBotToken)
	}
	if c.ChannelID == "" {
		c.ChannelID = os.Getenv(EnvChannelID)
	}
	if c.Username == "" {
		c.Username = os.Getenv(EnvUsername)
	}
	if c.AvatarURL == "" {
		c.AvatarURL = os.Getenv(EnvAvatarURL)
	}
}

// Client provides methods to send messages to Discord.
type Client struct {
	config Config
	client *http.Client
}

// NewClient creates a new Discord client with the provided configuration.
// If credentials or webhook URLs are empty, environment variables are checked as fallback.
func NewClient(cfg Config) (*Client, error) {
	cfg.applyEnvDefaults()

	if cfg.WebhookURL == "" && (cfg.BotToken == "" || cfg.ChannelID == "") {
		return nil, errors.New("either WebhookURL or both BotToken and ChannelID must be provided")
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		timeout := cfg.Timeout
		if timeout <= 0 {
			timeout = DefaultTimeout
		}
		httpClient = &http.Client{
			Timeout: timeout,
		}
	}

	return &Client{
		config: cfg,
		client: httpClient,
	}, nil
}

// Send sends a message to Discord Text Channel using the configured destination (Webhook or Bot).
func (c *Client) Send(ctx context.Context, message string) error {
	if strings.TrimSpace(message) == "" {
		return errors.New("message cannot be empty")
	}
	return sendWithRetry(ctx, c.client, c.config, message)
}

// Send sends a message to Discord with the provided Config.
func Send(ctx context.Context, cfg Config, message string) error {
	client, err := NewClient(cfg)
	if err != nil {
		return err
	}
	return client.Send(ctx, message)
}

// SendFromEnv sends a message using configuration loaded entirely from environment variables.
func SendFromEnv(ctx context.Context, message string, opts ...Option) error {
	cfg := NewConfigFromEnv()
	for _, opt := range opts {
		opt(&cfg)
	}
	return Send(ctx, cfg, message)
}

// Option defines a functional option for configuring message sending.
type Option func(*Config)

// WithUsername sets a custom username for Webhook messages.
func WithUsername(username string) Option {
	return func(cfg *Config) {
		cfg.Username = username
	}
}

// WithAvatarURL sets a custom avatar URL for Webhook messages.
func WithAvatarURL(avatarURL string) Option {
	return func(cfg *Config) {
		cfg.AvatarURL = avatarURL
	}
}

// WithTimeout sets a custom HTTP timeout.
func WithTimeout(timeout time.Duration) Option {
	return func(cfg *Config) {
		cfg.Timeout = timeout
	}
}

// WithRetry sets the number of retry attempts on failure.
func WithRetry(retry int) Option {
	return func(cfg *Config) {
		cfg.Retry = retry
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(client *http.Client) Option {
	return func(cfg *Config) {
		cfg.HTTPClient = client
	}
}

// SendWebhook sends a message to a Discord Webhook URL.
func SendWebhook(ctx context.Context, webhookURL, message string, opts ...Option) error {
	cfg := Config{
		WebhookURL: webhookURL,
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return Send(ctx, cfg, message)
}

// SendBotMessage sends a message to a Discord Channel using Bot Token and Channel ID.
func SendBotMessage(ctx context.Context, botToken, channelID, message string, opts ...Option) error {
	cfg := Config{
		BotToken:  botToken,
		ChannelID: channelID,
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return Send(ctx, cfg, message)
}
