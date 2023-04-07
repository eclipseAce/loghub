package msg

import "encoding/json"

type MsgFlags uint64

const (
	MsgAttr_Tx = 1
)

func NewMsgFlags(attr uint8, ds uint8, sn uint32) MsgFlags {
	var f MsgFlags
	f |= MsgFlags(ds) << 56
	f |= MsgFlags(attr) << 48
	f |= MsgFlags(sn)
	return f
}

func (f MsgFlags) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		TX bool
		DS uint8
		SN uint32
	}{
		TX: (MsgAttr_Tx<<48)&f > 0,
		DS: uint8(f >> 56),
		SN: uint32(f),
	})
}
