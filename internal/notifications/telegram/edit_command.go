package telegram

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
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
	LinkHost  string
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
			h.sendReply(ctx, chatID, "We don't recognize this Telegram chat. If you've already connected Telegram during onboarding, make sure you're messaging from the same account. Otherwise, finish onboarding first.")
			return nil
		}
		h.logger.Warn("edit command: failed to resolve user by chat", "chat_id", chatID, "error", err)
		h.sendReply(ctx, chatID, "We couldn't look up your account right now. Please try again later.")
		return nil
	}

	base, err := h.buildLinkBase()
	if err != nil {
		h.logger.Warn("edit command: management link host is not configured", "chat_id", chatID, "error", err)
		h.sendReply(ctx, chatID, "We couldn't create a reachable management link right now. Please ask the operator to configure TELEGRAM_LINK_HOST with the reachable IP address and port.")
		return nil
	}

	grant, rawToken, err := h.auth.IssueAccessGrant(ctx, user.ID, DefaultTelegramAccessGrantTTL)
	if err != nil {
		h.logger.Warn("edit command: failed to issue access grant", "user_id", user.ID, "chat_id", chatID, "error", err)
		h.sendReply(ctx, chatID, "We couldn't create a sign-in link right now. Please try again later.")
		return nil
	}

	link := base + "/login?token=" + url.QueryEscape(rawToken)
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

	if err := h.bot.SendMessage(ctx, SendMessageRequest{ChatID: chatID, Text: text, ReplyMarkup: markup}); err == nil {
		h.logger.Info(
			"telegram edit link issued",
			"user_id", user.ID,
			"chat_id", chatID,
			"grant_selector", grant.Selector,
			"expires_at", grant.ExpiresAt.Format(time.RFC3339),
		)
		return nil
	}

	// If the inline-button message fails (e.g. Telegram rejects a localhost URL in dev),
	// fall back to a plain-text reply so the user still receives the link. Do not propagate
	// the error, because Telegram will retry the update and we have already issued the grant.
	h.logger.Warn(
		"edit command: inline-button reply failed, falling back to plain text",
		"user_id", user.ID,
		"chat_id", chatID,
		"grant_selector", grant.Selector,
		"error", err,
	)
	fallback := fmt.Sprintf(
		"Here is your private link to manage your routines. It expires in %d minutes and can only be used once. Do not forward it.\n\n%s",
		int(DefaultTelegramAccessGrantTTL.Minutes()),
		link,
	)
	if fallbackErr := h.sendReply(ctx, chatID, fallback); fallbackErr != nil {
		h.logger.Warn("edit command: plain-text fallback also failed", "user_id", user.ID, "chat_id", chatID, "error", fallbackErr)
	}
	return nil
}

func (h *EditCommandHandler) sendReply(ctx context.Context, chatID, text string) error {
	if err := h.bot.SendMessage(ctx, SendMessageRequest{ChatID: chatID, Text: text}); err != nil {
		h.logger.Warn("edit command: failed to send reply", "chat_id", chatID, "error", err)
		return err
	}
	return nil
}

func (h *EditCommandHandler) buildLinkBase() (string, error) {
	parsed, err := url.Parse(h.webOrigin)
	if err != nil || parsed.Scheme == "" || parsed.Hostname() == "" {
		return "", errors.New("WEB_ORIGIN must be an absolute HTTP(S) URL")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", errors.New("WEB_ORIGIN must use http or https")
	}
	parsed.Path = strings.TrimRight(parsed.Path, "/")
	parsed.RawQuery = ""
	parsed.Fragment = ""
	parsed.User = nil

	if h.LinkHost != "" {
		linkHost, err := parseLinkHost(h.LinkHost)
		if err != nil {
			return "", err
		}
		if isLocalHost(parsed.Hostname()) && !isUsableConfiguredIP(linkHost.Hostname()) {
			return "", errors.New("TELEGRAM_LINK_HOST must be a reachable non-loopback IP address for a local WEB_ORIGIN")
		}
		parsed.Host = linkHost.Host
	} else if isLocalHost(parsed.Hostname()) {
		return "", errors.New("TELEGRAM_LINK_HOST is required when WEB_ORIGIN uses localhost or a loopback address")
	}

	return strings.TrimRight(parsed.String(), "/"), nil
}

func parseLinkHost(value string) (*url.URL, error) {
	if strings.Contains(value, "://") {
		return nil, errors.New("TELEGRAM_LINK_HOST must be host:port without a scheme")
	}
	parsed, err := url.Parse("//" + strings.TrimSpace(value))
	if err != nil || parsed.Hostname() == "" || parsed.Port() == "" || parsed.Path != "" || parsed.RawQuery != "" || parsed.Fragment != "" || parsed.User != nil {
		return nil, errors.New("TELEGRAM_LINK_HOST must include a reachable host and port")
	}
	return parsed, nil
}

func isLocalHost(host string) bool {
	ip := net.ParseIP(host)
	return strings.EqualFold(host, "localhost") || ip != nil && ip.IsLoopback()
}

func isUsableConfiguredIP(host string) bool {
	ip := net.ParseIP(host)
	return ip != nil && !ip.IsLoopback() && !ip.IsUnspecified() && !ip.IsMulticast()
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
