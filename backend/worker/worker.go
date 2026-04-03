package worker

import (
	"context"
	"sync"
	"time"

	"gorm.io/gorm"

	"prompt-vault/middleware"
	"prompt-vault/models"
)

// WorkerConfig holds the configuration for the worker pool.
type WorkerConfig struct {
	PoolSize     int           // Number of concurrent workers
	PollInterval time.Duration // How often to poll the database for pending tasks
	MaxRetries   int           // Maximum number of retries for failed tasks
}

// DefaultWorkerConfig returns the default worker configuration.
func DefaultWorkerConfig() WorkerConfig {
	return WorkerConfig{
		PoolSize:     5,
		PollInterval: 3 * time.Second,
		MaxRetries:   3,
	}
}

// Pool manages a pool of workers that process tasks from the database.
type Pool struct {
	cfg        WorkerConfig
	db         *gorm.DB
	executor   *TaskExecutor
	taskChan   chan *models.Task
	stopChan   chan struct{}
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
	mu         sync.Mutex
	running    bool
}

// NewPool creates a new worker pool.
func NewPool(cfg WorkerConfig, db *gorm.DB) *Pool {
	ctx, cancel := context.WithCancel(context.Background())
	return &Pool{
		cfg:        cfg,
		db:         db,
		executor:   NewTaskExecutor(),
		taskChan:   make(chan *models.Task, cfg.PollInterval*10),
		stopChan:   make(chan struct{}),
		ctx:        ctx,
		cancel:     cancel,
	}
}

// Start begins the worker pool.
func (p *Pool) Start() error {
	p.mu.Lock()
	if p.running {
		p.mu.Unlock()
		return nil
	}
	p.running = true
	p.mu.Unlock()

	middleware.Info("worker pool: starting", map[string]interface{}{
		"pool_size":     p.cfg.PoolSize,
		"poll_interval": p.cfg.PollInterval.String(),
		"max_retries":  p.cfg.MaxRetries,
	})

	// Recover any running tasks from previous instance
	if err := p.recoverRunningTasks(); err != nil {
		middleware.Warn("worker pool: failed to recover running tasks", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Start worker goroutines
	for i := 0; i < p.cfg.PoolSize; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}

	// Start poller goroutine
	p.wg.Add(1)
	go p.poller()

	middleware.Info("worker pool: started", nil)
	return nil
}

// Stop gracefully stops the worker pool.
func (p *Pool) Stop() {
	p.mu.Lock()
	if !p.running {
		p.mu.Unlock()
		return
	}
	p.running = false
	p.mu.Unlock()

	middleware.Info("worker pool: stopping", nil)
	p.cancel()
	close(p.stopChan)
	p.wg.Wait()
	middleware.Info("worker pool: stopped", nil)
}

// worker is the main loop for a worker goroutine.
func (p *Pool) worker(id int) {
	defer p.wg.Done()

	middleware.Debug("worker: started", map[string]interface{}{
		"worker_id": id,
	})

	for {
		select {
		case <-p.ctx.Done():
			middleware.Debug("worker: stopping", map[string]interface{}{
				"worker_id": id,
			})
			return
		case task, ok := <-p.taskChan:
			if !ok {
				return
			}
			p.processTask(task, id)
		}
	}
}

// poller periodically checks the database for pending tasks.
func (p *Pool) poller() {
	defer p.wg.Done()

	ticker := time.NewTicker(p.cfg.PollInterval)
	defer ticker.Stop()

	middleware.Debug("worker: poller started", map[string]interface{}{
		"interval": p.cfg.PollInterval.String(),
	})

	for {
		select {
		case <-p.ctx.Done():
			middleware.Debug("worker: poller stopping", nil)
			return
		case <-ticker.C:
			p.pollTasks()
		}
	}
}

// pollTasks fetches pending tasks from the database and sends them to workers.
func (p *Pool) pollTasks() {
	var tasks []models.Task
	now := time.Now()

	// Find tasks that are pending and due (run_at <= now)
	err := p.db.Where("status = ? AND run_at <= ?", models.TaskStatusPending, now).
		Order("run_at ASC").
		Limit(p.cfg.PoolSize * 2).
		Find(&tasks).Error

	if err != nil {
		middleware.Error("worker: failed to poll tasks", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	for _, task := range tasks {
		// Atomically transition task to running
		updated := p.db.Model(&models.Task{}).
			Where("id = ? AND status = ?", task.ID, models.TaskStatusPending).
			Update("status", models.TaskStatusRunning)

		if updated.RowsAffected == 0 {
			// Task was already claimed by another poll
			continue
		}

		// Set started_at
		now := time.Now()
		p.db.Model(&task).Updates(map[string]interface{}{
			"status":     models.TaskStatusRunning,
			"started_at": now,
		})

		// Make a copy for the channel
		taskCopy := task
		select {
		case p.taskChan <- &taskCopy:
			middleware.Debug("worker: dispatched task", map[string]interface{}{
				"task_id": task.ID,
				"type":    task.Type,
			})
		default:
			middleware.Warn("worker: task channel full, dropping task", map[string]interface{}{
				"task_id": task.ID,
			})
		}
	}
}

// processTask processes a single task.
func (p *Pool) processTask(task *models.Task, workerID int) {
	middleware.Debug("worker: processing task", map[string]interface{}{
		"task_id":   task.ID,
		"type":      task.Type,
		"worker_id": workerID,
	})

	// Update function for the executor to update task progress
	updateFn := func(t *models.Task) error {
		updates := map[string]interface{}{
			"progress": t.Progress,
			"status":   t.Status,
		}
		if t.StartedAt != nil {
			updates["started_at"] = t.StartedAt
		}
		if t.CompletedAt != nil {
			updates["completed_at"] = t.CompletedAt
		}
		if t.Result != "" {
			updates["result"] = t.Result
		}
		if t.Error != "" {
			updates["error"] = t.Error
		}
		return p.db.Model(&models.Task{}).Where("id = ?", t.ID).Updates(updates).Error
	}

	// Execute the task
	if err := p.executor.Execute(task, updateFn); err != nil {
		middleware.Error("worker: task failed", map[string]interface{}{
			"task_id": task.ID,
			"error":   err.Error(),
		})

		// Increment retry count and reschedule or mark as failed
		task.RetryCount++
		if task.RetryCount >= p.cfg.MaxRetries {
			p.db.Model(&models.Task{}).Where("id = ?", task.ID).Updates(map[string]interface{}{
				"status":      models.TaskStatusFailed,
				"error":       err.Error(),
				"retry_count": task.RetryCount,
			})
			middleware.Error("worker: task permanently failed after max retries", map[string]interface{}{
				"task_id":     task.ID,
				"retry_count": task.RetryCount,
			})
		} else {
			// Reschedule with exponential backoff
			backoff := time.Duration(task.RetryCount*task.RetryCount) * time.Second
			p.db.Model(&models.Task{}).Where("id = ?", task.ID).Updates(map[string]interface{}{
				"status":      models.TaskStatusPending,
				"run_at":      time.Now().Add(backoff),
				"retry_count": task.RetryCount,
			})
			middleware.Warn("worker: task rescheduled", map[string]interface{}{
				"task_id":     task.ID,
				"retry_count": task.RetryCount,
				"backoff":     backoff.String(),
			})
		}
		return
	}

	middleware.Info("worker: task completed", map[string]interface{}{
		"task_id": task.ID,
		"type":    task.Type,
	})
}

// recoverRunningTasks marks any tasks that were running when the server crashed as pending.
func (p *Pool) recoverRunningTasks() error {
	now := time.Now()
	result := p.db.Model(&models.Task{}).
		Where("status = ?", models.TaskStatusRunning).
		Updates(map[string]interface{}{
			"status":  models.TaskStatusPending,
			"run_at":  now,
		})

	if result.RowsAffected > 0 {
		middleware.Info("worker pool: recovered running tasks", map[string]interface{}{
			"count": result.RowsAffected,
		})
	}

	return nil
}

// IsRunning returns whether the pool is currently running.
func (p *Pool) IsRunning() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.running
}
