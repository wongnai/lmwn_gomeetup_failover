package workerpool

import (
	"fmt"
	"sync"
)

// WorkerPool defines a simple worker pool
type WorkerPool struct {
	taskQueue chan func()
	wg        sync.WaitGroup
	once      sync.Once
}

// NewWorkerPool initializes a worker pool with a fixed number of workers
func NewWorkerPool(workerCount, queueSize int) *WorkerPool {
	pool := &WorkerPool{
		taskQueue: make(chan func(), queueSize),
	}

	for i := 0; i < workerCount; i++ {
		pool.wg.Add(1)
		go pool.worker(i)
	}

	return pool
}

// worker executes tasks from the queue
func (wp *WorkerPool) worker(id int) {
	defer wp.wg.Done()
	for task := range wp.taskQueue {
		fmt.Printf("Worker %d executing task...\n", id)
		task()
	}
}

// Submit adds a new task to the worker pool queue
func (wp *WorkerPool) Submit(task func()) {
	wp.taskQueue <- task
}

// Shutdown gracefully stops the worker pool
func (wp *WorkerPool) Shutdown() {
	wp.once.Do(func() {
		close(wp.taskQueue)
	})
	wp.wg.Wait()
}
