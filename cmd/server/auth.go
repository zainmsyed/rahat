package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/rahat/rahat/internal/auth"
	usr "github.com/rahat/rahat/internal/users"
)

const sessionCookieName = "rahat_session"

var errOriginNotAllowed = errors.New("request origin not allowed")

type authHandler struct {
	auth       *auth.Service
	users      *usr.Service
	webOrigin  string
	appEnv     string
	devOrigins []string
}

type authExchangeRequest struct {
	Token string `json:"token"`
}

type authIssueAccessLinkResult struct {
	UserID      string `json:"user_id"`
	DisplayName string `json:"display_name"`
	Path        string `json:"path"`
	URL         string `json:"url"`
	ExpiresAt   string `json:"expires_at"`
}

type sessionResponse struct {
	Authenticated bool              `json:"authenticated"`
	User          *auth.SessionUser `json:"user,omitempty"`
}

type authenticatedUser struct {
	Session auth.WebSession
	User    usr.User
}

func (h *authHandler) register(mux *http.ServeMux) {
	mux.HandleFunc("POST /auth/access-link/exchange", h.handleExchangeAccessLink)
	mux.HandleFunc("GET /auth/session", h.handleCurrentSession)
	mux.HandleFunc("POST /auth/logout", h.handleLogout)
}

func (h *authHandler) handleExchangeAccessLink(w http.ResponseWriter, r *http.Request) {
	if err := h.requireTrustedOrigin(r); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	var req authExchangeRequest
	if err := decodeJSON(r, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	session, rawSessionToken, err := h.auth.ExchangeAccessGrant(r.Context(), req.Token)
	if err != nil {
		writeAuthError(w, err)
		return
	}
	user, err := h.users.GetByID(r.Context(), session.UserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	h.writeSessionCookie(w, rawSessionToken, session.ExpiresAt)
	writeJSON(w, http.StatusOK, sessionResponse{Authenticated: true, User: sessionUserResponse(user)})
}

func (h *authHandler) handleCurrentSession(w http.ResponseWriter, r *http.Request) {
	current, err := h.resolveAuthenticatedUser(r.Context(), r)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidToken) || errors.Is(err, auth.ErrExpired) || errors.Is(err, auth.ErrRevoked) {
			clearSessionCookie(w, h.cookieSecure())
			writeJSON(w, http.StatusUnauthorized, sessionResponse{Authenticated: false})
			return
		}
		if errors.Is(err, http.ErrNoCookie) {
			writeJSON(w, http.StatusUnauthorized, sessionResponse{Authenticated: false})
			return
		}
		writeAuthError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, sessionResponse{Authenticated: true, User: sessionUserResponse(current.User)})
}

func (h *authHandler) handleLogout(w http.ResponseWriter, r *http.Request) {
	if err := h.requireTrustedOrigin(r); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	if cookie, err := r.Cookie(sessionCookieName); err == nil {
		_ = h.auth.RevokeSession(r.Context(), cookie.Value)
	}
	clearSessionCookie(w, h.cookieSecure())
	w.WriteHeader(http.StatusNoContent)
}

func (h *authHandler) resolveAuthenticatedUser(ctx context.Context, r *http.Request) (authenticatedUser, error) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		return authenticatedUser{}, err
	}
	session, err := h.auth.VerifySession(ctx, cookie.Value)
	if err != nil {
		return authenticatedUser{}, err
	}
	user, err := h.users.GetByID(ctx, session.UserID)
	if err != nil {
		return authenticatedUser{}, err
	}
	return authenticatedUser{Session: session, User: user}, nil
}

func (h *authHandler) requireAuthenticatedUser(w http.ResponseWriter, r *http.Request) (authenticatedUser, bool) {
	current, err := h.resolveAuthenticatedUser(r.Context(), r)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			http.Error(w, "missing session", http.StatusUnauthorized)
			return authenticatedUser{}, false
		}
		if errors.Is(err, auth.ErrExpired) || errors.Is(err, auth.ErrInvalidToken) || errors.Is(err, auth.ErrRevoked) {
			clearSessionCookie(w, h.cookieSecure())
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return authenticatedUser{}, false
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return authenticatedUser{}, false
	}
	return current, true
}

func (h *authHandler) issueAccessLink(ctx context.Context, userLookup string, ttl time.Duration) (authIssueAccessLinkResult, error) {
	var user usr.User
	var err error
	if strings.Contains(userLookup, "@") {
		user, err = h.users.GetByEmail(ctx, strings.TrimSpace(userLookup))
	} else {
		user, err = h.users.GetByID(ctx, strings.TrimSpace(userLookup))
	}
	if err != nil {
		return authIssueAccessLinkResult{}, err
	}
	grant, rawToken, err := h.auth.IssueAccessGrant(ctx, user.ID, ttl)
	if err != nil {
		return authIssueAccessLinkResult{}, err
	}
	path := "/login?token=" + rawToken
	base := strings.TrimRight(h.webOrigin, "/")
	return authIssueAccessLinkResult{UserID: user.ID, DisplayName: user.DisplayName, Path: path, URL: base + path, ExpiresAt: grant.ExpiresAt.Format(time.RFC3339)}, nil
}

func (h *authHandler) requireTrustedOrigin(r *http.Request) error {
	if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
		return nil
	}
	origin := strings.TrimSpace(r.Header.Get("Origin"))
	if origin == "" {
		return errOriginNotAllowed
	}
	if origin == strings.TrimRight(h.webOrigin, "/") {
		return nil
	}
	if h.appEnv == "development" {
		for _, allowed := range h.devOrigins {
			if origin == strings.TrimRight(allowed, "/") {
				return nil
			}
		}
	}
	return errOriginNotAllowed
}

func (h *authHandler) writeSessionCookie(w http.ResponseWriter, rawToken string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    rawToken,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   h.cookieSecure(),
		Expires:  expiresAt,
		MaxAge:   max(1, int(time.Until(expiresAt).Seconds())),
	})
}

func (h *authHandler) cookieSecure() bool {
	return strings.EqualFold(h.appEnv, "production")
}

func clearSessionCookie(w http.ResponseWriter, secure bool) {
	http.SetCookie(w, &http.Cookie{Name: sessionCookieName, Value: "", Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode, Secure: secure, MaxAge: -1, Expires: time.Unix(0, 0)})
}

func sessionUserResponse(user usr.User) *auth.SessionUser {
	return &auth.SessionUser{ID: user.ID, DisplayName: user.DisplayName, Timezone: user.Timezone, DailyTimeBudgetMinutes: user.DailyTimeBudgetMinutes, Email: user.Email}
}

func writeAuthError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, auth.ErrUnavailable):
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
	case errors.Is(err, auth.ErrInvalidToken), errors.Is(err, auth.ErrExpired), errors.Is(err, auth.ErrUsed), errors.Is(err, auth.ErrRevoked):
		http.Error(w, err.Error(), http.StatusUnauthorized)
	default:
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func requireAuthenticatedUserForRoute(handler *authHandler, next func(http.ResponseWriter, *http.Request, authenticatedUser)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		current, ok := handler.requireAuthenticatedUser(w, r)
		if !ok {
			return
		}
		next(w, r, current)
	}
}

func writeIssueLinkJSON(result authIssueAccessLinkResult) ([]byte, error) {
	return json.MarshalIndent(result, "", "  ")
}
