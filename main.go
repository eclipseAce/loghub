package main

import (
	"bytes"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/elastic/go-lumber/server"
	"github.com/gin-gonic/gin"
	"github.com/jessevdk/go-flags"
	bolt "go.etcd.io/bbolt"
)

func main() {
	var opts struct {
		DBFile       string `short:"d" long:"db-file" default:"t808data.db" description:"BoltDB file"`
		DBBucket     string `short:"b" long:"db-bucket" default:"packets" description:"BoltDB bucket"`
		DBBulkSize   uint   `short:"s" long:"bulk-size" default:"2000" description:"BoltDB bulk put size"`
		BindLogstash string `short:"l" long:"bind-logstash" default:":5044" description:"[host]:port Logstash bind address"`
		BindWeb      string `short:"w" long:"bind-web" default:":6060" description:"[host]:port Web bind address"`
	}

	_, err := flags.ParseArgs(&opts, os.Args)
	if err != nil {
		log.Fatalln(err)
	}

	db, err := bolt.Open(opts.DBFile, 0666, nil)
	if err != nil {
		log.Fatalln(err)
	}
	defer db.Close()

	if err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(opts.DBBucket))
		return err
	}); err != nil {
		log.Fatalln(err)
	}

	s, err := server.ListenAndServe(opts.BindLogstash, server.V1(true), server.V2(true))
	if err != nil {
		log.Fatal(err)
	}
	defer s.Close()

	bp := newBulkPut(db, opts.DBBulkSize)

	go func() {
		for batch := range s.ReceiveChan() {
			for _, event := range batch.Events {
				if err := func() error {
					pr, err := NewEntry(event)
					if err != nil {
						return err
					}
					key, value, err := pr.Bytes()
					if err != nil {
						return err
					}
					bp.Put([]byte(opts.DBBucket), key, value)
					return nil
				}(); err != nil {
					log.Println(err)
				}
			}
			batch.ACK()
		}
	}()

	r := gin.Default()
	r.GET("/query", func(c *gin.Context) {
		var params struct {
			IccId string    `form:"iccId" binding:"required"`
			Since time.Time `form:"since" time_format:"2006-01-02 15:04:05" binding:"required"`
			Until time.Time `form:"until" time_format:"2006-01-02 15:04:05" binding:"required"`
			Id    uint16    `form:"id"`
		}
		if err := c.BindQuery(&params); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		results := make([]*Entry, 0)
		if err := db.View(func(tx *bolt.Tx) error {
			c := tx.Bucket([]byte(opts.DBBucket)).Cursor()
			min := EntryKey(params.IccId, params.Since, 0)
			max := EntryKey(params.IccId, params.Until, 99999999)
			for k, v := c.Seek(min); k != nil && bytes.Compare(k, max) <= 0; k, v = c.Next() {
				pr, err := ParseEntry(k, v)
				if err != nil {
					log.Println(err)
					continue
				}
				if params.Id != 0 && pr.Id != params.Id {
					continue
				}
				results = append(results, pr)
			}
			return nil
		}); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"error": nil, "results": results})
	})

	go r.Run(opts.BindWeb)

	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)
	<-ctx.Done()
	bp.Flush()
}

type bulkPut struct {
	db      *bolt.DB
	bkvChan chan [][]byte
}

func newBulkPut(db *bolt.DB, bulkSize uint) *bulkPut {
	bp := &bulkPut{
		db:      db,
		bkvChan: make(chan [][]byte, bulkSize),
	}
	go func() {
		for {
			bp.Flush()
			time.Sleep(time.Second)
		}
	}()
	return bp
}

func (bp *bulkPut) Put(bucket, key, value []byte) {
	if len(bp.bkvChan) == cap(bp.bkvChan) {
		bp.Flush()
	}
	bp.bkvChan <- [][]byte{bucket, key, value}
}

func (bp *bulkPut) Flush() {
	bp.db.Update(func(tx *bolt.Tx) error {
		n := len(bp.bkvChan)
		for i := 0; i < n; i++ {
			bkv := <-bp.bkvChan
			if err := tx.Bucket(bkv[0]).Put(bkv[1], bkv[2]); err != nil {
				log.Println(err)
			}
		}
		return nil
	})
}
