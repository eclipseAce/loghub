package web

import (
	"loghub/msg"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
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
	Mileage   float64                 `json:"mileage"`
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
		Mileage:      b.ParsedExtInfo.Mileage,
	}
	for i, mb := range b.ExtInfo {
		body.ExtInfo[i] = &msgBody_0200_ExtInfo{
			ID:   mb.ID,
			Data: mb.Data,
		}
	}
	return body, nil
}

func newEntryFilter_0200(c *gin.Context) entryFilterFunc {
	qs := strings.Split(c.Request.URL.Query().Get("extIds"), ",")
	if qs[0] == "" {
		return func(msg any) bool { return true }
	}
	extIds := make([]uint8, len(qs))
	for _, q := range qs {
		if extId, err := strconv.Atoi(q); err == nil {
			extIds = append(extIds, uint8(extId))
		}
	}
	return func(msg any) bool {
		msg0200, ok := msg.(*msgBody_0200)
		if !ok {
			return true
		}
		for _, extId := range extIds {
			for _, extInfo := range msg0200.ExtInfo {
				if extId == extInfo.ID {
					return true
				}
			}
		}
		return false
	}
}
