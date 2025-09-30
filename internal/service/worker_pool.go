package service

import (
	"context"
	"sync"

	"filedownloader-20240926/internal/domain"
	"filedownloader-20240926/pkg/logger"
)

type DownloadTask struct {
	File   *domain.File
	TaskID string
}

type WorkerPool struct {
	workers    int
	downloader *Downloader
	taskChan   chan DownloadTask
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	tm         *TaskManager
	once       sync.Once
}

// NewWorkerPool creates a new worker pool with specified number of workers
func NewWorkerPool(workers int, tm *TaskManager) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	return &WorkerPool{
		workers:    workers,
		downloader: NewDownloader(),
		taskChan:   make(chan DownloadTask, workers*2),
		ctx:        ctx,
		cancel:     cancel,
		tm:         tm,
	}
}

// NewWorkerPoolWithContext creates WorkerPool with external context
func NewWorkerPoolWithContext(ctx context.Context, workers int, tm *TaskManager) *WorkerPool {
	workerCtx, cancel := context.WithCancel(ctx)
	return &WorkerPool{
		workers:    workers,
		downloader: NewDownloader(),
		taskChan:   make(chan DownloadTask, workers*2),
		ctx:        workerCtx,
		cancel:     cancel,
		tm:         tm,
	}
}

// Start starts all workers in the pool
func (wp *WorkerPool) Start() {
	logger.Logger.Info("Starting workers", "count", wp.workers)

	for i := 0; i < wp.workers; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}
}

// Stop stops all workers in the pool
func (wp *WorkerPool) Stop() {
	logger.Logger.Info("Stopping workers")
	wp.cancel()

	wp.once.Do(func() {
		close(wp.taskChan)
	})

	wp.wg.Wait()
	logger.Logger.Info("All workers stopped")
}

// worker processes tasks for a single worker
func (wp *WorkerPool) worker(id int) {
	defer wp.wg.Done()

	logger.Logger.Debug("Worker started", "worker_id", id)

	for {
		select {
		case task, ok := <-wp.taskChan:
			if !ok {
				logger.Logger.Debug("Worker channel closed", "worker_id", id)
				return
			}

			wp.processTask(task)

		case <-wp.ctx.Done():
			logger.Logger.Debug("Worker context cancelled", "worker_id", id)
			return
		}
	}
}

// processTask processes a single download task
func (wp *WorkerPool) processTask(task DownloadTask) {
	file := task.File
	logger.Logger.Debug("Processing file", "url", file.URL, "task_id", task.TaskID)

	file.Status = domain.StatusDownloading

	size, err := wp.downloader.GetFileSize(file.URL)
	if err != nil {
		logger.Logger.Error("Failed to get file size", "url", file.URL, "error", err)
		file.Status = domain.StatusFailed
		return
	}
	file.Size = size

	filename := wp.downloader.ExtractFilename(file.URL)
	savedName, err := wp.downloader.DownloadFile(file.URL, filename)
	if err != nil {
		logger.Logger.Error("Download failed", "url", file.URL, "error", err)
		file.Status = domain.StatusFailed
		return
	}

	file.Status = domain.StatusCompleted
	file.Downloaded = file.Size
	file.Filename = savedName

	logger.Logger.Info("Download completed", "url", file.URL, "size", file.Size, "filename", filename)

	wp.updateTaskProgress(task.TaskID)
}

// AddTask adds a task to the queue
func (wp *WorkerPool) AddTask(task DownloadTask) {
	select {
	case wp.taskChan <- task:
		logger.Logger.Debug("Task added to queue", "url", task.File.URL, "task_id", task.TaskID)
	case <-wp.ctx.Done():
		logger.Logger.Warn("Worker pool stopped, cannot add task")
	default:
		logger.Logger.Warn("Task queue full, dropping task")
	}
}

// ProcessFiles processes a list of files
func (wp *WorkerPool) ProcessFiles(taskID string, files []domain.File) {
	logger.Logger.Info("Processing files", "task_id", taskID, "files_count", len(files))

	for i := range files {
		downloadTask := DownloadTask{
			File:   &files[i],
			TaskID: taskID,
		}
		wp.AddTask(downloadTask)
	}
}

// updateTaskProgress updates the progress of a task based on file completion status
func (wp *WorkerPool) updateTaskProgress(taskID string) {
	if wp.tm == nil {
		return
	}
	task, ok := wp.tm.GetTask(taskID)
	if !ok {
		return
	}

	var totalSize int64
	var downloaded int64
	allCompleted := true
	anyInProgress := false
	for i := range task.Files {
		totalSize += task.Files[i].Size
		downloaded += task.Files[i].Downloaded
		if task.Files[i].Status != domain.StatusCompleted {
			allCompleted = false
		}
		if task.Files[i].Status == domain.StatusDownloading {
			anyInProgress = true
		}
	}

	if totalSize > 0 {
		task.Progress = int(float64(downloaded) / float64(totalSize) * 100)
	} else {
		completed := 0
		for i := range task.Files {
			if task.Files[i].Status == domain.StatusCompleted {
				completed++
			}
		}
		if len(task.Files) > 0 {
			task.Progress = int(float64(completed) / float64(len(task.Files)) * 100)
		}
	}

	switch {
	case allCompleted:
		task.Status = domain.StatusCompleted
	case anyInProgress:
		task.Status = domain.StatusDownloading
	default:
		if task.Progress > 0 {
			task.Status = domain.StatusDownloading
		} else {
			task.Status = domain.StatusPending
		}
	}

	_ = wp.tm.UpdateTask(task)
}
