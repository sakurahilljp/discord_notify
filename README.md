# discord_notify

DiscordのText Channelにメッセージを送信するための軽量なGolang製CLIツールです。  
**Webhook方式** および **Discord Bot Token方式** の両方に対応しており、コマンド引数や標準入力（パイプ）からのメッセージ送信をサポートしています。

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

コマンドの実行結果やログをパイプ経由でDiscordに直接送信できます。

```bash
echo "サーバーのディスク使用率が80%を超えました。" | ./discord_notify -w "https://discord.com/api/webhooks/..."
```

```bash
cat build.log | ./discord_notify
```

---

## ⚙️ コマンドラインオプション & 環境変数一覧

| オプション (短縮) | オプション (ロング) | 対応する環境変数 | 説明 |
|---|---|---|---|
| `-w` | `-webhook` | `DISCORD_WEBHOOK_URL` | Discord Webhook URL |
| `-t` | `-token` | `DISCORD_BOT_TOKEN` | Discord Bot Token |
| `-c` | `-channel` | `DISCORD_CHANNEL_ID` | 送信先のText Channel ID |
| `-m` | `-message` | - | 送信メッセージ |
| `-u` | `-username` | `DISCORD_USERNAME` | Webhook表示ユーザー名 |
| `-a` | `-avatar` | `DISCORD_AVATAR_URL` | Webhook表示アバター画像URL |
| `-v` | `-verbose` | - | 詳細な送信完了メッセージを出力 |

---

## 🧪 テストの実行

```bash
go test -v ./...
```
