package example

import "github.com/re-cinq/go-bus"

type MetericsCollectedEvent struct {
	// Event unique identifier so that the same event ID
	// will be processed sequentially whereas different
	// event IDs will be processed in parallel
	Id     string
	Cpu    string
	Memory string
}

// The topic this event is about
func (e MetericsCollectedEvent) Topic() bus.Topic {
	return MetricsCollectedTopic
}

// Returns the unique name of the instance or service
func (e MetericsCollectedEvent) Identifier() string {
	return e.Id
}
