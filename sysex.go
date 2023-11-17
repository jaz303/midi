package midi

const (
	msgTypeData = (3 << 28)

	packetBytesShift = 16
	statusShift      = 20

	sysExStatusMask = 0xF << statusShift
	sysExComplete   = 0
	sysExStart      = 1 << statusShift
	sysExContinue   = 2 << statusShift
	sysExEnd        = 3 << statusShift
)

func SysExV1ToUMP(dst []Word, data []byte) []Word {
	data = data[1 : len(data)-1] // strip 0xF0 and 0xF7, UMP doesn't require them

	var status Word = sysExStart
	for len(data) > 0 {
		bytesInPacket := min(6, len(data))

		var w1 Word = msgTypeData | status | Word(bytesInPacket<<packetBytesShift)
		var w2 Word = 0

		switch bytesInPacket {
		case 6:
			w2 |= Word(data[5])
			fallthrough
		case 5:
			w2 |= Word(data[4]) << 8
			fallthrough
		case 4:
			w2 |= Word(data[3]) << 16
			fallthrough
		case 3:
			w2 |= Word(data[2]) << 24
			fallthrough
		case 2:
			w1 |= Word(data[1])
			fallthrough
		case 1:
			w1 |= Word(data[0]) << 8
		}

		dst = append(dst, w1, w2)

		status = sysExContinue
		data = data[bytesInPacket:]
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
