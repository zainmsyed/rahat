package telegram

import (
	"context"
	"log/slog"
	"testing"
	"time"
)

type fakeRuntimeClient struct {
	updatesCalls int
	updates      []Update
	blockUntil   <-chan struct{}
}

func (f *fakeRuntimeClient) SendMessage(context.Context, SendMessageRequest) error { return nil }
func (f *fakeRuntimeClient) SetWebhook(context.Context, string, string) error      { return nil }
func (f *fakeRuntimeClient) DeleteWebhook(context.Context) error                   { return nil }
func (f *fakeRuntimeClient) GetUpdates(ctx context.Context, req GetUpdatesRequest) ([]Update, error) {
	f.updatesCalls++
	if len(f.updates) > 0 {
		updates := f.updates
		f.updates = nil
		return updates, nil
	}
	if f.blockUntil == nil {
		<-ctx.Done()
		return nil, ctx.Err()
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-f.blockUntil:
		return nil, nil
	}
}

type fakeCallbackHandler struct {
	calls []string
	done  chan struct{}
}

func (f *fakeCallbackHandler) HandleCallback(_ context.Context, data string) error {
	f.calls = append(f.calls, data)
	if f.done != nil {
		close(f.done)
		f.done = nil
	}
	return nil
}

func TestPollerProcessesCallbackQuery(t *testing.T) {
	client := &fakeRuntimeClient{updates: []Update{{UpdateID: 42, CallbackQuery: &CallbackQuery{Data: "d:occ-1"}}}}
	handler := &fakeCallbackHandler{done: make(chan struct{})}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	poller := NewPoller(client, handler, slog.Default())
	go poller.Run(ctx)

	select {
	case <-handler.done:
	case <-time.After(2 * time.Second):
		t.Fatal("poller did not deliver callback update")
	}
	cancel()

	if len(handler.calls) != 1 || handler.calls[0] != "d:occ-1" {
		t.Fatalf("callback calls = %#v, want [\"d:occ-1\"]", handler.calls)
	}
	if client.updatesCalls == 0 {
		t.Fatal("expected getUpdates to be called")
	}
}
