package tasks

import "time"

type CadenceType string

const (
	CadenceTypeInterval CadenceType = "interval"
	CadenceTypeCount    CadenceType = "count"
)

type Priority string

const (
	PriorityHigh   Priority = "high"
	PriorityMedium Priority = "medium"
	PriorityLow    Priority = "low"
)

type TimeOfDayPreference string

const (
	TimeOfDayAny       TimeOfDayPreference = "any"
	TimeOfDayMorning   TimeOfDayPreference = "morning"
	TimeOfDayAfternoon TimeOfDayPreference = "afternoon"
	TimeOfDayEvening   TimeOfDayPreference = "evening"
)

type SubtaskGapRule struct {
	MinGapAfterPreviousMinutes int
}

type Task struct {
	ID                  string
	UserID              string
	Name                string
	Description         string
	DurationMinutes     int
	CadenceType         CadenceType
	CadenceValue        int
	Priority            Priority
	TimeOfDayPreference TimeOfDayPreference
	IsMultistep         bool
	IsPaused            bool
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type Subtask struct {
	ID                  string
	TaskID              string
	StepOrder           int
	Name                string
	DurationMinutes     int
	TimeOfDayPreference TimeOfDayPreference
	GapRule             SubtaskGapRule
	CreatedAt           time.Time
}

type TaskWithSubtasks struct {
	Task     Task
	Subtasks []Subtask
}

type StarterTaskTemplate struct {
	ID                  string
	Slug                string
	Name                string
	Description         string
	DurationMinutes     int
	CadenceType         CadenceType
	CadenceValue        int
	Priority            Priority
	TimeOfDayPreference TimeOfDayPreference
	IsMultistep         bool
	SortOrder           int
	Subtasks            []StarterSubtaskTemplate
}

type StarterSubtaskTemplate struct {
	ID                         string
	StarterTaskTemplateID      string
	StepOrder                  int
	Name                       string
	DurationMinutes            int
	TimeOfDayPreference        TimeOfDayPreference
	MinGapAfterPreviousMinutes int
}
