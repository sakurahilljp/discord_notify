# discord_notify

A Go package and CLI tool to send short messages to Discord text channels via Webhooks or Bot Token.

Command-line argument parsing is built using [`github.com/sakurahilljp/docopt-go`](https://github.com/sakurahilljp/docopt-go).

---

## 📦 Using as a Go Library / Package

You can import and use `github.com/sakurahilljp/discord_notify/discord` directly in your Go applications.

### Installation
```bash
go get github.com/sakurahilljp/discord_notify
```

### 1. Send using Environment Variables

The `discord` package supports the same environment variables as the CLI. You can send messages with minimal code or inspect and override configuration:

#### Method A: `discord.SendFromEnv` (Shortest one-liner)
Reads credentials and defaults directly from environment variables. You can optionally pass functional options like `WithRetry`, `WithTimeout`, etc.
```go
package main

import (
	"context"
	"log"
	"time"

	"github.com/sakurahilljp/discord_notify/discord"
)

func main() {
	ctx := context.Background()

	// Automatically reads DISCORD_WEBHOOK_URL or DISCORD_BOT_TOKEN & DISCORD_CHANNEL_ID
	err := discord.SendFromEnv(ctx, "Build succeeded!",
		discord.WithRetry(3),
		discord.WithTimeout(5*time.Second),
	)
	if err != nil {
		log.Fatalf("failed to send: %v", err)
	}
}
```

#### Method B: `discord.NewConfigFromEnv` (Explicit configuration loading)
Loads configuration populated from environment variables, which can then be inspected or modified before sending.
```go
package main

import (
	"context"
	"log"

	"github.com/sakurahilljp/discord_notify/discord"
)

func main() {
	ctx := context.Background()

	// Load from environment variables
	cfg := discord.NewConfigFromEnv()
	cfg.Retry = 3
	cfg.Username = "Custom Worker" // override username

	if err := discord.Send(ctx, cfg, "Task completed."); err != nil {
		log.Fatalf("failed to send: %v", err)
	}
}
```

#### Method C: Automatic Environment Fallback in `NewClient` / `Send`
When using `discord.NewClient(cfg)` or `discord.Send(ctx, cfg, msg)`, any fields left empty (such as `WebhookURL`, `BotToken`, `ChannelID`, `Username`, `AvatarURL`) will automatically fall back to the corresponding environment variables.
```go
package main

import (
	"context"
	"log"

	"github.com/sakurahilljp/discord_notify/discord"
)

func main() {
	ctx := context.Background()

	// WebhookURL / Bot credentials will automatically fall back to environment variables
	cfg := discord.Config{
		Retry: 2,
	}
	if err := discord.Send(ctx, cfg, "Message with env fallback."); err != nil {
		log.Fatalf("failed to send: %v", err)
	}
}
```

---

### 2. Send via Webhook (Explicit URL)
```go
package main

import (
	"context"
	"log"
	"time"

	"github.com/sakurahilljp/discord_notify/discord"
)

func main() {
	ctx := context.Background()

	// 1. One-liner with functional options
	err := discord.SendWebhook(ctx, "https://discord.com/api/webhooks/...", "Build completed!",
		discord.WithUsername("CI Bot"),
		discord.WithRetry(3),
		discord.WithTimeout(5*time.Second),
	)
	if err != nil {
		log.Fatalf("failed to send: %v", err)
	}

	// 2. Using Config struct
	cfg := discord.Config{
		WebhookURL: "https://discord.com/api/webhooks/...",
		Username:   "Server Monitor",
		Retry:      3,
	}
	if err := discord.Send(ctx, cfg, "Server restarted."); err != nil {
		log.Fatalf("failed to send: %v", err)
	}

	// 3. Creating a reusable Client instance
	client, err := discord.NewClient(cfg)
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}
	_ = client.Send(ctx, "Message 1")
	_ = client.Send(ctx, "Message 2")
}
```

---

### 3. Send via Discord Bot API (Explicit Token & Channel)
```go
package main

import (
	"context"
	"log"

	"github.com/sakurahilljp/discord_notify/discord"
)

func main() {
	ctx := context.Background()

	err := discord.SendBotMessage(ctx, "YOUR_BOT_TOKEN", "YOUR_CHANNEL_ID", "Hello from bot!",
		discord.WithRetry(2),
	)
	if err != nil {
		log.Fatalf("failed to send: %v", err)
	}
}
```

---

### 4. Load from YAML Configuration & Profiles
You can load configuration from a YAML file for specific profiles (e.g. `dev`, `prod`):

```go
package main

import (
	"context"
	"log"

	"github.com/sakurahilljp/discord_notify/discord"
)

func main() {
	ctx := context.Background()

	// Load configuration for profile "prod"
	cfg, err := discord.LoadYAMLProfile("config.yaml", "prod")
	if err != nil {
		log.Fatalf("failed to load YAML config: %v", err)
	}

	if err := discord.Send(ctx, cfg, "Release deployed!"); err != nil {
		log.Fatalf("failed to send: %v", err)
	}
}
```

---

### 5. Load from `.env` File
```go
package main

import (
	"context"
	"log"

	"github.com/sakurahilljp/discord_notify/discord"
)

func main() {
	// Load environment variables from .env if present
	_ = discord.LoadDotEnvIfExists(".env")

	ctx := context.Background()
	_ = discord.SendFromEnv(ctx, "Notification with .env loaded.")
}
```

---

### 6. Send with Image or File Attachments
You can attach images (PNG, JPEG, GIF, WebP) or files using `WithFile`:

```go
package main

import (
	"context"
	"log"

	"github.com/sakurahilljp/discord_notify/discord"
)

func main() {
	ctx := context.Background()

	// Send message with image attachment
	err := discord.SendWebhook(ctx, "https://discord.com/api/webhooks/...", "Test report attached",
		discord.WithFile("./screenshot.png"),
	)
	if err != nil {
		log.Fatalf("failed to send: %v", err)
	}

	// Send image only (without message text)
	err = discord.SendFromEnv(ctx, "",
		discord.WithFile("./chart.png"),
	)
	if err != nil {
		log.Fatalf("failed to send: %v", err)
	}
}
```

---

## 🚀 CLI Installation & Building

### Using Makefile (Recommended)
```bash
make build    # Build binary to ./discord_notify
make test     # Run unit tests
make fmt      # Format source code
make lint     # Run go vet
make clean    # Remove built binary
```

### Install CLI binary globally
```bash
go install ./cmd/discord_notify
```

---

## 📋 CLI Usage

### Show Help / Version
```bash
./discord_notify --help
./discord_notify --version
```

### 1. Using Webhook (Easiest)

Obtain a Webhook URL from Discord channel settings (`Edit Channel` -> `Integrations` -> `Webhooks`).

#### Send via command argument
```bash
./discord_notify -w "https://discord.com/api/webhooks/..." "Hello from CLI!"
```

#### Send using environment variable
```bash
export DISCORD_WEBHOOK_URL="https://discord.com/api/webhooks/..."

./discord_notify "Build process completed successfully."
```

#### Customize sender username and avatar (Webhook only)
```bash
./discord_notify -w "https://discord.com/api/webhooks/..." \
  -u "CI Bot" \
  -a "https://example.com/avatar.png" \
  "Tests passed successfully."
```

---

### 2. Using Bot Token + Channel ID

Use a Bot Token generated in the Discord Developer Portal along with the target Text Channel ID.

#### Send via command options
```bash
./discord_notify -t "YOUR_BOT_TOKEN" -c "YOUR_CHANNEL_ID" "Notification from bot."
```

#### Send using environment variables
```bash
export DISCORD_BOT_TOKEN="YOUR_BOT_TOKEN"
export DISCORD_CHANNEL_ID="YOUR_CHANNEL_ID"

./discord_notify "Deployment finished."
```

---

### 3. Using YAML Configuration File & Profiles
You can manage multiple notification destinations in a YAML file (`.discord_notify.yaml`, `discord_notify.yaml`, or `~/.config/discord_notify/config.yaml`).

> [!TIP]
> A template configuration file [`discord_notify.example.yaml`](./discord_notify.example.yaml) and [`.env.example`](./.env.example) are included in this repository.

#### Example YAML Configuration (`.discord_notify.yaml`):
```yaml
default_profile: dev

profiles:
  dev:
    webhook_url: "https://discord.com/api/webhooks/..."
    username: "Dev Notifier"
    timeout: "5s"
    retry: 1

  prod:
    webhook_url: "https://discord.com/api/webhooks/..."
    username: "Release Bot"
    retry: 3

  bot-alerts:
    bot_token: "YOUR_BOT_TOKEN"
    channel_id: "YOUR_CHANNEL_ID"
    timeout: "15s"
```

#### Send using default profile
```bash
./discord_notify "Build succeeded!"
```

#### Send using a specific profile
```bash
./discord_notify -p prod "Release v1.3.0 deployed."
./discord_notify --config ./custom-config.yaml -p bot-alerts "Database alert."
```

---

### 4. Using `.env` File
`discord_notify` automatically checks and loads `.env` in the current working directory. You can also specify a custom path or disable it:

```bash
# Load from specific .env file
./discord_notify --env-file /etc/notify.env "Server restarted."

# Disable loading .env
./discord_notify --no-env "Message without .env"
```

---

### 5. Piping from standard input (stdin)

Pipe log outputs or script results directly to Discord.

```bash
echo "Disk usage exceeded 80%." | ./discord_notify -w "https://discord.com/api/webhooks/..."
```

---

### 6. Sending Image or File Attachments
Use `-f` or `--file` to attach an image (PNG, JPEG, GIF, WebP) or file:

```bash
# Send an image with a message caption
./discord_notify -w "https://discord.com/api/webhooks/..." -f ./screenshot.png "Test failure screenshot"

# Send an image only (without message text)
./discord_notify -p dev -f ./graph.png

# Pipe logs and attach a report file simultaneously
cat summary.txt | ./discord_notify -f ./coverage.html
```

---

## ⚙️ Supported Environment Variables

Both the CLI and the `discord` Go package support the following environment variables:

| Variable | Target Config Field | Description |
|---|---|---|
| `DISCORD_WEBHOOK_URL` | `Config.WebhookURL` | Discord Webhook URL |
| `DISCORD_BOT_TOKEN` | `Config.BotToken` | Discord Bot Token |
| `DISCORD_CHANNEL_ID` | `Config.ChannelID` | Discord Text Channel ID |
| `DISCORD_USERNAME` | `Config.Username` | Sender username (Webhook only) |
| `DISCORD_AVATAR_URL` | `Config.AvatarURL` | Avatar image URL (Webhook only) |

---

## 🧪 Running Tests

```bash
make test
# or
go test -v ./...
```

---

## 🤝 Contributing

This project follows the **GitHub Flow** model. Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details on branching conventions and workflow.

---

## 📜 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
