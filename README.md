# stoppable

This package implements stoppable tasks.

API documentation: https://godoc.org/github.com/vovanec/stoppable

Example usages:

1. Wait task for finish using polling.

    ```

	task := NewTask(WithStopTimeout(...optional stop timeout...))
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
	```


2. Wait task for finish using channel.

    ```

	task := NewTask()
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
	```

Inspired by https://github.com/matryer/runner
