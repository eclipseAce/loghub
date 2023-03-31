package web

import (
	"loghub/msg"
	"loghub/webui"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.ReleaseMode)
}

func Serve(bind string, db *msg.MsgDB) {
	r := gin.Default()

	r.StaticFS("/ui", http.FS(webui.Assets()))

	r.GET("/", func(ctx *gin.Context) {
		ctx.Redirect(http.StatusMovedPermanently, "/ui")
	})

	r.GET("/api/query", func(c *gin.Context) {
		var params struct {
			SimNo string    `form:"simNo" binding:"required"`
			Since time.Time `form:"since" time_format:"2006-01-02 15:04:05" binding:"required"`
			Until time.Time `form:"until" time_format:"2006-01-02 15:04:05" binding:"required"`
			Id    uint16    `form:"id"`
		}
		if err := c.BindQuery(&params); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		results, err := db.Query(params.SimNo, params.Since, params.Until, func(m *msg.Msg) bool {
			return params.Id == 0 || params.Id == m.MsgID
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"error": nil, "result": results})
	})

	go r.Run(bind)
}