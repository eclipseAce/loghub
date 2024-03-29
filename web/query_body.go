package web

import (
	"bytes"
	"fmt"
	"loghub/msg"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type msgBody_Base struct {
	Timestamp time.Time `json:"timestamp"`
	Warnings  []string  `json:"warnings"`
}

type msgEntry struct {
	Key   *msg.MsgKey
	Value *msg.Msg
}

type decodeBodyFunc func(base *msgBody_Base, raw []byte) (any, error)
type entryFilterFunc func(msg any) bool

func decodeEntries(entries []*msgEntry) any {
	base := &msgBody_Base{
		Timestamp: entries[0].Key.Timestamp,
		Warnings:  make([]string, 0),
	}
	buf := &bytes.Buffer{}
	for _, me := range entries {
		buf.Write(me.Value.Body)
	}
	var decode decodeBodyFunc
	switch entries[0].Key.MsgID {
	case 0x0200:
		decode = decodeBody_0200
	case 0x0705:
		decode = decodeBody_0705
	default:
		decode = decodeBody_unknown
	}
	body, err := decode(base, buf.Bytes())
	if err != nil {
		base.Warnings = append(base.Warnings, err.Error())
		body, _ = decodeBody_unknown(base, buf.Bytes())
	}
	return body
}

func newEntryFilter(msgID uint16, c *gin.Context) entryFilterFunc {
	switch msgID {
	case 0x0200:
		return newEntryFilter_0200(c)
	default:
		return func(msg any) bool { return true }
	}
}

func queryBody(mdb *msg.MsgDB, c *gin.Context) (res any, code int, err error) {
	var params struct {
		SimNo string    `form:"simNo" binding:"required"`
		Since time.Time `form:"since" time_format:"2006-01-02 15:04:05" binding:"required"`
		Until time.Time `form:"until" time_format:"2006-01-02 15:04:05" binding:"required"`
		DS    uint8     `form:"ds"`
		MsgID uint16    `form:"msgId"`
	}
	if err := c.BindQuery(&params); err != nil {
		return nil, http.StatusBadRequest, err
	}
	list := make([]any, 0)
	entries := make([]*msgEntry, 0)
	filter := newEntryFilter(params.MsgID, c)
	if err := mdb.Iterate(params.SimNo, params.Since, func(mi *msg.MsgItem) error {
		mk, err := mi.Key()
		if err != nil {
			return fmt.Errorf("decode msgKey: %w", err)
		}
		if mk.MsgID != params.MsgID || mk.DS != params.DS {
			return nil
		}
		if len(entries) == 0 && mk.Timestamp.After(params.Until) {
			return msg.ErrStopIteration
		}
		if n := len(entries); n > 0 {
			mkFirst, mkLast := entries[0].Key, entries[n-1].Key
			if mkFirst.MsgID != mk.MsgID || mkFirst.PartTotal != mk.PartTotal || mkLast.PartIndex+1 != mk.PartIndex {
				entries = make([]*msgEntry, 0)
			}
		}
		m, err := mi.Value()
		if err != nil {
			return fmt.Errorf(" decode msg: %w", err)
		}
		entries = append(entries, &msgEntry{Key: mk, Value: m})
		if len(entries) == int(entries[0].Key.PartTotal) {
			if item := decodeEntries(entries); filter(item) {
				list = append(list, item)
			}
			entries = make([]*msgEntry, 0)
		}
		return nil
	}); err != nil {
		return nil, http.StatusInternalServerError, err
	}
	return list, http.StatusOK, nil
}
