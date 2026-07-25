package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rahat/rahat/internal/auth"
	"github.com/rahat/rahat/internal/db"
	usr "github.com/rahat/rahat/internal/users"
)

func newTestAuthHandler(t *testing.T) (*authHandler, *auth.Service, *usr.Service) {
	t.Helper()
	sqlDB, err := db.OpenSQLite(context.Background(), filepath.Join(t.TempDir(), "rahat.sqlite3"))
	if err != nil {
		t.Fatalf("OpenSQLite() error = %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	userService := usr.NewService(usr.NewRepository(sqlDB))
	authService := auth.NewService(sqlDB, auth.NewRepository(sqlDB), "test-web-session-secret", 30*24*time.Hour)
	return &authHandler{auth: authService, users: userService, webOrigin: "http://localhost:5200", appEnv: "development"}, authService, userService
}

func TestExchangeAccessLinkSetsCookieAndCurrentSession(t *testing.T) {
	h, authService, userService := newTestAuthHandler(t)
	user, err := userService.Create(context.Background(), usr.User{DisplayName: "Tester", Timezone: "UTC", DailyTimeBudgetMinutes: 30, Email: "tester@example.com"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	_, token, err := authService.IssueAccessGrant(context.Background(), user.ID, time.Hour)
	if err != nil {
		t.Fatalf("IssueAccessGrant() error = %v", err)
	}
	mux := http.NewServeMux()
	h.register(mux)

	body, _ := json.Marshal(authExchangeRequest{Token: token})
	req := httptest.NewRequest(http.MethodPost, "/auth/access-link/exchange", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "http://localhost:5200")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("exchange status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	cookie := rec.Result().Cookies()[0]
	if cookie.Name != sessionCookieName || !cookie.HttpOnly || cookie.SameSite != http.SameSiteLaxMode || cookie.Secure {
		t.Fatalf("unexpected cookie: %+v", cookie)
	}

	req = httptest.NewRequest(http.MethodGet, "/auth/session", http.NoBody)
	req.AddCookie(cookie)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("session status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var current sessionResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &current); err != nil {
		t.Fatalf("decode session response: %v", err)
	}
	if !current.Authenticated || current.User == nil || current.User.ID != user.ID {
		t.Fatalf("unexpected session response: %+v", current)
	}
}

func TestAccessLinkCannotBeUsedTwice(t *testing.T) {
	h, authService, userService := newTestAuthHandler(t)
	user, err := userService.Create(context.Background(), usr.User{DisplayName: "Tester", Timezone: "UTC", DailyTimeBudgetMinutes: 30})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	_, token, err := authService.IssueAccessGrant(context.Background(), user.ID, time.Hour)
	if err != nil {
		t.Fatalf("IssueAccessGrant() error = %v", err)
	}
	mux := http.NewServeMux()
	h.register(mux)
	body, _ := json.Marshal(authExchangeRequest{Token: token})
	for idx := 0; idx < 2; idx++ {
		req := httptest.NewRequest(http.MethodPost, "/auth/access-link/exchange", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Origin", "http://localhost:5200")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if idx == 0 && rec.Code != http.StatusOK {
			t.Fatalf("first exchange status = %d, want %d", rec.Code, http.StatusOK)
		}
		if idx == 1 && rec.Code != http.StatusUnauthorized {
			t.Fatalf("second exchange status = %d, want %d", rec.Code, http.StatusUnauthorized)
		}
	}
}

func TestCurrentSessionRejectsMissingOrRevokedSession(t *testing.T) {
	h, authService, userService := newTestAuthHandler(t)
	user, err := userService.Create(context.Background(), usr.User{DisplayName: "Tester", Timezone: "UTC", DailyTimeBudgetMinutes: 30})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	_, sessionToken, err := authService.CreateSessionForUser(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("CreateSessionForUser() error = %v", err)
	}
	mux := http.NewServeMux()
	h.register(mux)

	req := httptest.NewRequest(http.MethodGet, "/auth/session", http.NoBody)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("missing session status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}

	if err := authService.RevokeSession(context.Background(), sessionToken); err != nil {
		t.Fatalf("RevokeSession() error = %v", err)
	}
	req = httptest.NewRequest(http.MethodGet, "/auth/session", http.NoBody)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: sessionToken})
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("revoked session status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestLogoutRevokesSessionAndClearsCookie(t *testing.T) {
	h, authService, userService := newTestAuthHandler(t)
	user, err := userService.Create(context.Background(), usr.User{DisplayName: "Tester", Timezone: "UTC", DailyTimeBudgetMinutes: 30})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	_, sessionToken, err := authService.CreateSessionForUser(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("CreateSessionForUser() error = %v", err)
	}
	mux := http.NewServeMux()
	h.register(mux)

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Origin", "http://localhost:5200")
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: sessionToken})
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("logout status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if !strings.Contains(rec.Header().Get("Set-Cookie"), "Max-Age=0") && !strings.Contains(rec.Header().Get("Set-Cookie"), "Max-Age=-1") {
		t.Fatalf("expected clearing cookie, got %q", rec.Header().Get("Set-Cookie"))
	}
}

func TestProtectedRouteUsesAuthenticatedUserNotRawUserID(t *testing.T) {
	h, authService, userService := newTestAuthHandler(t)
	userOne, err := userService.Create(context.Background(), usr.User{DisplayName: "One", Timezone: "UTC", DailyTimeBudgetMinutes: 30})
	if err != nil {
		t.Fatalf("Create() user one error = %v", err)
	}
	userTwo, err := userService.Create(context.Background(), usr.User{DisplayName: "Two", Timezone: "UTC", DailyTimeBudgetMinutes: 30})
	if err != nil {
		t.Fatalf("Create() user two error = %v", err)
	}
	_, sessionToken, err := authService.CreateSessionForUser(context.Background(), userOne.ID)
	if err != nil {
		t.Fatalf("CreateSessionForUser() error = %v", err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /protected", requireAuthenticatedUserForRoute(h, func(w http.ResponseWriter, r *http.Request, current authenticatedUser) {
		writeJSON(w, http.StatusOK, map[string]string{"user_id": current.User.ID, "requested_user_id": r.URL.Query().Get("user_id")})
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected?user_id="+userTwo.ID, http.NoBody)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: sessionToken})
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("protected status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), userOne.ID) || strings.Contains(rec.Body.String(), `"user_id":"`+userTwo.ID+`"`) {
		t.Fatalf("unexpected protected body: %s", rec.Body.String())
	}
}

func TestExchangeAccessLinkAlwaysCreatesSessionForGrantUser(t *testing.T) {
	h, authService, userService := newTestAuthHandler(t)
	userOne, err := userService.Create(context.Background(), usr.User{DisplayName: "One", Timezone: "UTC", DailyTimeBudgetMinutes: 30})
	if err != nil {
		t.Fatalf("Create() user one error = %v", err)
	}
	userTwo, err := userService.Create(context.Background(), usr.User{DisplayName: "Two", Timezone: "UTC", DailyTimeBudgetMinutes: 30})
	if err != nil {
		t.Fatalf("Create() user two error = %v", err)
	}
	_, userTwoSession, err := authService.CreateSessionForUser(context.Background(), userTwo.ID)
	if err != nil {
		t.Fatalf("CreateSessionForUser() error = %v", err)
	}
	_, grantToken, err := authService.IssueAccessGrant(context.Background(), userOne.ID, time.Hour)
	if err != nil {
		t.Fatalf("IssueAccessGrant() error = %v", err)
	}

	mux := http.NewServeMux()
	h.register(mux)
	body, _ := json.Marshal(authExchangeRequest{Token: grantToken})
	req := httptest.NewRequest(http.MethodPost, "/auth/access-link/exchange", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "http://localhost:5200")
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: userTwoSession})
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("exchange status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var resp sessionResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode session response: %v", err)
	}
	if !resp.Authenticated || resp.User == nil || resp.User.ID != userOne.ID {
		t.Fatalf("expected session for grant user %q, got %+v", userOne.ID, resp)
	}
}

func TestRequireTrustedOriginAllowsDevOrigins(t *testing.T) {
	h := &authHandler{
		auth:       nil,
		users:      nil,
		webOrigin:  "http://localhost:5200",
		appEnv:     "development",
		devOrigins: []string{"http://127.0.0.1:5200", "http://192.168.1.50:5200"},
	}
	localReq := httptest.NewRequest(http.MethodPost, "/auth/access-link/exchange", http.NoBody)
	localReq.Header.Set("Origin", "http://localhost:5200")
	if err := h.requireTrustedOrigin(localReq); err != nil {
		t.Fatalf("expected localhost origin to be allowed, got %v", err)
	}
	ipReq := httptest.NewRequest(http.MethodPost, "/auth/access-link/exchange", http.NoBody)
	ipReq.Header.Set("Origin", "http://192.168.1.50:5200")
	if err := h.requireTrustedOrigin(ipReq); err != nil {
		t.Fatalf("expected dev IP origin to be allowed, got %v", err)
	}
	badReq := httptest.NewRequest(http.MethodPost, "/auth/access-link/exchange", http.NoBody)
	badReq.Header.Set("Origin", "http://evil.example.com")
	if err := h.requireTrustedOrigin(badReq); err == nil {
		t.Fatal("expected untrusted origin to be rejected")
	}
}

func TestProductionCookieUsesSecureFlag(t *testing.T) {
	h := &authHandler{appEnv: "production"}
	rec := httptest.NewRecorder()
	h.writeSessionCookie(rec, "token", time.Now().Add(time.Hour))
	if !strings.Contains(rec.Header().Get("Set-Cookie"), "Secure") {
		t.Fatalf("expected Secure cookie, got %q", rec.Header().Get("Set-Cookie"))
	}
}
