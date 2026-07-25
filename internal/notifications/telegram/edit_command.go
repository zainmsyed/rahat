package telegram

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/rahat/rahat/internal/auth"
	"github.com/rahat/rahat/internal/users"
)

// DefaultTelegramAccessGrantTTL is the lifetime of a self-service link issued via /edit.
const DefaultTelegramAccessGrantTTL = 15 * time.Minute

// EditCommandHandler processes Telegram /edit messages and replies with a single-use,
// short-lived web access link bound to the Rahat user linked to the incoming chat.
type EditCommandHandler struct {
	auth      *auth.Service
	users     *users.Service
	bot       BotClient
	webOrigin string
	logger    *slog.Logger
}

// NewEditCommandHandler creates an EditCommandHandler.
func NewEditCommandHandler(authService *auth.Service, userService *users.Service, bot BotClient, webOrigin string, logger *slog.Logger) *EditCommandHandler {
	if logger == nil {
		logger = slog.Default()
	}
	return &EditCommandHandler{
		auth:      authService,
		users:     userService,
		bot:       bot,
		webOrigin: strings.TrimRight(webOrigin, "/"),
		logger:    logger,
	}
}

// HandleMessage implements the MessageHandler interface. It only acts on private
// chat messages that look like the /edit command.
func (h *EditCommandHandler) HandleMessage(ctx context.Context, msg *Message) error {
	if msg == nil || msg.Chat == nil || msg.Chat.Type != "private" {
		return nil
	}
	if !isEditCommand(msg.Text) {
		return nil
	}

	chatID := fmt.Sprintf("%d", msg.Chat.ID)
	user, err := h.users.GetByTelegramChatID(ctx, chatID)
	if err != nil {
		if errors.Is(err, users.ErrNotFound) {
			return h.reply(ctx, chatID, "We don't recognize this Telegram chat. If you've already connected Telegram during onboarding, make sure you're messaging from the same account. Otherwise, finish onboarding first.")
		}
		h.logger.Warn("edit command: failed to resolve user by chat", "chat_id", chatID, "error", err)
		return h.reply(ctx, chatID, "We couldn't look up your account right now. Please try again later.")
	}

	grant, rawToken, err := h.auth.IssueAccessGrant(ctx, user.ID, DefaultTelegramAccessGrantTTL)
	if err != nil {
		h.logger.Warn("edit command: failed to issue access grant", "user_id", user.ID, "chat_id", chatID, "error", err)
		return h.reply(ctx, chatID, "We couldn't create a sign-in link right now. Please try again later.")
	}

	link := h.buildLink(rawToken)
	text := fmt.Sprintf(
		"Here is your private link to manage your routines. It expires in %d minutes and can only be used once. Do not forward it.",
		int(DefaultTelegramAccessGrantTTL.Minutes()),
	)
	markup := &ReplyMarkup{
		InlineKeyboard: [][]InlineButton{{{
			Text: "Manage my routines",
			URL:  link,
		}}},
	}

	if err := h.bot.SendMessage(ctx, SendMessageRequest{ChatID: chatID, Text: text, ReplyMarkup: markup}); err != nil {
		h.logger.Warn("edit command: failed to send reply", "user_id", user.ID, "chat_id", chatID, "error", err)
		return err
	}

	h.logger.Info(
		"telegram edit link issued",
		"user_id", user.ID,
		"chat_id", chatID,
		"grant_selector", grant.Selector,
		"expires_at", grant.ExpiresAt.Format(time.RFC3339),
	)
	return nil
}

func (h *EditCommandHandler) reply(ctx context.Context, chatID, text string) error {
	return h.bot.SendMessage(ctx, SendMessageRequest{ChatID: chatID, Text: text})
}

func (h *EditCommandHandler) buildLink(rawToken string) string {
	return h.webOrigin + "/login?token=" + url.QueryEscape(rawToken)
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
