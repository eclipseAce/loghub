package web

import (
	"loghub/msg"
	"time"
)

type msgBody_0200 struct {
	*msgBody_Base
	Alarm     uint32                  `json:"alarm"`
	Status    uint32                  `json:"status"`
	Latitude  float64                 `json:"latitude"`
	Longitude float64                 `json:"longitude"`
	Altitude  uint16                  `json:"altitude"`
	Speed     float64                 `json:"speed"`
	Direction uint16                  `json:"direction"`
	Time      time.Time               `json:"time"`
	ExtInfo   []*msgBody_0200_ExtInfo `json:"extInfo"`
}

type msgBody_0200_ExtInfo struct {
	ID   uint8  `json:"id"`
	Data []byte `json:"data"`
}

func decodeBody_0200(base *msgBody_Base, raw []byte) (any, error) {
	b, err := msg.DecodeBody_0200(raw)
	if err != nil {
		return nil, err
	}
	body := &msgBody_0200{
		msgBody_Base: base,
		Alarm:        b.Alarm,
		Status:       b.Status,
		Latitude:     b.Latitude,
		Longitude:    b.Longitude,
		Altitude:     b.Altitude,
		Speed:        b.Speed,
		Direction:    b.Direction,
		Time:         b.Time,
		ExtInfo:      make([]*msgBody_0200_ExtInfo, len(b.ExtInfo)),
	}
	for i, mb := range b.ExtInfo {
		body.ExtInfo[i] = &msgBody_0200_ExtInfo{
			ID:   mb.ID,
			Data: mb.Data,
		}
	}
	return body, nil
}
