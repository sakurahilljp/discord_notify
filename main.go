package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/sakurahilljp/docopt-go"
)

const version = "1.0.0"

const usage = `discord_notify - A simple CLI tool to send short messages to Discord text channels.

Usage:
  discord_notify [options] [<message>]
  discord_notify -h | --help
  discord_notify --version

Options:
  -h --help             Show this help message and exit.
  --version             Show version and exit.
  -w --webhook=<url>    Discord Webhook URL.
  -t --token=<token>    Discord Bot Token.
  -c --channel=<id>     Discord Channel ID.
  -m --message=<msg>    Message to send.
  -u --username=<name>  Sender username (Webhook only).
  -a --avatar=<url>     Avatar image URL (Webhook only).
  -i --ignore-errors    Ignore send errors and exit with code 0 (prints warning).
  --timeout=<duration>  HTTP request timeout [default: 10s].
  --retry=<count>       Number of retry attempts on failure [default: 0].
  -v --verbose          Show verbose output log.

Environment Variables:
  DISCORD_WEBHOOK_URL   Discord Webhook URL
  DISCORD_BOT_TOKEN     Discord Bot Token
  DISCORD_CHANNEL_ID    Discord Channel ID
  DISCORD_USERNAME      Sender username (Webhook only)
  DISCORD_AVATAR_URL    Avatar image URL (Webhook only)
`

type Config struct {
	WebhookURL   string
	BotToken     string
	ChannelID    string
	Message      string
	Username     string
	AvatarURL    string
	IgnoreErrors bool
	Timeout      time.Duration
	Retry        int
	Verbose      bool
}

type WebhookPayload struct {
	Content   string `json:"content"`
	Username  string `json:"username,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

type BotPayload struct {
	Content string `json:"content"`
}

type DiscordErrorResponse struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

type DiscordRateLimitResponse struct {
	Message    string  `json:"message"`
	RetryAfter float64 `json:"retry_after"`
	Global     bool    `json:"global"`
}

func main() {
	cfg, err := parseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := sendMessage(cfg); err != nil {
		if cfg.IgnoreErrors {
			fmt.Fprintf(os.Stderr, "Warning: Send failed: %v\n", err)
			os.Exit(0)
		}
		fmt.Fprintf(os.Stderr, "Send failed: %v\n", err)
		os.Exit(1)
	}

	if cfg.Verbose {
		fmt.Println("Message sent successfully.")
	}
}

func parseArgs(argv []string) (*Config, error) {
	parser := &docopt.Parser{
		HelpHandler: docopt.PrintHelpOnly,
	}

	opts, err := parser.ParseArgs(usage, argv, version)
	if err != nil {
		return nil, err
	}

	cfg := &Config{}

	webhook, _ := opts.String("--webhook")
	token, _ := opts.String("--token")
	channel, _ := opts.String("--channel")
	messageOpt, _ := opts.String("--message")
	messageArg, _ := opts.String("<message>")
	username, _ := opts.String("--username")
	avatar, _ := opts.String("--avatar")
	ignoreErrors, _ := opts.Bool("--ignore-errors")
	timeoutStr, _ := opts.String("--timeout")
	retryStr, _ := opts.String("--retry")
	verbose, _ := opts.Bool("--verbose")

	// Parse --timeout
	if timeoutStr != "" {
		d, err := time.ParseDuration(timeoutStr)
		if err != nil {
			return nil, fmt.Errorf("invalid --timeout duration %q: %w", timeoutStr, err)
		}
		cfg.Timeout = d
	} else {
		cfg.Timeout = 10 * time.Second
	}

	// Parse --retry
	if retryStr != "" {
		r, err := strconv.Atoi(retryStr)
		if err != nil || r < 0 {
			return nil, fmt.Errorf("invalid --retry count %q: must be a non-negative integer", retryStr)
		}
		cfg.Retry = r
	}

	// 1. Fallback to environment variables if flags are empty
	cfg.WebhookURL = getFirstNonEmpty(webhook, os.Getenv("DISCORD_WEBHOOK_URL"))
	cfg.BotToken = getFirstNonEmpty(token, os.Getenv("DISCORD_BOT_TOKEN"))
	cfg.ChannelID = getFirstNonEmpty(channel, os.Getenv("DISCORD_CHANNEL_ID"))
	cfg.Username = getFirstNonEmpty(username, os.Getenv("DISCORD_USERNAME"))
	cfg.AvatarURL = getFirstNonEmpty(avatar, os.Getenv("DISCORD_AVATAR_URL"))
	cfg.IgnoreErrors = ignoreErrors
	cfg.Verbose = verbose

	// 2. Resolve message priority: --message > <message> positional argument > stdin
	if messageOpt != "" {
		cfg.Message = messageOpt
	} else if messageArg != "" {
		cfg.Message = messageArg
	} else {
		stdinMsg, err := readStdin()
		if err != nil {
			return nil, fmt.Errorf("failed to read from standard input: %w", err)
		}
		cfg.Message = stdinMsg
	}

	// Validate message
	if strings.TrimSpace(cfg.Message) == "" {
		return nil, errors.New("no message specified")
	}

	// Validate credentials
	if cfg.WebhookURL == "" && (cfg.BotToken == "" || cfg.ChannelID == "") {
		return nil, errors.New("must specify either a Webhook URL or both Bot Token and Channel ID")
	}

	return cfg, nil
}

func getFirstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func readStdin() (string, error) {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return "", err
	}

	if (stat.Mode() & os.ModeCharDevice) == 0 {
		bytes, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(bytes)), nil
	}

	return "", nil
}

func sendMessage(cfg *Config) error {
	client := &http.Client{
		Timeout: cfg.Timeout,
	}

	maxAttempts := cfg.Retry + 1
	var lastErr error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		var retryAfter time.Duration
		var err error

		if cfg.WebhookURL != "" {
			retryAfter, err = sendWebhookMessage(client, cfg)
		} else {
			retryAfter, err = sendBotMessage(client, cfg)
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

			if cfg.Verbose {
				fmt.Fprintf(os.Stderr, "Attempt %d/%d failed: %v. Retrying in %v...\n", attempt, maxAttempts, err, sleepDuration)
			}

			time.Sleep(sleepDuration)
		}
	}

	return lastErr
}

func sendWebhookMessage(client *http.Client, cfg *Config) (time.Duration, error) {
	payload := WebhookPayload{
		Content:   cfg.Message,
		Username:  cfg.Username,
		AvatarURL: cfg.AvatarURL,
	}

	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return 0, fmt.Errorf("failed to encode JSON payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, cfg.WebhookURL, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return 0, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

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
		var apiErr DiscordErrorResponse
		if json.Unmarshal(body, &apiErr) == nil && apiErr.Message != "" {
			return 0, fmt.Errorf("Discord API error (status %d): %s (code: %d)", resp.StatusCode, apiErr.Message, apiErr.Code)
		}
		return 0, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return 0, nil
}

func sendBotMessage(client *http.Client, cfg *Config) (time.Duration, error) {
	apiURL := fmt.Sprintf("https://discord.com/api/v10/channels/%s/messages", cfg.ChannelID)

	payload := BotPayload{
		Content: cfg.Message,
	}

	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return 0, fmt.Errorf("failed to encode JSON payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return 0, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bot "+cfg.BotToken)

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
		var apiErr DiscordErrorResponse
		if json.Unmarshal(body, &apiErr) == nil && apiErr.Message != "" {
			return 0, fmt.Errorf("Discord API error (status %d): %s (code: %d)", resp.StatusCode, apiErr.Message, apiErr.Code)
		}
		return 0, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return 0, nil
}

func parseRetryAfter(resp *http.Response) time.Duration {
	if headerVal := resp.Header.Get("Retry-After"); headerVal != "" {
		if sec, err := strconv.ParseFloat(headerVal, 64); err == nil && sec > 0 {
			return time.Duration(sec * float64(time.Second))
		}
	}

	body, err := io.ReadAll(resp.Body)
	if err == nil {
		var rl DiscordRateLimitResponse
		if json.Unmarshal(body, &rl) == nil && rl.RetryAfter > 0 {
			return time.Duration(rl.RetryAfter * float64(time.Second))
		}
	}

	return 0
}
