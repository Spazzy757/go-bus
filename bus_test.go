package bus

import (
	"fmt"
	"runtime"
	"sync"
	"testing"
)

const testTopic Topic = 1

type testHandler struct {
	counter int
}

type testEvent struct {
	Id string
}

func (e testEvent) Identifier() string {
	return e.Id
}

func (e testEvent) Topic() Topic {
	return testTopic
}

func (h *testHandler) Apply(event Event) {

	// Make sure we got the right event
	if _, ok := event.(testEvent); ok {
		h.counter++
	}
}

func TestEventBus(t *testing.T) {

	// Max messages
	max := 1_000_000

	// Init and start the bus
	testBus := NewEventBus(1024, runtime.NumCPU())
	testBus.Start()

	// Build the handler
	handler := testHandler{counter: 0}

	// Subscribe the handler to the topic
	testBus.Subscribe(testTopic, &handler)

	wg := sync.WaitGroup{}

	// Publish 1000 events
	for i := 0; i < max; i++ {
		wg.Add(1)
		go func() {
			testBus.Publish(testEvent{})
			wg.Done()
		}()

	}

	fmt.Println("publishing....")

	wg.Wait()

	// Stop the bus
	testBus.Stop()

	if handler.counter != max {
		t.Fail()
	}

}
