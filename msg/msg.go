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
	"strconv"
	"strings"
	"time"

	"github.com/dgraph-io/badger/v4"
)

type Msg struct {
	SN           uint64
	SimNo        string
	MsgID        uint16
	MsgSN        uint16
	MsgVersion   int16
	MsgEncrypted bool
	PartTotal    uint16
	PartIndex    uint16
	Body         []byte
	BadChecksum  bool
	BadBodyLen   bool
	Raw          []byte
	Timestamp    time.Time
}

var msgLogRegex = regexp.MustCompile(
	`^(?P<timestamp>\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}) GpsDataService:\d+ - \([0-9A-F]+\)收到报文类型：\d+,报文内容：(?P<payload>[a-f0-9]+)$`,
)

const keyTimestampLayout = "20060102150405"

func NewMsgKey(simNo string, timestamp time.Time, sn uint64) []byte {
	return []byte(fmt.Sprintf("%s:%s:%08x", simNo, timestamp.Format(keyTimestampLayout), sn))
}

func ParseMsgKey(key []byte) (simNo string, timestamp time.Time, sn uint64, err error) {
	keyParts := strings.Split(string(key), ":")
	if len(keyParts) != 3 {
		err = errors.New("invalid key")
		return
	}
	timestamp, err = time.Parse(keyTimestampLayout, keyParts[1])
	if err != nil {
		return
	}
	sn, err = strconv.ParseUint(keyParts[2], 16, 0)
	if err != nil {
		return
	}
	return
}

func NewMsgFromLog(msgLog string, nextSeq func() (uint64, error)) (*Msg, error) {
	matches := msgLogRegex.FindStringSubmatch(msgLog)
	if matches == nil {
		return nil, fmt.Errorf("invalid message: %s", msgLog)
	}
	fields := make(map[string]string)
	for i, name := range msgLogRegex.SubexpNames() {
		if i != 0 && name != "" {
			fields[name] = matches[i]
		}
	}
	timestamp, err := time.Parse("2006-01-02 15:04:05", fields["timestamp"])
	if err != nil {
		return nil, err
	}
	payload, err := hex.DecodeString(fields["payload"])
	if err != nil {
		return nil, err
	}
	sn, err := nextSeq()
	if err != nil {
		return nil, fmt.Errorf("seq next: %w", err)
	}
	m := &Msg{SN: sn, Raw: payload, Timestamp: timestamp}
	if err := m.decode(); err != nil {
		return nil, fmt.Errorf("bad payload (%s) %w", fields["payload"], err)
	}
	return m, nil
}

func NewMsgFromItem(item *badger.Item) (m *Msg, err error) {
	m = &Msg{}
	_, m.Timestamp, m.SN, err = ParseMsgKey(item.Key())
	if err != nil {
		return nil, err
	}
	if err := item.Value(func(val []byte) error {
		gz, err := gzip.NewReader(bytes.NewReader(val))
		if err != nil {
			return err
		}
		m.Raw, err = io.ReadAll(gz)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	if err := m.decode(); err != nil {
		return nil, err
	}
	return m, nil
}

func (m *Msg) ToEntry() (*badger.Entry, error) {
	val := &bytes.Buffer{}
	gz := gzip.NewWriter(val)
	if _, err := gz.Write(m.Raw); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}
	return badger.NewEntry(NewMsgKey(m.SimNo, m.Timestamp, m.SN), val.Bytes()), nil
}

func (m *Msg) decode() error {
	buf := new(bytes.Buffer)
	for i := 1; i+1 < len(m.Raw); i++ {
		b1, b2 := m.Raw[i], m.Raw[i+1]
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

	var checksum uint8
	for _, b := range buf.Bytes() {
		checksum ^= b
	}
	if checksum != 0 {
		m.BadChecksum = true
	}

	if err := binary.Read(buf, binary.BigEndian, &m.MsgID); err != nil {
		return err
	}
	var attribute uint16
	if err := binary.Read(buf, binary.BigEndian, &attribute); err != nil {
		return err
	}
	m.MsgEncrypted = (attribute & 0x0400) != 0

	var iccIdData []byte
	if (attribute & 0x4000) != 0 { // versioned flag
		var version uint8
		if err := binary.Read(buf, binary.BigEndian, &version); err != nil {
			return err
		}
		m.MsgVersion = int16(version)
		iccIdData = make([]byte, 10)
	} else {
		m.MsgVersion = -1
		iccIdData = make([]byte, 6)
	}

	if err := binary.Read(buf, binary.BigEndian, iccIdData); err != nil {
		return err
	}
	m.SimNo = strings.TrimLeft(hex.EncodeToString(iccIdData), "0")
	if err := binary.Read(buf, binary.BigEndian, &m.MsgSN); err != nil {
		return err
	}
	if (attribute & 0x2000) != 0 { // splitted flag
		if err := binary.Read(buf, binary.BigEndian, &m.PartTotal); err != nil {
			return err
		}
		if err := binary.Read(buf, binary.BigEndian, &m.PartIndex); err != nil {
			return err
		}
	} else {
		m.PartTotal = 1
		m.PartIndex = 0
	}
	remain := buf.Len()
	if int(attribute&0x03FF) != remain-1 {
		m.BadBodyLen = true
	}
	if remain > 1 {
		m.Body = make([]byte, remain-1)
		binary.Read(buf, binary.BigEndian, m.Body)
	} else {
		if remain == 0 {
			m.BadChecksum = true
		}
		m.Body = []byte{}
	}
	// last byte is checksum, ignore
	return nil
}
