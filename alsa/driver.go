//go:build linux

package alsa

// #cgo LDFLAGS: -lasound
// #include "binding.h"
import "C"

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"
	"unsafe"

	"github.com/jaz303/midi"
	"github.com/jaz303/midi/ump"
)

const (
	driverName = "alsa"
)

var ErrALSA = errors.New("failed with status")

func alsaError(funcName string, exitCode C.int) error {
	return fmt.Errorf("%s() %w %d", funcName, ErrALSA, exitCode)
}

// MARK: Init

func init() {
	midi.Register(&midi.Stub{
		Name:      driverName,
		Available: true,
		CreateDriver: func(clientName string) (midi.Driver, error) {
			d := &driver{
				queueID: -1,
			}

			err := C.snd_seq_open(&d.seq, C.CString("default"), C.SND_SEQ_OPEN_DUPLEX, 0)
			if err < 0 {
				return nil, alsaError("snd_seq_open", err)
			}

			C.snd_seq_set_client_name(d.seq, C.CString(clientName))
			d.clientID = C.uchar(C.snd_seq_client_id(d.seq))

			d.queueID = C.snd_seq_alloc_named_queue(d.seq, C.CString("default"))
			if d.queueID < 0 {
				d.Close()
				return nil, alsaError("snd_seq_alloc_named_queue", d.queueID)
			}

			d.onReceive = midi.NopHandler

			err = C.snd_seq_control_queue(d.seq, d.queueID, C.SND_SEQ_EVENT_START, 0, nil)
			if err < 0 {
				return nil, alsaError("snd_seq_control_queue", err)
			}

			go d.readLoop()

			// d.client = C.allocateClient()         // allocate C struct for client
			// d.pinner.Pin(d)                       // pin driver so client C struct can reference it safely
			// d.client.goDriver = unsafe.Pointer(d) // back reference from C client struct to the driver so event handler can relay events to correct driver instance
			// d.onReceive = midi.NopHandler         // default event handler

			// result := C.init(d.client)
			// if result != 0 {
			// 	C.shutdown(d.client)
			// 	d.pinner.Unpin()
			// 	return nil, fmt.Errorf("(cgo) init failed with error %d", result)
			// }

			//d.client.wasInit = 1

			return d, nil
		},
	})
}

type driver struct {
	seq       *C.snd_seq_t
	queueID   C.int
	clientID  C.uchar
	onReceive midi.ReceiveEventHandler
}

func (d *driver) Name() string {
	return driverName
}

func (d *driver) Close() error {
	// TODO: free queue
	C.snd_seq_close(d.seq)
	return nil
}

func (d *driver) SetReceiveHandler(handler midi.ReceiveEventHandler) {
	if handler == nil {
		handler = midi.NopHandler
	}
	d.onReceive = handler
}

// MARK: Open Input

func (d *driver) OpenInput(ent midi.Entity) error {
	portID := C.snd_seq_create_simple_port(d.seq,
		C.CString(fmt.Sprintf("output %d", ent)),
		C.SND_SEQ_PORT_CAP_WRITE|C.SND_SEQ_PORT_CAP_SUBS_WRITE,
		C.SND_SEQ_PORT_TYPE_MIDI_GENERIC,
	)

	if portID < 0 {
		return alsaError("snd_seq_create_simple_port", portID)
	}

	var sender, dest C.snd_seq_addr_t
	sender.client = entityClientID(ent)
	sender.port = entityPortID(ent)
	dest.client = d.clientID
	dest.port = C.uchar(portID)

	subscribePtr := (*C.snd_seq_port_subscribe_t)(C.malloc(C.snd_seq_port_subscribe_sizeof()))

	C.snd_seq_port_subscribe_set_sender(subscribePtr, &sender)
	C.snd_seq_port_subscribe_set_dest(subscribePtr, &dest)
	C.snd_seq_port_subscribe_set_time_update(subscribePtr, 1)
	C.snd_seq_port_subscribe_set_time_real(subscribePtr, 1)
	C.snd_seq_port_subscribe_set_queue(subscribePtr, d.queueID)

	if status := C.snd_seq_subscribe_port(d.seq, subscribePtr); status < 0 {
		C.free(unsafe.Pointer(subscribePtr))
		C.snd_seq_delete_port(d.seq, portID)
		return alsaError("snd_seq_subscribe_port", status)
	}

	return nil
}

// MARK: Open Output

func (d *driver) OpenOutput(midi.Entity) error {
	return nil
}

func (d *driver) Send(time.Time, midi.Entity, []ump.Word) error {
	return nil
}

func (d *driver) SendSysEx(midi.Entity, []ump.Word) error {
	return nil
}

func (d *driver) SendSysExV1(midi.Entity, []byte) error {
	return nil
}

// MARK: Enumerate

func (d *driver) Enumerate() (*midi.Node, error) {
	root := &midi.Node{
		Type: midi.Root,
	}

	clientInfoPtr := C.malloc(C.snd_seq_client_info_sizeof())
	if clientInfoPtr == nil {
		return nil, fmt.Errorf("failed to allocate client info")
	}

	defer C.free(clientInfoPtr)

	portInfoPtr := C.malloc(C.snd_seq_port_info_sizeof())
	if portInfoPtr == nil {
		return nil, fmt.Errorf("failed to allocate port info")
	}

	defer C.free(portInfoPtr)

	clientInfo := (*C.snd_seq_client_info_t)(clientInfoPtr)
	portInfo := (*C.snd_seq_port_info_t)(portInfoPtr)

	C.snd_seq_client_info_set_client(clientInfo, -1)
	for C.snd_seq_query_next_client(d.seq, clientInfo) >= 0 {
		clientID := C.snd_seq_client_info_get_client(clientInfo)
		if clientID == C.snd_seq_client_id(d.seq) {
			continue
		}

		device := &midi.Node{
			Type:         midi.Device,
			Manufacturer: "",
			Model:        "",
			Name:         C.GoString(C.snd_seq_client_info_get_name(clientInfo)),
			Driver:       d,
			Entity:       makeClientEntity(clientID),
		}

		// clientName :=
		// clientPortCount := C.snd_seq_client_info_get_num_ports(clientInfo)
		// clientType := C.snd_seq_client_info_get_type(clientInfo)

		// log.Printf("Client: %d name=%s type=%d ports=%d", clientID, clientName, clientType, clientPortCount)

		C.snd_seq_port_info_set_client(portInfo, clientID)
		C.snd_seq_port_info_set_port(portInfo, -1)
		for C.snd_seq_query_next_port(d.seq, portInfo) >= 0 {
			portID := C.snd_seq_port_info_get_port(portInfo)
			portName := C.GoString(C.snd_seq_port_info_get_name(portInfo))
			portType := C.snd_seq_port_info_get_type(portInfo)
			portCaps := C.snd_seq_port_info_get_capability(portInfo)

			portGroup := &midi.Node{
				Type:         midi.PortGroup,
				Manufacturer: "",
				Model:        "",
				Name:         portName,
				Driver:       d,
				Entity:       makePortEntity(clientID, portID),
				Metadata: map[string]any{
					fmt.Sprintf("%s:type", driverName): getPortTypeString(int(portType)),
					fmt.Sprintf("%s:caps", driverName): getPortCapString(int(portCaps)),
				},
			}

			isInput := portCaps&C.SND_SEQ_PORT_CAP_READ > 0
			if isInput {
				portGroup.Children = append(portGroup.Children, &midi.Node{
					Type:         midi.Input,
					Manufacturer: "",
					Model:        "",
					Name:         fmt.Sprintf("%s input", portName),
					Driver:       d,
					Entity:       makePortEntity(clientID, portID),
				})
			}

			isOutput := portCaps&C.SND_SEQ_PORT_CAP_WRITE > 0
			if isOutput {
				portGroup.Children = append(portGroup.Children, &midi.Node{
					Type:         midi.Output,
					Manufacturer: "",
					Model:        "",
					Name:         fmt.Sprintf("%s output", portName),
					Driver:       d,
					Entity:       makePortEntity(clientID, portID),
				})
			}

			device.Children = append(device.Children, portGroup)
		}

		root.Children = append(root.Children, device)
	}

	return root, nil
}

// MARK: Read Loop

func (d *driver) readLoop() {
	var event *C.snd_seq_event_t
	var words []ump.Word

	for {
		var batchTimestamp time.Time
		var batchEntity midi.Entity

		remain := C.snd_seq_event_input(d.seq, &event)
		if remain < 0 {
			// TODO: I think there are some error codes here that only constitute
			// a warning (e.g. underrun)
			log.Printf("read loop returned error %d, exiting", remain)
			return
		}

		batchTimestamp = time.Time{} // TODO
		batchEntity = makePortEntityFromEvent(event)
		words = append(words, parseEvent(event))

		for remain > 1 {
			remain = C.snd_seq_event_input(d.seq, &event)
			if remain < 0 {
				// TODO: I think there are some error codes here that only constitute
				// a warning (e.g. underrun)
				log.Printf("read loop returned error %d, exiting", remain)
				return
			}

			thisTimestamp := time.Time{}
			thisEntity := makePortEntityFromEvent(event)
			if thisTimestamp != batchTimestamp || thisEntity != batchEntity {
				d.onReceive(batchTimestamp, batchEntity, words)
				batchTimestamp = thisTimestamp
				batchEntity = thisEntity
				words = words[:0]
			}

			words = append(words, parseEvent(event))
		}

		d.onReceive(batchTimestamp, batchEntity, words)
		words = words[:0]
	}
}

func parseEvent(evt *C.snd_seq_event_t) ump.Word {
	switch evt._type {
	case C.SND_SEQ_EVENT_NOTEON:
		return ump.NoteOn(byte(evt.data[0]), int8(evt.data[1]), int8(evt.data[2]))
	case C.SND_SEQ_EVENT_NOTEOFF:
		return ump.NoteOff(byte(evt.data[0]), int8(evt.data[1]), int8(evt.data[2]))
	case C.SND_SEQ_EVENT_CONTROLLER:
		// TODO(jason): are there endianness issues here?
		return ump.ControlChange(byte(evt.data[0]), int8(evt.data[4]), int8(evt.data[8]))
	case C.SND_SEQ_EVENT_PGMCHANGE:
		// TODO
	case C.SND_SEQ_EVENT_CHANPRESS:
		// TODO
	case C.SND_SEQ_EVENT_PITCHBEND:
		// TODO
	}

	return 0
}

var portCapBits = []string{
	"READ",
	"WRITE",
	"SYNC_READ",
	"SYNC_WRITE",
	"DUPLEX",
	"SUBS_READ",
	"SUBS_WRITE",
	"NO_EXPORT",
	"INACTIVE",
	"UMP_ENDPOINT",
}

func getPortCapString(portCaps int) string {
	return stringFromBitmask(portCaps, portCapBits)
}

var portTypeBits = []string{
	"SPECIFIC",
	"GENERIC",
	"GM",
	"GS",
	"XG",
	"MT32",
	"GM2",
	"UMP",
	"",
	"",
	"SYNTH",
	"DIRECT_SAMPLE",
	"SAMPLE",
	"",
	"",
	"",
	"HARDWARE",
	"SOFTWARE",
	"SYNTHESIZER",
	"PORT",
	"APPLICATION",
}

func getPortTypeString(portType int) string {
	return stringFromBitmask(portType, portTypeBits)
}

func stringFromBitmask(val int, bitNames []string) string {
	out := strings.Builder{}
	for i, n := range bitNames {
		if n == "" {
			continue
		}
		if val&(1<<i) > 0 {
			out.WriteString(":" + n)
		}
	}
	if out.Len() > 0 {
		out.WriteRune(':')
	}
	return out.String()
}

func makeClientEntity(clientID C.int) midi.Entity {
	return midi.Entity(uint32(clientID) << 8)
}

func makePortEntity(clientID C.int, portID C.int) midi.Entity {
	return midi.Entity((uint32(clientID) << 8) | uint32(portID))
}

func makePortEntityFromEvent(evt *C.snd_seq_event_t) midi.Entity {
	return midi.Entity((uint32(evt.source.client) << 8) | (uint32(evt.source.port)))
}

func entityClientID(e midi.Entity) C.uchar { return C.uchar(uint32(e) >> 8) }
func entityPortID(e midi.Entity) C.uchar   { return C.uchar(e) }
