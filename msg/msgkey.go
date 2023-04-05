package msg

import (
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
	if len(b) != 10+8+8 {
		return nil, errors.New("invalid key bytes")
	}
	return &MsgKey{
		SimNo:     strings.TrimLeft(hex.EncodeToString(b[:10]), "0"),
		Timestamp: time.Unix(int64(binary.BigEndian.Uint64(b[10:10+8])), 0),
		SN:        binary.BigEndian.Uint64(b[10+8:]),
	}, nil
}

func (mk *MsgKey) Encode() ([]byte, error) {
	simNo, err := hex.DecodeString(strings.Repeat("0", 10*2-len(mk.SimNo)) + mk.SimNo)
	if err != nil {
		return nil, fmt.Errorf("invalid simNo: %w", err)
	}
	b := make([]byte, 10+8+8)
	copy(b, simNo)
	binary.BigEndian.PutUint64(b[10:], uint64(mk.Timestamp.UTC().Unix()))
	binary.BigEndian.PutUint64(b[10+8:], mk.SN)
	return b, nil
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
