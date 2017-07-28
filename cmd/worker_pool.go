package cmd

import bosherr "github.com/cloudfoundry/bosh-utils/errors"

type WorkerPool struct {
	WorkerCount int
}

// Runs the given set of tasks in parallel using the configured number of worker go routines
// Will stop adding new tasks if a task throws an error, but will wait for in-flight tasks to finish
func (w WorkerPool) ParallelDo(tasks ...func() (interface{}, error)) ([]interface{}, error) {
	jobs := make(chan func() (interface{}, error))
	results := make(chan interface{}, len(tasks))
	errs := make(chan error, len(tasks))
	done := make(chan bool)

	for i := 0; i < w.WorkerCount; i++ {
		w.spawnWorker(jobs, results, errs, done)
	}

	for _, task := range tasks {
		select {
		case jobs <- task:
			// add another job
		case err := <-errs:
			// stop adding jobs
			errs <- err
			break
		}
	}
	close(jobs)

	for i := 0; i < w.WorkerCount; i++ {
		<-done
	}
	close(results)
	close(errs)

	combinedResults := []interface{}{}
	for result := range results {
		combinedResults = append(combinedResults, result)
	}

	var combinedErr error
	for err := range errs {
		if combinedErr == nil {
			combinedErr = err
		} else {
			combinedErr = bosherr.WrapError(combinedErr, err.Error())
		}
	}

	if combinedErr != nil {
		return nil, combinedErr
	}

	return combinedResults, nil
}

func (w WorkerPool) spawnWorker(tasks <-chan func() (interface{}, error), results chan<- interface{}, errs chan<- error, done chan<- bool) {
	go func() {
		for task := range tasks {
			result, err := task()
			if err != nil {
				errs <- err
				break
			} else {
				results <- result
			}
		}

		done <- true
	}()
}
