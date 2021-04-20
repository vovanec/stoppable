// Package stoppable implements stoppable tasks. Example usages:
//
//	// 1. Using polling.
//	task := NewTask(...optional stop timeout...)
//	go task.Run(func(c TaskChecker) error {
//		for !c.ShouldStop() {
//			// ... do some work here
//		}
//		// task has been told to stop...
//		return nil // or some error
//	})
//
//	// ..... after some time .....
//	if err := task.Stop(); err != nil {
//		// handle error
//	}
//
//
//	// 2. Using channels.
//	task := NewTask(...optional stop timeout...)
//	go task.Run(func(c TaskChecker) error {
//		for {
//			select {
//			case <- c.StopChan():
//				// task has been told to stop...
//				return nil
//			// case data := <- someWorkChannel:
//			// ... do work ...
//			}
//
//		}
//	})
//
//	// ..... after some time .....
//	task.TellStop()
//	if err := <-task.Wait(); err != nil {
//		// handle error
//	}
//
//Inspired by https://github.com/matryer/runner

package stoppable