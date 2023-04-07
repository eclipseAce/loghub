package msg

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

type Msg struct {
	Raw       []byte
	Flags     MsgFlags
	Timestamp time.Time
	MsgID     uint16
	MsgSN     uint16
	SimNo     string
	Version   int16
	Encrypted bool
	PartTotal uint16
	PartIndex uint16
	Body      any
	Warnings  []string
}

const (
	MsgAttr_Tx = 1
)

var (
	ErrEmptyMsg = errors.New("empty msg")
	ErrBadMsg   = errors.New("bad msg")
)

func Decode(raw []byte, flags MsgFlags, timestamp time.Time) (*Msg, error) {
	m := &Msg{Flags: flags, Timestamp: timestamp, Raw: raw, Warnings: make([]string, 0)}

	if len(raw) < 2 || raw[0] != 0x7E || raw[len(raw)-1] != 0x7E {
		return nil, ErrBadMsg
	}
	if len(raw) == 2 {
		return nil, ErrEmptyMsg
	}

	// unescape
	buf := new(bytes.Buffer)
	for i := 1; i+1 < len(raw); i++ {
		b1, b2 := raw[i], raw[i+1]
		switch {
		case b1 == 0x7D && b2 == 0x01:
			buf.WriteByte(0x7D)
			i++
		case b1 == 0x7D && b2 == 0x02:
			buf.WriteByte(0x7E)
			i++
		default:
			buf.WriteByte(b1)
		}
	}

	// checksum
	var checksum uint8
	for _, b := range buf.Bytes() {
		checksum ^= b
	}
	if checksum != 0 {
		m.Warnings = append(m.Warnings, "bad checksum")
	}

	// read msg id
	if err := binary.Read(buf, binary.BigEndian, &m.MsgID); err != nil {
		return nil, fmt.Errorf("invalid msgId: %w", err)
	}

	// read msg attributes
	var attribute uint16
	if err := binary.Read(buf, binary.BigEndian, &attribute); err != nil {
		return nil, fmt.Errorf("invalid msgAttr: %w", err)
	}

	// read encrypt bits
	m.Encrypted = (attribute & 0x0400) != 0

	// read version
	if (attribute & 0x4000) != 0 {
		var version uint8
		if err := binary.Read(buf, binary.BigEndian, &version); err != nil {
			return nil, fmt.Errorf("invalid version: %w", err)
		}
		m.Version = int16(version)
	} else {
		m.Version = -1
	}

	// read simNo
	simNoData := make([]byte, 6)
	if m.Version != -1 {
		simNoData = make([]byte, 10)
	}
	if err := binary.Read(buf, binary.BigEndian, simNoData); err != nil {
		return nil, fmt.Errorf("invalid simNo: %w", err)
	}
	m.SimNo = strings.TrimLeft(hex.EncodeToString(simNoData), "0")

	// read msg sn
	if err := binary.Read(buf, binary.BigEndian, &m.MsgSN); err != nil {
		return nil, fmt.Errorf("invalid msgSn: %w", err)
	}

	// read split info
	if (attribute & 0x2000) != 0 {
		if err := binary.Read(buf, binary.BigEndian, &m.PartTotal); err != nil {
			return nil, fmt.Errorf("invalid msgPartTotal: %w", err)
		}
		if err := binary.Read(buf, binary.BigEndian, &m.PartIndex); err != nil {
			return nil, fmt.Errorf("invalid msgPartIndex: %w", err)
		}
	} else {
		m.PartTotal = 1
		m.PartIndex = 1
	}

	// read msg body
	remain := buf.Len()
	if int(attribute&0x03FF) != remain-1 {
		m.Warnings = append(m.Warnings, "bad body length")
	}
	var body []byte
	if remain > 1 {
		body = make([]byte, remain-1)
		if err := binary.Read(buf, binary.BigEndian, body); err != nil {
			return nil, fmt.Errorf("invalid msgBody: %w", err)
		}
	} else {
		body = []byte{}
		if remain == 0 {
			m.Warnings = append(m.Warnings, "missing checksum")
		}
	}

	var err error
	switch m.MsgID {
	case 0x0200:
		m.Body, err = DecodeBody_0200(body)
	default:
		m.Body = body
	}
	if err != nil {
		m.Warnings = append(m.Warnings, "bad msg body")
		m.Body = body
	}

	// last byte is checksum, ignore
	return m, nil
}

var logPattern = regexp.MustCompile(`^(?P<timestamp>\d{14}) (?P<xfer>Rx|Tx) (?P<payload>[a-f0-9]+)$`)

func DecodeLog(log string, ds uint8, sn uint32) (*Msg, error) {
	matches := logPattern.FindStringSubmatch(log)
	if matches == nil {
		return nil, errors.New("invalid log format")
	}
	fields := make(map[string]string)
	for i, name := range logPattern.SubexpNames() {
		if i != 0 && name != "" {
			fields[name] = matches[i]
		}
	}
	timestamp, err := time.ParseInLocation("20060102150405", fields["timestamp"], time.Local)
	if err != nil {
		return nil, fmt.Errorf("invalid log timestamp: %w", err)
	}
	payload, err := hex.DecodeString(fields["payload"])
	if err != nil {
		return nil, fmt.Errorf("invalid log payload: %w", err)
	}
	var attr uint8
	if fields["xfer"] == "Tx" {
		attr |= MsgAttr_Tx
	}
	flags := NewMsgFlags(attr, ds, sn)
	return Decode(payload, flags, timestamp)
}

func DecodeEntry(key, val []byte) (*Msg, error) {
	mk, err := DecodeKey(key)
	if err != nil {
		return nil, err
	}
	return Decode(val, mk.Flags, mk.Timestamp)
}

func (m *Msg) Key() *MsgKey {
	return &MsgKey{SimNo: m.SimNo, Timestamp: m.Timestamp, Flags: m.Flags}
}

func (m *Msg) Encode() (key, val []byte, err error) {
	key, err = m.Key().Encode()
	if err != nil {
		return nil, nil, err
	}
	val = m.Raw
	return
}
