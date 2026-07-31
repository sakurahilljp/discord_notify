package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

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
	cfg, err := parseFlags()
	if err != nil {
		fmt.Fprintf(os.Stderr, "エラー: %v\n\n", err)
		flag.Usage()
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

func parseFlags() (*Config, error) {
	cfg := &Config{}

	var webhookFlag, tokenFlag, channelFlag, messageFlag, usernameFlag, avatarFlag string
	var verboseFlag bool

	flag.StringVar(&webhookFlag, "w", "", "Discord Webhook URL (短縮)")
	flag.StringVar(&webhookFlag, "webhook", "", "Discord Webhook URL")

	flag.StringVar(&tokenFlag, "t", "", "Discord Bot Token (短縮)")
	flag.StringVar(&tokenFlag, "token", "", "Discord Bot Token")

	flag.StringVar(&channelFlag, "c", "", "Discord Channel ID (短縮)")
	flag.StringVar(&channelFlag, "channel", "", "Discord Channel ID")

	flag.StringVar(&messageFlag, "m", "", "送信するメッセージ (短縮)")
	flag.StringVar(&messageFlag, "message", "", "送信するメッセージ")

	flag.StringVar(&usernameFlag, "u", "", "送信者名 [Webhook用] (短縮)")
	flag.StringVar(&usernameFlag, "username", "", "送信者名 [Webhook用]")

	flag.StringVar(&avatarFlag, "a", "", "アバター画像URL [Webhook用] (短縮)")
	flag.StringVar(&avatarFlag, "avatar", "", "アバター画像URL [Webhook用]")

	flag.BoolVar(&verboseFlag, "v", false, "詳細ログを表示 (短縮)")
	flag.BoolVar(&verboseFlag, "verbose", false, "詳細ログを表示")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "使用方法:\n")
		fmt.Fprintf(os.Stderr, "  discord_notify [オプション] [メッセージ]\n")
		fmt.Fprintf(os.Stderr, "  echo \"メッセージ\" | discord_notify [オプション]\n\n")
		fmt.Fprintf(os.Stderr, "環境変数:\n")
		fmt.Fprintf(os.Stderr, "  DISCORD_WEBHOOK_URL : Discord Webhook URL\n")
		fmt.Fprintf(os.Stderr, "  DISCORD_BOT_TOKEN   : Discord Bot Token\n")
		fmt.Fprintf(os.Stderr, "  DISCORD_CHANNEL_ID  : 送信先 Text Channel ID\n")
		fmt.Fprintf(os.Stderr, "  DISCORD_USERNAME    : 表示するユーザー名 (Webhook用)\n")
		fmt.Fprintf(os.Stderr, "  DISCORD_AVATAR_URL  : 表示するアバターURL (Webhook用)\n\n")
		fmt.Fprintf(os.Stderr, "オプション:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	// 1. 環境変数からのデフォルト値取得
	cfg.WebhookURL = getFirstNonEmpty(webhookFlag, os.Getenv("DISCORD_WEBHOOK_URL"))
	cfg.BotToken = getFirstNonEmpty(tokenFlag, os.Getenv("DISCORD_BOT_TOKEN"))
	cfg.ChannelID = getFirstNonEmpty(channelFlag, os.Getenv("DISCORD_CHANNEL_ID"))
	cfg.Username = getFirstNonEmpty(usernameFlag, os.Getenv("DISCORD_USERNAME"))
	cfg.AvatarURL = getFirstNonEmpty(avatarFlag, os.Getenv("DISCORD_AVATAR_URL"))
	cfg.Verbose = verboseFlag

	// 2. メッセージの取得優先度: -m / -message > 位置引数 > 標準入力
	if messageFlag != "" {
		cfg.Message = messageFlag
	} else if flag.NArg() > 0 {
		cfg.Message = strings.Join(flag.Args(), " ")
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

	// ターミナル入力ではなくパイプ等からの入力があるか確認
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
