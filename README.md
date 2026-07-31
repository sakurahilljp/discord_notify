# discord_notify

DiscordのText Channelにメッセージを送信するための軽量なGolang製CLIツールです。  
コマンドライン引数解析には [`github.com/sakurahilljp/docopt-go`](https://github.com/sakurahilljp/docopt-go) を使用しています。

---

## 🚀 インストール / ビルド

### ソースコードからビルド
```bash
git clone <repository_url>
cd discord_notify
go build -o discord_notify .
```

### システムにインストール
```bash
go install .
```

---

## 📋 使い方

### ヘルプ表示 / バージョン表示
```bash
./discord_notify --help
./discord_notify --version
```

### 1. Webhook を使用する場合（最も簡単）

Discordの「チャンネル設定」→「連携」→「ウェブフック」からWebhook URLを取得して使用します。

#### コマンドライン引数で送信
```bash
./discord_notify -w "https://discord.com/api/webhooks/..." "こんにちは！"
```

#### 環境変数を事前に設定して送信
```bash
export DISCORD_WEBHOOK_URL="https://discord.com/api/webhooks/..."

./discord_notify "ビルド処理が完了しました。"
```

#### カスタムのユーザー名とアバター画像を指定して送信 (Webhook限定)
```bash
./discord_notify -w "https://discord.com/api/webhooks/..." \
  -u "CI Bot" \
  -a "https://example.com/avatar.png" \
  "テストが成功しました。"
```

---

### 2. Bot Token + Channel ID を使用する場合

Discord Developer Portalで作成したBot Tokenと送信先のText Channel IDを使用します。

#### コマンドライン引数で送信
```bash
./discord_notify -t "YOUR_BOT_TOKEN" -c "YOUR_CHANNEL_ID" "Botからの通知です。"
```

#### 環境変数で設定して送信
```bash
export DISCORD_BOT_TOKEN="YOUR_BOT_TOKEN"
export DISCORD_CHANNEL_ID="YOUR_CHANNEL_ID"

./discord_notify "デプロイが成功しました。"
```

---

### 3. パイプ（標準入力）からの送信

```bash
echo "サーバーのディスク使用率が80%を超えました。" | ./discord_notify -w "https://discord.com/api/webhooks/..."
```

---

## ⚙️ オプション & 環境変数一覧

```
Options:
  -h --help             Show help message.
  --version             Show version.
  -w --webhook=<url>    Discord Webhook URL.
  -t --token=<token>    Discord Bot Token.
  -c --channel=<id>     Discord Channel ID.
  -m --message=<msg>    Message to send.
  -u --username=<name>  Sender username (Webhook only).
  -a --avatar=<url>     Avatar image URL (Webhook only).
  -v --verbose          Show verbose output log.
```
