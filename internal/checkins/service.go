package checkins

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/rahat/rahat/internal/events"
	preferences "github.com/rahat/rahat/internal/notifications/preferences"
	ntg "github.com/rahat/rahat/internal/notifications/telegram"
	"github.com/rahat/rahat/internal/occurrences"
	"github.com/rahat/rahat/internal/tasks"
	"github.com/rahat/rahat/internal/users"
)

type Service struct {
	bot         ntg.BotClient
	users       *users.Service
	tasks       *tasks.Service
	occurrences *occurrences.Service
	events      *events.Service
	prefs       *preferences.Service
	now         func() time.Time
}

func NewService(bot ntg.BotClient, usersService *users.Service, tasksService *tasks.Service, occurrenceService *occurrences.Service, eventService *events.Service, prefService *preferences.Service) *Service {
	return &Service{bot: bot, users: usersService, tasks: tasksService, occurrences: occurrenceService, events: eventService, prefs: prefService, now: func() time.Time { return time.Now().UTC() }}
}

func (s *Service) HandleCallback(ctx context.Context, data string) error {
	parts := strings.Split(data, ":")
	if len(parts) < 2 {
		return fmt.Errorf("invalid callback data")
	}
	action, target := parts[0], parts[1]
	switch action {
	case "pt":
		return s.pauseGlobal(ctx, target, 24*time.Hour, "Pause everything today")
	case "pw":
		return s.pauseGlobal(ctx, target, 7*24*time.Hour, "Pause this week")
	case "pk":
		return s.pauseTaskByTaskID(ctx, target)
	case "d", "n", "z", "r", "k":
		return s.handleOccurrenceActionByID(ctx, action, target)
	default:
		return fmt.Errorf("unsupported callback action %s", action)
	}
}

func (s *Service) handleOccurrenceActionByID(ctx context.Context, action, occurrenceID string) error {
	user, occurrence, label, err := s.loadOccurrenceContextByOccurrenceID(ctx, occurrenceID)
	if err != nil {
		return err
	}
	userID := user.ID
	action = map[string]string{"d": "done", "n": "notyet", "z": "snooze", "r": "reschedule", "k": "skip"}[action]
	now := s.now()

	switch action {
	case "done":
		occurrence.Status = occurrences.StatusCompleted
		occurrence.CompletedAt = &now
		occurrence.ConsecutiveNoCount = 0
		if _, err := s.occurrences.Update(ctx, occurrence); err != nil {
			return err
		}
		if err := s.sendMessage(ctx, user.TelegramChatID, fmt.Sprintf("Marked done: %s", label), nil); err != nil {
			return err
		}
		return s.logEvent(ctx, userID, occurrenceID, "user_response", "done", map[string]any{"label": label})
	case "notyet":
		occurrence.ConsecutiveNoCount++
		if _, err := s.occurrences.Update(ctx, occurrence); err != nil {
			return err
		}
		if err := s.logEvent(ctx, userID, occurrenceID, "user_response", "not_yet", map[string]any{"count": occurrence.ConsecutiveNoCount}); err != nil {
			return err
		}
		if occurrence.ConsecutiveNoCount >= 2 {
			return s.offerAdaptiveOptions(ctx, user, occurrence, label)
		}
		return s.sendMessage(ctx, user.TelegramChatID, fmt.Sprintf("No problem — we’ll check in again on %s.", label), nil)
	case "snooze":
		snoozeUntil := now.Add(72 * time.Hour)
		occurrence.Status = occurrences.StatusPending
		occurrence.SnoozedUntilAt = &snoozeUntil
		occurrence.ScheduledForDate = snoozeUntil.Format("2006-01-02")
		occurrence.ConsecutiveNoCount = 0
		if _, err := s.occurrences.Update(ctx, occurrence); err != nil {
			return err
		}
		if err := s.sendMessage(ctx, user.TelegramChatID, fmt.Sprintf("Okay — snoozed %s for a few days.", label), nil); err != nil {
			return err
		}
		return s.logEvent(ctx, userID, occurrenceID, "user_response", "snooze", map[string]any{"until": occurrence.ScheduledForDate})
	case "reschedule":
		rescheduledFor := now.Add(7 * 24 * time.Hour)
		occurrence.Status = occurrences.StatusPending
		occurrence.ScheduledForDate = rescheduledFor.Format("2006-01-02")
		occurrence.ConsecutiveNoCount = 0
		if _, err := s.occurrences.Update(ctx, occurrence); err != nil {
			return err
		}
		if err := s.sendMessage(ctx, user.TelegramChatID, fmt.Sprintf("Got it — I’ll bring %s back later.", label), nil); err != nil {
			return err
		}
		return s.logEvent(ctx, userID, occurrenceID, "user_response", "reschedule", map[string]any{"until": occurrence.ScheduledForDate})
	case "skip":
		occurrence.Status = occurrences.StatusSkipped
		occurrence.SkippedAt = &now
		occurrence.ConsecutiveNoCount = 0
		if _, err := s.occurrences.Update(ctx, occurrence); err != nil {
			return err
		}
		if err := s.sendMessage(ctx, user.TelegramChatID, fmt.Sprintf("Okay — we’ll skip %s for now.", label), nil); err != nil {
			return err
		}
		return s.logEvent(ctx, userID, occurrenceID, "user_response", "skip", map[string]any{})
	default:
		return fmt.Errorf("unsupported occurrence action %s", action)
	}
}

func (s *Service) offerAdaptiveOptions(ctx context.Context, user users.User, occurrence occurrences.Occurrence, label string) error {
	offers, err := s.countOccurrenceEvents(ctx, user.ID, occurrence.ID, "reschedule_offered")
	if err != nil {
		return err
	}
	if offers >= 2 {
		now := s.now()
		occurrence.Status = occurrences.StatusSkipped
		occurrence.SkippedAt = &now
		occurrence.ConsecutiveNoCount = 0
		if _, err := s.occurrences.Update(ctx, occurrence); err != nil {
			return err
		}
		if err := s.sendMessage(ctx, user.TelegramChatID, fmt.Sprintf("This hasn’t been landing, so I’m skipping %s for now and we’ll pick it back up later.", label), nil); err != nil {
			return err
		}
		return s.logEvent(ctx, user.ID, occurrence.ID, "system", "skip_after_reschedule_cap", map[string]any{"offers": offers})
	}
	markup := &ntg.ReplyMarkup{InlineKeyboard: [][]ntg.InlineButton{
		{{Text: "Snooze 3 days", CallbackData: ntg.SnoozeAction(user.ID, occurrence.ID)}},
		{{Text: "Reschedule", CallbackData: ntg.RescheduleAction(user.ID, occurrence.ID)}, {Text: "Skip for now", CallbackData: ntg.SkipAction(user.ID, occurrence.ID)}},
	}}
	if err := s.sendMessage(ctx, user.TelegramChatID, fmt.Sprintf("%s hasn’t been landing — want more time, a reschedule, or to snooze it for a few days?", label), markup); err != nil {
		return err
	}
	return s.logEvent(ctx, user.ID, occurrence.ID, "reschedule_offered", "adaptive_checkin", map[string]any{"offers": offers + 1})
}

func (s *Service) pauseGlobal(ctx context.Context, userID string, duration time.Duration, reason string) error {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	start := s.now()
	if _, err := s.prefs.CreatePause(ctx, preferences.Pause{UserID: userID, Scope: "global", Reason: reason, StartsAt: start, EndsAt: start.Add(duration)}); err != nil {
		return err
	}
	if err := s.sendMessage(ctx, user.TelegramChatID, reason+" confirmed.", nil); err != nil {
		return err
	}
	return s.logEvent(ctx, userID, "", "user_response", "pause_global", map[string]any{"reason": reason})
}

func (s *Service) pauseTaskByTaskID(ctx context.Context, taskID string) error {
	lookup, err := s.tasks.GetTaskWithSubtasks(ctx, taskID)
	if err != nil {
		return err
	}
	userID := lookup.Task.UserID
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	start := s.now()
	if _, err := s.prefs.CreatePause(ctx, preferences.Pause{UserID: userID, TaskID: taskID, Scope: "task", Reason: "Pause task", StartsAt: start, EndsAt: start.Add(7 * 24 * time.Hour)}); err != nil {
		return err
	}
	if err := s.sendMessage(ctx, user.TelegramChatID, fmt.Sprintf("Okay — pausing %s for now.", lookup.Task.Name), nil); err != nil {
		return err
	}
	return s.logEvent(ctx, userID, "", "user_response", "pause_task", map[string]any{"task_id": taskID})
}

func (s *Service) loadOccurrenceContextByOccurrenceID(ctx context.Context, occurrenceID string) (users.User, occurrences.Occurrence, string, error) {
	occurrence, err := s.occurrences.GetByID(ctx, occurrenceID)
	if err != nil {
		return users.User{}, occurrences.Occurrence{}, "", err
	}
	user, err := s.users.GetByID(ctx, occurrence.UserID)
	if err != nil {
		return users.User{}, occurrences.Occurrence{}, "", err
	}
	lookup, err := s.tasks.GetTaskWithSubtasks(ctx, occurrence.TaskID)
	if err != nil {
		return users.User{}, occurrences.Occurrence{}, "", err
	}
	label := lookup.Task.Name
	if occurrence.SubtaskID != "" {
		for _, subtask := range lookup.Subtasks {
			if subtask.ID == occurrence.SubtaskID {
				label = lookup.Task.Name + " — " + subtask.Name
				break
			}
		}
	}
	return user, occurrence, label, nil
}

func (s *Service) countOccurrenceEvents(ctx context.Context, userID, occurrenceID, eventType string) (int, error) {
	items, err := s.events.ListByUser(ctx, userID)
	if err != nil {
		return 0, err
	}
	count := 0
	for _, item := range items {
		if item.OccurrenceID == occurrenceID && item.EventType == eventType {
			count++
		}
	}
	return count, nil
}

func (s *Service) sendMessage(ctx context.Context, chatID, text string, markup *ntg.ReplyMarkup) error {
	return s.bot.SendMessage(ctx, ntg.SendMessageRequest{ChatID: chatID, Text: text, ReplyMarkup: markup})
}

func (s *Service) logEvent(ctx context.Context, userID, occurrenceID, eventType, messageType string, payload map[string]any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal event payload: %w", err)
	}
	_, err = s.events.Create(ctx, events.EventLog{UserID: userID, OccurrenceID: occurrenceID, Channel: "telegram", EventType: eventType, MessageType: messageType, PayloadJSON: string(body)})
	return err
}
