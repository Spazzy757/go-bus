package bus

import "math"

// This is the event will send
type poisonPillEvent struct{}

// Defines the poison pill with topic MaxUint32
const poisonPill Topic = math.MaxUint32

func (p poisonPillEvent) Identifier() string {
	return ""
}

func (p poisonPillEvent) Topic() Topic {
	return poisonPill
}
