package midi

type Word uint32

type Protocol byte

const (
	MIDI1 = Protocol(1)
	MIDI2 = Protocol(2)
)
