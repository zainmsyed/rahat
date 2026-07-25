package auth

import "time"

type AccessGrant struct {
	ID        string
	UserID    string
	Selector  string
	TokenHash string
	ExpiresAt time.Time
	UsedAt    *time.Time
	CreatedAt time.Time
}

type WebSession struct {
	ID         string
	UserID     string
	Selector   string
	TokenHash  string
	ExpiresAt  time.Time
	RevokedAt  *time.Time
	CreatedAt  time.Time
	LastSeenAt time.Time
}

type SessionUser struct {
	ID                     string `json:"id"`
	DisplayName            string `json:"display_name"`
	Timezone               string `json:"timezone"`
	DailyTimeBudgetMinutes int    `json:"daily_time_budget_minutes"`
	Email                  string `json:"email"`
}
