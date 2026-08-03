package telegram

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rahat/rahat/internal/auth"
	"github.com/rahat/rahat/internal/db"
	usr "github.com/rahat/rahat/internal/users"
)

type editFakeBot struct {
	messages []SendMessageRequest
}

func (f *editFakeBot) SendMessage(_ context.Context, req SendMessageRequest) error {
	f.messages = append(f.messages, req)
	return nil
}

type failingThenOkBot struct {
	messages      []SendMessageRequest
	failRemaining int
}

func (f *failingThenOkBot) SendMessage(_ context.Context, req SendMessageRequest) error {
	if f.failRemaining > 0 {
		f.failRemaining--
		return errors.New("sendMessage failed")
	}
	f.messages = append(f.messages, req)
	return nil
}

func newTestEditHandler(t *testing.T) (*EditCommandHandler, *editFakeBot, *auth.Service, *usr.Service) {
	t.Helper()
	sqlDB, err := db.OpenSQLite(context.Background(), filepath.Join(t.TempDir(), "rahat.sqlite3"))
	if err != nil {
		t.Fatalf("OpenSQLite() error = %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	userService := usr.NewService(usr.NewRepository(sqlDB))
	authService := auth.NewService(sqlDB, auth.NewRepository(sqlDB), "test-secret", 30*24*time.Hour)
	bot := &editFakeBot{}
	handler := NewEditCommandHandler(authService, userService, bot, "http://localhost:8080", nil)
	handler.LinkHost = "192.168.1.20:8080"
	return handler, bot, authService, userService
}

func TestEditCommandIssuesLinkForLinkedChat(t *testing.T) {
	handler, bot, authService, userService := newTestEditHandler(t)
	ctx := context.Background()

	user, err := userService.Create(ctx, usr.User{DisplayName: "Tester", Timezone: "UTC", DailyTimeBudgetMinutes: 30})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if _, err := userService.LinkTelegramChat(ctx, user.ID, "111"); err != nil {
		t.Fatalf("LinkTelegramChat() error = %v", err)
	}

	if err := handler.HandleMessage(ctx, &Message{Text: "/edit", Chat: &Chat{ID: 111, Type: "private"}}); err != nil {
		t.Fatalf("HandleMessage() error = %v", err)
	}

	if len(bot.messages) != 1 {
		t.Fatalf("expected 1 reply, got %d", len(bot.messages))
	}
	msg := bot.messages[0]
	if msg.ChatID != "111" {
		t.Fatalf("chat_id = %q, want 111", msg.ChatID)
	}
	if !strings.Contains(msg.Text, "private link") {
		t.Fatalf("unexpected text: %q", msg.Text)
	}
	if len(msg.ReplyMarkup.InlineKeyboard) != 1 || len(msg.ReplyMarkup.InlineKeyboard[0]) != 1 {
		t.Fatalf("expected single inline button, got %+v", msg.ReplyMarkup.InlineKeyboard)
	}
	button := msg.ReplyMarkup.InlineKeyboard[0][0]
	if button.Text != "Manage my routines" {
		t.Fatalf("button text = %q, want Manage my routines", button.Text)
	}
	if !strings.HasPrefix(button.URL, "http://192.168.1.20:8080/login?token=") {
		t.Fatalf("unexpected button url: %q", button.URL)
	}

	rawToken := strings.TrimPrefix(button.URL, "http://192.168.1.20:8080/login?token=")
	session, _, err := authService.ExchangeAccessGrant(ctx, rawToken)
	if err != nil {
		t.Fatalf("ExchangeAccessGrant() error = %v", err)
	}
	if session.UserID != user.ID {
		t.Fatalf("session user = %q, want %q", session.UserID, user.ID)
	}
}

func TestEditCommandDifferentChatsIssueDifferentLinks(t *testing.T) {
	handler, bot, authService, userService := newTestEditHandler(t)
	ctx := context.Background()

	first, err := userService.Create(ctx, usr.User{DisplayName: "First", Timezone: "UTC", DailyTimeBudgetMinutes: 30})
	if err != nil {
		t.Fatalf("Create() first error = %v", err)
	}
	second, err := userService.Create(ctx, usr.User{DisplayName: "Second", Timezone: "UTC", DailyTimeBudgetMinutes: 30})
	if err != nil {
		t.Fatalf("Create() second error = %v", err)
	}
	if _, err := userService.LinkTelegramChat(ctx, first.ID, "1001"); err != nil {
		t.Fatalf("link first error = %v", err)
	}
	if _, err := userService.LinkTelegramChat(ctx, second.ID, "1002"); err != nil {
		t.Fatalf("link second error = %v", err)
	}

	var tokens []string
	for _, chatID := range []int64{1001, 1002} {
		bot.messages = nil
		if err := handler.HandleMessage(ctx, &Message{Text: "/edit", Chat: &Chat{ID: chatID, Type: "private"}}); err != nil {
			t.Fatalf("HandleMessage(%d) error = %v", chatID, err)
		}
		if len(bot.messages) != 1 {
			t.Fatalf("expected 1 reply for chat %d, got %d", chatID, len(bot.messages))
		}
		button := bot.messages[0].ReplyMarkup.InlineKeyboard[0][0]
		tokens = append(tokens, button.URL)
	}
	if tokens[0] == tokens[1] {
		t.Fatal("expected different links for different chats")
	}

	sessionOne, _, err := authService.ExchangeAccessGrant(ctx, strings.TrimPrefix(tokens[0], "http://192.168.1.20:8080/login?token="))
	if err != nil {
		t.Fatalf("exchange first error = %v", err)
	}
	sessionTwo, _, err := authService.ExchangeAccessGrant(ctx, strings.TrimPrefix(tokens[1], "http://192.168.1.20:8080/login?token="))
	if err != nil {
		t.Fatalf("exchange second error = %v", err)
	}
	if sessionOne.UserID == sessionTwo.UserID {
		t.Fatal("expected sessions for different users")
	}
}

func TestEditCommandUnknownChatGetsSafeGuidance(t *testing.T) {
	handler, bot, _, _ := newTestEditHandler(t)
	ctx := context.Background()

	if err := handler.HandleMessage(ctx, &Message{Text: "/edit", Chat: &Chat{ID: 999, Type: "private"}}); err != nil {
		t.Fatalf("HandleMessage() error = %v", err)
	}

	if len(bot.messages) != 1 {
		t.Fatalf("expected 1 reply, got %d", len(bot.messages))
	}
	msg := bot.messages[0]
	if msg.Text == "" {
		t.Fatal("expected a non-empty guidance reply")
	}
	if msg.ReplyMarkup != nil {
		t.Fatal("expected no inline keyboard for unknown chat")
	}
}

func TestEditCommandLinkSingleUse(t *testing.T) {
	handler, bot, authService, userService := newTestEditHandler(t)
	ctx := context.Background()

	user, err := userService.Create(ctx, usr.User{DisplayName: "Tester", Timezone: "UTC", DailyTimeBudgetMinutes: 30})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if _, err := userService.LinkTelegramChat(ctx, user.ID, "333"); err != nil {
		t.Fatalf("LinkTelegramChat() error = %v", err)
	}

	_ = handler.HandleMessage(ctx, &Message{Text: "/edit", Chat: &Chat{ID: 333, Type: "private"}})
	button := bot.messages[0].ReplyMarkup.InlineKeyboard[0][0]
	rawToken := strings.TrimPrefix(button.URL, "http://192.168.1.20:8080/login?token=")

	if _, _, err := authService.ExchangeAccessGrant(ctx, rawToken); err != nil {
		t.Fatalf("first exchange error = %v", err)
	}
	if _, _, err := authService.ExchangeAccessGrant(ctx, rawToken); !errors.Is(err, auth.ErrUsed) {
		t.Fatalf("expected ErrUsed, got %v", err)
	}
}

func TestEditCommandFallbackToPlainText(t *testing.T) {
	ctx := context.Background()
	sqlDB, err := db.OpenSQLite(context.Background(), filepath.Join(t.TempDir(), "rahat.sqlite3"))
	if err != nil {
		t.Fatalf("OpenSQLite() error = %v", err)
	}
	defer sqlDB.Close()
	userService := usr.NewService(usr.NewRepository(sqlDB))
	authService := auth.NewService(sqlDB, auth.NewRepository(sqlDB), "test-secret", 30*24*time.Hour)

	user, err := userService.Create(ctx, usr.User{DisplayName: "Tester", Timezone: "UTC", DailyTimeBudgetMinutes: 30})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if _, err := userService.LinkTelegramChat(ctx, user.ID, "444"); err != nil {
		t.Fatalf("LinkTelegramChat() error = %v", err)
	}

	bot := &failingThenOkBot{failRemaining: 1}
	handler := NewEditCommandHandler(authService, userService, bot, "http://localhost:8080", nil)
	handler.LinkHost = "192.168.1.20:8080"
	if err := handler.HandleMessage(ctx, &Message{Text: "/edit", Chat: &Chat{ID: 444, Type: "private"}}); err != nil {
		t.Fatalf("HandleMessage() error = %v", err)
	}

	if len(bot.messages) != 1 {
		t.Fatalf("expected 1 fallback message, got %d", len(bot.messages))
	}
	msg := bot.messages[0]
	if msg.ReplyMarkup != nil {
		t.Fatal("expected no inline keyboard in fallback message")
	}
	if !strings.Contains(msg.Text, "http://192.168.1.20:8080/login?token=") {
		t.Fatalf("fallback message missing link: %q", msg.Text)
	}
	if !strings.Contains(msg.Text, "Do not forward it") {
		t.Fatalf("fallback message missing warning: %q", msg.Text)
	}
}

func TestEditCommandRequiresExplicitHostForLocalOrigin(t *testing.T) {
	handler, bot, _, userService := newTestEditHandler(t)
	ctx := context.Background()
	handler.LinkHost = ""

	user, err := userService.Create(ctx, usr.User{DisplayName: "Tester", Timezone: "UTC", DailyTimeBudgetMinutes: 30})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if _, err := userService.LinkTelegramChat(ctx, user.ID, "555"); err != nil {
		t.Fatalf("LinkTelegramChat() error = %v", err)
	}

	if err := handler.HandleMessage(ctx, &Message{Text: "/edit", Chat: &Chat{ID: 555, Type: "private"}}); err != nil {
		t.Fatalf("HandleMessage() error = %v", err)
	}
	if len(bot.messages) != 1 {
		t.Fatalf("expected one configuration guidance message, got %d", len(bot.messages))
	}
	if bot.messages[0].ReplyMarkup != nil {
		t.Fatal("expected no button when local link host is not configured")
	}
	if !strings.Contains(bot.messages[0].Text, "TELEGRAM_LINK_HOST") {
		t.Fatalf("expected host configuration guidance, got %q", bot.messages[0].Text)
	}
}

func TestEditCommandRejectsLoopbackLinkHost(t *testing.T) {
	handler, bot, _, userService := newTestEditHandler(t)
	ctx := context.Background()
	handler.LinkHost = "127.0.0.1:8080"

	user, err := userService.Create(ctx, usr.User{DisplayName: "Tester", Timezone: "UTC", DailyTimeBudgetMinutes: 30})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if _, err := userService.LinkTelegramChat(ctx, user.ID, "557"); err != nil {
		t.Fatalf("LinkTelegramChat() error = %v", err)
	}

	if err := handler.HandleMessage(ctx, &Message{Text: "/edit", Chat: &Chat{ID: 557, Type: "private"}}); err != nil {
		t.Fatalf("HandleMessage() error = %v", err)
	}
	if len(bot.messages) != 1 || bot.messages[0].ReplyMarkup != nil {
		t.Fatalf("expected configuration guidance without a button, got %+v", bot.messages)
	}
}

func TestEditCommandUsesPublicWebOrigin(t *testing.T) {
	handler, bot, _, userService := newTestEditHandler(t)
	ctx := context.Background()
	handler.webOrigin = "https://rahat.example.com"
	handler.LinkHost = ""

	user, err := userService.Create(ctx, usr.User{DisplayName: "Tester", Timezone: "UTC", DailyTimeBudgetMinutes: 30})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if _, err := userService.LinkTelegramChat(ctx, user.ID, "556"); err != nil {
		t.Fatalf("LinkTelegramChat() error = %v", err)
	}

	if err := handler.HandleMessage(ctx, &Message{Text: "/edit", Chat: &Chat{ID: 556, Type: "private"}}); err != nil {
		t.Fatalf("HandleMessage() error = %v", err)
	}
	button := bot.messages[0].ReplyMarkup.InlineKeyboard[0][0]
	if !strings.HasPrefix(button.URL, "https://rahat.example.com/login?token=") {
		t.Fatalf("unexpected public button URL: %q", button.URL)
	}
}

func TestEditCommandGroupChatIgnored(t *testing.T) {
	handler, bot, _, _ := newTestEditHandler(t)
	ctx := context.Background()

	if err := handler.HandleMessage(ctx, &Message{Text: "/edit", Chat: &Chat{ID: -1, Type: "group"}}); err != nil {
		t.Fatalf("HandleMessage() error = %v", err)
	}
	if len(bot.messages) != 0 {
		t.Fatalf("expected no reply in group chat, got %+v", bot.messages)
	}
}

func TestIsEditCommand(t *testing.T) {
	cases := []struct {
		text string
		want bool
	}{
		{"/edit", true},
		{"  /edit  ", true},
		{"/edit@RahatBot", true},
		{"/edit extra", true},
		{"/start", false},
		{"", false},
		{"edit", false},
	}
	for _, tc := range cases {
		if got := isEditCommand(tc.text); got != tc.want {
			t.Fatalf("isEditCommand(%q) = %v, want %v", tc.text, got, tc.want)
		}
	}
}
