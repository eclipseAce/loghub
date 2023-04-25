package web

import (
	"loghub/msg"
	"loghub/webui"
	"net/http"

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

	r.GET("/api/query", handleRequest(db, queryRaw))
	r.GET("/api/queryBody", handleRequest(db, queryBody))

	go r.Run(bind)
}

type handleFunc func(mdb *msg.MsgDB, c *gin.Context) (res any, code int, err error)

func handleRequest(mdb *msg.MsgDB, h handleFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		res, code, err := h(mdb, c)
		c.JSON(code, gin.H{"error": err, "result": res})
	}
}
