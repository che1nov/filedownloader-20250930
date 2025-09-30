package handler

import (
	"encoding/json"
	"net/http"

	"filedownloader-20240926/internal/domain"
	"filedownloader-20240926/internal/service"
	"filedownloader-20240926/pkg/logger"

	"github.com/gorilla/mux"
)

type TaskHandler struct {
	taskManager *service.TaskManager
	wp          *service.WorkerPool
}

// NewTaskHandler creates a new task handler instance
func NewTaskHandler(tm *service.TaskManager, wp *service.WorkerPool) *TaskHandler {
	return &TaskHandler{taskManager: tm, wp: wp}
}

// CreateTask handles HTTP request to create a new download task
func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Logger.Error("Failed to decode request", "error", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if len(req.URLs) == 0 {
		logger.Logger.Warn("Empty URLs array")
		http.Error(w, "URLs array cannot be empty", http.StatusBadRequest)
		return
	}

	task, err := h.taskManager.CreateTask(req.URLs)
	if err != nil {
		logger.Logger.Error("Failed to create task", "error", err)
		http.Error(w, "Failed to create task", http.StatusInternalServerError)
		return
	}

	if h.wp != nil {
		h.wp.ProcessFiles(task.ID, task.Files)
	}

	logger.Logger.Info("Created task", "task_id", task.ID, "urls_count", len(req.URLs))

	resp := domain.CreateTaskResponse{TaskID: task.ID}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetTaskStatus handles HTTP request to get task status
func (h *TaskHandler) GetTaskStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["id"]

	task, exists := h.taskManager.GetTask(taskID)
	if !exists {
		logger.Logger.Warn("Task not found", "task_id", taskID)
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	resp := domain.TaskStatusResponse{
		ID:       task.ID,
		Status:   string(task.Status),
		Progress: task.Progress,
		Files:    task.Files,
	}

	logger.Logger.Debug("Returning task status", "task_id", taskID, "status", task.Status)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
