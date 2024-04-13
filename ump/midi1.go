package ump

const (
	Clock         = Word(MsgTypeSystem | (0xF8 << 16))
	Start         = Word(MsgTypeSystem | (0xFA << 16))
	Continue      = Word(MsgTypeSystem | (0xFB << 16))
	Stop          = Word(MsgTypeSystem | (0xFC << 16))
	ActiveSensing = Word(MsgTypeSystem | (0xFE << 16))
	Reset         = Word(MsgTypeSystem | (0xFF << 16))
)

const (
	noteOff         = MsgTypeMIDIv1 | (0b1000 << 20)
	noteOn          = MsgTypeMIDIv1 | (0b1001 << 20)
	polyPressure    = MsgTypeMIDIv1 | (0b1010 << 20)
	controlChange   = MsgTypeMIDIv1 | (0b1011 << 20)
	programChange   = MsgTypeMIDIv1 | (0b1100 << 20)
	channelPressure = MsgTypeMIDIv1 | (0b1101 << 20)
	pitchBend       = MsgTypeMIDIv1 | (0b1110 << 20)

	channelShift = 16
)

func NoteOff(channel uint8, note int8, velocity int8) Word {
	return noteOff |
		(Word(channel&0x0F) << channelShift) |
		(Word(note) << 8) |
		Word(velocity)
}

func NoteOn(channel uint8, note int8, velocity int8) Word {
	return noteOn |
		(Word(channel&0x0F) << channelShift) |
		(Word(note) << 8) |
		Word(velocity)
}

// TODO: poly pressure

func ControlChange(channel uint8, controller, value int8) Word {
	return controlChange |
		(Word(channel&0x0F) << channelShift) |
		(Word(controller) << 8) |
		Word(value)
}

// TODO: program change
// TODO: channel pressure
// TODO: pitch bend
