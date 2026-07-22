package telegram

import (
	"encoding/json"
	"net/http"

	ntg "github.com/rahat/rahat/internal/notifications/telegram"
)

type Handler struct {
	secret string
	svc    ntg.CallbackHandler
}

func NewHandler(secret string, svc ntg.CallbackHandler) *Handler {
	return &Handler{secret: secret, svc: svc}
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
	if update.CallbackQuery != nil && update.CallbackQuery.Data != "" {
		if err := h.svc.HandleCallback(r.Context(), update.CallbackQuery.Data); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}
