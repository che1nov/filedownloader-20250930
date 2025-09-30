package service

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"filedownloader-20240926/internal/domain"
	"filedownloader-20240926/internal/repository"
)

type TaskManager struct {
	tasks   map[string]*domain.Task
	storage *repository.TaskStorage
	mutex   sync.RWMutex
}

// NewTaskManager creates a new task manager instance
func NewTaskManager() *TaskManager {
	tm := &TaskManager{
		tasks:   make(map[string]*domain.Task),
		storage: repository.NewTaskStorage(),
	}

	tm.loadExistingTasks()
	return tm
}

// loadExistingTasks loads all tasks from state on startup
func (tm *TaskManager) loadExistingTasks() {
	tasks, err := tm.storage.LoadAllTasks()
	if err != nil {
		log.Printf("Failed to load existing tasks: %v", err)
		return
	}

	tm.mutex.Lock()
	tm.tasks = tasks
	tm.mutex.Unlock()

	log.Printf("Loaded %d existing tasks", len(tasks))
}

// CreateTask creates a new task
func (tm *TaskManager) CreateTask(urls []string) (*domain.Task, error) {
	taskID := generateTaskID()

	var files []domain.File
	for _, url := range urls {
		files = append(files, domain.File{
			URL:      url,
			Filename: extractFilename(url),
			Status:   domain.StatusPending,
		})
	}

	task := &domain.Task{
		ID:       taskID,
		URLs:     urls,
		Status:   domain.StatusPending,
		Files:    files,
		Progress: 0,
	}
	tm.mutex.Lock()
	tm.tasks[taskID] = task
	tm.mutex.Unlock()

	if err := tm.storage.SaveTask(task); err != nil {
		log.Printf("Failed to save task %s: %v", taskID, err)
		return nil, err
	}

	return task, nil
}

// GetTask returns task by ID
func (tm *TaskManager) GetTask(taskID string) (*domain.Task, bool) {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	task, exists := tm.tasks[taskID]
	return task, exists
}

// UpdateTask updates task
func (tm *TaskManager) UpdateTask(task *domain.Task) error {
	tm.mutex.Lock()
	tm.tasks[task.ID] = task
	tm.mutex.Unlock()

	if err := tm.storage.UpdateTask(task); err != nil {
		log.Printf("Failed to update task %s: %v", task.ID, err)
		return err
	}

	return nil
}

// GetAllTasks returns all tasks
func (tm *TaskManager) GetAllTasks() map[string]*domain.Task {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	result := make(map[string]*domain.Task)
	for k, v := range tm.tasks {
		result[k] = v
	}
	return result
}

// generateTaskID generates task ID
func generateTaskID() string {
	return fmt.Sprintf("task_%d", time.Now().UnixNano())
}

// extractFilename extracts filename from URL
func extractFilename(url string) string {
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		filename := parts[len(parts)-1]
		if filename == "" {
			return "unknown_file"
		}
		return filename
	}
	return "unknown_file"
}
