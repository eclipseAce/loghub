package msg

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
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
	return &MsgKey{
		SimNo:     hex.EncodeToString(b[:off]),
		Timestamp: time.Unix(int64(binary.BigEndian.Uint64(b[off:off+8])), 0),
		SN:        binary.BigEndian.Uint64(b[off+8 : off+8+8]),
	}, nil
}

func (mk *MsgKey) Encode() ([]byte, error) {
	simNo, err := hex.DecodeString(strings.Repeat("0", len(mk.SimNo)%2) + mk.SimNo)
	if err != nil {
		return nil, fmt.Errorf("invalid simNo: %w", err)
	}
	buf := &bytes.Buffer{}
	buf.Write(simNo)
	binary.Write(buf, binary.BigEndian, uint64(mk.Timestamp.UTC().Unix()))
	binary.Write(buf, binary.BigEndian, mk.SN)
	return buf.Bytes(), nil
}

func EncodeKeyRange(simNo string, since, until time.Time) (sinceKey, untilKey []byte, err error) {
	sk := &MsgKey{SimNo: simNo, Timestamp: since, SN: uint64(0)}
	uk := &MsgKey{SimNo: simNo, Timestamp: until, SN: ^uint64(0)}
	if sinceKey, err = sk.Encode(); err != nil {
		return nil, nil, err
	}
	if untilKey, err = uk.Encode(); err != nil {
		return nil, nil, err
	}
	return
}
