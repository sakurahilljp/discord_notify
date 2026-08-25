# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.2.0] - 2026-08-26

### Changed
- Changed Go module path to `github.com/sakurahilljp/discord_notify`.

## [1.1.0] - 2026-08-24

### Added
- Adopted **GitHub Flow** development model, documented in [`AGENTS.md`](./AGENTS.md) and [`CONTRIBUTING.md`](./CONTRIBUTING.md).
- Created `discord` Go package (`discord_notify/discord`) for sending Discord messages as a reusable library.
- Exported `Config`, `Client`, `Send`, `SendWebhook`, `SendBotMessage`, `SendFromEnv`, `NewConfigFromEnv` with functional options (`WithUsername`, `WithAvatarURL`, `WithTimeout`, `WithRetry`, `WithHTTPClient`).
- Support environment variable fallback and loading (`DISCORD_WEBHOOK_URL`, `DISCORD_BOT_TOKEN`, `DISCORD_CHANNEL_ID`, `DISCORD_USERNAME`, `DISCORD_AVATAR_URL`) in `discord` package.
- Support `context.Context` cancellation and timeouts in all package send functions.
- Comprehensive unit tests in `discord/discord_test.go`.

### Changed
- Refactored CLI into `cmd/discord_notify/` structure following standard Go project layout.
- Updated `Makefile` and `README.md` to support both library usage and CLI building.

## [1.0.0] - 2026-07-31

### Added
- Initial release of `discord_notify` CLI tool written in Go.
- Discord Webhook support for sending short messages.
- Discord Bot Token + Channel ID support for sending messages.
- Standard input (pipe) support for message streaming.
- CLI argument parsing powered by `github.com/sakurahilljp/docopt-go`.
- `-i` / `--ignore-errors` option to output warning and exit with code 0 on failure.
- `--timeout` option to configure HTTP request timeout.
- `--retry` option with exponential backoff and Discord Rate Limit (HTTP 429 `Retry-After`) handling.
- `Makefile` for automated build, test, lint, and formatting.
- MIT License.
- Complete English documentation and CLI help screens.

### Fixed
- Fixed bug where `-h` / `--help` and `--version` continued execution and caused `no message specified` error.
