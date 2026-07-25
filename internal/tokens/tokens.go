package tokens

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrUnavailable = errors.New("lookahead tokens are not configured")
	ErrInvalid     = errors.New("lookahead token is invalid")
	ErrExpired     = errors.New("lookahead token is expired")
)

type Manager struct {
	secret []byte
	now    func() time.Time
}

type Claims struct {
	UserID    string `json:"user_id"`
	ExpiresAt int64  `json:"expires_at"`
	Nonce     string `json:"nonce"`
}

func NewManager(secret string) *Manager {
	return &Manager{secret: []byte(strings.TrimSpace(secret)), now: func() time.Time { return time.Now().UTC() }}
}

func (m *Manager) Available() bool {
	return m != nil && len(m.secret) >= 16
}

func (m *Manager) Issue(userID string, ttl time.Duration) (string, error) {
	if !m.Available() {
		return "", ErrUnavailable
	}
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return "", fmt.Errorf("user id is required")
	}
	if ttl <= 0 {
		ttl = 30 * 24 * time.Hour
	}
	nonce, err := randomNonce()
	if err != nil {
		return "", err
	}
	claims := Claims{UserID: userID, ExpiresAt: m.now().Add(ttl).Unix(), Nonce: nonce}
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("marshal token claims: %w", err)
	}
	payloadPart := base64.RawURLEncoding.EncodeToString(payload)
	signature := m.sign(payloadPart)
	return payloadPart + "." + signature, nil
}

func (m *Manager) Verify(token string) (Claims, error) {
	if !m.Available() {
		return Claims{}, ErrUnavailable
	}
	parts := strings.Split(strings.TrimSpace(token), ".")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return Claims{}, ErrInvalid
	}
	want := m.sign(parts[0])
	if !hmac.Equal([]byte(want), []byte(parts[1])) {
		return Claims{}, ErrInvalid
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return Claims{}, ErrInvalid
	}
	var claims Claims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return Claims{}, ErrInvalid
	}
	if strings.TrimSpace(claims.UserID) == "" {
		return Claims{}, ErrInvalid
	}
	if claims.ExpiresAt <= m.now().Unix() {
		return Claims{}, ErrExpired
	}
	return claims, nil
}

func (m *Manager) sign(payloadPart string) string {
	mac := hmac.New(sha256.New, m.secret)
	_, _ = mac.Write([]byte(payloadPart))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func randomNonce() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate token nonce: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
