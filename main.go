package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"log/slog"
	"os"
	"time"
)

func main() {
	modelsToken := mustEnv("GH_PAT_CLASSIC_TOKEN")
	telegramToken := mustEnv("TELEGRAM_BOT_TOKEN")
	chatID := mustEnv("TELEGRAM_CHAT_ID")
	githubToken := os.Getenv("GITHUB_TOKEN")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	versions := readVersions()
	var hasError bool

	for _, repo := range githubRepos {
		lastVersion := versions[repo]

		release, err := retry(ctx, "fetch-release", func(ctx context.Context) (*Release, error) {
			return fetchLatestRelease(ctx, repo, githubToken)
		})
		if err != nil {
			slog.Error("fetch release failed", "repo", repo, "error", err)
			hasError = true
			continue
		}

		if release.TagName == lastVersion {
			slog.Info("no new version", "repo", repo, "version", lastVersion)
			continue
		}

		slog.Info("new version detected", "repo", repo, "new", release.TagName, "old", lastVersion)

		summary, err := retry(ctx, "interpret-release", func(ctx context.Context) (string, error) {
			return interpretRelease(ctx, modelsToken, repo, release)
		})
		if err != nil {
			slog.Error("interpret release failed", "repo", repo, "error", err)
			hasError = true
			continue
		}

		msg := formatMessage(repo, release, summary)
		if err := retryDo(ctx, "send-telegram", func(ctx context.Context) error {
			return sendTelegram(ctx, telegramToken, chatID, msg)
		}); err != nil {
			slog.Error("send telegram failed", "repo", repo, "error", err)
			hasError = true
			continue
		}

		versions[repo] = release.TagName
	}

	if err := writeVersions(versions); err != nil {
		slog.Error("write versions failed", "error", err)
		os.Exit(1)
	}

	if hasError {
		slog.Error("completed with errors")
		os.Exit(1)
	}

	slog.Info("done")
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		slog.Error("required env not set", "key", key)
		os.Exit(1)
	}
	return v
}

func readVersions() map[string]string {
	data, err := os.ReadFile(versionFile)
	if err != nil {
		return make(map[string]string)
	}
	var versions map[string]string
	if err := json.Unmarshal(data, &versions); err != nil {
		return make(map[string]string)
	}
	return versions
}

func writeVersions(versions map[string]string) error {
	data, err := json.MarshalIndent(versions, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(versionFile, append(data, '\n'), 0644)
}

func formatMessage(repo string, r *Release, summary string) string {
	header := fmt.Sprintf(msgHeader, html.EscapeString(repo), html.EscapeString(r.TagName))
	footer := fmt.Sprintf(msgFooter, r.HTMLURL)
	suffix := "\n\n" + footer

	escaped := html.EscapeString(summary)
	maxLen := telegramMsgLimit - len([]rune(header)) - len([]rune(suffix)) - 5
	runes := []rune(escaped)
	if len(runes) > maxLen {
		escaped = string(runes[:maxLen]) + "..."
	}
	return header + "\n\n" + escaped + suffix
}
