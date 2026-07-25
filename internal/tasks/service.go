package tasks

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateTaskWithSubtasks(ctx context.Context, task Task, subtasks []Subtask) (TaskWithSubtasks, error) {
	tx, err := s.repo.db.BeginTx(ctx, nil)
	if err != nil {
		return TaskWithSubtasks{}, fmt.Errorf("begin task create transaction: %w", err)
	}

	createdTask, err := createTask(ctx, tx, task)
	if err != nil {
		_ = tx.Rollback()
		return TaskWithSubtasks{}, err
	}

	createdSubtasks := make([]Subtask, 0, len(subtasks))
	for _, subtask := range subtasks {
		subtask.TaskID = createdTask.ID
		createdSubtask, err := createSubtask(ctx, tx, subtask)
		if err != nil {
			_ = tx.Rollback()
			return TaskWithSubtasks{}, err
		}
		createdSubtasks = append(createdSubtasks, createdSubtask)
	}

	if err := tx.Commit(); err != nil {
		return TaskWithSubtasks{}, fmt.Errorf("commit task create transaction: %w", err)
	}

	return TaskWithSubtasks{Task: createdTask, Subtasks: createdSubtasks}, nil
}

func (s *Service) UpdateTask(ctx context.Context, task Task) (Task, error) {
	return s.repo.UpdateTask(ctx, task)
}

func (s *Service) ReplaceTaskWithSubtasks(ctx context.Context, task Task, subtasks []Subtask) (TaskWithSubtasks, error) {
	tx, err := s.repo.db.BeginTx(ctx, nil)
	if err != nil {
		return TaskWithSubtasks{}, fmt.Errorf("begin task replace transaction: %w", err)
	}

	creating := task.ID == ""
	if creating {
		createdTask, err := createTask(ctx, tx, task)
		if err != nil {
			_ = tx.Rollback()
			return TaskWithSubtasks{}, err
		}
		task = createdTask
	} else {
		updatedTask, err := updateTask(ctx, tx, task)
		if err != nil {
			_ = tx.Rollback()
			return TaskWithSubtasks{}, err
		}
		task = updatedTask

		existing, err := listSubtasksByTask(ctx, tx, task.ID)
		if err != nil {
			_ = tx.Rollback()
			return TaskWithSubtasks{}, err
		}
		existingByID := make(map[string]Subtask, len(existing))
		for _, subtask := range existing {
			existingByID[subtask.ID] = subtask
		}
		retained := make(map[string]bool, len(subtasks))
		for _, subtask := range subtasks {
			if subtask.ID == "" {
				continue
			}
			if retained[subtask.ID] {
				_ = tx.Rollback()
				return TaskWithSubtasks{}, fmt.Errorf("subtask %s appears more than once", subtask.ID)
			}
			if _, ok := existingByID[subtask.ID]; !ok {
				_ = tx.Rollback()
				return TaskWithSubtasks{}, fmt.Errorf("subtask %s does not belong to task %s", subtask.ID, task.ID)
			}
			retained[subtask.ID] = true
		}

		temporaryOrderBase := time.Now().UTC().UnixNano()
		for index, existingSubtask := range existing {
			temporaryOrder := temporaryOrderBase + int64(index)
			if retained[existingSubtask.ID] {
				if err := setSubtaskStepOrder(ctx, tx, existingSubtask.ID, temporaryOrder); err != nil {
					_ = tx.Rollback()
					return TaskWithSubtasks{}, err
				}
				continue
			}
			if err := archiveSubtask(ctx, tx, existingSubtask.ID, -temporaryOrder); err != nil {
				_ = tx.Rollback()
				return TaskWithSubtasks{}, err
			}
		}
	}

	for _, subtask := range subtasks {
		subtask.TaskID = task.ID
		if creating {
			subtask.ID = ""
		}
		if subtask.ID == "" {
			if _, err := createSubtask(ctx, tx, subtask); err != nil {
				_ = tx.Rollback()
				return TaskWithSubtasks{}, err
			}
			continue
		}
		if _, err := updateSubtask(ctx, tx, subtask); err != nil {
			_ = tx.Rollback()
			return TaskWithSubtasks{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return TaskWithSubtasks{}, fmt.Errorf("commit task replace transaction: %w", err)
	}
	return s.GetTaskWithSubtasks(ctx, task.ID)
}

func (s *Service) DeleteTask(ctx context.Context, taskID string) error {
	return s.repo.DeleteTask(ctx, taskID)
}

func (s *Service) ArchiveTask(ctx context.Context, taskID string) error {
	return s.repo.ArchiveTask(ctx, taskID)
}

func (s *Service) GetTaskWithSubtasks(ctx context.Context, taskID string) (TaskWithSubtasks, error) {
	task, err := s.repo.GetTaskByID(ctx, taskID)
	if err != nil {
		return TaskWithSubtasks{}, err
	}
	subtasks, err := s.repo.ListSubtasksByTask(ctx, taskID)
	if err != nil {
		return TaskWithSubtasks{}, err
	}
	return TaskWithSubtasks{Task: task, Subtasks: subtasks}, nil
}

func (s *Service) ListTaskWithSubtasksByUser(ctx context.Context, userID string) ([]TaskWithSubtasks, error) {
	return s.repo.ListTaskWithSubtasksByUser(ctx, userID)
}

func (s *Service) ListTasksByUser(ctx context.Context, userID string) ([]Task, error) {
	return s.repo.ListTasksByUser(ctx, userID)
}

func (s *Service) ListTaskWithSubtasksByUserIncludingArchived(ctx context.Context, userID string) ([]TaskWithSubtasks, error) {
	tasksForUser, err := s.repo.ListTasksByUserIncludingArchived(ctx, userID)
	if err != nil {
		return nil, err
	}
	result := make([]TaskWithSubtasks, 0, len(tasksForUser))
	for _, task := range tasksForUser {
		subtasks, err := s.repo.ListSubtasksByTask(ctx, task.ID)
		if err != nil {
			return nil, err
		}
		result = append(result, TaskWithSubtasks{Task: task, Subtasks: subtasks})
	}
	return result, nil
}

func (s *Service) UpdateSubtask(ctx context.Context, subtask Subtask) (Subtask, error) {
	return s.repo.UpdateSubtask(ctx, subtask)
}

func (s *Service) DeleteSubtask(ctx context.Context, subtaskID string) error {
	return s.repo.DeleteSubtask(ctx, subtaskID)
}

func (s *Service) PauseTask(ctx context.Context, taskID string, paused bool) (Task, error) {
	return s.repo.SetTaskPaused(ctx, taskID, paused)
}

func (s *Service) ListStarterTaskTemplates(ctx context.Context) ([]StarterTaskTemplate, error) {
	return s.repo.ListStarterTaskTemplates(ctx)
}

func (s *Service) CreateTaskFromStarterTemplate(ctx context.Context, userID, templateID string) (TaskWithSubtasks, error) {
	templates, err := s.repo.ListStarterTaskTemplates(ctx)
	if err != nil {
		return TaskWithSubtasks{}, err
	}
	for _, tmpl := range templates {
		if tmpl.ID != templateID {
			continue
		}
		task := Task{
			UserID:              userID,
			Name:                tmpl.Name,
			Description:         tmpl.Description,
			DurationMinutes:     tmpl.DurationMinutes,
			CadenceType:         tmpl.CadenceType,
			CadenceValue:        tmpl.CadenceValue,
			Priority:            tmpl.Priority,
			TimeOfDayPreference: tmpl.TimeOfDayPreference,
			IsMultistep:         tmpl.IsMultistep,
		}
		subtasks := make([]Subtask, 0, len(tmpl.Subtasks))
		for _, starter := range tmpl.Subtasks {
			subtasks = append(subtasks, Subtask{StepOrder: starter.StepOrder, Name: starter.Name, DurationMinutes: starter.DurationMinutes, TimeOfDayPreference: starter.TimeOfDayPreference, DependencyType: starter.DependencyType, GapRule: SubtaskGapRule{MinGapAfterPreviousMinutes: starter.MinGapAfterPreviousMinutes}})
		}
		return s.CreateTaskWithSubtasks(ctx, task, subtasks)
	}
	return TaskWithSubtasks{}, errors.New("starter task template not found")
}
