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

func NopHandler(time time.Time, entity Entity, words []Word) {}

type Entity uintptr

type Driver interface {
	Close() error
	SetReceiveHandler(ReceiveEventHandler)
	OpenInput(Entity) error
	OpenOutput(Entity) error
	Send(time.Time, Entity, []Word) error
	SendSysEx(Entity, []byte) error
	Enumerate() (*Node, error)
}
