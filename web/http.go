package web

import (
	"fmt"
	"loghub/msg"
	"loghub/webui"
	"net/http"
	"strconv"
	"strings"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.ReleaseMode)
}

func Serve(bind string, db *msg.MsgDB) {
	r := gin.Default()

	r.Use(gzip.Gzip(gzip.DefaultCompression))

	r.StaticFS("/ui", http.FS(webui.Assets()))

	r.GET("/", func(ctx *gin.Context) {
		ctx.Redirect(http.StatusMovedPermanently, "/ui")
	})

	r.GET("/api/query", func(c *gin.Context) {
		var params struct {
			SimNo   string    `form:"simNo" binding:"required"`
			Since   time.Time `form:"since" time_format:"2006-01-02 15:04:05" binding:"required"`
			Until   time.Time `form:"until" time_format:"2006-01-02 15:04:05" binding:"required"`
			MsgIDs  string    `form:"msgIds"`
			MsgXfer string    `form:"msgXfer"`
		}
		if err := c.BindQuery(&params); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		msgs, msgIds, err := queryMsg(db, params.SimNo, params.Since, params.Until, []msgKeyFilterFunc{
			newMsgIdsFilter(params.MsgIDs),
			newMsgXferFilter(params.MsgXfer),
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"error": nil,
			"result": gin.H{
				"msgs":   msgs,
				"msgIds": msgIds,
			},
		})
	})

	go r.Run(bind)
}

type msgJSON struct {
	Raw       []byte    `json:"raw"`
	TX        bool      `json:"tx"`
	DS        uint8     `json:"ds"`
	SN        uint32    `json:"sn"`
	Timestamp time.Time `json:"timestamp"`
	MsgID     uint16    `json:"msgId"`
	MsgSN     uint16    `json:"msgSn"`
	SimNo     string    `json:"simNo"`
	Version   int16     `json:"version"`
	Encrypted bool      `json:"encrypted"`
	PartTotal uint16    `json:"partTotal"`
	PartIndex uint16    `json:"partIndex"`
	Warnings  []string  `json:"warnings"`
}

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

func queryMsg(mdb *msg.MsgDB, simNo string, since, until time.Time, filters []msgKeyFilterFunc) ([]*msgJSON, []uint16, error) {
	msgs := make([]*msgJSON, 0)
	msgIds := mapset.NewThreadUnsafeSet[uint16]()
	if err := mdb.Iterate(simNo, since, func(mi *msg.MsgItem) error {
		mk, err := mi.Key()
		if err != nil {
			return fmt.Errorf("decode msgKey: %w", err)
		}
		if mk.Timestamp.After(until) {
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
		msgs = append(msgs, &msgJSON{
			Raw:       m.Raw,
			TX:        mk.TX,
			DS:        mk.DS,
			SN:        mk.SN,
			Timestamp: mk.Timestamp,
			MsgID:     m.MsgID,
			MsgSN:     m.MsgSN,
			SimNo:     m.SimNo,
			Version:   m.Version,
			Encrypted: m.Encrypted,
			PartTotal: m.PartIndex,
			PartIndex: m.PartTotal,
			Warnings:  m.Warnings,
		})
		return nil
	}); err != nil {
		return nil, nil, err
	}
	return msgs, msgIds.ToSlice(), nil
}
