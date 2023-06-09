package msg

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

const SimNoBytes = 10

type MsgKey struct {
	SimNo     string
	Timestamp time.Time
	DS        uint8
	TX        bool
	SN        uint32
	MsgID     uint16
	PartIndex uint16
	PartTotal uint16
}

type MsgKeyLayout struct {
	SimNo     [SimNoBytes]byte
	Timestamp uint64
	DS        uint8
	Flags     uint8
	Reserved  uint16
	SN        uint32
	MsgID     uint16
	PartIndex uint16
	PartTotal uint16
}

const (
	MsgKeyFlag_Tx = (1 << iota)
)

func DecodeKey(b []byte) (*MsgKey, error) {
	var s MsgKeyLayout
	if err := binary.Read(bytes.NewReader(b), binary.BigEndian, &s); err != nil {
		return nil, err
	}
	mk := &MsgKey{
		SimNo:     strings.TrimLeft(hex.EncodeToString(s.SimNo[:]), "0"),
		Timestamp: time.Unix(int64(s.Timestamp), 0),
		DS:        s.DS,
		TX:        (s.Flags & MsgKeyFlag_Tx) != 0,
		SN:        s.SN,
		MsgID:     s.MsgID,
		PartIndex: s.PartIndex,
		PartTotal: s.PartTotal,
	}
	return mk, nil
}

func (mk *MsgKey) Encode() ([]byte, error) {
	s := MsgKeyLayout{
		Timestamp: uint64(mk.Timestamp.UTC().Unix()),
		DS:        mk.DS,
		SN:        mk.SN,
		MsgID:     mk.MsgID,
		PartIndex: mk.PartIndex,
		PartTotal: mk.PartTotal,
	}
	if len(mk.SimNo) > SimNoBytes*2 {
		return nil, fmt.Errorf("simNo too long (>%d chars)", SimNoBytes*2)
	}
	simNo, err := hex.DecodeString(strings.Repeat("0", SimNoBytes*2-len(mk.SimNo)) + mk.SimNo)
	if err != nil {
		return nil, fmt.Errorf("simNo contains non-hex chars: %w", err)
	}
	if err != nil {
		return nil, err
	}
	copy(s.SimNo[:], simNo)
	if mk.TX {
		s.Flags |= MsgKeyFlag_Tx
	}
	buf := &bytes.Buffer{}
	_ = binary.Write(buf, binary.BigEndian, &s)
	return buf.Bytes(), nil
}
