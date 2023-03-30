package main

import (
	"bytes"
	"compress/gzip"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"loghub/t808"
	"regexp"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/elastic/go-lumber/server"
)

func closeOnError(fn func() error, err *error) error {
	if err != nil && *err != nil {
		return fn()
	}
	return nil
}

var logPatternRe = regexp.MustCompile(
	`^(?P<timestamp>\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}) GpsDataService:\d+ - \([0-9A-F]+\)收到报文类型：\d+,报文内容：(?P<payload>[a-f0-9]+)$`,
)

type Msg struct {
	SeqNo     uint64
	SimNo     string
	MsgID     uint16
	MsgNo     uint16
	Version   int16
	Timestamp time.Time
	RawBytes  []byte
}

type MsgDB struct {
	db        *badger.DB
	s         server.Server
	seq       *badger.Sequence
	counter   uint64
	entryChan chan *badger.Entry
	closeChan chan struct{}
	closeWait sync.WaitGroup
}

func NewMsgDB(path string, bind string, bulkSize uint) (mdb *MsgDB, err error) {
	db, err := badger.Open(badger.DefaultOptions(path))
	if err != nil {
		return nil, err
	}
	defer closeOnError(db.Close, &err)

	seq, err := db.GetSequence([]byte("msgseq"), 10000)
	if err != nil {
		return nil, err
	}
	defer closeOnError(seq.Release, &err)

	s, err := server.ListenAndServe(bind, server.V1(true), server.V2(true))
	if err != nil {
		return nil, err
	}
	defer closeOnError(s.Close, &err)

	mdb = &MsgDB{
		db:        db,
		s:         s,
		seq:       seq,
		entryChan: make(chan *badger.Entry, bulkSize),
		closeChan: make(chan struct{}),
	}

	go mdb.scheduleTask()
	go mdb.receiveTask()
	go mdb.statTask()

	return mdb, nil
}

func (mdb *MsgDB) scheduleTask() {
	mdb.closeWait.Add(1)
	defer mdb.closeWait.Done()
	tkGC := time.NewTicker(5 * time.Minute)
	defer tkGC.Stop()
	tkFlush := time.NewTicker(time.Second)
	defer tkFlush.Stop()
	for {
		select {
		case <-tkGC.C:
			mdb.db.RunValueLogGC(0.7)
		case <-tkFlush.C:
			mdb.flush()
		case <-mdb.closeChan:
			mdb.flush()
			mdb.db.RunValueLogGC(0.7)
			return
		}
	}
}

func (mdb *MsgDB) receiveTask() {
	mdb.closeWait.Add(1)
	defer mdb.closeWait.Done()
	recvChan := mdb.s.ReceiveChan()
	for {
		select {
		case batch := <-recvChan:
			for _, event := range batch.Events {
				if err := mdb.handleEvent(event); err != nil {
					log.Println(fmt.Errorf("handleEvent: %w", err))
				}
				atomic.AddUint64(&mdb.counter, 1)
			}
			batch.ACK()
		case <-mdb.closeChan:
			return
		}
	}
}

func (mdb *MsgDB) statTask() {
	mdb.closeWait.Add(1)
	defer mdb.closeWait.Done()
	tk := time.NewTicker(time.Second)
	defer tk.Stop()
	for {
		select {
		case <-tk.C:
			log.Printf("messages received last second: %d", atomic.SwapUint64(&mdb.counter, 0))
		case <-mdb.closeChan:
			return
		}
	}
}

func (mdb *MsgDB) handleEvent(event any) error {
	m, err := mdb.createMsg(event)
	if err != nil {
		return err
	}
	val := &bytes.Buffer{}
	gz := gzip.NewWriter(val)
	enc := gob.NewEncoder(gz)
	if err := enc.Encode(m); err != nil {
		return err
	}
	gz.Close()
	e := badger.NewEntry(
		mdb.createKey(m.SimNo, m.Timestamp, m.SeqNo),
		val.Bytes(),
	)
	if len(mdb.entryChan) == cap(mdb.entryChan) {
		mdb.flush()
	}
	mdb.entryChan <- e.WithTTL(48 * time.Hour)
	return nil
}

func (mdb *MsgDB) createKey(simNo string, timestamp time.Time, seqNo uint64) []byte {
	return []byte(fmt.Sprintf("%s:%s:%08x", simNo, timestamp.Format("20060102150405"), seqNo))
}

func (mdb *MsgDB) createMsg(event any) (*Msg, error) {
	rawMsg := event.(map[string]any)["message"].(string)

	matches := logPatternRe.FindStringSubmatch(rawMsg)
	if matches == nil {
		return nil, fmt.Errorf("invalid message: %s", rawMsg)
	}
	fields := make(map[string]string)
	for i, name := range logPatternRe.SubexpNames() {
		if i != 0 && name != "" {
			fields[name] = matches[i]
		}
	}

	timestamp, err := time.Parse("2006-01-02 15:04:05", fields["timestamp"])
	if err != nil {
		return nil, err
	}

	payload, err := hex.DecodeString(fields["payload"])
	if err != nil {
		return nil, err
	}

	packet, err := t808.BytesToPacket(payload)
	if err != nil {
		return nil, fmt.Errorf("bytesToPacket: %w", err)
	}

	seqNo, err := mdb.seq.Next()
	if err != nil {
		return nil, fmt.Errorf("seq next: %w", err)
	}
	m := &Msg{
		SeqNo:     seqNo,
		SimNo:     packet.IccId,
		MsgID:     packet.Id,
		MsgNo:     packet.SerialNo,
		Timestamp: timestamp,
		RawBytes:  payload,
		Version:   -1,
	}
	if packet.Versioned {
		m.Version = int16(packet.Version)
	}
	return m, nil
}

func (mdb *MsgDB) flush() {
	if err := mdb.db.Update(func(txn *badger.Txn) error {
		for i := 0; i < cap(mdb.entryChan); i++ {
			select {
			case e := <-mdb.entryChan:
				if err := txn.SetEntry(e); err != nil {
					log.Println(fmt.Errorf("db setEntry: %w", err))
				}
			default:
				return nil
			}
		}
		return nil
	}); err != nil {
		log.Println(fmt.Errorf("db update: %w", err))
	}
}

func (mdb *MsgDB) Close() error {
	close(mdb.closeChan)
	mdb.closeWait.Wait()
	mdb.s.Close()
	mdb.seq.Release()
	mdb.db.Close()
	return nil
}

func (mdb *MsgDB) Query(simNo string, since, until time.Time, filter func(*Msg) bool) ([]*Msg, error) {
	results := make([]*Msg, 0)
	if err := mdb.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		sinceKey := mdb.createKey(simNo, since, 0x00000000)
		untilKey := mdb.createKey(simNo, until, 0xFFFFFFFF)
		for it.Seek(sinceKey); ; it.Next() {
			item := it.Item()
			key := item.Key()
			if bytes.Compare(key, untilKey) > 0 {
				break
			}
			if err := item.Value(func(val []byte) error {
				msg := &Msg{}
				gz, err := gzip.NewReader(bytes.NewReader(val))
				if err != nil {
					return err
				}
				dec := gob.NewDecoder(gz)
				if err := dec.Decode(msg); err != nil {
					return err
				}
				if filter == nil || filter(msg) {
					results = append(results, msg)
				}
				return nil
			}); err != nil {
				log.Println(fmt.Errorf("db read value: %w", err))
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return results, nil
}
