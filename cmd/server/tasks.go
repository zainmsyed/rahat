package main

import (
	"errors"
	"net/http"
	"strings"

	taskpkg "github.com/rahat/rahat/internal/tasks"
)

type taskManagementHandler struct {
	auth  *authHandler
	tasks *taskpkg.Service
}

type taskPauseRequest struct {
	Paused bool `json:"paused"`
}

func (h *taskManagementHandler) register(mux *http.ServeMux) {
	mux.HandleFunc("GET /tasks", h.handleList)
	mux.HandleFunc("POST /tasks", h.handleCreate)
	mux.HandleFunc("PUT /tasks/{taskID}", h.handleUpdate)
	mux.HandleFunc("POST /tasks/{taskID}/pause", h.handlePause)
	mux.HandleFunc("DELETE /tasks/{taskID}", h.handleArchive)
}

func (h *taskManagementHandler) handleList(w http.ResponseWriter, r *http.Request) {
	current, ok := h.auth.requireAuthenticatedUser(w, r)
	if !ok { return }
	tasks, err := h.tasks.ListTaskWithSubtasksByUserIncludingArchived(r.Context(), current.User.ID)
	if err != nil { http.Error(w, err.Error(), http.StatusBadRequest); return }
	writeJSON(w, http.StatusOK, toTaskResponses(tasks))
}

func (h *taskManagementHandler) handleCreate(w http.ResponseWriter, r *http.Request) {
	current, ok := h.auth.requireAuthenticatedUser(w, r)
	if !ok { return }
	var req onboardingTaskRequest
	if err := decodeJSON(r, &req); err != nil { http.Error(w, err.Error(), http.StatusBadRequest); return }
	task, subtasks, err := validateTaskRequest(current.User.ID, req)
	if err != nil { http.Error(w, err.Error(), http.StatusBadRequest); return }
	created, err := h.tasks.ReplaceTaskWithSubtasks(r.Context(), task, subtasks)
	if err != nil { http.Error(w, err.Error(), http.StatusBadRequest); return }
	writeJSON(w, http.StatusCreated, toTaskResponse(created))
}

func (h *taskManagementHandler) handleUpdate(w http.ResponseWriter, r *http.Request) {
	current, ok := h.auth.requireAuthenticatedUser(w, r)
	if !ok { return }
	taskID := strings.TrimSpace(r.PathValue("taskID"))
	existing, err := h.requireOwnedTask(r, current.User.ID, taskID)
	if err != nil { http.Error(w, err.Error(), http.StatusBadRequest); return }
	if existing.Task.ArchivedAt != nil { http.Error(w, "removed tasks cannot be edited", http.StatusBadRequest); return }
	var req onboardingTaskRequest
	if err := decodeJSON(r, &req); err != nil { http.Error(w, err.Error(), http.StatusBadRequest); return }
	task, subtasks, err := validateTaskRequest(current.User.ID, req)
	if err != nil { http.Error(w, err.Error(), http.StatusBadRequest); return }
	task.ID = taskID
	task.IsPaused = existing.Task.IsPaused
	updated, err := h.tasks.ReplaceTaskWithSubtasks(r.Context(), task, subtasks)
	if err != nil { http.Error(w, err.Error(), http.StatusBadRequest); return }
	writeJSON(w, http.StatusOK, toTaskResponse(updated))
}

func (h *taskManagementHandler) handlePause(w http.ResponseWriter, r *http.Request) {
	current, ok := h.auth.requireAuthenticatedUser(w, r)
	if !ok { return }
	taskID := strings.TrimSpace(r.PathValue("taskID"))
	existing, err := h.requireOwnedTask(r, current.User.ID, taskID)
	if err != nil { http.Error(w, err.Error(), http.StatusBadRequest); return }
	if existing.Task.ArchivedAt != nil { http.Error(w, "removed tasks cannot be paused or resumed", http.StatusBadRequest); return }
	var req taskPauseRequest
	if err := decodeJSON(r, &req); err != nil { http.Error(w, err.Error(), http.StatusBadRequest); return }
	if _, err := h.tasks.PauseTask(r.Context(), taskID, req.Paused); err != nil { http.Error(w, err.Error(), http.StatusBadRequest); return }
	updated, err := h.tasks.GetTaskWithSubtasks(r.Context(), taskID)
	if err != nil { http.Error(w, err.Error(), http.StatusBadRequest); return }
	writeJSON(w, http.StatusOK, toTaskResponse(updated))
}

func (h *taskManagementHandler) handleArchive(w http.ResponseWriter, r *http.Request) {
	current, ok := h.auth.requireAuthenticatedUser(w, r)
	if !ok { return }
	taskID := strings.TrimSpace(r.PathValue("taskID"))
	if _, err := h.requireOwnedTask(r, current.User.ID, taskID); err != nil { http.Error(w, err.Error(), http.StatusBadRequest); return }
	if err := h.tasks.ArchiveTask(r.Context(), taskID); err != nil { http.Error(w, err.Error(), http.StatusBadRequest); return }
	w.WriteHeader(http.StatusNoContent)
}

func (h *taskManagementHandler) requireOwnedTask(r *http.Request, userID, taskID string) (taskpkg.TaskWithSubtasks, error) {
	if taskID == "" { return taskpkg.TaskWithSubtasks{}, errors.New("missing task id") }
	taskDef, err := h.tasks.GetTaskWithSubtasks(r.Context(), taskID)
	if err != nil { return taskpkg.TaskWithSubtasks{}, err }
	if taskDef.Task.UserID != userID { return taskpkg.TaskWithSubtasks{}, errors.New("task does not belong to this user") }
	return taskDef, nil
}
