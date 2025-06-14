package main

type Worker struct {
	id        int
	taskQueue chan func()
	quitChan  chan bool
}

func New(id int) *Worker {
	worker := &Worker{
		id:        id,
		taskQueue: make(chan func()),
		quitChan:  make(chan bool),
	}
	go worker.Start()
	return worker
}

func (w *Worker) Stop() {
	w.stop()
}

func (w *Worker) stop() {
	w.quitChan <- true
}

func (w *Worker) Start() {
	for {
		select {
		case task := <-w.taskQueue:
			go doTask(task)
		case <-w.quitChan:
			close(w.taskQueue)
			return
		}
	}
}

func (w *Worker) Submit(task func()) {
	if task != nil {
		w.taskQueue <- task
	}
}

func doTask(task func()) {
	task()
}
