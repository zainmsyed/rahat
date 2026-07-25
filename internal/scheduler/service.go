package scheduler

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/rahat/rahat/internal/occurrences"
	"github.com/rahat/rahat/internal/store"
	"github.com/rahat/rahat/internal/tasks"
	daytime "github.com/rahat/rahat/internal/time"
	"github.com/rahat/rahat/internal/users"
)

const rolloverCap = 3

type Service struct {
	users       *users.Service
	tasks       *tasks.Service
	occurrences *occurrences.Service
	checkpoints *store.ScheduleCheckpointRepository
	blocks      *store.CalendarBlockRepository
	clock       Clock
}

func NewService(usersService *users.Service, tasksService *tasks.Service, occurrenceService *occurrences.Service, checkpointRepo *store.ScheduleCheckpointRepository, blockRepo *store.CalendarBlockRepository) *Service {
	return &Service{
		users:       usersService,
		tasks:       tasksService,
		occurrences: occurrenceService,
		checkpoints: checkpointRepo,
		blocks:      blockRepo,
		clock:       realClock{},
	}
}

func (s *Service) PlanDay(ctx context.Context, userID string, day time.Time) (PlanResult, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return PlanResult{}, fmt.Errorf("load user: %w", err)
	}
	tasksWithSubtasks, err := s.tasks.ListTaskWithSubtasksByUser(ctx, userID)
	if err != nil {
		return PlanResult{}, fmt.Errorf("load tasks: %w", err)
	}
	allOccurrences, err := s.occurrences.ListByUser(ctx, userID)
	if err != nil {
		return PlanResult{}, fmt.Errorf("load occurrences: %w", err)
	}

	planDate := day.UTC()
	planDateStr := store.FormatDate(planDate)
	backlog, current, history := splitOccurrences(allOccurrences, planDateStr)
	calendarBlocks, err := s.blocks.ListByUserAndDate(ctx, userID, planDateStr)
	if err != nil {
		return PlanResult{}, fmt.Errorf("load calendar blocks: %w", err)
	}
	constraints := buildCalendarConstraints(calendarBlocks)

	candidates, err := s.buildCandidates(ctx, tasksWithSubtasks, history, backlog, current, planDate)
	if err != nil {
		return PlanResult{}, err
	}

	windowBudgets := splitWindowBudgets(user.DailyTimeBudgetMinutes, candidates, constraints)
	applyCalendarBudgets(windowBudgets, constraints)
	scheduled, overflowed, skipped := fitCandidates(candidates, windowBudgets, planDate, constraints)

	persistedScheduled := make([]occurrences.Occurrence, 0, len(scheduled))
	for _, candidate := range scheduled {
		candidate.Occurrence.Status = occurrences.StatusScheduled
		candidate.Occurrence.ScheduledForDate = planDateStr
		candidate.Occurrence.ScheduledTimeOfDay = candidate.Window
		candidate.Occurrence.UpdatedAt = s.clock.Now()

		persisted, err := persistOccurrence(ctx, s.occurrences, candidate.Occurrence)
		if err != nil {
			return PlanResult{}, err
		}
		persistedScheduled = append(persistedScheduled, persisted)
	}

	persistedOverflowed := make([]occurrences.Occurrence, 0, len(overflowed))
	for _, candidate := range overflowed {
		nextDay := planDate.Add(24 * time.Hour)
		candidate.Occurrence.Status = occurrences.StatusPending
		candidate.Occurrence.ScheduledForDate = store.FormatDate(nextDay)
		candidate.Occurrence.ScheduledTimeOfDay = candidate.Window
		candidate.Occurrence.RolloverCount++
		candidate.Occurrence.UpdatedAt = s.clock.Now()

		persisted, err := persistOccurrence(ctx, s.occurrences, candidate.Occurrence)
		if err != nil {
			return PlanResult{}, err
		}
		persistedOverflowed = append(persistedOverflowed, persisted)
	}

	persistedSkipped := make([]occurrences.Occurrence, 0, len(skipped))
	for _, candidate := range skipped {
		now := s.clock.Now()
		candidate.Occurrence.Status = occurrences.StatusSkipped
		candidate.Occurrence.SkippedAt = &now
		candidate.Occurrence.RolloverCount++
		candidate.Occurrence.UpdatedAt = now

		persisted, err := persistOccurrence(ctx, s.occurrences, candidate.Occurrence)
		if err != nil {
			return PlanResult{}, err
		}
		persistedSkipped = append(persistedSkipped, persisted)
	}

	checkpoint, err := s.checkpoints.Upsert(ctx, store.ScheduleCheckpoint{
		UserID:                   userID,
		ScheduleDate:             planDateStr,
		NextCheckpointAt:         nextCheckpoint(planDate, persistedScheduled),
		ScheduledOccurrenceCount: len(persistedScheduled),
		GeneratedAt:              s.clock.Now(),
	})
	if err != nil {
		return PlanResult{}, fmt.Errorf("persist schedule checkpoint: %w", err)
	}

	return PlanResult{
		Date:                 planDateStr,
		Scheduled:            persistedScheduled,
		Overflowed:           persistedOverflowed,
		Skipped:              persistedSkipped,
		Checkpoint:           checkpoint,
		WindowBudgetsMinutes: windowBudgets,
		BlockedWindows:       constraints.BlockedWindows,
		SmallTaskOnlyReason:  constraints.SmallTaskOnlyReason,
	}, nil
}

func (s *Service) PreviewDay(ctx context.Context, userID string, day time.Time) (PlanResult, error) {
	user, taskDefs, allOccurrences, err := s.loadPreviewInputs(ctx, userID)
	if err != nil {
		return PlanResult{}, err
	}
	result, _, err := s.previewDayWithOccurrences(ctx, user, taskDefs, allOccurrences, day)
	return result, err
}

func (s *Service) PreviewRange(ctx context.Context, userID string, startDay time.Time, days int) ([]PlanResult, error) {
	if days < 1 {
		return nil, nil
	}
	user, taskDefs, allOccurrences, err := s.loadPreviewInputs(ctx, userID)
	if err != nil {
		return nil, err
	}
	results := make([]PlanResult, 0, days)
	state := cloneOccurrences(allOccurrences)
	for i := 0; i < days; i++ {
		result, nextState, err := s.previewDayWithOccurrences(ctx, user, taskDefs, state, startDay.Add(time.Duration(i)*24*time.Hour))
		if err != nil {
			return nil, err
		}
		results = append(results, result)
		state = nextState
	}
	return results, nil
}

func (s *Service) loadPreviewInputs(ctx context.Context, userID string) (users.User, []tasks.TaskWithSubtasks, []occurrences.Occurrence, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return users.User{}, nil, nil, fmt.Errorf("load user: %w", err)
	}
	tasksWithSubtasks, err := s.tasks.ListTaskWithSubtasksByUser(ctx, userID)
	if err != nil {
		return users.User{}, nil, nil, fmt.Errorf("load tasks: %w", err)
	}
	allOccurrences, err := s.occurrences.ListByUser(ctx, userID)
	if err != nil {
		return users.User{}, nil, nil, fmt.Errorf("load occurrences: %w", err)
	}
	return user, tasksWithSubtasks, allOccurrences, nil
}

func (s *Service) previewDayWithOccurrences(ctx context.Context, user users.User, taskDefs []tasks.TaskWithSubtasks, allOccurrences []occurrences.Occurrence, day time.Time) (PlanResult, []occurrences.Occurrence, error) {
	planDate := day.UTC()
	planDateStr := store.FormatDate(planDate)
	backlog, current, history := splitOccurrences(allOccurrences, planDateStr)
	calendarBlocks, err := s.blocks.ListByUserAndDate(ctx, user.ID, planDateStr)
	if err != nil {
		return PlanResult{}, nil, fmt.Errorf("load calendar blocks: %w", err)
	}
	constraints := buildCalendarConstraints(calendarBlocks)

	candidates, err := s.buildCandidates(ctx, taskDefs, history, backlog, current, planDate)
	if err != nil {
		return PlanResult{}, nil, err
	}

	windowBudgets := splitWindowBudgets(user.DailyTimeBudgetMinutes, candidates, constraints)
	applyCalendarBudgets(windowBudgets, constraints)
	scheduled, overflowed, skipped := fitCandidates(candidates, windowBudgets, planDate, constraints)

	previewScheduled := make([]occurrences.Occurrence, 0, len(scheduled))
	for _, candidate := range scheduled {
		occurrence := candidate.Occurrence
		occurrence.Status = occurrences.StatusScheduled
		occurrence.ScheduledForDate = planDateStr
		occurrence.ScheduledTimeOfDay = candidate.Window
		previewScheduled = append(previewScheduled, occurrence)
	}

	previewOverflowed := make([]occurrences.Occurrence, 0, len(overflowed))
	for _, candidate := range overflowed {
		occurrence := candidate.Occurrence
		occurrence.Status = occurrences.StatusPending
		occurrence.ScheduledForDate = store.FormatDate(planDate.Add(24 * time.Hour))
		occurrence.ScheduledTimeOfDay = candidate.Window
		occurrence.RolloverCount++
		previewOverflowed = append(previewOverflowed, occurrence)
	}

	previewSkipped := make([]occurrences.Occurrence, 0, len(skipped))
	for _, candidate := range skipped {
		occurrence := candidate.Occurrence
		occurrence.Status = occurrences.StatusSkipped
		occurrence.RolloverCount++
		now := s.clock.Now()
		occurrence.SkippedAt = &now
		previewSkipped = append(previewSkipped, occurrence)
	}

	checkpoint := store.ScheduleCheckpoint{
		UserID:                   user.ID,
		ScheduleDate:             planDateStr,
		NextCheckpointAt:         nextCheckpoint(planDate, previewScheduled),
		ScheduledOccurrenceCount: len(previewScheduled),
		GeneratedAt:              s.clock.Now(),
	}

	result := PlanResult{
		Date:                 planDateStr,
		Scheduled:            previewScheduled,
		Overflowed:           previewOverflowed,
		Skipped:              previewSkipped,
		Checkpoint:           checkpoint,
		WindowBudgetsMinutes: windowBudgets,
		BlockedWindows:       constraints.BlockedWindows,
		SmallTaskOnlyReason:  constraints.SmallTaskOnlyReason,
	}
	return result, applyPreviewResults(allOccurrences, previewScheduled, previewOverflowed, previewSkipped), nil
}

type scheduledCandidate struct {
	Occurrence occurrences.Occurrence
	Task       tasks.Task
	Subtask    *tasks.Subtask
	Window     tasks.TimeOfDayPreference
	Duration   int
	OverdueAge int
	SortRank   int
}

func (s *Service) buildCandidates(ctx context.Context, taskDefs []tasks.TaskWithSubtasks, history, backlog, current []occurrences.Occurrence, planDate time.Time) ([]scheduledCandidate, error) {
	planDateStr := store.FormatDate(planDate)
	openKeys := map[string]bool{}
	openUnits := map[string]bool{}
	for _, item := range append(backlog, current...) {
		openKeys[occurrenceKey(item.TaskID, item.SubtaskID, item.OriginalScheduledForDate)] = true
		openUnits[unitKey(item.TaskID, item.SubtaskID)] = true
	}

	var candidates []scheduledCandidate
	for _, open := range append(backlog, current...) {
		taskDef, subtaskDef, ok := findDefinition(taskDefs, open.TaskID, open.SubtaskID)
		if !ok {
			continue
		}
		window := effectiveWindow(taskDef, subtaskDef)
		candidates = append(candidates, candidateFromOccurrence(open, taskDef, subtaskDef, window, planDateStr))
	}

	for _, taskDef := range taskDefs {
		if taskDef.Task.IsPaused {
			continue
		}
		if !isDue(taskDef, history, planDate) {
			continue
		}

		if len(taskDef.Subtasks) == 0 {
			if openUnits[unitKey(taskDef.Task.ID, "")] || openKeys[occurrenceKey(taskDef.Task.ID, "", planDateStr)] {
				continue
			}
			candidate := scheduledCandidate{
				Occurrence: occurrences.Occurrence{
					UserID:                   taskDef.Task.UserID,
					TaskID:                   taskDef.Task.ID,
					Status:                   occurrences.StatusPending,
					ScheduledForDate:         planDateStr,
					OriginalScheduledForDate: planDateStr,
					ScheduledTimeOfDay:       taskDef.Task.TimeOfDayPreference,
				},
				Task:       taskDef.Task,
				Window:     taskDef.Task.TimeOfDayPreference,
				Duration:   taskDef.Task.DurationMinutes,
				OverdueAge: 0,
				SortRank:   priorityRank(taskDef.Task.Priority),
			}
			if candidate.Window == tasks.TimeOfDayAny {
				candidate.Window = tasks.TimeOfDayMorning
			}
			candidates = append(candidates, candidate)
			continue
		}

		for _, subtaskDef := range taskDef.Subtasks {
			key := occurrenceKey(taskDef.Task.ID, subtaskDef.ID, planDateStr)
			if openUnits[unitKey(taskDef.Task.ID, subtaskDef.ID)] || openKeys[key] {
				continue
			}
			subtaskCopy := subtaskDef
			candidate := scheduledCandidate{
				Occurrence: occurrences.Occurrence{
					UserID:                   taskDef.Task.UserID,
					TaskID:                   taskDef.Task.ID,
					SubtaskID:                subtaskDef.ID,
					Status:                   occurrences.StatusPending,
					ScheduledForDate:         planDateStr,
					OriginalScheduledForDate: planDateStr,
					ScheduledTimeOfDay:       subtaskDef.TimeOfDayPreference,
				},
				Task:       taskDef.Task,
				Subtask:    &subtaskCopy,
				Window:     subtaskDef.TimeOfDayPreference,
				Duration:   subtaskDef.DurationMinutes,
				OverdueAge: 0,
				SortRank:   priorityRank(taskDef.Task.Priority),
			}
			candidates = append(candidates, candidate)
		}
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].Window != candidates[j].Window {
			return daytime.Order(string(candidates[i].Window)) < daytime.Order(string(candidates[j].Window))
		}
		if candidates[i].SortRank != candidates[j].SortRank {
			return candidates[i].SortRank < candidates[j].SortRank
		}
		if candidates[i].OverdueAge != candidates[j].OverdueAge {
			return candidates[i].OverdueAge > candidates[j].OverdueAge
		}
		if candidates[i].Subtask != nil && candidates[j].Subtask != nil && candidates[i].Subtask.TaskID == candidates[j].Subtask.TaskID {
			return candidates[i].Subtask.StepOrder < candidates[j].Subtask.StepOrder
		}
		return candidates[i].Occurrence.TaskID < candidates[j].Occurrence.TaskID
	})

	return candidates, nil
}

func cloneOccurrences(all []occurrences.Occurrence) []occurrences.Occurrence {
	cloned := make([]occurrences.Occurrence, 0, len(all))
	for _, occurrence := range all {
		cloned = append(cloned, cloneOccurrence(occurrence))
	}
	return cloned
}

func cloneOccurrence(occurrence occurrences.Occurrence) occurrences.Occurrence {
	cloned := occurrence
	if occurrence.ReadyAt != nil {
		readyAt := occurrence.ReadyAt.UTC()
		cloned.ReadyAt = &readyAt
	}
	if occurrence.SkippedAt != nil {
		skippedAt := occurrence.SkippedAt.UTC()
		cloned.SkippedAt = &skippedAt
	}
	return cloned
}

func applyPreviewResults(existing, scheduled, overflowed, skipped []occurrences.Occurrence) []occurrences.Occurrence {
	state := cloneOccurrences(existing)
	for _, occurrence := range scheduled {
		state = upsertPreviewOccurrence(state, occurrence)
	}
	for _, occurrence := range overflowed {
		state = upsertPreviewOccurrence(state, occurrence)
	}
	for _, occurrence := range skipped {
		state = upsertPreviewOccurrence(state, occurrence)
	}
	return state
}

func upsertPreviewOccurrence(state []occurrences.Occurrence, updated occurrences.Occurrence) []occurrences.Occurrence {
	for idx, occurrence := range state {
		if updated.ID != "" && occurrence.ID == updated.ID {
			state[idx] = cloneOccurrence(updated)
			return state
		}
		if occurrence.TaskID == updated.TaskID && occurrence.SubtaskID == updated.SubtaskID && occurrence.OriginalScheduledForDate == updated.OriginalScheduledForDate {
			state[idx] = cloneOccurrence(updated)
			return state
		}
	}
	return append(state, cloneOccurrence(updated))
}

func splitOccurrences(all []occurrences.Occurrence, planDate string) (backlog, current, history []occurrences.Occurrence) {
	for _, occurrence := range all {
		switch occurrence.Status {
		case occurrences.StatusCompleted, occurrences.StatusSkipped:
			history = append(history, occurrence)
		default:
			if occurrence.ScheduledForDate < planDate {
				backlog = append(backlog, occurrence)
			} else if occurrence.ScheduledForDate == planDate {
				current = append(current, occurrence)
			}
		}
	}
	return backlog, current, history
}

func candidateFromOccurrence(occ occurrences.Occurrence, taskDef tasks.Task, subtaskDef *tasks.Subtask, window tasks.TimeOfDayPreference, planDate string) scheduledCandidate {
	overdue := 0
	if occ.OriginalScheduledForDate < planDate {
		d1, _ := time.Parse(store.DateLayout, occ.OriginalScheduledForDate)
		d2, _ := time.Parse(store.DateLayout, planDate)
		overdue = int(d2.Sub(d1).Hours() / 24)
	}
	duration := taskDef.DurationMinutes
	if subtaskDef != nil {
		duration = subtaskDef.DurationMinutes
	}
	if window == tasks.TimeOfDayAny {
		window = tasks.TimeOfDayMorning
	}
	return scheduledCandidate{Occurrence: occ, Task: taskDef, Subtask: subtaskDef, Window: window, Duration: duration, OverdueAge: overdue, SortRank: priorityRank(taskDef.Priority)}
}

func findDefinition(taskDefs []tasks.TaskWithSubtasks, taskID, subtaskID string) (tasks.Task, *tasks.Subtask, bool) {
	for _, taskDef := range taskDefs {
		if taskDef.Task.ID != taskID {
			continue
		}
		if subtaskID == "" {
			return taskDef.Task, nil, true
		}
		for _, subtask := range taskDef.Subtasks {
			if subtask.ID == subtaskID {
				subtaskCopy := subtask
				return taskDef.Task, &subtaskCopy, true
			}
		}
		return taskDef.Task, nil, true
	}
	return tasks.Task{}, nil, false
}

func effectiveWindow(taskDef tasks.Task, subtaskDef *tasks.Subtask) tasks.TimeOfDayPreference {
	if subtaskDef != nil && subtaskDef.TimeOfDayPreference != tasks.TimeOfDayAny {
		return subtaskDef.TimeOfDayPreference
	}
	if taskDef.TimeOfDayPreference != tasks.TimeOfDayAny {
		return taskDef.TimeOfDayPreference
	}
	return tasks.TimeOfDayMorning
}

func isDue(taskDef tasks.TaskWithSubtasks, history []occurrences.Occurrence, planDate time.Time) bool {
	switch taskDef.Task.CadenceType {
	case tasks.CadenceTypeInterval:
		anchor, ok := latestAnchorDate(taskDef, history)
		if !ok {
			return true
		}
		return !planDate.Before(anchor.AddDate(0, 0, taskDef.Task.CadenceValue))
	case tasks.CadenceTypeCount:
		weekStart := startOfWeek(planDate)
		weekEnd := weekStart.AddDate(0, 0, 7)
		count := 0
		var lastAnchor time.Time
		seenAnchors := map[string]bool{}
		for _, occurrence := range history {
			if occurrence.TaskID != taskDef.Task.ID {
				continue
			}
			anchor := resolvedAnchorDate(occurrence)
			anchorKey := store.FormatDate(anchor)
			if !anchor.Before(weekStart) && anchor.Before(weekEnd) && !seenAnchors[anchorKey] {
				seenAnchors[anchorKey] = true
				count++
				if anchor.After(lastAnchor) {
					lastAnchor = anchor
				}
			}
		}
		if count >= taskDef.Task.CadenceValue {
			return false
		}
		remaining := taskDef.Task.CadenceValue - count
		daysLeft := int(weekEnd.Sub(planDate).Hours() / 24)
		if remaining >= daysLeft {
			return true
		}
		if count == 0 {
			return true
		}
		spacing := max(1, 7/taskDef.Task.CadenceValue)
		return !planDate.Before(lastAnchor.AddDate(0, 0, spacing))
	default:
		return false
	}
}

func latestAnchorDate(taskDef tasks.TaskWithSubtasks, history []occurrences.Occurrence) (time.Time, bool) {
	var latest time.Time
	found := false
	for _, occurrence := range history {
		if occurrence.TaskID != taskDef.Task.ID {
			continue
		}
		anchor := resolvedAnchorDate(occurrence)
		if !found || anchor.After(latest) {
			latest = anchor
			found = true
		}
	}
	return latest, found
}

func resolvedAnchorDate(occurrence occurrences.Occurrence) time.Time {
	value := occurrence.ScheduledForDate
	if occurrence.Status == occurrences.StatusSkipped {
		value = occurrence.OriginalScheduledForDate
	}
	parsed, _ := time.Parse(store.DateLayout, value)
	return parsed.UTC()
}

func startOfWeek(day time.Time) time.Time {
	weekday := int(day.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	return time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, -(weekday - 1))
}

func splitWindowBudgets(total int, candidates []scheduledCandidate, constraints calendarConstraints) map[string]int {
	budgets := map[string]int{"morning": 0, "afternoon": 0, "evening": 0}
	if total <= 0 {
		return budgets
	}
	demand := map[string]int{"morning": 0, "afternoon": 0, "evening": 0}
	for _, candidate := range candidates {
		window := string(candidate.Window)
		if window == string(tasks.TimeOfDayAny) || window == "" {
			window = "morning"
		}
		demand[window] += candidate.Duration
	}

	totalDemand := demand["morning"] + demand["afternoon"] + demand["evening"]
	if totalDemand == 0 {
		budgets["morning"] = total / 3
		budgets["afternoon"] = total / 3
		budgets["evening"] = total - budgets["morning"] - budgets["afternoon"]
		return budgets
	}

	// If everything fits in the daily budget, allocate each window exactly the
	// demand it needs, then distribute the leftover minutes to windows that are
	// not blocked by the calendar. This prevents artificial per-window caps from
	// making feasible combinations fail.
	if totalDemand <= total {
		budgets["morning"] = demand["morning"]
		budgets["afternoon"] = demand["afternoon"]
		budgets["evening"] = demand["evening"]
		leftover := total - totalDemand
		if leftover > 0 {
			unblocked := []string{}
			for _, window := range []string{"morning", "afternoon", "evening"} {
				if !constraints.ZeroBudgetWindows[window] {
					unblocked = append(unblocked, window)
				}
			}
			if len(unblocked) > 0 {
				per := leftover / len(unblocked)
				rem := leftover % len(unblocked)
				for i, window := range unblocked {
					budgets[window] += per
					if i < rem {
						budgets[window]++
					}
				}
			}
		}
		return budgets
	}

	// When demand exceeds the daily budget, fall back to a proportional split so
	// each window gets a fair share of the limited time.
	allocated := 0
	for idx, window := range []string{"morning", "afternoon", "evening"} {
		if idx == 2 {
			budgets[window] = total - allocated
			break
		}
		share := (total * demand[window]) / totalDemand
		budgets[window] = share
		allocated += share
	}
	if budgets["morning"] == 0 && demand["morning"] > 0 && total > 0 {
		budgets["morning"] = 1
		budgets["evening"] = max(0, budgets["evening"]-1)
	}
	return budgets
}

type calendarConstraints struct {
	BlockedWindows      map[string][]string
	ZeroBudgetWindows   map[string]bool
	SmallTaskOnlyReason string
}

func buildCalendarConstraints(blocks []store.CalendarBlock) calendarConstraints {
	constraints := calendarConstraints{BlockedWindows: map[string][]string{"morning": {}, "afternoon": {}, "evening": {}}, ZeroBudgetWindows: map[string]bool{}}
	hasLargeDayConstraint := false
	for _, block := range blocks {
		reason := block.Title
		if reason == "" {
			reason = "calendar event"
		}
		reason = fmt.Sprintf("%s (%s)", reason, block.Classification)
		switch block.Classification {
		case "medium":
			if block.Window == "morning" || block.Window == "afternoon" || block.Window == "evening" {
				constraints.BlockedWindows[block.Window] = append(constraints.BlockedWindows[block.Window], reason)
				constraints.ZeroBudgetWindows[block.Window] = true
			}
		case "large":
			if block.Window == "all-day" || block.IsAllDay {
				hasLargeDayConstraint = true
				for _, window := range []string{"morning", "afternoon", "evening"} {
					constraints.BlockedWindows[window] = append(constraints.BlockedWindows[window], reason)
				}
			} else if block.Window == "morning" || block.Window == "afternoon" || block.Window == "evening" {
				constraints.BlockedWindows[block.Window] = append(constraints.BlockedWindows[block.Window], reason)
				constraints.ZeroBudgetWindows[block.Window] = true
			}
		}
	}
	if hasLargeDayConstraint {
		constraints.SmallTaskOnlyReason = "Large calendar commitment today; only shorter tasks should be scheduled."
	}
	return constraints
}

func applyCalendarBudgets(budgets map[string]int, constraints calendarConstraints) {
	for window := range constraints.ZeroBudgetWindows {
		budgets[window] = 0
	}
}

func fitCandidates(candidates []scheduledCandidate, budgets map[string]int, planDate time.Time, constraints calendarConstraints) (scheduled, overflowed, skipped []scheduledCandidate) {
	used := map[string]int{"morning": 0, "afternoon": 0, "evening": 0}
	stepWindows := map[string]int{}
	stepReadyAt := map[string]time.Time{}
	processedSubtaskTasks := map[string]bool{}

	for _, candidate := range candidates {
		if candidate.Subtask != nil {
			if processedSubtaskTasks[candidate.Task.ID] {
				continue
			}
			processedSubtaskTasks[candidate.Task.ID] = true
			group := collectSubtaskCandidates(candidates, candidate.Task.ID)
			required, softFollowups := splitRequiredAndSoftFollowups(group)
			requiredScheduled, ok := tryFitCandidateGroup(required, used, stepWindows, stepReadyAt, budgets, planDate, constraints)
			if !ok {
				groupOverflowed, groupSkipped := deferCandidateGroup(group)
				overflowed = append(overflowed, groupOverflowed...)
				skipped = append(skipped, groupSkipped...)
				continue
			}
			scheduled = append(scheduled, requiredScheduled...)
			for _, followup := range softFollowups {
				fitted, ok := tryFitCandidate(followup, used, stepWindows, stepReadyAt, budgets, planDate, constraints)
				if ok {
					scheduled = append(scheduled, fitted)
					continue
				}
				followupOverflowed, followupSkipped := deferCandidateGroup([]scheduledCandidate{followup})
				overflowed = append(overflowed, followupOverflowed...)
				skipped = append(skipped, followupSkipped...)
			}
			continue
		}

		fitted, ok := tryFitCandidate(candidate, used, stepWindows, stepReadyAt, budgets, planDate, constraints)
		if ok {
			scheduled = append(scheduled, fitted)
			continue
		}
		candidateOverflowed, candidateSkipped := deferCandidateGroup([]scheduledCandidate{candidate})
		overflowed = append(overflowed, candidateOverflowed...)
		skipped = append(skipped, candidateSkipped...)
	}

	return scheduled, overflowed, skipped
}

func collectSubtaskCandidates(candidates []scheduledCandidate, taskID string) []scheduledCandidate {
	group := []scheduledCandidate{}
	for _, candidate := range candidates {
		if candidate.Subtask != nil && candidate.Task.ID == taskID {
			group = append(group, candidate)
		}
	}
	sort.SliceStable(group, func(i, j int) bool {
		return group[i].Subtask.StepOrder < group[j].Subtask.StepOrder
	})
	return group
}

func splitRequiredAndSoftFollowups(group []scheduledCandidate) (required, softFollowups []scheduledCandidate) {
	for _, candidate := range group {
		if candidate.Subtask != nil && candidate.Subtask.DependencyType == tasks.SubtaskDependencySoftFollowup {
			softFollowups = append(softFollowups, candidate)
			continue
		}
		required = append(required, candidate)
	}
	if len(required) == 0 {
		return group, nil
	}
	return required, softFollowups
}

func tryFitCandidateGroup(group []scheduledCandidate, used, stepWindows map[string]int, stepReadyAt map[string]time.Time, budgets map[string]int, planDate time.Time, constraints calendarConstraints) ([]scheduledCandidate, bool) {
	simUsed := copyIntMap(used)
	simStepWindows := copyIntMap(stepWindows)
	simStepReadyAt := copyTimeMap(stepReadyAt)
	scheduled := make([]scheduledCandidate, 0, len(group))
	for _, candidate := range group {
		fitted, ok := tryFitCandidate(candidate, simUsed, simStepWindows, simStepReadyAt, budgets, planDate, constraints)
		if !ok {
			return nil, false
		}
		scheduled = append(scheduled, fitted)
	}
	copyIntMapInto(used, simUsed)
	copyIntMapInto(stepWindows, simStepWindows)
	copyTimeMapInto(stepReadyAt, simStepReadyAt)
	return scheduled, true
}

func tryFitCandidate(candidate scheduledCandidate, used, stepWindows map[string]int, stepReadyAt map[string]time.Time, budgets map[string]int, planDate time.Time, constraints calendarConstraints) (scheduledCandidate, bool) {
	candidate = prepareCandidate(candidate, stepWindows, stepReadyAt, planDate)
	preferred := string(candidate.Window)
	for _, window := range fallbackWindows(preferred) {
		if fitted, ok := tryFitInWindow(candidate, window, used, stepWindows, stepReadyAt, budgets, planDate, constraints); ok {
			return fitted, true
		}
	}
	return candidate, false
}

func fallbackWindows(preferred string) []string {
	order := []string{"morning", "afternoon", "evening"}
	result := make([]string, 0, len(order))
	if preferred != "" {
		result = append(result, preferred)
	}
	for _, window := range order {
		if window != preferred {
			result = append(result, window)
		}
	}
	return result
}

func tryFitInWindow(candidate scheduledCandidate, window string, used, stepWindows map[string]int, stepReadyAt map[string]time.Time, budgets map[string]int, planDate time.Time, constraints calendarConstraints) (scheduledCandidate, bool) {
	// Subtask chains cannot move to a window earlier than a previously scheduled step.
	if candidate.Subtask != nil {
		if prevOrder, ok := stepWindows[candidate.Task.ID]; ok {
			if daytime.Order(window) < prevOrder {
				return candidate, false
			}
		}
	}

	readyAt := daytime.StartTime(planDate, window)
	if candidate.Subtask != nil {
		if prevReady, ok := stepReadyAt[candidate.Task.ID]; ok {
			minReady := prevReady.Add(time.Duration(candidate.Subtask.GapRule.MinGapAfterPreviousMinutes) * time.Minute)
			if minReady.After(readyAt) {
				readyAt = minReady
			}
			// If the required gap pushes the step out of this window, the window cannot hold it.
			if shiftedWindow, ok := daytime.WindowForTime(planDate, readyAt); !ok || shiftedWindow != window {
				return candidate, false
			}
		}
	}

	if constraints.SmallTaskOnlyReason != "" && candidate.Duration > 15 {
		return candidate, false
	}
	if used[window]+candidate.Duration > budgets[window] {
		return candidate, false
	}

	used[window] += candidate.Duration
	candidate.Window = tasks.TimeOfDayPreference(window)
	candidate.Occurrence.ReadyAt = &readyAt
	if candidate.Subtask != nil {
		stepWindows[candidate.Task.ID] = daytime.Order(window)
		stepReadyAt[candidate.Task.ID] = readyAt
	}
	return candidate, true
}

func prepareCandidate(candidate scheduledCandidate, stepWindows map[string]int, stepReadyAt map[string]time.Time, planDate time.Time) scheduledCandidate {
	window := string(candidate.Window)
	if window == "" || window == string(tasks.TimeOfDayAny) {
		window = "morning"
	}
	candidate.Window = tasks.TimeOfDayPreference(window)

	readyAt := daytime.StartTime(planDate, window)
	if candidate.Subtask != nil {
		prevWindow, ok := stepWindows[candidate.Task.ID]
		currentWindow := daytime.Order(window)
		if candidate.Subtask.StepOrder > 1 && ok && currentWindow < prevWindow {
			currentWindow = prevWindow
			switch currentWindow {
			case 0:
				window = "morning"
			case 1:
				window = "afternoon"
			default:
				window = "evening"
			}
			candidate.Window = tasks.TimeOfDayPreference(window)
			readyAt = daytime.StartTime(planDate, window)
		}
		if prevReady, ok := stepReadyAt[candidate.Task.ID]; ok {
			minReady := prevReady.Add(time.Duration(candidate.Subtask.GapRule.MinGapAfterPreviousMinutes) * time.Minute)
			if minReady.After(readyAt) {
				readyAt = minReady
			}
			if shiftedWindow, ok := daytime.WindowForTime(planDate, readyAt); ok {
				window = shiftedWindow
				candidate.Window = tasks.TimeOfDayPreference(window)
			} else {
				window = "evening"
				candidate.Window = tasks.TimeOfDayPreference(window)
				readyAt = daytime.EndTime(planDate, window)
			}
		}
	}
	candidate.Occurrence.ReadyAt = &readyAt
	return candidate
}

func deferCandidateGroup(group []scheduledCandidate) (overflowed, skipped []scheduledCandidate) {
	for _, candidate := range group {
		nextRollover := candidate.Occurrence.RolloverCount + 1
		if candidate.Task.Priority != tasks.PriorityHigh && nextRollover >= rolloverCap {
			skipped = append(skipped, candidate)
			continue
		}
		overflowed = append(overflowed, candidate)
	}
	return overflowed, skipped
}

func copyIntMap(source map[string]int) map[string]int {
	result := map[string]int{}
	for key, value := range source {
		result[key] = value
	}
	return result
}

func copyIntMapInto(target, source map[string]int) {
	for key, value := range source {
		target[key] = value
	}
}

func copyTimeMap(source map[string]time.Time) map[string]time.Time {
	result := map[string]time.Time{}
	for key, value := range source {
		result[key] = value
	}
	return result
}

func copyTimeMapInto(target, source map[string]time.Time) {
	for key, value := range source {
		target[key] = value
	}
}

func persistOccurrence(ctx context.Context, service *occurrences.Service, occurrence occurrences.Occurrence) (occurrences.Occurrence, error) {
	if occurrence.ID == "" {
		return service.Create(ctx, occurrence)
	}
	return service.Update(ctx, occurrence)
}

func nextCheckpoint(day time.Time, scheduled []occurrences.Occurrence) *time.Time {
	if len(scheduled) == 0 {
		return nil
	}
	sort.SliceStable(scheduled, func(i, j int) bool {
		left := scheduled[i].ReadyAt
		right := scheduled[j].ReadyAt
		if left != nil && right != nil && !left.Equal(*right) {
			return left.Before(*right)
		}
		return daytime.Order(string(scheduled[i].ScheduledTimeOfDay)) < daytime.Order(string(scheduled[j].ScheduledTimeOfDay))
	})
	if scheduled[0].ReadyAt != nil {
		ready := scheduled[0].ReadyAt.UTC()
		return &ready
	}
	next := daytime.StartTime(day, string(scheduled[0].ScheduledTimeOfDay))
	return &next
}

func occurrenceKey(taskID, subtaskID, original string) string {
	return taskID + "|" + subtaskID + "|" + original
}

func unitKey(taskID, subtaskID string) string {
	return taskID + "|" + subtaskID
}

func priorityRank(priority tasks.Priority) int {
	switch priority {
	case tasks.PriorityHigh:
		return 0
	case tasks.PriorityMedium:
		return 1
	default:
		return 2
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
