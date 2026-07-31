package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/sakurahilljp/docopt-go"
)

const version = "1.0.0"

const usage = `discord_notify - Short message sender CLI for Discord.

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
  -v --verbose          Show verbose output log.

Environment Variables:
  DISCORD_WEBHOOK_URL   Discord Webhook URL
  DISCORD_BOT_TOKEN     Discord Bot Token
  DISCORD_CHANNEL_ID    Discord Channel ID
  DISCORD_USERNAME      Sender username (Webhook only)
  DISCORD_AVATAR_URL    Avatar image URL (Webhook only)
`

type Config struct {
	WebhookURL string
	BotToken   string
	ChannelID  string
	Message    string
	Username   string
	AvatarURL  string
	Verbose    bool
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

func main() {
	cfg, err := parseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "エラー: %v\n", err)
		os.Exit(1)
	}

	if err := sendMessage(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "送信失敗: %v\n", err)
		os.Exit(1)
	}

	if cfg.Verbose {
		fmt.Println("メッセージを正常に送信しました。")
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
	verbose, _ := opts.Bool("--verbose")

	// 1. 環境変数からのデフォルト値取得
	cfg.WebhookURL = getFirstNonEmpty(webhook, os.Getenv("DISCORD_WEBHOOK_URL"))
	cfg.BotToken = getFirstNonEmpty(token, os.Getenv("DISCORD_BOT_TOKEN"))
	cfg.ChannelID = getFirstNonEmpty(channel, os.Getenv("DISCORD_CHANNEL_ID"))
	cfg.Username = getFirstNonEmpty(username, os.Getenv("DISCORD_USERNAME"))
	cfg.AvatarURL = getFirstNonEmpty(avatar, os.Getenv("DISCORD_AVATAR_URL"))
	cfg.Verbose = verbose

	// 2. メッセージの優先度: --message > <message> 位置引数 > 標準入力
	if messageOpt != "" {
		cfg.Message = messageOpt
	} else if messageArg != "" {
		cfg.Message = messageArg
	} else {
		stdinMsg, err := readStdin()
		if err != nil {
			return nil, fmt.Errorf("標準入力の読み込みに失敗しました: %w", err)
		}
		cfg.Message = stdinMsg
	}

	// メッセージバリデーション
	if strings.TrimSpace(cfg.Message) == "" {
		return nil, errors.New("送信するメッセージが指定されていません")
	}

	// 認証情報バリデーション
	if cfg.WebhookURL == "" && (cfg.BotToken == "" || cfg.ChannelID == "") {
		return nil, errors.New("Webhook URL または (Bot Token と Channel ID のペア) のいずれかを指定してください")
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
		Timeout: 10 * time.Second,
	}

	if cfg.WebhookURL != "" {
		return sendWebhookMessage(client, cfg)
	}
	return sendBotMessage(client, cfg)
}

func sendWebhookMessage(client *http.Client, cfg *Config) error {
	payload := WebhookPayload{
		Content:   cfg.Message,
		Username:  cfg.Username,
		AvatarURL: cfg.AvatarURL,
	}

	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("JSONのエンコードに失敗しました: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, cfg.WebhookURL, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return fmt.Errorf("リクエストの作成に失敗しました: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTPリクエストエラー: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		var apiErr DiscordErrorResponse
		if json.Unmarshal(body, &apiErr) == nil && apiErr.Message != "" {
			return fmt.Errorf("Discord API エラー (ステータス %d): %s (コード: %d)", resp.StatusCode, apiErr.Message, apiErr.Code)
		}
		return fmt.Errorf("送信失敗 (ステータス %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

func sendBotMessage(client *http.Client, cfg *Config) error {
	apiURL := fmt.Sprintf("https://discord.com/api/v10/channels/%s/messages", cfg.ChannelID)

	payload := BotPayload{
		Content: cfg.Message,
	}

	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("JSONのエンコードに失敗しました: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return fmt.Errorf("リクエストの作成に失敗しました: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bot "+cfg.BotToken)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTPリクエストエラー: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		var apiErr DiscordErrorResponse
		if json.Unmarshal(body, &apiErr) == nil && apiErr.Message != "" {
			return fmt.Errorf("Discord API エラー (ステータス %d): %s (コード: %d)", resp.StatusCode, apiErr.Message, apiErr.Code)
		}
		return fmt.Errorf("送信失敗 (ステータス %d): %s", resp.StatusCode, string(body))
	}

	return nil
}
