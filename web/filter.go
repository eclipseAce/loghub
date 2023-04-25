package web

import (
	"loghub/msg"
	"strconv"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
)

type msgKeyFilterFunc func(*msg.MsgKey) bool

func newMsgIdsFilter(val string) msgKeyFilterFunc {
	s := mapset.NewThreadUnsafeSet[uint16]()
	for _, it := range strings.Split(val, ",") {
		if id, err := strconv.Atoi(it); err == nil {
			s.Add(uint16(id))
		}
	}
	return func(mk *msg.MsgKey) bool {
		return s.Cardinality() == 0 || s.Contains(mk.MsgID)
	}
}

func newMsgXferFilter(val string) msgKeyFilterFunc {
	tx, rx := false, false
	for _, xfer := range strings.Split(val, ",") {
		switch xfer {
		case "tx":
			tx = true
		case "rx":
			rx = true
		}
	}
	if !tx && !rx {
		tx = true
		rx = true
	}
	return func(mk *msg.MsgKey) bool {
		return !(mk.TX && !tx || !mk.TX && !rx)
	}
}
