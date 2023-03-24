package main

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
	"t808logger/protocol"
	"time"
)

var logPatternRe = regexp.MustCompile(
	`^(?P<host>[^ ]+) (?P<timestamp>\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}) GpsDataService:\d+ - \(\d+\)收到报文类型：\d+,报文内容：(?P<payload>[a-f0-9]+)$`,
)

var timestampSN uint64

func init() {
	go func() {
		for {
			time.Sleep(time.Second)
			atomic.StoreUint64(&timestampSN, 0)
		}
	}()
}

type Entry struct {
	*protocol.Packet
	Timestamp   time.Time
	TimestampSN uint64
}

func NewEntry(line string) (pr *Entry, err error) {
	m := logPatternRe.FindStringSubmatch(line)
	if m == nil {
		return nil, errors.New("invalid payload")
	}
	fields := make(map[string]string)
	for i, name := range logPatternRe.SubexpNames() {
		if i != 0 && name != "" {
			fields[name] = m[i]
		}
	}
	pr = &Entry{}
	pr.Timestamp, err = time.Parse("2006-01-02 15:04:05", fields["timestamp"])
	if err != nil {
		return nil, err
	}
	payload, err := hex.DecodeString(fields["payload"])
	if err != nil {
		return nil, err
	}
	pr.Packet, err = protocol.BytesToPacket(payload)
	if err != nil {
		return nil, err
	}
	pr.TimestampSN = atomic.AddUint64(&timestampSN, 1)
	return pr, nil
}

func ParseEntry(key, value []byte) (pr *Entry, err error) {
	packet := &protocol.Packet{}
	dec := gob.NewDecoder(bytes.NewReader(value))
	if err := dec.Decode(packet); err != nil {
		return nil, err
	}
	keyParts := strings.Split(string(key), ":")
	if len(keyParts) != 3 {
		return nil, errors.New("invalid key")
	}
	timestamp, err := time.Parse("20060102150405", keyParts[1])
	if err != nil {
		return nil, err
	}
	timestampSN, err := strconv.ParseUint(keyParts[2], 10, 0)
	if err != nil {
		return nil, err
	}
	return &Entry{Packet: packet, Timestamp: timestamp, TimestampSN: timestampSN}, nil
}

func EntryKey(iccId string, timestamp time.Time, timestampSN uint64) []byte {
	return []byte(fmt.Sprintf("%s:%s:%08d", iccId, timestamp.Format("20060102150405"), timestampSN))
}

func (pr *Entry) Bytes() (key, value []byte, err error) {
	buf := &bytes.Buffer{}
	enc := gob.NewEncoder(buf)
	if err = enc.Encode(pr.Packet); err != nil {
		return nil, nil, err
	}
	key = EntryKey(pr.Packet.IccId, pr.Timestamp, atomic.AddUint64(&timestampSN, 1))
	value = buf.Bytes()
	return
}
