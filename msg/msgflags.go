package msg

type MsgFlags uint64

func NewMsgFlags(attr uint8, ds uint8, sn uint32) MsgFlags {
	var f MsgFlags
	f |= MsgFlags(ds) << 56
	f |= MsgFlags(attr) << 48
	f |= MsgFlags(sn)
	return f
}
