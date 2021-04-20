package stoppable

import (
	"math"
	"sync"
	"sync/atomic"
	"time"
)

const defStopTimeout = math.MaxInt64

type TaskChecker interface {
	// ShouldStop returns true when currently running task is told to stop.
	ShouldStop() bool
	// StopChan returns a channel that is closed when currently running task is told to stop.
	StopChan() <-chan struct{}
}

// TaskFunc is a function executed by Task, in order to stop gracefully it should
// use one of methods of TaskChecker instance as a signal to stop execution.
type TaskFunc func(TaskChecker) error

// Task represents a runnable task that can be stopped.
type Task interface {
	// Run starts execution of a TaskFunc function. This method blocks until fh is completed
	// so typically needs to be executed in a separate goroutine.
	Run(TaskFunc)
	// TellStop tells TaskFunc function that is being executed to stop execution.
	TellStop()
	// Wait returns a channel that waits for TaskFunc function to finish and returns its error.
	Wait() <-chan error
	// Stop is a convenience function which calls TellStop, waits for TaskFunc function
	// to complete and returns its error.
	Stop() error
	// IsStopped returns true if task is already finished, otherwise false.
	IsStopped() bool
}

type task struct {
	mu          sync.Mutex
	stopTimeout time.Duration
	doneChan    chan struct{}
	stopChan    chan struct{}
	waitChan	chan error
	stopped     int32
	err         error
}

type TaskOption func (t *task)

// WithStopTimeout adds stop timeout for a task - how long task should be awaited
// for normal completion before Stop returns. Stop timeout of 0 means infinite timeout.
func WithStopTimeout(d time.Duration) TaskOption {
	return func(t *task) {
		if d < 0 {
			panic("task stop timeout must be positive or zero duration")
		}

		if d > 0 {
			t.stopTimeout = d
		}
	}
}

// NewTask returns an instance of Task
func NewTask(opts ...TaskOption) Task {

	t := &task{
		doneChan:    make(chan struct{}),
		stopChan:    make(chan struct{}),
		waitChan:    make(chan error),
		stopTimeout: defStopTimeout,
	}

	for _, opt := range opts {
		opt(t)
	}

	return t
}


func (t *task) Run(fn TaskFunc) {

	var err error

	defer func() {
		t.mu.Lock()
		t.err = err
		t.mu.Unlock()

		atomic.StoreInt32(&t.stopped, 1)
		close(t.doneChan)
	}()

	go func() {
		<-t.doneChan
		t.waitChan <- t.err
	}()

	err = fn(&taskChecker{task: t})
}

func (t *task) TellStop() {

	atomic.StoreInt32(&t.stopped, 1)
	close(t.stopChan)
}

func (t *task) Stop() error {

	if t.IsStopped() {
		return t.err
	}

	t.TellStop()

	select {
	case err := <-t.Wait():
		return err
	case <-time.After(t.stopTimeout):
		return nil
	}
}

func (t *task) IsStopped() bool {
	if atomic.LoadInt32(&t.stopped) > 0 {
		return true
	}
	return false
}

func (t *task) Wait() <-chan error {
	return t.waitChan
}

type taskChecker struct {
	task *task
}

func (t *taskChecker) StopChan() <-chan struct{} {
	return t.task.stopChan
}

func (t *taskChecker) ShouldStop() bool {
	return t.task.IsStopped()
}
