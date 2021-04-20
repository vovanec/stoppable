package stoppable

import (
	"errors"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var unexpectedErr = errors.New("something unexpected happened")


func TestNewTask(t *testing.T) {

	var (
		task   = NewTask(WithStopTimeout(time.Second * 2))
		cycles int
	)

	go task.Run(func(checker TaskChecker) error {

		for {
			if checker.ShouldStop() {
				break
			}
			time.Sleep(time.Millisecond * 100)
			cycles += 1
		}
		return nil
	})

	time.Sleep(time.Millisecond * 1000)
	assert.False(t, task.IsStopped())

	err := task.Stop()

	assert.Nil(t, err)
	assert.Equal(t, 10, cycles)
	assert.True(t, task.IsStopped())
}

func TestTaskError(t *testing.T) {

	task := NewTask(WithStopTimeout(time.Second * 2))

	go task.Run(func(_ TaskChecker) error {
		return errors.New("unexpected error")
	})

	assert.False(t, task.IsStopped())
	time.Sleep(time.Millisecond * 100)
	assert.True(t, task.IsStopped())

	err := task.Stop()

	assert.NotNil(t, err)
	assert.True(t, task.IsStopped())
}

func TestTask_StopWithTimeout(t *testing.T) {

	stopTimeout := time.Millisecond * 100
	taskTimeout := time.Millisecond * 500
	task := NewTask(WithStopTimeout(stopTimeout))

	tStart := time.Now()
	go task.Run(func(_ TaskChecker) error {
		time.Sleep(taskTimeout)
		return nil
	})

	assert.False(t, task.IsStopped())

	err := task.Stop()

	assert.Nil(t, err)
	elapsed := time.Now().Sub(tStart)
	assert.True(t, elapsed >= stopTimeout && elapsed < taskTimeout)
	assert.True(t, task.IsStopped())
}

func TestTaskWait(t *testing.T) {

	var (
		task   = NewTask(WithStopTimeout(time.Second * 2))
	)

	go task.Run(func(checker TaskChecker) error {

		time.Sleep(time.Millisecond * 500)
		return unexpectedErr
	})

	assert.False(t, task.IsStopped())
	err := <- task.Wait()
	assert.Equal(t, unexpectedErr, err)
	assert.True(t, task.IsStopped())
}

func ExampleTask() {

	task := NewTask(WithStopTimeout(0))
	go task.Run(func(c TaskChecker) error {
		for !c.ShouldStop() {
			time.Sleep(time.Millisecond * 100)
		}
		log.Println("task has been told to stop")
		return nil
	})

	time.Sleep(time.Millisecond * 100)
	if err := task.Stop(); err != nil {
		log.Fatalf("task returned error: %s", err)
	}
	fmt.Println(task.IsStopped())

	task = NewTask(WithStopTimeout(0))
	go task.Run(func(c TaskChecker) error {
		<-c.StopChan()
		log.Println("task has been told to stop")
		return nil
	})

	time.Sleep(time.Millisecond * 100)
	if err := task.Stop(); err != nil {
		log.Fatalf("task returned error: %s", err)
	}
	fmt.Println(task.IsStopped())

	//
	task = NewTask(WithStopTimeout(0))
	go task.Run(func(c TaskChecker) error {
		<-c.StopChan()
		log.Println("task has been told to stop")
		return nil
	})

	time.Sleep(time.Millisecond * 100)
	task.TellStop()
	if err := <-task.Wait(); err != nil {
		log.Fatalf("task returned error: %s", err)
	}
	fmt.Println(task.IsStopped())
	// Output:
	// true
	// true
	// true
}
