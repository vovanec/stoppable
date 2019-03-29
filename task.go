package stoppable

/* This package implements stoppable tasks. Example usages:

	// 1. Using polling.
	task := NewTask(...optional stop timeout...)
	go task.Run(func(c TaskChecker) error {
		for !c.ShouldStop() {
			// ... do some work here
		}
		// task has been told to stop...
		return nil // or some error
	})

	// ..... after some time .....
	if err := task.Stop(); err != nil {
		// handle error
	}


	// 2. Using channels.
	task := NewTask(...optional stop timeout...)
	go task.Run(func(c TaskChecker) error {
		for {
			select {
			case <- c.StopChan():
				// task has been told to stop...
				return nil
			// case data := <- someWorkChannel:
			// ... do work ...
			}

		}
	})

	// ..... after some time .....
	task.TellStop()
	if err := <-task.Wait(); err != nil {
		// handle error
	}

Inspired by https://github.com/matryer/runner
*/

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

// Instance of Task represents a task that can be stopped.
type Task interface {
	// Run starts execution of a TaskFunc function. This method blocks until fh is completed
	// so typically needs to be executed in a separate goroutine.
	Run(TaskFunc)
	// TellStop tells TaskFunc function that is being executed to stop execution.
	TellStop()
	// Wait returns a channel that waits for TaskFunc function to finish and returns its error.
	Wait() <-chan error
	// Stop is a convenience function which calls TellStop, waits for TaskFunc function to finish and return its error.
	Stop() error
	// IsStopped returns true if task is already finished, otherwise false.
	IsStopped() bool
}

type task struct {
	mu          sync.Mutex
	stopTimeout time.Duration
	doneChan    chan struct{}
	stopChan    chan struct{}
	stopped     int32
	err         error
}

// NewTask returns an instance of Task
func NewTask(stopTimeout time.Duration) Task {

	if stopTimeout == 0 {
		stopTimeout = defStopTimeout
	}

	return &task{
		doneChan:    make(chan struct{}),
		stopChan:    make(chan struct{}),
		stopTimeout: stopTimeout,
	}
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

	err = fn(&taskChecker{task: t})
}

func (t *task) TellStop() {

	atomic.StoreInt32(&t.stopped, 1)
	close(t.stopChan)
}

func (t *task) Wait() <-chan error {

	ch := make(chan error)

	go func() {
		<-t.doneChan
		ch <- t.err
	}()

	return ch
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

type taskChecker struct {
	task *task
}

func (t *taskChecker) StopChan() <-chan struct{} {
	return t.task.stopChan
}

func (t *taskChecker) ShouldStop() bool {
	return t.task.IsStopped()
}
