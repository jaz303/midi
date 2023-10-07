package darwin

// #cgo CFLAGS: -x objective-c
// #cgo CFLAGS: -Wint-to-void-pointer-cast
// #cgo LDFLAGS: -framework CoreMIDI
// #include <CoreMIDI/MIDIServices.h>
// #include <CoreFoundation/CFRunLoop.h>
// #include "binding.h"
// #include <stdio.h>
import "C"

import (
	"fmt"
	"time"
	"unsafe"

	"github.com/jaz303/midi"
)

func init() {
	midi.Register(&driver{})
}

type driver struct{}

// Callback function that receives all incoming MIDI events
var onReceive midi.ReceiveEventHandler

// These values tie the system local time to Core MIDI's
// internal time.
var (
	// Go time at driver init
	goEpoch time.Time

	// mach_absolute_time() at driver init
	machEpoch uint64

	// mach timebase for conver
	// Retrieved from mach_timebase_info()
	machTimebaseNumer uint64
	machTimebaseDenom uint64
)

func timestampToTime(ts uint64) time.Time {
	ticks := ts - machEpoch
	nanos := (ticks * machTimebaseNumer) / machTimebaseDenom
	return goEpoch.Add(time.Duration(nanos))
}

func timeToTimestamp(t time.Time) uint64 {
	nanos := uint64(t.Sub(goEpoch))
	ticks := (nanos * machTimebaseDenom) / machTimebaseNumer
	return machEpoch + ticks
}

//export OnReceive
func OnReceive(timestamp uint64, source unsafe.Pointer, words unsafe.Pointer, wordCount uint32) {
	if onReceive != nil {
		goWords := unsafe.Slice((*midi.Word)(words), wordCount)
		onReceive(timestampToTime(timestamp), midi.Entity(source), goWords)
	}
}

func (d *driver) Name() string {
	return "Core MIDI"
}

func (d *driver) Available() bool {
	return true
}

func (d *driver) Init(cfg *midi.DriverConfig) error {
	// TODO: only allow to be called once, error on duplicate

	d.initEpoch()

	onReceive = cfg.ReceiveHandler

	result := C.init()
	if result != 0 {
		return fmt.Errorf("C init failed with error %d", result)
	}

	return nil
}

func (d *driver) OpenInput(p midi.Entity) error {
	result := C.openInput(C.uint(p), unsafe.Pointer(p))
	if result != 0 {
		return fmt.Errorf("C init failed with error %d", result)
	}

	return nil
}

func (d *driver) OpenOutput(p midi.Entity) error {
	// no-op, there's no need to open an output on Core MIDI
	return nil
}

func (d *driver) Send(ts time.Time, dest midi.Entity, words []midi.Word) error {
	if len(words) == 0 {
		return nil
	}

	C.send(
		C.uint(dest),
		C.ulonglong(timeToTimestamp(ts)),
		(*C.uint)(unsafe.Pointer(&words[0])),
		C.uint(len(words)),
	)

	return nil
}

func (d *driver) Enumerate() (*midi.Node, error) {
	root := &midi.Node{
		Type: midi.Root,
	}

	deviceCount := midiGetNumberOfDevices()
	for i := 0; i < deviceCount; i++ {
		device := midiGetDevice(i)

		offline, err := midiObjectGetBoolProperty(device, propOffline)
		if err != nil {
			return nil, err
		} else if offline {
			continue
		}

		deviceNode := d.nodeForRef(midi.Device, device)

		entityCount := midiDeviceGetNumberOfEntities(device)
		for j := 0; j < entityCount; j++ {
			entity := midiDeviceGetEntity(device, j)
			entityNode := d.nodeForRef(midi.PortGroup, entity)

			sourceCount := midiEntityGetNumberOfSources(entity)
			for k := 0; k < sourceCount; k++ {
				source := midiEntityGetSource(entity, k)
				entityNode.Children = append(entityNode.Children, d.nodeForRef(midi.Input, source))
			}
			destCount := midiEntityGetNumberOfDestinations(entity)
			for k := 0; k < destCount; k++ {
				dest := midiEntityGetDestination(entity, k)
				entityNode.Children = append(entityNode.Children, d.nodeForRef(midi.Output, dest))
			}

			deviceNode.Children = append(deviceNode.Children, entityNode)
		}

		root.Children = append(root.Children, deviceNode)
	}

	return root, nil
}

func (d *driver) initEpoch() {
	var timebase C.mach_timebase_info_data_t
	C.mach_timebase_info(&timebase)

	goEpoch = time.Now()
	machEpoch = uint64(C.mach_absolute_time())
	machTimebaseNumer = uint64(timebase.numer)
	machTimebaseDenom = uint64(timebase.denom)
}

func (d *driver) nodeForRef(typ midi.NodeType, ref midiObjectRef) *midi.Node {
	manufacturer, _ := midiObjectGetStringProperty(ref, propManufacturer)
	model, _ := midiObjectGetStringProperty(ref, propModel)
	name, _ := midiObjectGetStringProperty(ref, propName)

	return &midi.Node{
		Type:         typ,
		Manufacturer: manufacturer,
		Model:        model,
		Name:         name,

		Driver: d,
		Entity: midi.Entity(ref),
	}
}
