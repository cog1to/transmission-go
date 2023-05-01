package worker

type WorkerList []Worker

func (list WorkerList) Start() {
	for _, worker := range list {
		worker.Start()
	}
}

func (list WorkerList) Stop() {
	for _, worker := range list {
		worker.Stop()
	}
}

func WithSuspended(workers WorkerList, block func()) {
	workers.Stop()
	block()
	workers.Start()
}
