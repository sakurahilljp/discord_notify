package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/sakurahilljp/docopt-go"

	"github.com/sakurahilljp/discord_notify/discord"
)

const version = "1.2.0"

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

type cliConfig struct {
	discordConfig discord.Config
	message       string
	ignoreErrors  bool
	verbose       bool
}

func main() {
	cfg, err := parseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()
	if err := discord.Send(ctx, cfg.discordConfig, cfg.message); err != nil {
		if cfg.ignoreErrors {
			fmt.Fprintf(os.Stderr, "Warning: Send failed: %v\n", err)
			os.Exit(0)
		}
		fmt.Fprintf(os.Stderr, "Send failed: %v\n", err)
		os.Exit(1)
	}

	if cfg.verbose {
		fmt.Println("Message sent successfully.")
	}
}

func parseArgs(argv []string) (*cliConfig, error) {
	opts, err := docopt.ParseArgs(usage, argv, version)
	if err != nil {
		return nil, err
	}

	cfg := &cliConfig{
		discordConfig: discord.NewConfigFromEnv(),
	}

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
		cfg.discordConfig.Timeout = d
	}

	// Parse --retry
	if retryStr != "" {
		r, err := strconv.Atoi(retryStr)
		if err != nil || r < 0 {
			return nil, fmt.Errorf("invalid --retry count %q: must be a non-negative integer", retryStr)
		}
		cfg.discordConfig.Retry = r
	}

	// 1. Override environment variables with explicit CLI flags
	if webhook != "" {
		cfg.discordConfig.WebhookURL = webhook
	}
	if token != "" {
		cfg.discordConfig.BotToken = token
	}
	if channel != "" {
		cfg.discordConfig.ChannelID = channel
	}
	if username != "" {
		cfg.discordConfig.Username = username
	}
	if avatar != "" {
		cfg.discordConfig.AvatarURL = avatar
	}
	cfg.ignoreErrors = ignoreErrors
	cfg.verbose = verbose

	// 2. Resolve message priority: --message > <message> positional argument > stdin
	if messageOpt != "" {
		cfg.message = messageOpt
	} else if messageArg != "" {
		cfg.message = messageArg
	} else {
		stdinMsg, err := readStdin()
		if err != nil {
			return nil, fmt.Errorf("failed to read from standard input: %w", err)
		}
		cfg.message = stdinMsg
	}

	// Validate message
	if strings.TrimSpace(cfg.message) == "" {
		return nil, errors.New("no message specified")
	}

	// Validate credentials
	if cfg.discordConfig.WebhookURL == "" && (cfg.discordConfig.BotToken == "" || cfg.discordConfig.ChannelID == "") {
		return nil, errors.New("must specify either a Webhook URL or both Bot Token and Channel ID")
	}

	return cfg, nil
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
