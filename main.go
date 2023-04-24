package main

import (
	"context"
	"log"
	"loghub/msg"
	"loghub/web"
	"os"
	"os/signal"

	_ "net/http/pprof"

	"github.com/jessevdk/go-flags"
)

func main() {
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

	db, err := msg.OpenDB(opts.DataDir, opts.BulkSize)
	if err != nil {
		log.Fatalln(err)
	}
	defer db.Close()

	if err := db.Listen(opts.BindLogstash); err != nil {
		log.Fatalln(err)
	}

	web.Serve(opts.BindWeb, db)

	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)
	<-ctx.Done()
}
