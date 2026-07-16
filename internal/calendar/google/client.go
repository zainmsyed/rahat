package google

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	cal "github.com/rahat/rahat/internal/calendar"
	"github.com/rahat/rahat/internal/store"
	"golang.org/x/oauth2"
	googleoauth "golang.org/x/oauth2/google"
)

const calendarReadOnlyScope = "https://www.googleapis.com/auth/calendar.readonly"

type Client struct {
	config     *oauth2.Config
	calendarID string
	httpClient *http.Client
}

func NewClient(clientID, clientSecret, redirectURL, calendarID string) *Client {
	if calendarID == "" {
		calendarID = "primary"
	}
	return &Client{
		config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes:       []string{calendarReadOnlyScope},
			Endpoint:     googleoauth.Endpoint,
		},
		calendarID: calendarID,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) AuthCodeURL(state string) string {
	return c.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func (c *Client) ExchangeCode(ctx context.Context, code string) (cal.OAuthToken, error) {
	token, err := c.config.Exchange(ctx, code)
	if err != nil {
		return cal.OAuthToken{}, err
	}
	return oauthTokenFromGoogle(token), nil
}

func (c *Client) ListEvents(ctx context.Context, conn store.CalendarConnection, day time.Time, location *time.Location) ([]cal.Event, error) {
	startOfDay := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, location)
	endOfDay := startOfDay.Add(24 * time.Hour)
	token := &oauth2.Token{AccessToken: conn.AccessToken, RefreshToken: conn.RefreshToken, TokenType: conn.TokenType, Expiry: derefTime(conn.Expiry)}
	httpClient := oauth2.NewClient(ctx, c.config.TokenSource(ctx, token))
	httpClient.Timeout = c.httpClient.Timeout

	base := fmt.Sprintf("https://www.googleapis.com/calendar/v3/calendars/%s/events", url.PathEscape(conn.CalendarID))
	query := url.Values{}
	query.Set("singleEvents", "true")
	query.Set("orderBy", "startTime")
	query.Set("timeMin", startOfDay.Format(time.RFC3339))
	query.Set("timeMax", endOfDay.Format(time.RFC3339))
	query.Set("timeZone", location.String())
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"?"+query.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("build google calendar request: %w", err)
	}
	resp, err := httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("execute google calendar request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("google calendar events returned status %d", resp.StatusCode)
	}
	var payload eventsResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode google calendar response: %w", err)
	}
	result := make([]cal.Event, 0, len(payload.Items))
	for _, item := range payload.Items {
		event, ok, err := item.toEvent(location)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		result = append(result, event)
	}
	return result, nil
}

type eventsResponse struct {
	Items []eventItem `json:"items"`
}

type eventItem struct {
	ID      string     `json:"id"`
	Summary string     `json:"summary"`
	Start   eventPoint `json:"start"`
	End     eventPoint `json:"end"`
}

type eventPoint struct {
	Date     string `json:"date"`
	DateTime string `json:"dateTime"`
	TimeZone string `json:"timeZone"`
}

func (e eventItem) toEvent(location *time.Location) (cal.Event, bool, error) {
	if e.ID == "" {
		return cal.Event{}, false, nil
	}
	start, allDay, err := e.Start.parse(location)
	if err != nil {
		return cal.Event{}, false, fmt.Errorf("parse google event start: %w", err)
	}
	end, _, err := e.End.parse(location)
	if err != nil {
		return cal.Event{}, false, fmt.Errorf("parse google event end: %w", err)
	}
	return cal.Event{ID: e.ID, Summary: strings.TrimSpace(e.Summary), Start: start, End: end, AllDay: allDay, TimeZone: location.String()}, true, nil
}

func (p eventPoint) parse(location *time.Location) (time.Time, bool, error) {
	if p.Date != "" {
		parsed, err := time.ParseInLocation("2006-01-02", p.Date, location)
		return parsed, true, err
	}
	if p.DateTime != "" {
		parsed, err := time.Parse(time.RFC3339, p.DateTime)
		return parsed.In(location), false, err
	}
	return time.Time{}, false, fmt.Errorf("missing date and dateTime")
}

func oauthTokenFromGoogle(token *oauth2.Token) cal.OAuthToken {
	var expiry *time.Time
	if !token.Expiry.IsZero() {
		value := token.Expiry.UTC()
		expiry = &value
	}
	scope := ""
	if raw, ok := token.Extra("scope").(string); ok {
		scope = raw
	}
	return cal.OAuthToken{AccessToken: token.AccessToken, RefreshToken: token.RefreshToken, TokenType: token.TokenType, Expiry: expiry, Scope: scope}
}

func derefTime(value *time.Time) time.Time {
	if value == nil {
		return time.Time{}
	}
	return value.UTC()
}
