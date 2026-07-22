package calendar

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/rahat/rahat/internal/db"
	"github.com/rahat/rahat/internal/store"
	"github.com/rahat/rahat/internal/users"
)

type fakeOAuthClient struct {
	authURL string
	token   OAuthToken
	events  []Event
}

func (f *fakeOAuthClient) AuthCodeURL(state string) string { return f.authURL + state }
func (f *fakeOAuthClient) ExchangeCode(context.Context, string) (OAuthToken, error) {
	return f.token, nil
}
func (f *fakeOAuthClient) ListEvents(context.Context, store.CalendarConnection, time.Time, *time.Location) ([]Event, error) {
	return f.events, nil
}

func TestSyncGoogleDayConvertsEventsToLocalBlocks(t *testing.T) {
	ctx := context.Background()
	sqlDB := openTestDB(t)
	if err := store.ApplyMigrations(ctx, sqlDB); err != nil {
		t.Fatal(err)
	}
	userSvc := users.NewService(users.NewRepository(sqlDB))
	connRepo := store.NewCalendarConnectionRepository(sqlDB)
	blockRepo := store.NewCalendarBlockRepository(sqlDB)
	stateRepo := store.NewOAuthStateRepository(sqlDB)
	client := &fakeOAuthClient{events: []Event{{ID: "evt-1", Summary: "School pickup", Start: time.Date(2026, 7, 20, 18, 0, 0, 0, time.UTC), End: time.Date(2026, 7, 20, 20, 30, 0, 0, time.UTC)}}}
	svc := NewService(userSvc, connRepo, blockRepo, stateRepo, client)

	user, _ := userSvc.Create(ctx, users.User{DisplayName: "TZ", Timezone: "America/New_York", DailyTimeBudgetMinutes: 30})
	_, err := connRepo.Upsert(ctx, store.CalendarConnection{UserID: user.ID, Provider: "google", CalendarID: "primary", AccessToken: "token", Timezone: user.Timezone})
	if err != nil {
		t.Fatal(err)
	}
	blocks, err := svc.SyncGoogleDay(ctx, user.ID, time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) != 1 {
		t.Fatalf("len(blocks) = %d, want 1", len(blocks))
	}
	if blocks[0].LocalDate != "2026-07-20" {
		t.Fatalf("local_date = %s, want 2026-07-20", blocks[0].LocalDate)
	}
	if blocks[0].Window != "afternoon" {
		t.Fatalf("window = %s, want afternoon", blocks[0].Window)
	}
	if blocks[0].Classification != "medium" {
		t.Fatalf("classification = %s, want medium", blocks[0].Classification)
	}
}

func TestConnectGoogleRequiresValidOAuthState(t *testing.T) {
	ctx := context.Background()
	sqlDB := openTestDB(t)
	if err := store.ApplyMigrations(ctx, sqlDB); err != nil {
		t.Fatal(err)
	}
	userSvc := users.NewService(users.NewRepository(sqlDB))
	connRepo := store.NewCalendarConnectionRepository(sqlDB)
	blockRepo := store.NewCalendarBlockRepository(sqlDB)
	stateRepo := store.NewOAuthStateRepository(sqlDB)
	client := &fakeOAuthClient{authURL: "https://example.test/oauth?state=", token: OAuthToken{AccessToken: "token", RefreshToken: "refresh", TokenType: "Bearer", Scope: "calendar.readonly"}}
	svc := NewService(userSvc, connRepo, blockRepo, stateRepo, client)
	now := time.Date(2026, 7, 16, 0, 0, 0, 0, time.UTC)
	svc.now = func() time.Time { return now }

	user, _ := userSvc.Create(ctx, users.User{DisplayName: "OAuth", Timezone: "UTC", DailyTimeBudgetMinutes: 30})
	authURL, err := svc.GoogleAuthURL(ctx, user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if authURL == "" {
		t.Fatal("expected auth URL")
	}
	state := authURL[len("https://example.test/oauth?state="):]
	conn, err := svc.ConnectGoogle(ctx, state, "code-123")
	if err != nil {
		t.Fatal(err)
	}
	if conn.UserID != user.ID {
		t.Fatalf("conn.UserID = %s, want %s", conn.UserID, user.ID)
	}
	if _, err := svc.ConnectGoogle(ctx, state, "code-123"); err == nil {
		t.Fatal("expected reused state to fail")
	}
	if _, err := svc.ConnectGoogle(ctx, "missing-state", "code-123"); err == nil {
		t.Fatal("expected missing state to fail")
	}
	_, err = stateRepo.Create(ctx, store.OAuthState{UserID: user.ID, Provider: "google", StateToken: "expired", ExpiresAt: now.Add(-time.Minute)})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ConnectGoogle(ctx, "expired", "code-123"); err == nil {
		t.Fatal("expected expired state to fail")
	}
}

func TestIsGoogleConnectedAndDisconnect(t *testing.T) {
	ctx := context.Background()
	sqlDB := openTestDB(t)
	if err := store.ApplyMigrations(ctx, sqlDB); err != nil {
		t.Fatal(err)
	}
	userSvc := users.NewService(users.NewRepository(sqlDB))
	connRepo := store.NewCalendarConnectionRepository(sqlDB)
	blockRepo := store.NewCalendarBlockRepository(sqlDB)
	stateRepo := store.NewOAuthStateRepository(sqlDB)
	client := &fakeOAuthClient{authURL: "https://example.test/oauth?state="}
	svc := NewService(userSvc, connRepo, blockRepo, stateRepo, client)

	user, _ := userSvc.Create(ctx, users.User{DisplayName: "Conn", Timezone: "UTC", DailyTimeBudgetMinutes: 30})

	connected, err := svc.IsGoogleConnected(ctx, user.ID)
	if err != nil {
		t.Fatalf("IsGoogleConnected error = %v", err)
	}
	if connected {
		t.Fatal("expected not connected")
	}

	authURL, err := svc.GoogleAuthURL(ctx, user.ID)
	if err != nil {
		t.Fatal(err)
	}
	state := authURL[len("https://example.test/oauth?state="):]
	if _, err := svc.ConnectGoogle(ctx, state, "code-123"); err != nil {
		t.Fatalf("ConnectGoogle error = %v", err)
	}

	connected, err = svc.IsGoogleConnected(ctx, user.ID)
	if err != nil {
		t.Fatalf("IsGoogleConnected error = %v", err)
	}
	if !connected {
		t.Fatal("expected connected")
	}

	if err := svc.DisconnectGoogle(ctx, user.ID); err != nil {
		t.Fatalf("DisconnectGoogle error = %v", err)
	}

	connected, err = svc.IsGoogleConnected(ctx, user.ID)
	if err != nil {
		t.Fatalf("IsGoogleConnected error = %v", err)
	}
	if connected {
		t.Fatal("expected not connected after disconnect")
	}
}

func TestSyncGoogleDayStoresAllDayEventAsLargeBlock(t *testing.T) {
	ctx := context.Background()
	sqlDB := openTestDB(t)
	if err := store.ApplyMigrations(ctx, sqlDB); err != nil {
		t.Fatal(err)
	}
	userSvc := users.NewService(users.NewRepository(sqlDB))
	connRepo := store.NewCalendarConnectionRepository(sqlDB)
	blockRepo := store.NewCalendarBlockRepository(sqlDB)
	stateRepo := store.NewOAuthStateRepository(sqlDB)
	client := &fakeOAuthClient{events: []Event{{ID: "evt-2", Summary: "Travel", Start: time.Date(2026, 7, 21, 0, 0, 0, 0, time.UTC), End: time.Date(2026, 7, 22, 0, 0, 0, 0, time.UTC), AllDay: true}}}
	svc := NewService(userSvc, connRepo, blockRepo, stateRepo, client)

	user, _ := userSvc.Create(ctx, users.User{DisplayName: "All Day", Timezone: "UTC", DailyTimeBudgetMinutes: 30})
	_, err := connRepo.Upsert(ctx, store.CalendarConnection{UserID: user.ID, Provider: "google", CalendarID: "primary", AccessToken: "token", Timezone: user.Timezone})
	if err != nil {
		t.Fatal(err)
	}
	blocks, err := svc.SyncGoogleDay(ctx, user.ID, time.Date(2026, 7, 21, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) != 1 {
		t.Fatalf("len(blocks) = %d, want 1", len(blocks))
	}
	if !blocks[0].IsAllDay || blocks[0].Classification != "large" || blocks[0].Window != "all-day" {
		t.Fatalf("unexpected all-day block: %+v", blocks[0])
	}
}

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	sqlDB, err := db.OpenSQLite(context.Background(), filepath.Join(t.TempDir(), "rahat.sqlite3"))
	if err != nil {
		t.Fatalf("OpenSQLite() error = %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	return sqlDB
}
