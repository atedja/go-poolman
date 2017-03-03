package poolman

import (
	"runtime"
	"sync"
)

// Poolman manages background tasks by queuing them up and assigning them
// to background workers as they become available.
type Poolman struct {
	sync.Mutex
	workers  []*worker
	incoming chan *task
}

var Default, _ = New(runtime.NumCPU(), runtime.NumCPU()*2)

// Create a new Poolman instance
func New(workerCount int, queueSize int) (*Poolman, error) {
	if workerCount < 1 || queueSize < 1 {
		return nil, ErrInvalidWorkerCountOrQueueSize
	}

	pm := &Poolman{
		workers:  make([]*worker, workerCount),
		incoming: make(chan *task, queueSize),
	}

	// Spawn workers and run them
	for i := 0; i < workerCount; i++ {
		w := &worker{
			stopped: make(chan bool, 1),
		}
		go w.run(pm.incoming)
		pm.workers[i] = w
	}

	return pm, nil
}

// Add a new task to the pool. Non-blocking unless queue is full.
func (self *Poolman) AddTask(fn interface{}, args ...interface{}) error {
	tsk := &task{
		Fn:   fn,
		Args: args,
	}
	self.incoming <- tsk

	return nil
}

// Resize the number of workers. Must be at least 1.
// If new size is bigger, new workers are added, the old ones are untouched.
// If new size is smaller, Poolman will stop excess workers.
// This method does not increase the size of the queue.
func (self *Poolman) Resize(newSize int) error {
	if newSize < 1 {
		return ErrInvalidWorkerCountOrQueueSize
	}

	self.Lock()
	defer self.Unlock()

	workerCount := len(self.workers)

	if newSize == workerCount {
		return nil
	}

	// Move workers to the new array, and stopping the excess ones
	newWorkers := make([]*worker, newSize)
	for i, w := range self.workers {
		if i < newSize {
			newWorkers[i] = w
		} else {
			w.stop()
		}
	}

	// If new size is bigger, spawn new workers and run them
	if newSize > workerCount {
		for i := workerCount; i < newSize; i++ {
			w := &worker{
				stopped: make(chan bool, 1),
			}
			go w.run(self.incoming)
			newWorkers[i] = w
		}
	}

	self.workers = newWorkers

	return nil
}

// Close Poolman by stopping all workers. If there are remaining tasks in the queue, they won't get processed.
// This method is intended to be used when program is being terminated.
// You can reactivate Poolman by calling Resize(), and it's been tested to work, but please just..no.
func (self *Poolman) Close() {
	self.Lock()
	defer self.Unlock()

	for _, w := range self.workers {
		w.stop()
	}

	self.workers = nil
}
