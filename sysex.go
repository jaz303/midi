package midi

const (
	msgTypeData = (1 << 28)

	packetBytesShift = 16
	statusShift      = 20

	sysExStatusMask = 0xF << statusShift
	sysExComplete   = 0
	sysExStart      = 1 << statusShift
	sysExContinue   = 2 << statusShift
	sysExEnd        = 3 << statusShift
)

func SysExV1ToUMP(dst []Word, sysEx []byte) []Word {
	sysEx = sysEx[1 : len(sysEx)-1] // strip 0xF0 and 0xF7, UMP doesn't require them

	var status Word = sysExStart
	for len(sysEx) > 0 {
		bytesInPacket := min(6, len(sysEx))

		var w1 Word = msgTypeData | status | Word(bytesInPacket<<16)
		var w2 Word = 0

		switch bytesInPacket {
		case 6:
			w2 |= Word(sysEx[5])
			fallthrough
		case 5:
			w2 |= Word(sysEx[4]) << 8
			fallthrough
		case 4:
			w2 |= Word(sysEx[3]) << 16
			fallthrough
		case 3:
			w2 |= Word(sysEx[2]) << 24
			fallthrough
		case 2:
			w1 |= Word(sysEx[1])
			fallthrough
		case 1:
			w1 |= Word(sysEx[0]) << 8
		}

		dst = append(dst, w1, w2)

		status = sysExContinue
		sysEx = sysEx[bytesInPacket:]
	}

	if len(dst) == 2 {
		// Full message is in single packet
		dst[0] &^= sysExStatusMask
	} else if len(dst) > 2 {
		// Set end status on final packet
		// This is a bit hacky; the "correct" approach is to clear the
		// status field to 0x0 then set it to sysExEnd. However, sysExEnd
		// is a superset of the bits set sysExContinue, so just OR'ing it
		// in is fine.
		dst[len(dst)-2] |= sysExEnd
	}

	return dst
}
