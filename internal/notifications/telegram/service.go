package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/rahat/rahat/internal/events"
	"github.com/rahat/rahat/internal/occurrences"
	"github.com/rahat/rahat/internal/scheduler"
	"github.com/rahat/rahat/internal/store"
	"github.com/rahat/rahat/internal/tasks"
	daytime "github.com/rahat/rahat/internal/time"
	"github.com/rahat/rahat/internal/users"
)

type Service struct {
	bot                BotClient
	users              *users.Service
	tasks              *tasks.Service
	occurrences        *occurrences.Service
	events             *events.Service
	confirmations      *store.OnboardingConfirmationRepository
}

func NewService(bot BotClient, usersService *users.Service, tasksService *tasks.Service, occurrenceService *occurrences.Service, eventService *events.Service, confirmations *store.OnboardingConfirmationRepository) *Service {
	return &Service{bot: bot, users: usersService, tasks: tasksService, occurrences: occurrenceService, events: eventService, confirmations: confirmations}
}

func (s *Service) SendMorningBatch(ctx context.Context, userID string, day time.Time) error {
	user, items, err := s.loadScheduledItems(ctx, userID, day)
	if err != nil {
		return err
	}
	if user.TelegramChatID == "" {
		return fmt.Errorf("user %s has no telegram chat configured", userID)
	}

	lines := []string{"Here’s today’s list:"}
	for _, item := range items {
		lines = append(lines, fmt.Sprintf("- %s (%s)", item.Label, item.Window))
	}
	if len(items) == 0 {
		lines = append(lines, "- Nothing scheduled right now.")
	}

	markup := &ReplyMarkup{InlineKeyboard: [][]InlineButton{{
		{Text: "Pause everything today", CallbackData: PauseTodayAction(userID)},
		{Text: "Pause this week", CallbackData: PauseWeekAction(userID)},
	}}}

	if err := s.bot.SendMessage(ctx, SendMessageRequest{ChatID: user.TelegramChatID, Text: strings.Join(lines, "\n"), ReplyMarkup: markup}); err != nil {
		return err
	}
	return s.logEvent(ctx, userID, "", "message_sent", "daily_list", map[string]any{"date": day.Format("2006-01-02"), "count": len(items)})
}

func (s *Service) SendWindowReminders(ctx context.Context, userID string, day time.Time, window tasks.TimeOfDayPreference) error {
	user, items, err := s.loadScheduledItems(ctx, userID, day)
	if err != nil {
		return err
	}
	if user.TelegramChatID == "" {
		return fmt.Errorf("user %s has no telegram chat configured", userID)
	}

	for _, item := range items {
		if item.Occurrence.ScheduledTimeOfDay != window {
			continue
		}
		markup := &ReplyMarkup{InlineKeyboard: [][]InlineButton{
			{{Text: "Done", CallbackData: DoneAction(userID, item.Occurrence.ID)}, {Text: "Not Yet", CallbackData: NotYetAction(userID, item.Occurrence.ID)}},
			{{Text: "Pause this task", CallbackData: PauseTaskAction(userID, item.Occurrence.TaskID)}},
		}}
		text := fmt.Sprintf("%s window: %s", strings.Title(string(window)), item.Label)
		if err := s.bot.SendMessage(ctx, SendMessageRequest{ChatID: user.TelegramChatID, Text: text, ReplyMarkup: markup}); err != nil {
			return err
		}
		if err := s.logEvent(ctx, userID, item.Occurrence.ID, "message_sent", "window_reminder", map[string]any{"window": window, "task": item.Label}); err != nil {
			return err
		}
	}
	return nil
}

type scheduledItem struct {
	Occurrence occurrences.Occurrence
	Label      string
	Window     string
}

func (s *Service) loadScheduledItems(ctx context.Context, userID string, day time.Time) (users.User, []scheduledItem, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return users.User{}, nil, fmt.Errorf("load user: %w", err)
	}
	defs, err := s.tasks.ListTaskWithSubtasksByUser(ctx, userID)
	if err != nil {
		return users.User{}, nil, fmt.Errorf("load task definitions: %w", err)
	}
	occList, err := s.occurrences.ListByUser(ctx, userID)
	if err != nil {
		return users.User{}, nil, fmt.Errorf("load occurrences: %w", err)
	}
	labels := labelsByTaskUnit(defs)
	date := day.Format("2006-01-02")
	items := make([]scheduledItem, 0)
	for _, occurrence := range occList {
		if occurrence.Status != occurrences.StatusScheduled || occurrence.ScheduledForDate != date {
			continue
		}
		label := labels[occurrence.SubtaskID]
		if label == "" {
			label = labels[occurrence.TaskID]
		}
		if label == "" {
			label = occurrence.TaskID
		}
		items = append(items, scheduledItem{Occurrence: occurrence, Label: label, Window: string(occurrence.ScheduledTimeOfDay)})
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Occurrence.ScheduledTimeOfDay != items[j].Occurrence.ScheduledTimeOfDay {
			return items[i].Occurrence.ScheduledTimeOfDay < items[j].Occurrence.ScheduledTimeOfDay
		}
		return items[i].Label < items[j].Label
	})
	return user, items, nil
}

func labelsByTaskUnit(defs []tasks.TaskWithSubtasks) map[string]string {
	result := map[string]string{}
	for _, def := range defs {
		result[def.Task.ID] = def.Task.Name
		for _, subtask := range def.Subtasks {
			result[subtask.ID] = fmt.Sprintf("%s — %s", def.Task.Name, subtask.Name)
		}
	}
	return result
}

func (s *Service) logEvent(ctx context.Context, userID, occurrenceID, eventType, messageType string, payload map[string]any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal telegram event payload: %w", err)
	}
	_, err = s.events.Create(ctx, events.EventLog{UserID: userID, OccurrenceID: occurrenceID, Channel: "telegram", EventType: eventType, MessageType: messageType, PayloadJSON: string(body)})
	return err
}

// SendOnboardingConfirmation sends a one-time Telegram summary after onboarding.
// It is idempotent: repeated calls for the same user return the previously
// recorded result without sending another message.
func (s *Service) SendOnboardingConfirmation(ctx context.Context, userID string, planDate time.Time, plan scheduler.PlanResult, taskDefs []tasks.TaskWithSubtasks) (bool, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("load user: %w", err)
	}
	if user.TelegramChatID == "" {
		return false, nil
	}

	if s.confirmations != nil {
		existing, found, err := s.confirmations.Get(ctx, userID)
		if err != nil {
			return false, fmt.Errorf("check onboarding confirmation: %w", err)
		}
		if found {
			return existing.Delivered, nil
		}
	}

	text := buildOnboardingConfirmationText(user, planDate, taskDefs, plan)
	if err := s.bot.SendMessage(ctx, SendMessageRequest{ChatID: user.TelegramChatID, Text: text}); err != nil {
		if s.confirmations != nil {
			_ = s.confirmations.Record(ctx, userID, false, err.Error())
		}
		_ = s.logEvent(ctx, userID, "", "message_failed", "onboarding_confirmation", map[string]any{
			"date":    plan.Date,
			"success": false,
			"reason":  "send failed",
		})
		return false, err
	}

	if s.confirmations != nil {
		if err := s.confirmations.Record(ctx, userID, true, ""); err != nil {
			return false, fmt.Errorf("record onboarding confirmation: %w", err)
		}
	}
	_ = s.logEvent(ctx, userID, "", "message_sent", "onboarding_confirmation", map[string]any{
		"date":    plan.Date,
		"success": true,
	})
	return true, nil
}

func buildOnboardingConfirmationText(user users.User, planDate time.Time, taskDefs []tasks.TaskWithSubtasks, plan scheduler.PlanResult) string {
	labels := labelsByTaskUnit(taskDefs)

	scheduledItems := make([]scheduledItem, 0, len(plan.Scheduled))
	for _, occurrence := range plan.Scheduled {
		label := labels[occurrence.SubtaskID]
		if label == "" {
			label = labels[occurrence.TaskID]
		}
		if label == "" {
			label = occurrence.TaskID
		}
		scheduledItems = append(scheduledItems, scheduledItem{
			Occurrence: occurrence,
			Label:      label,
			Window:     string(occurrence.ScheduledTimeOfDay),
		})
	}
	sort.SliceStable(scheduledItems, func(i, j int) bool {
		leftReady := scheduledItems[i].Occurrence.ReadyAt
		rightReady := scheduledItems[j].Occurrence.ReadyAt
		if leftReady != nil && rightReady != nil && !leftReady.Equal(*rightReady) {
			return leftReady.Before(*rightReady)
		}
		if scheduledItems[i].Window != scheduledItems[j].Window {
			return daytime.Order(scheduledItems[i].Window) < daytime.Order(scheduledItems[j].Window)
		}
		return scheduledItems[i].Label < scheduledItems[j].Label
	})

	loc, err := time.LoadLocation(user.Timezone)
	if err != nil {
		loc = time.UTC
	}

	lines := []string{fmt.Sprintf("Welcome to Rahat, %s!", user.DisplayName), ""}

	lines = append(lines, "Your routines are saved:")
	for _, taskDef := range taskDefs {
		lines = append(lines, fmt.Sprintf("- %s", taskDef.Task.Name))
	}
	lines = append(lines, "")

	dateLabel := planDate.In(loc).Format("Monday, January 2")
	if parsedDate, err := time.ParseInLocation("2006-01-02", plan.Date, loc); err == nil {
		dateLabel = parsedDate.Format("Monday, January 2")
	}
	lines = append(lines, fmt.Sprintf("For %s:", dateLabel))

	grouped := map[string][]scheduledItem{"morning": {}, "afternoon": {}, "evening": {}}
	for _, item := range scheduledItems {
		window := item.Window
		if window == "" || window == string(tasks.TimeOfDayAny) {
			window = "morning"
		}
		grouped[window] = append(grouped[window], item)
	}

	hasScheduled := false
	for _, window := range []string{"morning", "afternoon", "evening"} {
		items := grouped[window]
		if len(items) == 0 {
			continue
		}
		hasScheduled = true
		lines = append(lines, fmt.Sprintf("%s:", strings.Title(window)))
		for _, item := range items {
			lines = append(lines, fmt.Sprintf("- %s", item.Label))
		}
	}
	if !hasScheduled {
		lines = append(lines, "- Nothing scheduled for this day yet.")
	}
	lines = append(lines, "")

	if len(plan.Overflowed) > 0 || len(plan.Skipped) > 0 {
		parts := []string{}
		if len(plan.Overflowed) > 0 {
			parts = append(parts, fmt.Sprintf("%d item(s) moved to a later day", len(plan.Overflowed)))
		}
		if len(plan.Skipped) > 0 {
			parts = append(parts, fmt.Sprintf("%d item(s) skipped because they couldn't fit", len(plan.Skipped)))
		}
		lines = append(lines, fmt.Sprintf("There wasn't room for everything today: %s.", strings.Join(parts, " and ")))
	} else {
		lines = append(lines, "Everything that fit today is scheduled.")
	}

	lines = append(lines, "", "Rahat will send reminders here when it's time. Send /edit whenever you need routine settings, especially on a new or signed-out device.")

	return strings.Join(lines, "\n")
}

func DoneAction(_ string, occurrenceID string) string   { return "d:" + occurrenceID }
func NotYetAction(_ string, occurrenceID string) string { return "n:" + occurrenceID }
func SnoozeAction(_ string, occurrenceID string) string { return "z:" + occurrenceID }
func RescheduleAction(_ string, occurrenceID string) string {
	return "r:" + occurrenceID
}
func SkipAction(_ string, occurrenceID string) string { return "k:" + occurrenceID }
func PauseTodayAction(userID string) string           { return "pt:" + userID }
func PauseWeekAction(userID string) string            { return "pw:" + userID }
func PauseTaskAction(_ string, taskID string) string  { return "pk:" + taskID }
