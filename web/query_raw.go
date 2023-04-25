package web

import (
	"fmt"
	"loghub/msg"
	"net/http"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/gin-gonic/gin"
)

type msgRawJSON struct {
	Timestamp time.Time `json:"timestamp"`
	Raw       []byte    `json:"raw"`
	TX        bool      `json:"tx"`
	DS        uint8     `json:"ds"`
	SN        uint32    `json:"sn"`
	MsgID     uint16    `json:"msgId"`
	MsgSN     uint16    `json:"msgSn"`
	Version   int16     `json:"version"`
	Encrypted bool      `json:"encrypted"`
	PartTotal uint16    `json:"partTotal"`
	PartIndex uint16    `json:"partIndex"`
	Warnings  []string  `json:"warnings"`
}

func queryRaw(mdb *msg.MsgDB, c *gin.Context) (res any, code int, err error) {
	var params struct {
		SimNo   string    `form:"simNo" binding:"required"`
		Since   time.Time `form:"since" time_format:"2006-01-02 15:04:05" binding:"required"`
		Until   time.Time `form:"until" time_format:"2006-01-02 15:04:05" binding:"required"`
		MsgIDs  string    `form:"msgIds"`
		MsgXfer string    `form:"msgXfer"`
	}
	if err := c.BindQuery(&params); err != nil {
		return nil, http.StatusBadRequest, err
	}
	filters := []msgKeyFilterFunc{
		newMsgIdsFilter(params.MsgIDs),
		newMsgXferFilter(params.MsgXfer),
	}
	msgs := make([]*msgRawJSON, 0)
	msgIds := mapset.NewThreadUnsafeSet[uint16]()
	if err := mdb.Iterate(params.SimNo, params.Since, func(mi *msg.MsgItem) error {
		mk, err := mi.Key()
		if err != nil {
			return fmt.Errorf("decode msgKey: %w", err)
		}
		if mk.Timestamp.After(params.Until) {
			return msg.ErrStopIteration
		}
		msgIds.Add(mk.MsgID)
		for _, filter := range filters {
			if !filter(mk) {
				return nil
			}
		}
		m, err := mi.Value()
		if err != nil {
			return fmt.Errorf(" decode msg: %w", err)
		}
		msgs = append(msgs, &msgRawJSON{
			Timestamp: mk.Timestamp,
			Raw:       m.Raw,
			TX:        mk.TX,
			DS:        mk.DS,
			SN:        mk.SN,
			MsgID:     m.MsgID,
			MsgSN:     m.MsgSN,
			Version:   m.Version,
			Encrypted: m.Encrypted,
			PartTotal: m.PartIndex,
			PartIndex: m.PartTotal,
			Warnings:  m.Warnings,
		})
		return nil
	}); err != nil {
		return nil, http.StatusInternalServerError, err
	}
	return gin.H{"msgs": msgs, "msgIds": msgIds.ToSlice()}, http.StatusOK, nil
}
