package example

import (
	"log"

	"github.com/re-cinq/go-bus"
)

type MyEventHandler struct {
}

func (h *MyEventHandler) Apply(event bus.Event) {

	// Make sure we got the right event
	if metricsCollected, ok := event.(MetericsCollectedEvent); ok {

		log.Printf("CPU %s - MEMORY %s", metricsCollected.Cpu, metricsCollected.Memory)
		return
	}

	log.Printf("EmissionCalculator got an unknown event: %+v", event)

}
