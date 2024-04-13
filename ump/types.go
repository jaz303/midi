package ump

type Word uint32

const (
	MsgTypeUtility = iota << 28
	MsgTypeSystem
	MsgTypeMIDIv1
	MsgTypeData
	MsgTypeMIDIv2
)
