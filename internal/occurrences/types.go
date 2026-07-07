package occurrences

import (
	"time"

	"github.com/rahat/rahat/internal/tasks"
)

type Status string

const (
	StatusPending   Status = "pending"
	StatusScheduled Status = "scheduled"
	StatusCompleted Status = "completed"
	StatusSkipped   Status = "skipped"
)

type Occurrence struct {
	ID                       string
	UserID                   string
	TaskID                   string
	SubtaskID                string
	Status                   Status
	ScheduledForDate         string
	OriginalScheduledForDate string
	ScheduledTimeOfDay       tasks.TimeOfDayPreference
	RolloverCount            int
	ConsecutiveNoCount       int
	SnoozedUntilAt           *time.Time
	CompletedAt              *time.Time
	SkippedAt                *time.Time
	CreatedAt                time.Time
	UpdatedAt                time.Time
}
