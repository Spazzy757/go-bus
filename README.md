### Build

[![build](https://github.com/re-cinq/go-bus/actions/workflows/build.yaml/badge.svg)](https://github.com/re-cinq/go-bus/actions/workflows/build.yaml)

### go-bus
It's a very simple pub sub implementation for Golang which allows to build an  
event driven architecture around it, by having handlers subscribing to specific topics and sending events of a specific topic to the bus.  

When an even of a specific topic is published to bus the bus loads all the subscribers and calls the `Apply` method of each subscriber and passes the event to it.  
This all happens in parallel

### Example

See the `example` folder