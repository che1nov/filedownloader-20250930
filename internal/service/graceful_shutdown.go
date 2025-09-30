package service

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type GracefulShutdown struct {
	server      *http.Server
	workerPool  *WorkerPool
	taskManager *TaskManager
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
}

// NewGracefulShutdown creates a new graceful shutdown handler
func NewGracefulShutdown(server *http.Server, workerPool *WorkerPool, taskManager *TaskManager) *GracefulShutdown {
	ctx, cancel := context.WithCancel(context.Background())
	return &GracefulShutdown{
		server:      server,
		workerPool:  workerPool,
		taskManager: taskManager,
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Start starts server with graceful shutdown
func (gs *GracefulShutdown) Start() error {
	gs.workerPool.Start()

	gs.wg.Add(1)
	go func() {
		defer gs.wg.Done()
		if err := gs.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Server error: %v", err)
		}
	}()

	gs.waitForSignals()

	gs.shutdown()

	return nil
}

// waitForSignals waits for signals for graceful shutdown
func (gs *GracefulShutdown) waitForSignals() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	log.Printf("Received signal: %v", sig)
}

// shutdown performs graceful shutdown
func (gs *GracefulShutdown) shutdown() {
	log.Println("Starting graceful shutdown...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := gs.server.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}
	gs.workerPool.Stop()
	gs.saveAllTasks()
	gs.wg.Wait()

	log.Println("Graceful shutdown completed")
}

// saveAllTasks saves state of all tasks
func (gs *GracefulShutdown) saveAllTasks() {
	log.Println("Saving all tasks state...")

	tasks := gs.taskManager.GetAllTasks()
	saved := 0

	for _, task := range tasks {
		if err := gs.taskManager.UpdateTask(task); err != nil {
			log.Printf("Failed to save task %s: %v", task.ID, err)
		} else {
			saved++
		}
	}

	log.Printf("Saved %d tasks", saved)
}

// GetContext returns context for use in other components
func (gs *GracefulShutdown) GetContext() context.Context {
	return gs.ctx
}
