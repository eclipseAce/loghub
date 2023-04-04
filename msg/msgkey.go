package msg

import (
	"bytes"
	"encoding/binary"
	"errors"
	"time"
)

type MsgKey struct {
	SimNo     string
	Timestamp time.Time
	SN        uint64
}

func DecodeKey(b []byte) (*MsgKey, error) {
	if len(b) < 16 {
		return nil, errors.New("invalid key length")
	}
	off := len(b) - 16
	simNo := string(b[:off])
	timestamp := binary.BigEndian.Uint64(b[off : off+8])
	sn := binary.BigEndian.Uint64(b[off+8 : off+8+8])
	return &MsgKey{SimNo: simNo, Timestamp: time.Unix(int64(timestamp), 0), SN: sn}, nil
}

func (mk *MsgKey) Encode() []byte {
	buf := &bytes.Buffer{}
	buf.Write([]byte(mk.SimNo))
	binary.Write(buf, binary.BigEndian, uint64(mk.Timestamp.UTC().Unix()))
	binary.Write(buf, binary.BigEndian, mk.SN)
	return buf.Bytes()
}

func EncodeKeyRange(simNo string, since, until time.Time) (sinceKey, untilKey []byte) {
	sinceKey = (&MsgKey{SimNo: simNo, Timestamp: since, SN: uint64(0)}).Encode()
	untilKey = (&MsgKey{SimNo: simNo, Timestamp: until, SN: ^uint64(0)}).Encode()
	return
}
