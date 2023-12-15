package bus

// worker is spawned in a go routine and
type worker struct {

	// The worker own channel
	data chan Event

	// Waiting for the worker to conclude its tasks
	shutdown chan bool
}

func newWorker() worker {
	return worker{
		data:     make(chan Event, 1),
		shutdown: make(chan bool, 1),
	}
}

func (w *worker) done() {
	// close the data channel
	close(w.data)

	// Notify the parent that this worker is fully shutdown
	w.shutdown <- true
}
