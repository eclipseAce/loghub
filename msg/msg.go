package msg

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"time"
)

type Msg struct {
	SN        uint64
	Raw       []byte
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

func Decode(raw []byte, timestamp time.Time, sn uint64) (*Msg, error) {
	m := &Msg{Timestamp: timestamp, SN: sn, Raw: raw, Warnings: make([]string, 0)}

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
		m.PartIndex = 0
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

var logPattern = regexp.MustCompile(
	`^(?P<timestamp>\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}) GpsDataService:\d+ - \([0-9A-F]+\)收到报文类型：\d+,报文内容：(?P<payload>[a-f0-9]+)$`,
)

func DecodeLog(log string, sn uint64) (*Msg, error) {
	matches := logPattern.FindStringSubmatch(log)
	if matches == nil {
		return nil, fmt.Errorf("invalid log: %s", log)
	}
	fields := make(map[string]string)
	for i, name := range logPattern.SubexpNames() {
		if i != 0 && name != "" {
			fields[name] = matches[i]
		}
	}
	timestamp, err := time.ParseInLocation("2006-01-02 15:04:05", fields["timestamp"], time.Local)
	if err != nil {
		return nil, fmt.Errorf("invalid log timestamp: %w", err)
	}
	payload, err := hex.DecodeString(fields["payload"])
	if err != nil {
		return nil, fmt.Errorf("invalid log payload: %w", err)
	}
	return Decode(payload, timestamp, sn)
}

func DecodeEntry(key, val []byte) (*Msg, error) {
	mk, err := DecodeKey(key)
	if err != nil {
		return nil, err
	}
	return Decode(val, mk.Timestamp, mk.SN)
}

func (m *Msg) Key() *MsgKey {
	return &MsgKey{SimNo: m.SimNo, Timestamp: m.Timestamp, SN: m.SN}
}

func (m *Msg) Encode() (key, val []byte, err error) {
	key, err = m.Key().Encode()
	if err != nil {
		return nil, nil, err
	}
	val = m.Raw
	return
}
