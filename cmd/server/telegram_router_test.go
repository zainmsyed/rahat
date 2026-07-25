package main

import (
	"context"
	"errors"
	"testing"

	ntg "github.com/rahat/rahat/internal/notifications/telegram"
)

type recordingMessageHandler struct {
	messages []*ntg.Message
}

func (r *recordingMessageHandler) HandleMessage(_ context.Context, msg *ntg.Message) error {
	r.messages = append(r.messages, msg)
	return nil
}

type errorMessageHandler struct {
	err error
}

func (e *errorMessageHandler) HandleMessage(_ context.Context, _ *ntg.Message) error {
	return e.err
}

func TestRouterRoutesEditToEditHandler(t *testing.T) {
	onboarding := &recordingMessageHandler{}
	edit := &recordingMessageHandler{}
	router := &telegramMessageRouter{onboarding: onboarding, edit: edit}

	msg := &ntg.Message{Text: "/edit", Chat: &ntg.Chat{ID: 1, Type: "private"}}
	if err := router.HandleMessage(context.Background(), msg); err != nil {
		t.Fatalf("HandleMessage error = %v", err)
	}
	if len(edit.messages) != 1 {
		t.Fatalf("expected edit handler to receive message, got %d", len(edit.messages))
	}
	if len(onboarding.messages) != 0 {
		t.Fatalf("expected onboarding handler to be skipped, got %d", len(onboarding.messages))
	}
}

func TestRouterRoutesStartToOnboardingHandler(t *testing.T) {
	onboarding := &recordingMessageHandler{}
	edit := &recordingMessageHandler{}
	router := &telegramMessageRouter{onboarding: onboarding, edit: edit}

	msg := &ntg.Message{Text: "/start ABC123", Chat: &ntg.Chat{ID: 1, Type: "private"}}
	if err := router.HandleMessage(context.Background(), msg); err != nil {
		t.Fatalf("HandleMessage error = %v", err)
	}
	if len(onboarding.messages) != 1 {
		t.Fatalf("expected onboarding handler to receive message, got %d", len(onboarding.messages))
	}
	if len(edit.messages) != 0 {
		t.Fatalf("expected edit handler to be skipped, got %d", len(edit.messages))
	}
}

func TestRouterPropagatesEditHandlerErrors(t *testing.T) {
	wantErr := errors.New("edit failed")
	router := &telegramMessageRouter{edit: &errorMessageHandler{err: wantErr}}

	msg := &ntg.Message{Text: "/edit", Chat: &ntg.Chat{ID: 1, Type: "private"}}
	if err := router.HandleMessage(context.Background(), msg); !errors.Is(err, wantErr) {
		t.Fatalf("expected error %v, got %v", wantErr, err)
	}
}

func TestRouterRecognizesEditWithBotUsername(t *testing.T) {
	edit := &recordingMessageHandler{}
	router := &telegramMessageRouter{edit: edit}

	msg := &ntg.Message{Text: "/edit@RahatBot", Chat: &ntg.Chat{ID: 1, Type: "private"}}
	if err := router.HandleMessage(context.Background(), msg); err != nil {
		t.Fatalf("HandleMessage error = %v", err)
	}
	if len(edit.messages) != 1 {
		t.Fatalf("expected edit handler to receive message, got %d", len(edit.messages))
	}
}
