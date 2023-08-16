package msg

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"time"
)

type MsgBody_0200 struct {
	Alarm     uint32
	Status    uint32
	Latitude  float64
	Longitude float64
	Altitude  uint16
	Speed     float64
	Direction uint16
	Time      time.Time
	ExtInfo   []*MsgBody_0200_ExtInfo
	Warnings  []string
}

type MsgBody_0200_ExtInfo struct {
	ID   uint8
	Data []byte
}

func DecodeBody_0200(raw []byte) (*MsgBody_0200, error) {
	var common struct {
		Alarm     uint32
		Status    uint32
		Latitude  uint32
		Longitude uint32
		Altitude  uint16
		Speed     uint16
		Direction uint16
		Time      [6]byte
	}
	warnings := make([]string, 0)
	buf := bytes.NewReader(raw)
	if err := binary.Read(buf, binary.BigEndian, &common); err != nil {
		return nil, err
	}
	extInfo := make([]*MsgBody_0200_ExtInfo, 0)
	for buf.Len() > 0 {
		var id uint8
		if err := binary.Read(buf, binary.BigEndian, &id); err != nil {
			return nil, err
		}
		// if id == 0xE0 {
		// 	continue // vendor custom, skip
		// }
		var length uint8
		if err := binary.Read(buf, binary.BigEndian, &length); err != nil {
			return nil, err
		}
		data := make([]byte, length)
		if err := binary.Read(buf, binary.BigEndian, data); err != nil {
			return nil, err
		}
		extInfo = append(extInfo, &MsgBody_0200_ExtInfo{
			ID:   id,
			Data: data,
		})
	}
	bcdTime := hex.EncodeToString(common.Time[:])
	time, err := time.ParseInLocation("20060102150405", "20"+bcdTime, time.Local)
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("bad time in 0200 body '%s': %v", bcdTime, err))
	}
	return &MsgBody_0200{
		Alarm:     common.Alarm,
		Status:    common.Status,
		Latitude:  float64(common.Latitude) / 1000000,
		Longitude: float64(common.Longitude) / 1000000,
		Altitude:  common.Altitude,
		Speed:     float64(common.Speed) / 10,
		Direction: common.Direction,
		Time:      time,
		ExtInfo:   extInfo,
		Warnings:  warnings,
	}, nil
}
