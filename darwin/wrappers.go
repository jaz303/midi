//go:build darwin

package darwin

// #cgo CFLAGS: -x objective-c
// #cgo LDFLAGS: -framework CoreMIDI
// #include <CoreMIDI/MIDIServices.h>
import "C"
import (
	"fmt"
	"sync"
)

const (
	propName = iota
	propModel
	propManufacturer
	propUniqueID
	propDeviceID
	propDisplayName
	propProtocolID
	propOffline
	propPrivate
)

var properties []C.CFStringRef

var initProperties = sync.OnceFunc(func() {
	properties = []C.CFStringRef{
		C.kMIDIPropertyName,
		C.kMIDIPropertyModel,
		C.kMIDIPropertyManufacturer,
		C.kMIDIPropertyUniqueID,
		C.kMIDIPropertyDeviceID,
		C.kMIDIPropertyDisplayName,
		C.kMIDIPropertyProtocolID,
		C.kMIDIPropertyOffline,
		C.kMIDIPropertyPrivate,
	}
})

func getProperty(p int) C.CFStringRef {
	initProperties()
	return properties[p]
}

type midiObjectRef uint32

func midiGetNumberOfDestinations() int {
	return int(C.MIDIGetNumberOfDestinations())
}

func midiGetNumberOfDevices() int {
	return int(C.MIDIGetNumberOfDevices())
}

func midiGetDevice(index int) midiObjectRef {
	return midiObjectRef(C.MIDIGetDevice(C.ulong(index)))
}

func midiDeviceGetNumberOfEntities(ref midiObjectRef) int {
	return int(C.MIDIDeviceGetNumberOfEntities(C.uint(ref)))
}

func midiDeviceGetEntity(ref midiObjectRef, index int) midiObjectRef {
	return midiObjectRef(C.MIDIDeviceGetEntity(C.uint(ref), C.ulong(index)))
}

func midiEntityGetNumberOfSources(ref midiObjectRef) int {
	return int(C.MIDIEntityGetNumberOfSources(C.uint(ref)))
}

func midiEntityGetSource(ref midiObjectRef, index int) midiObjectRef {
	return midiObjectRef(C.MIDIEntityGetSource(C.uint(ref), C.ulong(index)))
}

func midiEntityGetNumberOfDestinations(ref midiObjectRef) int {
	return int(C.MIDIEntityGetNumberOfDestinations(C.uint(ref)))
}

func midiEntityGetDestination(ref midiObjectRef, index int) midiObjectRef {
	return midiObjectRef(C.MIDIEntityGetDestination(C.uint(ref), C.ulong(index)))
}

func midiObjectGetStringProperty(ref midiObjectRef, property int) (string, error) {
	var val C.CFStringRef
	status := C.MIDIObjectGetStringProperty(C.uint(ref), getProperty(property), &val)
	if status != 0 {
		return "", fmt.Errorf("get property failed (%d)", status)
	}
	ptr := C.CFStringGetCStringPtr(val, C.kCFStringEncodingASCII)
	return C.GoString(ptr), nil
}

func midiObjectGetIntegerProperty(ref midiObjectRef, property int) (int, error) {
	var val C.int
	status := C.MIDIObjectGetIntegerProperty(C.uint(ref), getProperty(property), &val)
	if status != 0 {
		return 0, fmt.Errorf("get property failed (%d)", status)
	}
	return int(val), nil
}

func midiObjectGetBoolProperty(ref midiObjectRef, property int) (bool, error) {
	val, err := midiObjectGetIntegerProperty(ref, property)
	if err != nil {
		return false, err
	}
	return val == 1, nil
}
