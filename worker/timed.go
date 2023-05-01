package worker

import (
	"time"
)

type RepeatingWorker struct {
	interval int
	job func()
	suspend chan bool
}

func Repeating(interval int, job func()) Worker {
	return &RepeatingWorker{ interval, job, make(chan bool) }
}

func (worker *RepeatingWorker) Start() {
	go func() {
		// Initial run.
		worker.job()

		for {
			select {
			case <-worker.suspend:
				return
			case <-time.After(time.Duration(worker.interval) * time.Second):
				worker.job()
			}
		}
	}()
}

func (worker *RepeatingWorker) Stop() {
	worker.suspend <- true
}
