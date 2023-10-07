package midi1

import (
	"github.com/jaz303/midi"
	m2 "github.com/jaz303/midi/midi2"
)

const (
	Clock         = midi.Word(m2.MsgTypeSystem | (0xF8 << 16))
	Start         = midi.Word(m2.MsgTypeSystem | (0xFA << 16))
	Continue      = midi.Word(m2.MsgTypeSystem | (0xFB << 16))
	Stop          = midi.Word(m2.MsgTypeSystem | (0xFC << 16))
	ActiveSensing = midi.Word(m2.MsgTypeSystem | (0xFE << 16))
	Reset         = midi.Word(m2.MsgTypeSystem | (0xFF << 16))
)

const (
	noteOff = m2.MsgTypeMIDIv1 | (0b1000 << 20)
	noteOn  = m2.MsgTypeMIDIv1 | (0b1001 << 20)

	channelShift = 16
)

func NoteOn(channel uint8, note uint8, velocity uint8) midi.Word {
	return noteOn |
		(midi.Word(channel&0x0F) << channelShift) |
		(midi.Word(note) << 8) |
		midi.Word(velocity)
}

func NoteOff(channel uint8, note uint8, velocity uint8) midi.Word {
	return noteOff |
		(midi.Word(channel&0x0F) << channelShift) |
		(midi.Word(note) << 8) |
		midi.Word(velocity)
}
