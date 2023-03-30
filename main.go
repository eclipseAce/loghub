package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jessevdk/go-flags"
)

func main() {
	gin.SetMode(gin.ReleaseMode)

	var opts struct {
		DataDir      string `short:"d" long:"data-dir" default:"data" description:"Data file directory"`
		BulkSize     uint   `short:"b" long:"bulk-size" default:"2000" description:"DB bulk set size"`
		BindLogstash string `short:"l" long:"bind-logstash" default:":5044" description:"[host]:port Logstash bind address"`
		BindWeb      string `short:"w" long:"bind-web" default:":6060" description:"[host]:port Web bind address"`
	}

	_, err := flags.ParseArgs(&opts, os.Args)
	if err != nil {
		log.Fatalln(err)
	}

	db, err := NewMsgDB(opts.DataDir, opts.BindLogstash, opts.BulkSize)
	if err != nil {
		log.Fatalln(err)
	}
	defer db.Close()

	r := gin.Default()
	r.GET("/query", func(c *gin.Context) {
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
		results, err := db.Query(params.SimNo, params.Since, params.Until, func(m *Msg) bool {
			return params.Id == 0 || params.Id == m.MsgID
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"error": nil, "results": results})
	})

	go r.Run(opts.BindWeb)

	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)
	<-ctx.Done()
}
