package users

import "time"

type User struct {
	ID                     string
	DisplayName            string
	Timezone               string
	DailyTimeBudgetMinutes int
	TelegramChatID         string
	Email                  string
	CreatedAt              time.Time
	UpdatedAt              time.Time
}
