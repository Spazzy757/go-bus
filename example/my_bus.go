package example

import (
	"runtime"

	"github.com/re-cinq/go-bus"
	klog "k8s.io/klog/v2"
)

func main() {

	// Init the bus
	myBus := bus.NewEventBus(12, runtime.NumCPU(), klog.NewKlogr())

	// Init the handler
	myHandler := MyEventHandler{}

	// Subscribe it and from this moment on all events published on the
	// EmissionsCalculatedTopic will reach this handler
	// You can have multiple handlers subscribed to the same topic as well
	myBus.Subscribe(EmissionsCalculatedTopic, &myHandler)
}
