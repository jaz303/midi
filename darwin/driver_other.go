//go:build !darwin

package darwin

import (
	"github.com/jaz303/midi"
)

func init() {
	midi.Register(&midi.Stub{
		Name:      "Core MIDI",
		Available: false,
		CreateDriver: func() (midi.Driver, error) {
			return nil, midi.ErrDriverNotAvailable
		},
	})
}
