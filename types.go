package midi

type Timestamp uint64

type Word uint32

type Protocol byte

const (
	MIDI1 = Protocol(1)
	MIDI2 = Protocol(2)
)
