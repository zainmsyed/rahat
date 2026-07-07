package preferences

import "time"

type Channel string

const (
	ChannelTelegram Channel = "telegram"
	ChannelEmail    Channel = "email"
	ChannelSMS      Channel = "sms"
)

type Preference struct {
	ID                  string
	UserID              string
	Channel             Channel
	Enabled             bool
	IsPrimary           bool
	SupportsInteractive bool
	RecapEnabled        bool
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type Pause struct {
	ID        string
	UserID    string
	TaskID    string
	Scope     string
	Reason    string
	StartsAt  time.Time
	EndsAt    time.Time
	CreatedAt time.Time
}
