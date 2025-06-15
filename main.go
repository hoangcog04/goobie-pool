package main

import (
	"sync"
	"time"
)

const idleTimeout = time.Second * 2

// WorkerPool is a fixed-size pool of workers that execute tasks.
// Tasks are distributed among workers as they become available up to the concurrency limit.
type WorkerPool struct {
	// A channel used to limit the number of active workers.
	// Each token represents an available worker slot.
	workerTokens chan struct{}
	// A channel holds incoming tasks waiting to be executed.
	taskQueue chan func()
	// It decouples task submission from execution by acting as a queue of ready-to-pick tasks.
	// It holds tasks that are ready to be picked up by any available worker.
	workerQueue chan func()
	// A channel is used to signal all workers to gracefully shut down.
	quitChan chan bool
	//
	mu      sync.Mutex
	stopped bool
}

func NewPool(maxWorkers int) *WorkerPool {
	worker := &WorkerPool{
		workerTokens: make(chan struct{}, maxWorkers),
		taskQueue:    make(chan func()),
		workerQueue:  make(chan func()),
		quitChan:     make(chan bool),
	}
	go worker.Run()
	return worker
}

func (wp *WorkerPool) Stop() {
	wp.stop()
}

func (wp *WorkerPool) stop() {
	wp.quitChan <- true
}

func (wp *WorkerPool) Run() {
	timeout := time.NewTimer(idleTimeout)

	for {
		select {
		case task := <-wp.taskQueue:
			select {
			case wp.workerQueue <- task:
			default:
				wp.workerTokens <- struct{}{} // fix: this operation can block and needs handling
				go runWorker(task, wp.workerTokens, wp.workerQueue)
			}
		case <-timeout.C:
			if len(wp.workerTokens) == cap(wp.workerTokens) {
				_ = wp.killAnyIdleWorker()
			}
			timeout.Reset(idleTimeout)
		case <-wp.quitChan:
			close(wp.taskQueue)
			return
		}
	}
}

func (wp *WorkerPool) Submit(task func()) {
	if task != nil {
		wp.taskQueue <- task
	}
}

func runWorker(task func(), workerTokens chan struct{}, workerQueue chan func()) {
	for task != nil {
		task()
		task = <-workerQueue
	}
	// release worker back into pool
	// or in other words, exit the goroutine
	<-workerTokens
}

func (wp *WorkerPool) Close() {
	wp.mu.Lock()
	wp.stopped = true
	wp.mu.Unlock()
	// fix: race condition here
	close(wp.taskQueue)
}

func (wp *WorkerPool) killAnyIdleWorker() bool {
	select {
	case wp.workerQueue <- nil:
		return true
	default:
		return false
	}
}
