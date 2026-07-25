package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	ntg "github.com/rahat/rahat/internal/notifications/telegram"
)

type fakeCallbackHandler struct {
	calls []string
}

func (f *fakeCallbackHandler) HandleCallback(_ context.Context, data string) error {
	f.calls = append(f.calls, data)
	return nil
}

type fakeMessageHandler struct {
	messages []*ntg.Message
}

func (f *fakeMessageHandler) HandleMessage(_ context.Context, msg *ntg.Message) error {
	f.messages = append(f.messages, msg)
	return nil
}

func TestWebhookRejectsInvalidSecret(t *testing.T) {
	callback := &fakeCallbackHandler{}
	message := &fakeMessageHandler{}
	handler := NewHandler("correct-secret", callback, message)

	req := httptest.NewRequest(http.MethodPost, "/webhooks/telegram", bytes.NewReader([]byte("{}")))
	req.Header.Set("X-Telegram-Bot-Api-Secret-Token", "wrong-secret")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestWebhookRoutesCallbackQuery(t *testing.T) {
	callback := &fakeCallbackHandler{}
	message := &fakeMessageHandler{}
	handler := NewHandler("", callback, message)

	update := ntg.Update{UpdateID: 1, CallbackQuery: &ntg.CallbackQuery{ID: "q1", Data: "d:occ-1"}}
	body, _ := json.Marshal(update)
	req := httptest.NewRequest(http.MethodPost, "/webhooks/telegram", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if len(callback.calls) != 1 || callback.calls[0] != "d:occ-1" {
		t.Fatalf("callback calls = %#v, want [d:occ-1]", callback.calls)
	}
	if len(message.messages) != 0 {
		t.Fatalf("unexpected message calls: %+v", message.messages)
	}
}

func TestWebhookRoutesMessage(t *testing.T) {
	callback := &fakeCallbackHandler{}
	message := &fakeMessageHandler{}
	handler := NewHandler("", callback, message)

	update := ntg.Update{UpdateID: 2, Message: &ntg.Message{Chat: &ntg.Chat{ID: 123, Type: "private"}, Text: "ABC123"}}
	body, _ := json.Marshal(update)
	req := httptest.NewRequest(http.MethodPost, "/webhooks/telegram", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if len(message.messages) != 1 {
		t.Fatalf("message calls = %d, want 1", len(message.messages))
	}
	if message.messages[0].Text != "ABC123" {
		t.Fatalf("message text = %q, want ABC123", message.messages[0].Text)
	}
	if len(callback.calls) != 0 {
		t.Fatalf("unexpected callback calls: %+v", callback.calls)
	}
}

func TestWebhookRoutesEditMessage(t *testing.T) {
	callback := &fakeCallbackHandler{}
	message := &fakeMessageHandler{}
	handler := NewHandler("", callback, message)

	update := ntg.Update{UpdateID: 4, Message: &ntg.Message{Chat: &ntg.Chat{ID: 123, Type: "private"}, Text: "/edit"}}
	body, _ := json.Marshal(update)
	req := httptest.NewRequest(http.MethodPost, "/webhooks/telegram", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if len(message.messages) != 1 {
		t.Fatalf("message calls = %d, want 1", len(message.messages))
	}
	if message.messages[0].Text != "/edit" {
		t.Fatalf("message text = %q, want /edit", message.messages[0].Text)
	}
	if len(callback.calls) != 0 {
		t.Fatalf("unexpected callback calls: %+v", callback.calls)
	}
}

func TestWebhookReturnsBadRequestWhenCallbackHandlerErrors(t *testing.T) {
	message := &fakeMessageHandler{}
	callbackErr := errors.New("callback failed")
	handler := NewHandler("", &errorCallbackHandler{err: callbackErr}, message)

	update := ntg.Update{UpdateID: 3, CallbackQuery: &ntg.CallbackQuery{ID: "q2", Data: "d:occ-2"}}
	body, _ := json.Marshal(update)
	req := httptest.NewRequest(http.MethodPost, "/webhooks/telegram", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

type errorCallbackHandler struct {
	err error
}

func (e *errorCallbackHandler) HandleCallback(_ context.Context, _ string) error {
	return e.err
}
