package web

import (
	"fmt"
	"loghub/msg"
	"loghub/webui"
	"net/http"
	"time"

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
			SimNo string    `form:"simNo" binding:"required"`
			Since time.Time `form:"since" time_format:"2006-01-02 15:04:05" binding:"required"`
			Until time.Time `form:"until" time_format:"2006-01-02 15:04:05" binding:"required"`
		}
		if err := c.BindQuery(&params); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		results, err := queryMsg(db, params.SimNo, params.Since, params.Until)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"error": nil, "result": results})
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

func queryMsg(mdb *msg.MsgDB, simNo string, since, until time.Time) ([]*msgJSON, error) {
	results := make([]*msgJSON, 0)
	if err := mdb.Iterate(simNo, since, func(mi *msg.MsgItem) error {
		mk, err := mi.Key()
		if err != nil {
			return fmt.Errorf("decode msgKey: %w", err)
		}
		if mk.Timestamp.After(until) {
			return msg.ErrStopIteration
		}
		m, err := mi.Value()
		if err != nil {
			return fmt.Errorf(" decode msg: %w", err)
		}
		results = append(results, &msgJSON{
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
		return nil, err
	}
	return results, nil
}
