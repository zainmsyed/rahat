package tasks

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"time"

	"github.com/rahat/rahat/internal/store"
)

type dbExecutor interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateTask(ctx context.Context, task Task) (Task, error) {
	return createTask(ctx, r.db, task)
}

func createTask(ctx context.Context, exec dbExecutor, task Task) (Task, error) {
	now := time.Now().UTC()
	if task.ID == "" {
		task.ID = store.NewID()
	}
	if task.CreatedAt.IsZero() {
		task.CreatedAt = now
	}
	if task.UpdatedAt.IsZero() {
		task.UpdatedAt = now
	}

	if _, err := exec.ExecContext(ctx, `
		INSERT INTO tasks (id, user_id, name, description, duration_minutes, cadence_type, cadence_value, priority, time_of_day_preference, is_multistep, is_paused, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, task.ID, task.UserID, task.Name, task.Description, task.DurationMinutes, task.CadenceType, task.CadenceValue, task.Priority, task.TimeOfDayPreference, task.IsMultistep, task.IsPaused, store.FormatTime(task.CreatedAt), store.FormatTime(task.UpdatedAt)); err != nil {
		return Task{}, fmt.Errorf("create task: %w", err)
	}

	return task, nil
}

func (r *Repository) UpdateTask(ctx context.Context, task Task) (Task, error) {
	updated, err := updateTask(ctx, r.db, task)
	if err != nil {
		return Task{}, err
	}
	return r.GetTaskByID(ctx, updated.ID)
}

func updateTask(ctx context.Context, exec dbExecutor, task Task) (Task, error) {
	task.UpdatedAt = time.Now().UTC()
	if _, err := exec.ExecContext(ctx, `
		UPDATE tasks
		SET name = ?, description = ?, duration_minutes = ?, cadence_type = ?, cadence_value = ?, priority = ?, time_of_day_preference = ?, is_multistep = ?, is_paused = ?, updated_at = ?
		WHERE id = ?
	`, task.Name, task.Description, task.DurationMinutes, task.CadenceType, task.CadenceValue, task.Priority, task.TimeOfDayPreference, task.IsMultistep, task.IsPaused, store.FormatTime(task.UpdatedAt), task.ID); err != nil {
		return Task{}, fmt.Errorf("update task %s: %w", task.ID, err)
	}
	return task, nil
}

func (r *Repository) GetTaskByID(ctx context.Context, id string) (Task, error) {
	var task Task
	var createdAt string
	var updatedAt string
	var archivedAt sql.NullString
	if err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, name, description, duration_minutes, cadence_type, cadence_value, priority, time_of_day_preference, is_multistep, is_paused, archived_at, created_at, updated_at
		FROM tasks
		WHERE id = ?
	`, id).Scan(&task.ID, &task.UserID, &task.Name, &task.Description, &task.DurationMinutes, &task.CadenceType, &task.CadenceValue, &task.Priority, &task.TimeOfDayPreference, &task.IsMultistep, &task.IsPaused, &archivedAt, &createdAt, &updatedAt); err != nil {
		return Task{}, fmt.Errorf("get task %s: %w", id, err)
	}

	if err := parseTaskTimes(&task, createdAt, updatedAt, archivedAt); err != nil {
		return Task{}, err
	}

	return task, nil
}

func (r *Repository) DeleteTask(ctx context.Context, id string) error {
	return r.ArchiveTask(ctx, id)
}

func (r *Repository) ArchiveTask(ctx context.Context, id string) error {
	now := store.FormatTime(time.Now().UTC())
	if _, err := r.db.ExecContext(ctx, `UPDATE tasks SET archived_at = COALESCE(archived_at, ?), updated_at = ? WHERE id = ?`, now, now, id); err != nil {
		return fmt.Errorf("archive task %s: %w", id, err)
	}
	return nil
}

func (r *Repository) SetTaskPaused(ctx context.Context, id string, paused bool) (Task, error) {
	now := store.FormatTime(time.Now().UTC())
	if _, err := r.db.ExecContext(ctx, `UPDATE tasks SET is_paused = ?, updated_at = ? WHERE id = ? AND archived_at IS NULL`, paused, now, id); err != nil {
		return Task{}, fmt.Errorf("set task paused %s: %w", id, err)
	}
	return r.GetTaskByID(ctx, id)
}

func (r *Repository) ListTaskWithSubtasksByUser(ctx context.Context, userID string) ([]TaskWithSubtasks, error) {
	tasksForUser, err := r.ListTasksByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]TaskWithSubtasks, 0, len(tasksForUser))
	for _, task := range tasksForUser {
		subtasks, err := r.ListSubtasksByTask(ctx, task.ID)
		if err != nil {
			return nil, err
		}
		result = append(result, TaskWithSubtasks{Task: task, Subtasks: subtasks})
	}
	return result, nil
}

func (r *Repository) ListTasksByUser(ctx context.Context, userID string) ([]Task, error) {
	return r.listTasksByUser(ctx, userID, false)
}

func (r *Repository) ListTasksByUserIncludingArchived(ctx context.Context, userID string) ([]Task, error) {
	return r.listTasksByUser(ctx, userID, true)
}

func (r *Repository) listTasksByUser(ctx context.Context, userID string, includeArchived bool) ([]Task, error) {
	where := `WHERE user_id = ? AND archived_at IS NULL`
	if includeArchived {
		where = `WHERE user_id = ?`
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, name, description, duration_minutes, cadence_type, cadence_value, priority, time_of_day_preference, is_multistep, is_paused, archived_at, created_at, updated_at
		FROM tasks
		`+where+`
		ORDER BY archived_at IS NOT NULL, created_at, id
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list tasks for user %s: %w", userID, err)
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		var createdAt string
		var updatedAt string
		var archivedAt sql.NullString
		if err := rows.Scan(&task.ID, &task.UserID, &task.Name, &task.Description, &task.DurationMinutes, &task.CadenceType, &task.CadenceValue, &task.Priority, &task.TimeOfDayPreference, &task.IsMultistep, &task.IsPaused, &archivedAt, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("scan task row: %w", err)
		}
		if err := parseTaskTimes(&task, createdAt, updatedAt, archivedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	return tasks, rows.Err()
}

func parseTaskTimes(task *Task, createdAt, updatedAt string, archivedAt sql.NullString) error {
	var err error
	task.CreatedAt, err = store.ParseTime(createdAt)
	if err != nil {
		return fmt.Errorf("parse task created_at: %w", err)
	}
	task.UpdatedAt, err = store.ParseTime(updatedAt)
	if err != nil {
		return fmt.Errorf("parse task updated_at: %w", err)
	}
	if archivedAt.Valid && archivedAt.String != "" {
		parsed, err := store.ParseTime(archivedAt.String)
		if err != nil {
			return fmt.Errorf("parse task archived_at: %w", err)
		}
		task.ArchivedAt = &parsed
	}
	return nil
}

func (r *Repository) CreateSubtask(ctx context.Context, subtask Subtask) (Subtask, error) {
	return createSubtask(ctx, r.db, subtask)
}

func createSubtask(ctx context.Context, exec dbExecutor, subtask Subtask) (Subtask, error) {
	if subtask.ID == "" {
		subtask.ID = store.NewID()
	}
	if subtask.CreatedAt.IsZero() {
		subtask.CreatedAt = time.Now().UTC()
	}
	subtask.DependencyType = normalizeSubtaskDependency(subtask.DependencyType)

	if _, err := exec.ExecContext(ctx, `
		INSERT INTO subtasks (id, task_id, step_order, name, duration_minutes, time_of_day_preference, dependency_type, min_gap_after_previous_minutes, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, subtask.ID, subtask.TaskID, subtask.StepOrder, subtask.Name, subtask.DurationMinutes, subtask.TimeOfDayPreference, subtask.DependencyType, subtask.GapRule.MinGapAfterPreviousMinutes, store.FormatTime(subtask.CreatedAt)); err != nil {
		return Subtask{}, fmt.Errorf("create subtask: %w", err)
	}

	return subtask, nil
}

func (r *Repository) GetSubtaskByID(ctx context.Context, id string) (Subtask, error) {
	var subtask Subtask
	var createdAt string
	if err := r.db.QueryRowContext(ctx, `
		SELECT id, task_id, step_order, name, duration_minutes, time_of_day_preference, dependency_type, min_gap_after_previous_minutes, created_at
		FROM subtasks
		WHERE id = ?
	`, id).Scan(&subtask.ID, &subtask.TaskID, &subtask.StepOrder, &subtask.Name, &subtask.DurationMinutes, &subtask.TimeOfDayPreference, &subtask.DependencyType, &subtask.GapRule.MinGapAfterPreviousMinutes, &createdAt); err != nil {
		return Subtask{}, fmt.Errorf("get subtask %s: %w", id, err)
	}
	var err error
	subtask.CreatedAt, err = store.ParseTime(createdAt)
	if err != nil {
		return Subtask{}, fmt.Errorf("parse subtask created_at: %w", err)
	}
	return subtask, nil
}

func (r *Repository) UpdateSubtask(ctx context.Context, subtask Subtask) (Subtask, error) {
	subtask.DependencyType = normalizeSubtaskDependency(subtask.DependencyType)
	if _, err := r.db.ExecContext(ctx, `
		UPDATE subtasks
		SET step_order = ?, name = ?, duration_minutes = ?, time_of_day_preference = ?, dependency_type = ?, min_gap_after_previous_minutes = ?
		WHERE id = ?
	`, subtask.StepOrder, subtask.Name, subtask.DurationMinutes, subtask.TimeOfDayPreference, subtask.DependencyType, subtask.GapRule.MinGapAfterPreviousMinutes, subtask.ID); err != nil {
		return Subtask{}, fmt.Errorf("update subtask %s: %w", subtask.ID, err)
	}
	return r.GetSubtaskByID(ctx, subtask.ID)
}

func (r *Repository) DeleteSubtask(ctx context.Context, id string) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM subtasks WHERE id = ?`, id); err != nil {
		return fmt.Errorf("delete subtask %s: %w", id, err)
	}
	return nil
}

func deleteSubtasksByTask(ctx context.Context, exec dbExecutor, taskID string) error {
	if _, err := exec.ExecContext(ctx, `DELETE FROM subtasks WHERE task_id = ?`, taskID); err != nil {
		return fmt.Errorf("delete subtasks for task %s: %w", taskID, err)
	}
	return nil
}

func (r *Repository) ListSubtasksByTask(ctx context.Context, taskID string) ([]Subtask, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, task_id, step_order, name, duration_minutes, time_of_day_preference, dependency_type, min_gap_after_previous_minutes, created_at
		FROM subtasks
		WHERE task_id = ?
		ORDER BY step_order, id
	`, taskID)
	if err != nil {
		return nil, fmt.Errorf("list subtasks for task %s: %w", taskID, err)
	}
	defer rows.Close()

	var subtasks []Subtask
	for rows.Next() {
		var subtask Subtask
		var createdAt string
		if err := rows.Scan(&subtask.ID, &subtask.TaskID, &subtask.StepOrder, &subtask.Name, &subtask.DurationMinutes, &subtask.TimeOfDayPreference, &subtask.DependencyType, &subtask.GapRule.MinGapAfterPreviousMinutes, &createdAt); err != nil {
			return nil, fmt.Errorf("scan subtask row: %w", err)
		}
		if subtask.CreatedAt, err = store.ParseTime(createdAt); err != nil {
			return nil, fmt.Errorf("parse subtask created_at: %w", err)
		}
		subtasks = append(subtasks, subtask)
	}

	return subtasks, rows.Err()
}

func normalizeSubtaskDependency(value SubtaskDependencyType) SubtaskDependencyType {
	if value == SubtaskDependencySoftFollowup {
		return SubtaskDependencySoftFollowup
	}
	return SubtaskDependencyRequiredSameDay
}

func (r *Repository) ListStarterTaskTemplates(ctx context.Context) ([]StarterTaskTemplate, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, slug, name, description, duration_minutes, cadence_type, cadence_value, priority, time_of_day_preference, is_multistep, sort_order
		FROM starter_task_templates
		ORDER BY sort_order, slug
	`)
	if err != nil {
		return nil, fmt.Errorf("list starter task templates: %w", err)
	}
	defer rows.Close()

	var templates []StarterTaskTemplate
	for rows.Next() {
		var tmpl StarterTaskTemplate
		if err := rows.Scan(&tmpl.ID, &tmpl.Slug, &tmpl.Name, &tmpl.Description, &tmpl.DurationMinutes, &tmpl.CadenceType, &tmpl.CadenceValue, &tmpl.Priority, &tmpl.TimeOfDayPreference, &tmpl.IsMultistep, &tmpl.SortOrder); err != nil {
			return nil, fmt.Errorf("scan starter task template: %w", err)
		}
		templates = append(templates, tmpl)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	subtaskRows, err := r.db.QueryContext(ctx, `
		SELECT id, starter_task_template_id, step_order, name, duration_minutes, time_of_day_preference, dependency_type, min_gap_after_previous_minutes
		FROM starter_subtask_templates
		ORDER BY starter_task_template_id, step_order
	`)
	if err != nil {
		return nil, fmt.Errorf("list starter subtask templates: %w", err)
	}
	defer subtaskRows.Close()

	subtasksByTemplate := map[string][]StarterSubtaskTemplate{}
	for subtaskRows.Next() {
		var subtask StarterSubtaskTemplate
		if err := subtaskRows.Scan(&subtask.ID, &subtask.StarterTaskTemplateID, &subtask.StepOrder, &subtask.Name, &subtask.DurationMinutes, &subtask.TimeOfDayPreference, &subtask.DependencyType, &subtask.MinGapAfterPreviousMinutes); err != nil {
			return nil, fmt.Errorf("scan starter subtask template: %w", err)
		}
		subtasksByTemplate[subtask.StarterTaskTemplateID] = append(subtasksByTemplate[subtask.StarterTaskTemplateID], subtask)
	}
	if err := subtaskRows.Err(); err != nil {
		return nil, err
	}

	for i := range templates {
		templates[i].Subtasks = subtasksByTemplate[templates[i].ID]
		sort.SliceStable(templates[i].Subtasks, func(a, b int) bool {
			return templates[i].Subtasks[a].StepOrder < templates[i].Subtasks[b].StepOrder
		})
	}

	return templates, nil
}
