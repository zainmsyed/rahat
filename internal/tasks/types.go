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

type SubtaskDependencyType string

const (
	SubtaskDependencyRequiredSameDay SubtaskDependencyType = "required_same_day"
	SubtaskDependencySoftFollowup    SubtaskDependencyType = "soft_followup"
)

type SubtaskGapRule struct {
	MinGapAfterPreviousMinutes int
}

type DayPreference string

const (
	DayPreferenceAny     DayPreference = "any"
	DayPreferenceWeekday DayPreference = "weekday"
	DayPreferenceWeekend DayPreference = "weekend"
)

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
	DayPreference       DayPreference
	IsMultistep         bool
	IsPaused            bool
	ArchivedAt          *time.Time
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
	DependencyType      SubtaskDependencyType
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
	DayPreference       DayPreference
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
	DependencyType             SubtaskDependencyType
	MinGapAfterPreviousMinutes int
}
