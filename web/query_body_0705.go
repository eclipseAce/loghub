package web

import (
	"loghub/msg"
	"time"
)

type msgBody_0705 struct {
	*msgBody_Base
	Count uint16               `json:"count"`
	Time  time.Time            `json:"time"`
	Items []*msgBody_0705_Item `json:"items"`
}

type msgBody_0705_Item struct {
	ID    uint32 `json:"id"`
	Flags uint8  `json:"flags"`
	Data  []byte `json:"data"`
}

func decodeBody_0705(base *msgBody_Base, raw []byte) (any, error) {
	b, err := msg.DecodeBody_0705(raw)
	if err != nil {
		return nil, err
	}
	body := &msgBody_0705{
		msgBody_Base: base,
		Count:        b.Count,
		Time:         b.Time,
		Items:        make([]*msgBody_0705_Item, len(b.Items)),
	}
	for i, mb := range b.Items {
		body.Items[i] = &msgBody_0705_Item{
			ID:    mb.ID,
			Flags: mb.Flags,
			Data:  mb.Data,
		}
	}
	return body, nil
}
