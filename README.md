# Monitor Bot

A GitHub Release monitor bot that periodically checks repositories for new releases, generates AI-powered summaries via GitHub Models (GPT-4o-mini), and sends notifications to Telegram.

## How It Works

1. Reads last known versions from `last_versions.json`
2. Fetches the latest release for each monitored repository via GitHub API
3. When a new version is detected, calls AI to interpret the release notes
4. Sends the summary to a Telegram channel/group
5. Commits the updated `last_versions.json`

## Getting Started

### Prerequisites

- Go 1.23 or later
- A [GitHub Models](https://github.com/marketplace/models) API token
- A Telegram Bot (create one via [@BotFather](https://t.me/BotFather))
- The chat ID of your target Telegram group or channel

### 1. Fork or Clone

```bash
git clone https://github.com/zakir-web3/monitor-bot.git
cd monitor-bot
```

### 2. Configure Repositories

Edit the `githubRepos` list in `config.go` to monitor the repositories you care about:

```go
var githubRepos = []string{
    "ethereum/go-ethereum",
    "cosmos/cosmos-sdk",
}
```

### 3. Set Up Secrets

Go to your forked repository's **Settings > Secrets and variables > Actions**, and add the following secrets:

| Secret | Description |
|--------|-------------|
| `GH_PAT_CLASSIC_TOKEN` | GitHub personal access token (classic) for Models API |
| `TELEGRAM_BOT_TOKEN` | Telegram Bot token obtained from @BotFather |
| `TELEGRAM_CHAT_ID` | Target Telegram chat/group/channel ID |
| `GITHUB_TOKEN` | GitHub token (optional, increases API rate limit from 60 to 5000 req/hr) |

> **Tip:** To find your Telegram chat ID, send a message to your bot, then visit `https://api.telegram.org/bot<YOUR_TOKEN>/getUpdates` and look for the `chat.id` field.

### 4. Deploy via GitHub Actions

The workflow runs automatically at **09:30 CST (Beijing time)** every day. You can also trigger it manually from the **Actions** tab.

To change the schedule, edit the cron expression in `.github/workflows/monitor.yml`:

```yaml
on:
  schedule:
    - cron: '30 1 * * *'  # UTC time — adjust to your timezone
```

### 5. Run Locally (Optional)

```bash
export GH_PAT_CLASSIC_TOKEN="your-token"
export TELEGRAM_BOT_TOKEN="your-bot-token"
export TELEGRAM_CHAT_ID="your-chat-id"
go run .
```

## Customization

- **AI model / prompt** — Edit `config.go` to change the model, endpoint, or prompt template.
- **Message format** — Modify `msgHeader` / `msgFooter` in `config.go`. Messages are sent in HTML format.
- **Schedule** — Adjust the cron expression in `.github/workflows/monitor.yml`.
