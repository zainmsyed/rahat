package main

import (
	"context"
	"strings"

	ntg "github.com/rahat/rahat/internal/notifications/telegram"
)

// telegramMessageRouter routes incoming Telegram private-chat messages to the
// appropriate handler: /edit goes to the self-service access-link handler, and
// everything else (including /start onboarding codes) goes to onboarding.
type telegramMessageRouter struct {
	onboarding ntg.MessageHandler
	edit       ntg.MessageHandler
}

func (r *telegramMessageRouter) HandleMessage(ctx context.Context, msg *ntg.Message) error {
	if msg == nil || msg.Chat == nil {
		return nil
	}
	if r.edit != nil && isEditCommand(msg.Text) {
		return r.edit.HandleMessage(ctx, msg)
	}
	if r.onboarding != nil {
		return r.onboarding.HandleMessage(ctx, msg)
	}
	return nil
}

func isEditCommand(text string) bool {
	text = strings.TrimSpace(text)
	if text == "" {
		return false
	}
	parts := strings.Fields(text)
	cmd := strings.ToLower(parts[0])
	return cmd == "/edit" || strings.HasPrefix(cmd, "/edit@")
}
