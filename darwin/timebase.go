package darwin

// #cgo CFLAGS: -x objective-c
// #cgo CFLAGS: -Wint-to-void-pointer-cast
// #cgo LDFLAGS: -framework CoreMIDI
// #include "binding.h"
// #include <stdio.h>
import "C"

import "time"

// These values tie the system local time to Core MIDI's
// internal time.
var (
	// Go time at driver init
	goEpoch time.Time

	// mach_absolute_time() at init
	machEpoch uint64

	// mach timebase for conversion
	// Retrieved from mach_timebase_info()
	machTimebaseNumer uint64
	machTimebaseDenom uint64
)

func initTimebase() {
	var timebase C.mach_timebase_info_data_t
	C.mach_timebase_info(&timebase)

	goEpoch = time.Now()
	machEpoch = uint64(C.mach_absolute_time())
	machTimebaseNumer = uint64(timebase.numer)
	machTimebaseDenom = uint64(timebase.denom)
}

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
