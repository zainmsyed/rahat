package tasks

import "context"

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateTaskWithSubtasks(ctx context.Context, task Task, subtasks []Subtask) (TaskWithSubtasks, error) {
	createdTask, err := s.repo.CreateTask(ctx, task)
	if err != nil {
		return TaskWithSubtasks{}, err
	}

	createdSubtasks := make([]Subtask, 0, len(subtasks))
	for _, subtask := range subtasks {
		subtask.TaskID = createdTask.ID
		createdSubtask, err := s.repo.CreateSubtask(ctx, subtask)
		if err != nil {
			return TaskWithSubtasks{}, err
		}
		createdSubtasks = append(createdSubtasks, createdSubtask)
	}

	return TaskWithSubtasks{Task: createdTask, Subtasks: createdSubtasks}, nil
}

func (s *Service) UpdateTask(ctx context.Context, task Task) (Task, error) {
	return s.repo.UpdateTask(ctx, task)
}

func (s *Service) DeleteTask(ctx context.Context, taskID string) error {
	return s.repo.DeleteTask(ctx, taskID)
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

func (s *Service) ListTasksByUser(ctx context.Context, userID string) ([]Task, error) {
	return s.repo.ListTasksByUser(ctx, userID)
}

func (s *Service) UpdateSubtask(ctx context.Context, subtask Subtask) (Subtask, error) {
	return s.repo.UpdateSubtask(ctx, subtask)
}

func (s *Service) DeleteSubtask(ctx context.Context, subtaskID string) error {
	return s.repo.DeleteSubtask(ctx, subtaskID)
}

func (s *Service) PauseTask(ctx context.Context, taskID string, paused bool) (Task, error) {
	task, err := s.repo.GetTaskByID(ctx, taskID)
	if err != nil {
		return Task{}, err
	}
	task.IsPaused = paused
	return s.repo.UpdateTask(ctx, task)
}

func (s *Service) ListStarterTaskTemplates(ctx context.Context) ([]StarterTaskTemplate, error) {
	return s.repo.ListStarterTaskTemplates(ctx)
}
