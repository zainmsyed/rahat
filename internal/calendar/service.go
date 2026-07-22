package calendar

import (
	"context"
	"fmt"
	"time"

	"github.com/rahat/rahat/internal/store"
	"github.com/rahat/rahat/internal/users"
)

type OAuthClient interface {
	AuthCodeURL(state string) string
	ExchangeCode(context.Context, string) (OAuthToken, error)
	ListEvents(context.Context, store.CalendarConnection, time.Time, *time.Location) ([]Event, error)
}

type OAuthToken struct {
	AccessToken  string
	RefreshToken string
	TokenType    string
	Expiry       *time.Time
	Scope        string
}

type Event struct {
	ID       string
	Summary  string
	Start    time.Time
	End      time.Time
	AllDay   bool
	TimeZone string
}

type Service struct {
	users       *users.Service
	connections *store.CalendarConnectionRepository
	blocks      *store.CalendarBlockRepository
	states      *store.OAuthStateRepository
	google      OAuthClient
	now         func() time.Time
}

func NewService(usersService *users.Service, connectionRepo *store.CalendarConnectionRepository, blockRepo *store.CalendarBlockRepository, stateRepo *store.OAuthStateRepository, googleClient OAuthClient) *Service {
	return &Service{users: usersService, connections: connectionRepo, blocks: blockRepo, states: stateRepo, google: googleClient, now: func() time.Time { return time.Now().UTC() }}
}

func (s *Service) GoogleAuthURL(ctx context.Context, userID string) (string, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("load user: %w", err)
	}
	state, err := s.states.Create(ctx, store.OAuthState{UserID: user.ID, Provider: "google"})
	if err != nil {
		return "", fmt.Errorf("create google oauth state: %w", err)
	}
	return s.google.AuthCodeURL(state.StateToken), nil
}

func (s *Service) CompleteGoogleOAuth(ctx context.Context, userID string) (store.CalendarConnection, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return store.CalendarConnection{}, fmt.Errorf("load user: %w", err)
	}
	conn, err := s.connections.GetByUserAndProvider(ctx, userID, "google")
	if err == nil {
		if user.Timezone != "" && conn.Timezone != user.Timezone {
			conn.Timezone = user.Timezone
			return s.connections.Upsert(ctx, conn)
		}
		return conn, nil
	}
	return store.CalendarConnection{UserID: user.ID, Provider: "google", CalendarID: "primary", Timezone: user.Timezone}, nil
}

func (s *Service) ConnectGoogle(ctx context.Context, stateToken, code string) (store.CalendarConnection, error) {
	state, err := s.states.Consume(ctx, "google", stateToken, s.now())
	if err != nil {
		return store.CalendarConnection{}, fmt.Errorf("consume google oauth state: %w", err)
	}
	user, err := s.users.GetByID(ctx, state.UserID)
	if err != nil {
		return store.CalendarConnection{}, fmt.Errorf("load user: %w", err)
	}
	token, err := s.google.ExchangeCode(ctx, code)
	if err != nil {
		return store.CalendarConnection{}, fmt.Errorf("exchange google auth code: %w", err)
	}
	conn := store.CalendarConnection{
		UserID:       user.ID,
		Provider:     "google",
		CalendarID:   "primary",
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
		Expiry:       token.Expiry,
		Scope:        token.Scope,
		Timezone:     user.Timezone,
	}
	return s.connections.Upsert(ctx, conn)
}

func (s *Service) IsGoogleConnected(ctx context.Context, userID string) (bool, error) {
	_, err := s.connections.GetByUserAndProvider(ctx, userID, "google")
	if err != nil {
		return false, nil
	}
	return true, nil
}

func (s *Service) DisconnectGoogle(ctx context.Context, userID string) error {
	if err := s.connections.DeleteByUserAndProvider(ctx, userID, "google"); err != nil {
		return fmt.Errorf("disconnect google: %w", err)
	}
	return nil
}

func (s *Service) SyncGoogleDay(ctx context.Context, userID string, day time.Time) ([]store.CalendarBlock, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("load user: %w", err)
	}
	conn, err := s.connections.GetByUserAndProvider(ctx, userID, "google")
	if err != nil {
		return nil, fmt.Errorf("load google connection: %w", err)
	}
	location, err := time.LoadLocation(user.Timezone)
	if err != nil {
		location = time.UTC
	}
	localDay := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, location)
	events, err := s.google.ListEvents(ctx, conn, localDay, location)
	if err != nil {
		return nil, fmt.Errorf("list google calendar events: %w", err)
	}
	localDate := localDay.Format(store.DateLayout)
	blocks := make([]store.CalendarBlock, 0, len(events))
	for _, event := range events {
		block, ok := classifyEvent(userID, user.Timezone, localDate, event, location)
		if !ok {
			continue
		}
		blocks = append(blocks, block)
	}
	if err := s.blocks.ReplaceDay(ctx, userID, "google", localDate, blocks); err != nil {
		return nil, fmt.Errorf("persist calendar blocks: %w", err)
	}
	return s.blocks.ListByUserAndDate(ctx, userID, localDate)
}

func classifyEvent(userID, timezone, localDate string, event Event, location *time.Location) (store.CalendarBlock, bool) {
	start := event.Start.In(location)
	end := event.End.In(location)
	if end.Before(start) || end.Equal(start) {
		return store.CalendarBlock{}, false
	}
	classification := classifyDuration(end.Sub(start), event.AllDay)
	window := eventWindow(start, end, event.AllDay, location)
	detail := fmt.Sprintf("%s calendar event", classification)
	return store.CalendarBlock{
		UserID:          userID,
		Provider:        "google",
		ExternalEventID: event.ID,
		LocalDate:       localDate,
		Timezone:        timezone,
		Title:           event.Summary,
		Detail:          detail,
		StartAt:         &start,
		EndAt:           &end,
		IsAllDay:        event.AllDay,
		Classification:  classification,
		Window:          window,
	}, true
}

func classifyDuration(duration time.Duration, allDay bool) string {
	if allDay || duration >= 6*time.Hour {
		return "large"
	}
	if duration >= 2*time.Hour {
		return "medium"
	}
	return "small"
}

func eventWindow(start, end time.Time, allDay bool, location *time.Location) string {
	if allDay {
		return "all-day"
	}
	windows := []struct {
		name       string
		start, end int
	}{{"morning", 8, 12}, {"afternoon", 12, 16}, {"evening", 16, 21}}
	dayStart := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, location)
	for _, window := range windows {
		ws := dayStart.Add(time.Duration(window.start) * time.Hour)
		we := dayStart.Add(time.Duration(window.end) * time.Hour)
		if start.Before(we) && end.After(ws) {
			return window.name
		}
	}
	return "none"
}
