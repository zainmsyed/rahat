package scheduler

import (
	"time"

	"github.com/rahat/rahat/internal/occurrences"
	"github.com/rahat/rahat/internal/store"
)

type PlanResult struct {
	Date                 string
	Scheduled            []occurrences.Occurrence
	Overflowed           []occurrences.Occurrence
	Skipped              []occurrences.Occurrence
	Checkpoint           store.ScheduleCheckpoint
	WindowBudgetsMinutes map[string]int
	BlockedWindows       map[string][]string
	SmallTaskOnlyReason  string
	Reasons              map[string]string
}

type Clock interface {
	Now() time.Time
}

type realClock struct{}

func (realClock) Now() time.Time {
	return time.Now().UTC()
}
