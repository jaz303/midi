//go:build darwin

package darwin

// #cgo CFLAGS: -x objective-c
// #cgo CFLAGS: -Wint-to-void-pointer-cast
// #cgo LDFLAGS: -framework CoreMIDI
// #include "binding.h"
// #include <stdio.h>
import "C"

import (
	"errors"
	"fmt"
	"runtime"
	"time"
	"unsafe"

	"github.com/jaz303/midi"
)

var immediately time.Time

func init() {
	initTimebase()

	midi.Register(&midi.Stub{
		Name:      "Core MIDI",
		Available: true,
		CreateDriver: func() (midi.Driver, error) {
			d := new(driver)
			d.client = C.allocateClient()         // allocate C struct for client
			d.pinner.Pin(d)                       // pin driver so client C struct can reference it safely
			d.client.goDriver = unsafe.Pointer(d) // back reference from C client struct to the driver so event handler can relay events to correct driver instance
			d.onReceive = midi.NopHandler         // default event handler

			result := C.init(d.client)
			if result != 0 {
				C.shutdown(d.client)
				d.pinner.Unpin()
				return nil, fmt.Errorf("(cgo) init failed with error %d", result)
			}

			d.client.wasInit = 1

			return d, nil
		},
	})
}

type driver struct {
	client    *C.struct_client
	pinner    runtime.Pinner
	onReceive midi.ReceiveEventHandler
}

//export OnReceive
func OnReceive(driverPointer unsafe.Pointer, timestamp uint64, source unsafe.Pointer, words unsafe.Pointer, wordCount uint32) {
	driver := (*driver)(driverPointer)
	goWords := unsafe.Slice((*midi.Word)(words), wordCount)
	driver.onReceive(timestampToTime(timestamp), midi.Entity(source), goWords)
}

func (d *driver) Close() error {
	d.onReceive = midi.NopHandler
	C.shutdown(d.client)
	d.pinner.Unpin()
	return nil
}

func (d *driver) SetReceiveHandler(hnd midi.ReceiveEventHandler) {
	if hnd == nil {
		panic(errors.New("receive handler cannot be nil"))
	}
	d.onReceive = hnd
}

func (d *driver) OpenInput(p midi.Entity) error {
	result := C.openInput(d.client, C.uint(p))
	if result != 0 {
		return fmt.Errorf("(cgo) open input failed with error %d", result)
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

	res := C.send(
		d.client,
		C.uint(dest),
		C.ulonglong(timeToTimestamp(ts)),
		(*C.uint)(unsafe.Pointer(&words[0])),
		C.uint(len(words)),
	)

	if res != 0 {
		return fmt.Errorf("send to destination %d failed with OSStatus=%d", dest, res)
	}

	return nil
}

func (d *driver) SendSysEx(dest midi.Entity, data []midi.Word) error {
	return d.Send(immediately, dest, data)
}

func (d *driver) SendSysExV1(dest midi.Entity, data []byte) error {
	return d.Send(immediately, dest, midi.SysExV1ToUMP(nil, data))
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
