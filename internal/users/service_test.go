package users

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/rahat/rahat/internal/db"
)

func newTestUserService(t *testing.T) *Service {
	t.Helper()
	sqlDB, err := db.OpenSQLite(context.Background(), filepath.Join(t.TempDir(), "rahat.sqlite3"))
	if err != nil {
		t.Fatalf("OpenSQLite() error = %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	return NewService(NewRepository(sqlDB))
}

func TestLinkTelegramChatResolvesToUser(t *testing.T) {
	svc := newTestUserService(t)
	ctx := context.Background()

	user, err := svc.Create(ctx, User{DisplayName: "Tester", Timezone: "UTC", DailyTimeBudgetMinutes: 30})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	linked, err := svc.LinkTelegramChat(ctx, user.ID, "123456")
	if err != nil {
		t.Fatalf("LinkTelegramChat() error = %v", err)
	}
	if linked.TelegramChatID != "123456" {
		t.Fatalf("telegram_chat_id = %q, want 123456", linked.TelegramChatID)
	}

	found, err := svc.GetByTelegramChatID(ctx, "123456")
	if err != nil {
		t.Fatalf("GetByTelegramChatID() error = %v", err)
	}
	if found.ID != user.ID {
		t.Fatalf("resolved user = %q, want %q", found.ID, user.ID)
	}
}

func TestTelegramChatOneToOneEnforced(t *testing.T) {
	svc := newTestUserService(t)
	ctx := context.Background()

	first, err := svc.Create(ctx, User{DisplayName: "First", Timezone: "UTC", DailyTimeBudgetMinutes: 30})
	if err != nil {
		t.Fatalf("Create() first error = %v", err)
	}
	second, err := svc.Create(ctx, User{DisplayName: "Second", Timezone: "UTC", DailyTimeBudgetMinutes: 30})
	if err != nil {
		t.Fatalf("Create() second error = %v", err)
	}

	if _, err := svc.LinkTelegramChat(ctx, first.ID, "shared-chat"); err != nil {
		t.Fatalf("LinkTelegramChat() first error = %v", err)
	}
	if _, err := svc.LinkTelegramChat(ctx, second.ID, "shared-chat"); !errors.Is(err, ErrTelegramChatLinked) {
		t.Fatalf("expected ErrTelegramChatLinked, got %v", err)
	}
}

func TestTelegramChatIdempotentForSameUser(t *testing.T) {
	svc := newTestUserService(t)
	ctx := context.Background()

	user, err := svc.Create(ctx, User{DisplayName: "Tester", Timezone: "UTC", DailyTimeBudgetMinutes: 30})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if _, err := svc.LinkTelegramChat(ctx, user.ID, "chat-1"); err != nil {
		t.Fatalf("first link error = %v", err)
	}
	if _, err := svc.LinkTelegramChat(ctx, user.ID, "chat-1"); err != nil {
		t.Fatalf("second link error = %v", err)
	}
}

func TestGetByTelegramChatIDUnknownChat(t *testing.T) {
	svc := newTestUserService(t)
	ctx := context.Background()

	_, err := svc.GetByTelegramChatID(ctx, "no-such-chat")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
