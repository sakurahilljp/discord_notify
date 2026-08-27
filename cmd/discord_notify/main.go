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

const version = "1.3.0"

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
  -p --profile=<name>   Use a specific profile from YAML config.
  --config=<path>       Path to YAML configuration file.
  --env-file=<path>     Path to .env file to load environment variables from.
  --no-env              Disable automatic loading of .env file.
  -i --ignore-errors    Ignore send errors and exit with code 0 (prints warning).
  --timeout=<duration>  HTTP request timeout (e.g. 10s, 1m).
  --retry=<count>       Number of retry attempts on failure.
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

	envFile, _ := opts.String("--env-file")
	noEnv, _ := opts.Bool("--no-env")
	configPath, _ := opts.String("--config")
	profile, _ := opts.String("--profile")
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

	// 1. Load .env file
	if envFile != "" {
		if err := discord.LoadDotEnv(envFile); err != nil {
			return nil, fmt.Errorf("failed to load env file %q: %w", envFile, err)
		}
	} else if !noEnv {
		_ = discord.LoadDotEnvIfExists(".env")
	}

	// 2. Load YAML configuration
	var baseConfig discord.Config
	resolvedConfigPath := configPath
	if resolvedConfigPath == "" {
		if foundPath := discord.FindDefaultConfigFile(); foundPath != "" {
			resolvedConfigPath = foundPath
		}
	}

	if resolvedConfigPath != "" {
		loadedCfg, err := discord.LoadYAMLProfile(resolvedConfigPath, profile)
		if err != nil {
			return nil, err
		}
		baseConfig = loadedCfg
	} else if profile != "" {
		return nil, fmt.Errorf("profile %q specified but no config file found", profile)
	}

	// 3. Override with environment variables
	if v := os.Getenv(discord.EnvWebhookURL); v != "" {
		baseConfig.WebhookURL = v
	}
	if v := os.Getenv(discord.EnvBotToken); v != "" {
		baseConfig.BotToken = v
	}
	if v := os.Getenv(discord.EnvChannelID); v != "" {
		baseConfig.ChannelID = v
	}
	if v := os.Getenv(discord.EnvUsername); v != "" {
		baseConfig.Username = v
	}
	if v := os.Getenv(discord.EnvAvatarURL); v != "" {
		baseConfig.AvatarURL = v
	}

	// 4. Override with explicit CLI flags
	if webhook != "" {
		baseConfig.WebhookURL = webhook
	}
	if token != "" {
		baseConfig.BotToken = token
	}
	if channel != "" {
		baseConfig.ChannelID = channel
	}
	if username != "" {
		baseConfig.Username = username
	}
	if avatar != "" {
		baseConfig.AvatarURL = avatar
	}

	if timeoutStr != "" {
		d, err := time.ParseDuration(timeoutStr)
		if err != nil {
			return nil, fmt.Errorf("invalid --timeout duration %q: %w", timeoutStr, err)
		}
		baseConfig.Timeout = d
	}

	if retryStr != "" {
		r, err := strconv.Atoi(retryStr)
		if err != nil || r < 0 {
			return nil, fmt.Errorf("invalid --retry count %q: must be a non-negative integer", retryStr)
		}
		baseConfig.Retry = r
	}

	cfg := &cliConfig{
		discordConfig: baseConfig,
		ignoreErrors:  ignoreErrors,
		verbose:       verbose,
	}

	// 5. Resolve message priority: --message > <message> positional argument > stdin
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
