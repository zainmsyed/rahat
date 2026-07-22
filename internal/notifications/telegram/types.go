package telegram

import "context"

type SendMessageRequest struct {
	ChatID                string       `json:"chat_id"`
	Text                  string       `json:"text"`
	ParseMode             string       `json:"parse_mode,omitempty"`
	ReplyMarkup           *ReplyMarkup `json:"reply_markup,omitempty"`
	DisableWebPagePreview bool         `json:"disable_web_page_preview,omitempty"`
}

type ReplyMarkup struct {
	InlineKeyboard [][]InlineButton `json:"inline_keyboard"`
}

type InlineButton struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data,omitempty"`
}

type Update struct {
	UpdateID      int64          `json:"update_id"`
	CallbackQuery *CallbackQuery `json:"callback_query,omitempty"`
}

type CallbackQuery struct {
	ID   string `json:"id"`
	Data string `json:"data"`
}

type GetUpdatesRequest struct {
	Offset  int64 `json:"offset,omitempty"`
	Timeout int   `json:"timeout,omitempty"`
}

type BotClient interface {
	SendMessage(context.Context, SendMessageRequest) error
}

type RuntimeClient interface {
	BotClient
	SetWebhook(context.Context, string, string) error
	DeleteWebhook(context.Context) error
	GetUpdates(context.Context, GetUpdatesRequest) ([]Update, error)
}

type CallbackHandler interface {
	HandleCallback(context.Context, string) error
}
