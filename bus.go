package bus

import (
	"fmt"
	"hash/fnv"
	"sync"
	"time"
)

type EventBus struct {
	// The size of the main queue
	queueSize int

	// The amount of parallel workers
	workerPoolSize int

	// Storing of the worker references
	workers map[uint32]worker

	// Storing of the subscribers to specific topics
	subscribers map[Topic][]EventHandler

	// The main queue
	queue chan Event

	// Locking for updating the subscribers
	lock sync.RWMutex

	// Lock for publishing messages and shuttind down
	publishLock sync.RWMutex

	// Signaling whether the bus is shutting down, so no more messages will be accepted
	shutdown bool

	logger Logger
}

type option func(*EventBus)

func WithLogger(l Logger) option {
	return func(e *EventBus) {
		e.logger = l
	}
}

// NewEventBus instance
func NewEventBus(queueSize int, workerPoolSize int, opts ...option) Bus {

	// Init the event bus
	bus := EventBus{
		queueSize:      queueSize,
		workerPoolSize: workerPoolSize,
		workers:        make(map[uint32]worker, workerPoolSize),
		subscribers:    make(map[Topic][]EventHandler),
		queue:          make(chan Event, queueSize),
		lock:           sync.RWMutex{},
		publishLock:    sync.RWMutex{},
		shutdown:       false,
	}

	// set any options that might be set
	for _, o := range opts {
		o(&bus)
	}

	// Init the channels and start the workers
	for i := 0; i < workerPoolSize; i++ {
		// init the channel
		bus.workers[uint32(i)] = newWorker()

		// Start the worker
		bus.startWorker(uint32(i))
	}

	return &bus
}

func (bus *EventBus) Subscribe(topic Topic, subscriber EventHandler) {
	bus.lock.Lock()
	if topicSubscribers, ok := bus.subscribers[topic]; ok {
		topicSubscribers = append(topicSubscribers, subscriber)
		bus.subscribers[topic] = topicSubscribers
	} else {
		bus.subscribers[topic] = []EventHandler{subscriber}
	}
	bus.lock.Unlock()
}

// Start the Bus
func (bus *EventBus) Start() {
	go bus.process()
}

// Stop the Bus
func (bus *EventBus) Stop() {

	// shutdown
	bus.publishLock.Lock()

	// send the poison pill inside the lock so no other publish
	// happens
	bus.queue <- poisonPillEvent{}

	// Set the bus in shutdown state
	bus.shutdown = true

	// Unlock
	bus.publishLock.Unlock()

	bus.logger.Info("bus received shutdown signal")

	// wait for all the workers to be done processing the poison pill
	for index, worker := range bus.workers {
		<-worker.shutdown
		bus.logger.Info(fmt.Sprintf("worker %d shutdown completed", index))
	}

	bus.logger.Info("event bus shutdown completed")
}

// Publish a new event
func (bus *EventBus) Publish(event Event) {
	var shuttingDown bool

	bus.publishLock.RLock()
	shuttingDown = bus.shutdown
	bus.publishLock.RUnlock()

	if !shuttingDown {
		bus.queue <- event
	}

}

func (bus *EventBus) process() {

	// Process all the events sequentially
	for {

		select {
		// Get the event
		case event, more := <-bus.queue:
			if more {

				// Send the poison pill to all the workers
				// this is so that we can guarantee the workers will drain
				// their queues
				if poison, ok := event.(poisonPillEvent); ok {

					// Send the poison pill to all the workers
					for _, worker := range bus.workers {
						worker.data <- poison
					}

					// Exit the loop
					break

				} else {
					// Calculate the worker ID from the topic name
					workerId := bus.getWorkerId(event.Identifier())

					// Get the worker channel
					bus.lock.RLock()
					worker := bus.workers[workerId]
					bus.lock.RUnlock()

					// Send the event for processing
					worker.data <- event
				}

			}
			// timeout otherwise to allow the shutdown procedure
		case <-time.After(1 * time.Minute):
			// Good for checking if the bus size needs to be increased
			bus.logger.Info(fmt.Sprintf("messages queued: %d", len(bus.queue)))
		}
	}

}

// Calculate the worker ID from the event topic
func (bus *EventBus) getWorkerId(eventId string) uint32 {

	// Under the hood this init a constant so it's not expensive
	// to init an hash at each iteration
	// Ideally though we should try to improve this if possible
	// Watch out for concurrency in case of refactoring
	eventHash := fnv.New32a()

	// Write the data
	eventHash.Write([]byte(eventId))

	// Calculate the id
	id := eventHash.Sum32()

	// Modulo the pool size
	return id % uint32(bus.workerPoolSize)

}

// Worker which processes the event
func (bus *EventBus) startWorker(id uint32) {

	// Get the channel outside of the go routine
	// so we don't need to sync
	workerChan := bus.workers[id]

	go func() {

		// Loop through it or wait for a task
		for {
			// Get the event
			event, more := <-workerChan.data

			// If we do have data
			if more {

				// Checke if we got a poison pill and it so
				// it's time to swallow it...
				if _, ok := event.(poisonPillEvent); ok {
					// Exit the loop since we are dead
					break
				}

				// otherwise load all the subscribers
				if subscribers, ok := bus.subscribers[event.Topic()]; ok {

					// Create a Sync group to process the events in parallel
					var wg sync.WaitGroup

					// Loop through all subscribers
					for _, subscriber := range subscribers {
						// add to sync group
						wg.Add(1)
						go func(sub EventHandler) {
							// apply
							sub.Apply(event)

							// done
							wg.Done()
						}(subscriber)
					}

					// wait for all the event handlers to finish
					wg.Wait()

				}
			}
		}

		// Our work is literally done here
		workerChan.done()
	}()

}
