package midi2

const (
	MsgTypeUtility = iota << 28
	MsgTypeSystem
	MsgTypeMIDIv1
	MsgTypeData
	MsgTypeMIDIv2
)
