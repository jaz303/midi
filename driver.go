package midi

import "time"

// ReceiveEventHandler represents a callback function that receives incoming
// MIDI events from an Entity.
//
// A ReceiveEventHandler can be invoked by its owning Driver from any thread.
//
// words MUST NOT be retained beyond the lifetime of the callback invocation
// since it likely references the driver's C-owned memory. Additionally, no
// operation which can affect the slice's underlying storage (e.g. append())
// should be attempted. If the data needs to be persist beyond the lifetime
// of the callback, make a copy.
type ReceiveEventHandler func(time time.Time, entity Entity, words []Word)

type DriverConfig struct {
	ReceiveHandler ReceiveEventHandler
}

type Entity uintptr

type Driver interface {
	Name() string
	Available() bool
	Init(cfg *DriverConfig) error
	OpenInput(p Entity) error
	OpenOutput(p Entity) error
	Enumerate() (*Node, error)
}
