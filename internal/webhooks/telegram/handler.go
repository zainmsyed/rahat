package telegram

import (
	"encoding/json"
	"net/http"

	ntg "github.com/rahat/rahat/internal/notifications/telegram"
)

type Handler struct {
	secret          string
	callbackHandler ntg.CallbackHandler
	messageHandler  ntg.MessageHandler
}

func NewHandler(secret string, callbackHandler ntg.CallbackHandler, messageHandler ntg.MessageHandler) *Handler {
	return &Handler{secret: secret, callbackHandler: callbackHandler, messageHandler: messageHandler}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.secret != "" && r.Header.Get("X-Telegram-Bot-Api-Secret-Token") != h.secret {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	defer r.Body.Close()
	var update ntg.Update
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if update.CallbackQuery != nil && update.CallbackQuery.Data != "" && h.callbackHandler != nil {
		if err := h.callbackHandler.HandleCallback(r.Context(), update.CallbackQuery.Data); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	if update.Message != nil && update.Message.Chat != nil && h.messageHandler != nil {
		if err := h.messageHandler.HandleMessage(r.Context(), update.Message); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}
