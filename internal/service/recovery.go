package service

import (
	"log"

	"filedownloader-20240926/internal/domain"
)

// RecoverIncompleteTasks recovers incomplete tasks on startup
func (tm *TaskManager) RecoverIncompleteTasks() {
	log.Println("Recovering incomplete tasks...")

	tasks := tm.GetAllTasks()
	recovered := 0

	for _, task := range tasks {
		if task.Status == domain.StatusPending || task.Status == domain.StatusDownloading {
			log.Printf("Recovering task %s with status %s", task.ID, task.Status)
			originalStatus := task.Status
			task.Status = domain.StatusPending
			task.Progress = 0
			for i := range task.Files {
				if task.Files[i].Status != domain.StatusCompleted {
					task.Files[i].Status = domain.StatusPending
					task.Files[i].Downloaded = 0
				}
			}
			if err := tm.UpdateTask(task); err != nil {
				log.Printf("Failed to update recovered task %s: %v", task.ID, err)
			} else {
				log.Printf("Recovered task %s from %s to %s", task.ID, originalStatus, task.Status)
				recovered++
			}
		}
	}

	log.Printf("Recovered %d incomplete tasks", recovered)
}

// GetIncompleteTasks returns list of incomplete tasks
func (tm *TaskManager) GetIncompleteTasks() []*domain.Task {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	var incomplete []*domain.Task
	for _, task := range tm.tasks {
		if task.Status == domain.StatusPending || task.Status == domain.StatusDownloading {
			incomplete = append(incomplete, task)
		}
	}

	return incomplete
}

// ResumeTasks resumes processing of incomplete tasks
func (wp *WorkerPool) ResumeTasks(tasks []*domain.Task) {
	log.Printf("Resuming %d incomplete tasks", len(tasks))

	for _, task := range tasks {
		for i := range task.Files {
			if task.Files[i].Status != domain.StatusCompleted {
				downloadTask := DownloadTask{
					File:   &task.Files[i],
					TaskID: task.ID,
				}
				wp.AddTask(downloadTask)
			}
		}
	}
}
