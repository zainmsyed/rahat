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
	Message       *Message       `json:"message,omitempty"`
	CallbackQuery *CallbackQuery `json:"callback_query,omitempty"`
}

type Message struct {
	Chat *Chat  `json:"chat,omitempty"`
	Text string `json:"text"`
}

type Chat struct {
	ID   int64  `json:"id"`
	Type string `json:"type"`
}

type User struct {
	ID        int64  `json:"id"`
	IsBot     bool   `json:"is_bot"`
	FirstName string `json:"first_name"`
	Username  string `json:"username"`
}

type CallbackQuery struct {
	ID   string `json:"id"`
	Data string `json:"data"`
}

type GetUpdatesRequest struct {
	Offset  int64 `json:"offset,omitempty"`
	Timeout int   `json:"timeout,omitempty"`
}

type BotInfo struct {
	ID        int64  `json:"id"`
	IsBot     bool   `json:"is_bot"`
	FirstName string `json:"first_name"`
	Username  string `json:"username"`
}

type BotClient interface {
	SendMessage(context.Context, SendMessageRequest) error
}

type RuntimeClient interface {
	BotClient
	SetWebhook(context.Context, string, string) error
	DeleteWebhook(context.Context) error
	GetUpdates(context.Context, GetUpdatesRequest) ([]Update, error)
	GetMe(context.Context) (BotInfo, error)
}

type CallbackHandler interface {
	HandleCallback(context.Context, string) error
}

type MessageHandler interface {
	HandleMessage(context.Context, *Message) error
}
