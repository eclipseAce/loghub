package msg

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"
)

var logPattern = regexp.MustCompile(
	`^(?P<timestamp>\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}) GpsDataService:\d+ - \([0-9A-F]+\)收到报文类型：\d+,报文内容：(?P<payload>[a-f0-9]+)$`,
)

type MsgKey struct {
	SimNo     string
	Timestamp time.Time
	SN        uint64
}

func DecodeKey(b []byte) (*MsgKey, error) {
	if len(b) < 16 {
		return nil, errors.New("invalid key bytes")
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
		return nil, err
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

type Msg struct {
	SN        uint64
	Raw       []byte
	Timestamp time.Time

	MsgID       uint16
	MsgSN       uint16
	SimNo       string
	Version     int16
	Encrypted   bool
	PartTotal   uint16
	PartIndex   uint16
	Body        []byte
	BadChecksum bool
	BadBodyLen  bool
}

func Decode(raw []byte, timestamp time.Time, sn uint64) (*Msg, error) {
	m := &Msg{Timestamp: timestamp, SN: sn, Raw: raw}

	// unescape
	buf := new(bytes.Buffer)
	for i := 1; i+1 < len(raw); i++ {
		b1, b2 := raw[i], raw[i+1]
		switch {
		case b1 == 0x7D && b2 == 0x01:
			buf.WriteByte(0x7E)
			i++
		case b1 == 0x7D && b2 == 0x02:
			buf.WriteByte(0x7D)
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
		m.BadChecksum = true
	}

	// read msg id
	if err := binary.Read(buf, binary.BigEndian, &m.MsgID); err != nil {
		return nil, err
	}

	// read msg attributes
	var attribute uint16
	if err := binary.Read(buf, binary.BigEndian, &attribute); err != nil {
		return nil, err
	}

	// read encrypt bits
	m.Encrypted = (attribute & 0x0400) != 0

	// read version
	if (attribute & 0x4000) != 0 {
		var version uint8
		if err := binary.Read(buf, binary.BigEndian, &version); err != nil {
			return nil, err
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
		return nil, err
	}
	m.SimNo = strings.TrimLeft(hex.EncodeToString(simNoData), "0")

	// read msg sn
	if err := binary.Read(buf, binary.BigEndian, &m.MsgSN); err != nil {
		return nil, err
	}

	// read split info
	if (attribute & 0x2000) != 0 {
		if err := binary.Read(buf, binary.BigEndian, &m.PartTotal); err != nil {
			return nil, err
		}
		if err := binary.Read(buf, binary.BigEndian, &m.PartIndex); err != nil {
			return nil, err
		}
	} else {
		m.PartTotal = 1
		m.PartIndex = 0
	}

	// read msg body
	remain := buf.Len()
	if int(attribute&0x03FF) != remain-1 {
		m.BadBodyLen = true
	}
	if remain > 1 {
		m.Body = make([]byte, remain-1)
		if err := binary.Read(buf, binary.BigEndian, m.Body); err != nil {
			return nil, err
		}
	} else {
		if remain == 0 {
			m.BadChecksum = true
		}
		m.Body = []byte{}
	}

	// last byte is checksum, ignore
	return m, nil
}

func DecodeLog(log string, sn uint64) (*Msg, error) {
	matches := logPattern.FindStringSubmatch(log)
	if matches == nil {
		return nil, fmt.Errorf("invalid message: %s", log)
	}
	fields := make(map[string]string)
	for i, name := range logPattern.SubexpNames() {
		if i != 0 && name != "" {
			fields[name] = matches[i]
		}
	}
	timestamp, err := time.ParseInLocation("2006-01-02 15:04:05", fields["timestamp"], time.Local)
	if err != nil {
		return nil, err
	}
	payload, err := hex.DecodeString(fields["payload"])
	if err != nil {
		return nil, err
	}
	return Decode(payload, timestamp, sn)
}

func DecodeEntry(key, val []byte) (*Msg, error) {
	mk, err := DecodeKey(key)
	if err != nil {
		return nil, err
	}
	if len(val) >= 2 && val[0] == 0x1F && val[1] == 0x8B {
		gzr, err := gzip.NewReader(bytes.NewReader(val))
		if err != nil {
			return nil, err
		}
		val, err = io.ReadAll(gzr)
		if err != nil {
			return nil, err
		}
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
	if len(val) >= 1024 {
		buf := &bytes.Buffer{}
		gzw := gzip.NewWriter(buf)
		if _, err := gzw.Write(m.Raw); err != nil {
			return nil, nil, err
		}
		if err := gzw.Close(); err != nil {
			return nil, nil, err
		}
		val = buf.Bytes()
	}
	return
}
