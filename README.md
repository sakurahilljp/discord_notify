# discord_notify

A lightweight command-line interface (CLI) tool written in Go to send short messages to Discord text channels.

Command-line argument parsing is built using [`github.com/sakurahilljp/docopt-go`](https://github.com/sakurahilljp/docopt-go).

---

## 🚀 Installation & Building

### Build from source
```bash
git clone <repository_url>
cd discord_notify
go build -o discord_notify .
```

### Install globally
```bash
go install .
```

---

## 📋 Usage

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

### 3. Piping from standard input (stdin)

Pipe log outputs or script results directly to Discord.

```bash
echo "Disk usage exceeded 80%." | ./discord_notify -w "https://discord.com/api/webhooks/..."
```

---

## ⚙️ Options & Environment Variables

```
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
```

---

## 🧪 Running Tests

```bash
go test -v ./...
```

---

## 📜 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

