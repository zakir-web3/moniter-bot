package main

import (
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	modelsToken := mustEnv("GH_MODELS_TOKEN")
	telegramToken := mustEnv("TELEGRAM_BOT_TOKEN")
	chatID := mustEnv("TELEGRAM_CHAT_ID")

	lastVersion := readLastVersion()

	release, err := fetchLatestRelease()
	if err != nil {
		log.Fatalf("fetch release: %v", err)
	}

	if release.TagName == lastVersion {
		fmt.Printf("no new version (current: %s)\n", lastVersion)
		return
	}

	fmt.Printf("new version detected: %s (was: %s)\n", release.TagName, lastVersion)

	summary, err := interpretRelease(modelsToken, release)
	if err != nil {
		log.Fatalf("interpret release: %v", err)
	}

	msg := formatMessage(release, summary)
	if err := sendTelegram(telegramToken, chatID, msg); err != nil {
		log.Fatalf("send telegram: %v", err)
	}

	if err := writeLastVersion(release.TagName); err != nil {
		log.Fatalf("write version: %v", err)
	}

	fmt.Println("done")
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("env %s is required", key)
	}
	return v
}

func readLastVersion() string {
	data, err := os.ReadFile(versionFile)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func writeLastVersion(version string) error {
	return os.WriteFile(versionFile, []byte(version+"\n"), 0644)
}

func formatMessage(r *Release, summary string) string {
	msg := fmt.Sprintf(messageTmpl, r.TagName, summary, r.HTMLURL)
	if len([]rune(msg)) > telegramMsgLimit {
		runes := []rune(msg)
		msg = string(runes[:telegramMsgLimit]) + "..."
	}
	return msg
}
