package service

import (
	"testing"

	"filedownloader-20240926/internal/domain"
)

// TestTaskManagerCreateTask tests task creation with various inputs
func TestTaskManagerCreateTask(t *testing.T) {
	tests := []struct {
		name        string
		urls        []string
		expectError bool
	}{
		{
			name:        "single URL",
			urls:        []string{"http://example.com/file1.txt"},
			expectError: false,
		},
		{
			name:        "multiple URLs",
			urls:        []string{"http://example.com/file1.txt", "http://example.com/file2.txt"},
			expectError: false,
		},
		{
			name:        "empty URLs",
			urls:        []string{},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm := NewTaskManager()
			task, err := tm.CreateTask(tt.urls)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if task == nil {
				t.Errorf("expected task but got nil")
				return
			}

			if len(task.URLs) != len(tt.urls) {
				t.Errorf("expected %d URLs, got %d", len(tt.urls), len(task.URLs))
			}

			if task.Status != domain.StatusPending {
				t.Errorf("expected status %s, got %s", domain.StatusPending, task.Status)
			}
		})
	}
}

// TestTaskManagerGetTask tests task retrieval
func TestTaskManagerGetTask(t *testing.T) {
	tests := []struct {
		name   string
		taskID string
		exists bool
		setup  func(*TaskManager) string
	}{
		{
			name:   "existing task",
			exists: true,
			setup: func(tm *TaskManager) string {
				task, _ := tm.CreateTask([]string{"http://example.com/test.txt"})
				return task.ID
			},
		},
		{
			name:   "non-existing task",
			exists: false,
			setup: func(tm *TaskManager) string {
				return "non-existing-task-id"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm := NewTaskManager()
			taskID := tt.setup(tm)

			task, exists := tm.GetTask(taskID)

			if exists != tt.exists {
				t.Errorf("expected exists=%v, got %v", tt.exists, exists)
			}

			if tt.exists && task == nil {
				t.Errorf("expected task but got nil")
			}

			if !tt.exists && task != nil {
				t.Errorf("expected nil task but got %v", task)
			}
		})
	}
}

// TestTaskManagerUpdateTask tests task updates
func TestTaskManagerUpdateTask(t *testing.T) {
	tests := []struct {
		name        string
		modifyTask  func(*domain.Task)
		expectError bool
	}{
		{
			name:        "update status",
			expectError: false,
			modifyTask: func(task *domain.Task) {
				task.Status = domain.StatusCompleted
			},
		},
		{
			name:        "update progress",
			expectError: false,
			modifyTask: func(task *domain.Task) {
				task.Progress = 50
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm := NewTaskManager()
			task, err := tm.CreateTask([]string{"http://example.com/test.txt"})
			if err != nil {
				t.Fatalf("failed to create task: %v", err)
			}

			tt.modifyTask(task)

			err = tm.UpdateTask(task)
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Verify the update
			updatedTask, exists := tm.GetTask(task.ID)
			if !exists {
				t.Errorf("task not found after update")
				return
			}

			if updatedTask.Status != task.Status {
				t.Errorf("expected status %s, got %s", task.Status, updatedTask.Status)
			}
		})
	}
}
