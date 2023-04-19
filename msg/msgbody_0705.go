package msg

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"time"
)

type MsgBody_0705 struct {
	Count    uint16
	Time     time.Time
	Items    []*MsgBody_0705Item
	Warnings []string
}

type MsgBody_0705Item struct {
	ID    uint32
	Flags uint8
	Data  [8]byte
}

func DecodeBody_0705(raw []byte) (*MsgBody_0705, error) {
	var common struct {
		Count uint16
		Time  [5]byte
	}
	warnings := make([]string, 0)
	buf := bytes.NewReader(raw)
	if err := binary.Read(buf, binary.BigEndian, &common); err != nil {
		return nil, err
	}
	items := make([]*MsgBody_0705Item, 0)
	for buf.Len() >= 12 {
		var item struct {
			ID   uint32
			Data [8]byte
		}
		if err := binary.Read(buf, binary.BigEndian, &item); err != nil {
			return nil, err
		}
		items = append(items, &MsgBody_0705Item{
			ID:    item.ID & 0x1FFFFFFF,
			Flags: uint8((item.ID & 0xE0000000) >> 29),
			Data:  item.Data,
		})
	}
	if buf.Len() != 0 {
		warnings = append(warnings, "bad tailing bytes")
	}
	if len(items) != int(common.Count) {
		warnings = append(warnings, "count mismatch")
	}
	bcdTime := hex.EncodeToString(common.Time[:3])
	bcdTime += "." + hex.EncodeToString(common.Time[3:])[1:] // fix millis >= 1000
	time, err := time.ParseInLocation("150405.000", bcdTime, time.Local)
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("bad time in 0200 body '%s': %v", bcdTime, err))
	}
	return &MsgBody_0705{
		Count:    common.Count,
		Time:     time,
		Items:    items,
		Warnings: warnings,
	}, nil
}
