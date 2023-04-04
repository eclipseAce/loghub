package msg

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"time"
)

type MsgBody_0200 struct {
	Alarm      uint32
	Status     uint32
	Latitude   float64
	Longitude  float64
	Altitude   uint16
	Speed      float64
	Direction  uint16
	Time       time.Time
	AttachInfo *AttachInfo
}

type AttachInfo struct {
	Mileage        *float64
	Fuel           *float64
	RecorderSpeed  *float64
	AnalogValue0   *uint16
	AnalogValue1   *uint16
	SignalStrength *uint8
	Satellites     *uint8
	Raw            map[uint8][]byte
}

func (ai *AttachInfo) Add(id uint8, data []byte) error {
	if ai.Raw == nil {
		ai.Raw = make(map[uint8][]byte)
	}
	if ai.Raw[id] != nil {
		return fmt.Errorf("duplicated attachInfo id '%02X'", id)
	}
	ai.Raw[id] = data
	switch {
	case id == 0x01 && len(data) == 4:
		ai.Mileage = new(float64)
		*ai.Mileage = float64(binary.BigEndian.Uint32(data)) / 10

	case id == 0x02 && len(data) == 2:
		ai.Fuel = new(float64)
		*ai.Fuel = float64(binary.BigEndian.Uint16(data)) / 10

	case id == 0x03 && len(data) == 2:
		ai.RecorderSpeed = new(float64)
		*ai.RecorderSpeed = float64(binary.BigEndian.Uint16(data)) / 10

	case id == 0x2B && len(data) == 4:
		val := binary.BigEndian.Uint32(data)
		ai.AnalogValue0 = new(uint16)
		*ai.AnalogValue0 = uint16(val)
		ai.AnalogValue1 = new(uint16)
		*ai.AnalogValue1 = uint16(val >> 16)

	case id == 0x30 && len(data) == 1:
		ai.SignalStrength = new(uint8)
		*ai.SignalStrength = data[0]

	case id == 0x31 && len(data) == 1:
		ai.Satellites = new(uint8)
		*ai.Satellites = data[0]
	}
	return nil
}

func DecodeBody_0200(m *Msg) error {
	buf := bytes.NewReader(m.Body)
	common := struct {
		Alarm     uint32
		Status    uint32
		Latitude  uint32
		Longitude uint32
		Altitude  uint16
		Speed     uint16
		Direction uint16
		Time      [6]byte
	}{}
	if err := binary.Read(buf, binary.BigEndian, &common); err != nil {
		return err
	}
	attachInfo := &AttachInfo{}
	for buf.Len() > 0 {
		var id uint8
		if err := binary.Read(buf, binary.BigEndian, &id); err != nil {
			return err
		}
		if id == 0xE0 {
			continue // vendor custom, skip
		}
		var length uint8
		if err := binary.Read(buf, binary.BigEndian, &length); err != nil {
			return err
		}
		data := make([]byte, length)
		if err := binary.Read(buf, binary.BigEndian, data); err != nil {
			return err
		}
		if err := attachInfo.Add(id, data); err != nil {
			m.Warnings = append(m.Warnings, err.Error())
		}
	}
	bcdTime := hex.EncodeToString(common.Time[:])
	time, err := time.ParseInLocation("20060102150405", "20"+bcdTime, time.Local)
	if err != nil {
		m.Warnings = append(m.Warnings, fmt.Sprintf("bad time in 0200 body '%s': %v", bcdTime, err))
	}
	m.DecodedBody = &MsgBody_0200{
		Alarm:      common.Alarm,
		Status:     common.Status,
		Latitude:   float64(common.Latitude) / 1000000,
		Longitude:  float64(common.Longitude) / 1000000,
		Altitude:   common.Altitude,
		Speed:      float64(common.Speed) / 10,
		Direction:  common.Direction,
		Time:       time,
		AttachInfo: attachInfo,
	}
	return nil
}
