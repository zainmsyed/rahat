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
	"github.com/rahat/rahat/internal/tasks"
	"github.com/rahat/rahat/internal/users"
)

type Service struct {
	bot         BotClient
	users       *users.Service
	tasks       *tasks.Service
	occurrences *occurrences.Service
	events      *events.Service
}

func NewService(bot BotClient, usersService *users.Service, tasksService *tasks.Service, occurrenceService *occurrences.Service, eventService *events.Service) *Service {
	return &Service{bot: bot, users: usersService, tasks: tasksService, occurrences: occurrenceService, events: eventService}
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
