package main

import (
	"bufio"
	"bytes"
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gin-gonic/gin"
	bolt "go.etcd.io/bbolt"
)

func main() {
	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)

	db, err := bolt.Open("gpsdata.db", 0666, nil)
	if err != nil {
		log.Fatalln(err)
	}
	defer db.Close()

	if err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("gpsdata"))
		return err
	}); err != nil {
		log.Fatalln(err)
	}

	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:6000")
	if err != nil {
		log.Fatalln(err)
	}

	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Close()

	s := bufio.NewScanner(conn)
	go func() {
		for s.Scan() && ctx.Err() == nil {
			if err := func() error {
				pr, err := NewEntry(s.Text())
				if err != nil {
					return err
				}
				key, value, err := pr.Bytes()
				if err != nil {
					return err
				}
				return db.Update(func(tx *bolt.Tx) error {
					return tx.Bucket([]byte("gpsdata")).Put(key, value)
				})
			}(); err != nil {
				log.Println(err)
			}
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
			c := tx.Bucket([]byte("gpsdata")).Cursor()
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

	go r.Run(":6001")

	<-ctx.Done()
}
