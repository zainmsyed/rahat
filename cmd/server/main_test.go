package main

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"testing"
	"time"

	ntg "github.com/rahat/rahat/internal/notifications/telegram"
)

type fakeRuntimeClient struct {
	deleteWebhookCalls int
	setWebhookCalls    int
	setWebhookURL      string
	updatesCalled      chan struct{}
}

func (f *fakeRuntimeClient) SendMessage(context.Context, ntg.SendMessageRequest) error { return nil }
func (f *fakeRuntimeClient) SetWebhook(context.Context, string, string) error {
	f.setWebhookCalls++
	return nil
}
func (f *fakeRuntimeClient) DeleteWebhook(context.Context) error {
	f.deleteWebhookCalls++
	return nil
}
func (f *fakeRuntimeClient) GetUpdates(ctx context.Context, req ntg.GetUpdatesRequest) ([]ntg.Update, error) {
	if f.updatesCalled != nil {
		select {
		case <-f.updatesCalled:
		default:
			close(f.updatesCalled)
		}
	}
	<-ctx.Done()
	return nil, ctx.Err()
}

type fakeCallbackHandler struct{}

func (fakeCallbackHandler) HandleCallback(context.Context, string) error { return nil }

func TestConfigureTelegramTransportFallsBackToLongPollingWithoutWebhook(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := &fakeRuntimeClient{updatesCalled: make(chan struct{})}
	mux := http.NewServeMux()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	transport := configureTelegramTransport(ctx, logger, mux, client, "", "", fakeCallbackHandler{})
	if transport != "long_polling" {
		t.Fatalf("transport = %s, want long_polling", transport)
	}
	if client.deleteWebhookCalls != 1 {
		t.Fatalf("deleteWebhook calls = %d, want 1", client.deleteWebhookCalls)
	}

	select {
	case <-client.updatesCalled:
	case <-time.After(2 * time.Second):
		t.Fatal("expected long polling to start getUpdates")
	}
}

func TestShouldUseTelegramWebhook(t *testing.T) {
	tests := []struct {
		name   string
		url    string
		secret string
		want   bool
	}{
		{name: "domain backed https", url: "https://rahat.example.com/webhooks/telegram", secret: "secret", want: true},
		{name: "missing secret", url: "https://rahat.example.com/webhooks/telegram", want: false},
		{name: "localhost", url: "https://localhost/webhooks/telegram", secret: "secret", want: false},
		{name: "ip host", url: "https://127.0.0.1/webhooks/telegram", secret: "secret", want: false},
		{name: "http only", url: "http://rahat.example.com/webhooks/telegram", secret: "secret", want: false},
		{name: "wrong path", url: "https://rahat.example.com/telegram-hook", secret: "secret", want: false},
		{name: "empty path", url: "https://rahat.example.com", secret: "secret", want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := shouldUseTelegramWebhook(tc.url, tc.secret); got != tc.want {
				t.Fatalf("shouldUseTelegramWebhook(%q, %q) = %v, want %v", tc.url, tc.secret, got, tc.want)
			}
		})
	}
}
