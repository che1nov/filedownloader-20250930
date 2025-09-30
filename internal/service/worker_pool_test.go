package service

import (
	"context"
	"testing"
	"time"

	"filedownloader-20240926/internal/domain"
)

// TestWorkerPoolCreation tests worker pool creation with different configurations
func TestWorkerPoolCreation(t *testing.T) {
	tests := []struct {
		name    string
		workers int
	}{
		{
			name:    "single worker",
			workers: 1,
		},
		{
			name:    "multiple workers",
			workers: 3,
		},
		{
			name:    "many workers",
			workers: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm := NewTaskManager()
			wp := NewWorkerPool(tt.workers, tm)

			if wp == nil {
				t.Errorf("expected worker pool but got nil")
				return
			}

			if wp.workers != tt.workers {
				t.Errorf("expected %d workers, got %d", tt.workers, wp.workers)
			}

			if wp.downloader == nil {
				t.Errorf("expected downloader but got nil")
			}

			if wp.taskChan == nil {
				t.Errorf("expected task channel but got nil")
			}
		})
	}
}

// TestWorkerPoolStartStop tests worker pool start and stop functionality
func TestWorkerPoolStartStop(t *testing.T) {
	tests := []struct {
		name    string
		workers int
	}{
		{
			name:    "start and stop single worker",
			workers: 1,
		},
		{
			name:    "start and stop multiple workers",
			workers: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm := NewTaskManager()
			wp := NewWorkerPool(tt.workers, tm)

			wp.Start()

			time.Sleep(10 * time.Millisecond)

			wp.Stop()

			wp.Start()
			time.Sleep(10 * time.Millisecond)
			wp.Stop()
		})
	}
}

// TestWorkerPoolAddTask tests adding tasks to the worker pool
func TestWorkerPoolAddTask(t *testing.T) {
	tests := []struct {
		name     string
		files    []domain.File
		expected int
	}{
		{
			name: "single file",
			files: []domain.File{
				{URL: "http://example.com/file1.txt", Filename: "file1.txt", Status: domain.StatusPending},
			},
			expected: 1,
		},
		{
			name: "multiple files",
			files: []domain.File{
				{URL: "http://example.com/file1.txt", Filename: "file1.txt", Status: domain.StatusPending},
				{URL: "http://example.com/file2.txt", Filename: "file2.txt", Status: domain.StatusPending},
				{URL: "http://example.com/file3.txt", Filename: "file3.txt", Status: domain.StatusPending},
			},
			expected: 3,
		},
		{
			name:     "no files",
			files:    []domain.File{},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm := NewTaskManager()
			wp := NewWorkerPool(2, tm)
			wp.Start()
			defer wp.Stop()

			task, err := tm.CreateTask([]string{"http://example.com/test.txt"})
			if err != nil {
				t.Fatalf("failed to create task: %v", err)
			}

			wp.ProcessFiles(task.ID, tt.files)

			time.Sleep(50 * time.Millisecond)

			retrievedTask, exists := tm.GetTask(task.ID)
			if !exists {
				t.Errorf("expected task to exist")
				return
			}

			if len(retrievedTask.Files) == 0 {
				t.Errorf("expected task to have files")
			}
		})
	}
}

// TestWorkerPoolWithContext tests worker pool creation with external context
func TestWorkerPoolWithContext(t *testing.T) {
	tests := []struct {
		name    string
		workers int
	}{
		{
			name:    "context with single worker",
			workers: 1,
		},
		{
			name:    "context with multiple workers",
			workers: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm := NewTaskManager()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			wp := NewWorkerPoolWithContext(ctx, tt.workers, tm)

			if wp == nil {
				t.Errorf("expected worker pool but got nil")
				return
			}

			if wp.workers != tt.workers {
				t.Errorf("expected %d workers, got %d", tt.workers, wp.workers)
			}

			cancel()
			time.Sleep(10 * time.Millisecond)
		})
	}
}

// TestWorkerPoolProcessFiles tests processing files through worker pool
func TestWorkerPoolProcessFiles(t *testing.T) {
	tests := []struct {
		name  string
		files []domain.File
	}{
		{
			name: "process single file",
			files: []domain.File{
				{URL: "http://example.com/a.txt", Filename: "a.txt", Status: domain.StatusPending},
			},
		},
		{
			name: "process multiple files",
			files: []domain.File{
				{URL: "http://example.com/a.txt", Filename: "a.txt", Status: domain.StatusPending},
				{URL: "http://example.com/b.txt", Filename: "b.txt", Status: domain.StatusPending},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm := NewTaskManager()
			wp := NewWorkerPool(2, tm)
			wp.Start()
			defer wp.Stop()

			task, err := tm.CreateTask([]string{"http://example.com/test.txt"})
			if err != nil {
				t.Fatalf("failed to create task: %v", err)
			}

			go wp.ProcessFiles(task.ID, tt.files)

			time.Sleep(100 * time.Millisecond)

			retrievedTask, exists := tm.GetTask(task.ID)
			if !exists {
				t.Errorf("expected task to exist")
				return
			}

			if len(retrievedTask.Files) == 0 {
				t.Errorf("expected task to have files")
			}
		})
	}
}
