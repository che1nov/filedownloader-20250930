package repository

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"filedownloader-20240926/internal/domain"
)

type TaskStorage struct {
	stateDir string
	mutex    sync.RWMutex
}

// NewTaskStorage creates a new task storage instance
func NewTaskStorage() *TaskStorage {
	wd, err := os.Getwd()
	if err != nil {
		return &TaskStorage{
			stateDir: "./state",
		}
	}

	projectRoot := wd
	for {
		if _, err := os.Stat(filepath.Join(projectRoot, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(projectRoot)
		if parent == projectRoot {
			projectRoot = wd
			break
		}
		projectRoot = parent
	}

	return &TaskStorage{
		stateDir: filepath.Join(projectRoot, "state"),
	}
}

// SaveTask saves task to JSON file
func (ts *TaskStorage) SaveTask(task *domain.Task) error {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	if err := os.MkdirAll(ts.stateDir, 0755); err != nil {
		return fmt.Errorf("failed to create state dir: %w", err)
	}

	filePath := filepath.Join(ts.stateDir, task.ID+".json")

	data, err := json.MarshalIndent(task, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write task file: %w", err)
	}

	fmt.Printf("DEBUG: Saved task %s to %s\n", task.ID, filePath)
	return nil
}

// LoadTask loads task from JSON file
func (ts *TaskStorage) LoadTask(taskID string) (*domain.Task, error) {
	ts.mutex.RLock()
	defer ts.mutex.RUnlock()

	filePath := filepath.Join(ts.stateDir, taskID+".json")

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read task file: %w", err)
	}

	var task domain.Task
	if err := json.Unmarshal(data, &task); err != nil {
		return nil, fmt.Errorf("failed to unmarshal task: %w", err)
	}

	fmt.Printf("DEBUG: Loaded task %s from %s\n", taskID, filePath)
	return &task, nil
}

// LoadAllTasks loads all tasks from state directory
func (ts *TaskStorage) LoadAllTasks() (map[string]*domain.Task, error) {
	ts.mutex.RLock()
	defer ts.mutex.RUnlock()

	tasks := make(map[string]*domain.Task)

	entries, err := os.ReadDir(ts.stateDir)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("DEBUG: State dir %s does not exist, starting fresh\n", ts.stateDir)
			return tasks, nil
		}
		return nil, fmt.Errorf("failed to read state dir: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		name := entry.Name()
		taskID := name[:len(name)-5]

		task, err := ts.LoadTask(taskID)
		if err != nil {
			log.Printf("WARNING: Failed to load task %s: %v", taskID, err)
			continue
		}

		tasks[taskID] = task
	}

	fmt.Printf("DEBUG: Loaded %d tasks from state\n", len(tasks))
	return tasks, nil
}

// DeleteTask deletes task file
func (ts *TaskStorage) DeleteTask(taskID string) error {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	filePath := filepath.Join(ts.stateDir, taskID+".json")

	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to delete task file: %w", err)
	}

	fmt.Printf("DEBUG: Deleted task %s\n", taskID)
	return nil
}

// UpdateTask updates existing task
func (ts *TaskStorage) UpdateTask(task *domain.Task) error {
	return ts.SaveTask(task)
}
