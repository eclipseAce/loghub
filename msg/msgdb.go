package msg

import (
	"bytes"
	"fmt"
	"log"
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

type MsgDB struct {
	db        *badger.DB
	seq       *badger.Sequence
	counter   uint64
	entryChan chan *badger.Entry
	closeChan chan struct{}
	closeWait sync.WaitGroup
}

func NewMsgDB(path string, bulkSize uint) (mdb *MsgDB, err error) {
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

	mdb = &MsgDB{
		db:        db,
		seq:       seq,
		entryChan: make(chan *badger.Entry, bulkSize),
		closeChan: make(chan struct{}),
	}

	go mdb.scheduleTask()
	go mdb.statTask()

	return mdb, nil
}

func (mdb *MsgDB) scheduleTask() {
	mdb.closeWait.Add(1)
	defer mdb.closeWait.Done()
	tkGC := time.NewTicker(time.Hour)
	defer tkGC.Stop()
	tkFlush := time.NewTicker(time.Second)
	defer tkFlush.Stop()
	for {
		select {
		case <-tkGC.C:
			mdb.db.RunValueLogGC(0.5)
		case <-tkFlush.C:
			mdb.flush()
		case <-mdb.closeChan:
			mdb.flush()
			mdb.db.RunValueLogGC(0.5)
			return
		}
	}
}

func (mdb *MsgDB) receiveTask(s server.Server) {
	mdb.closeWait.Add(1)
	defer mdb.closeWait.Done()
	defer s.Close()
	recvChan := s.ReceiveChan()
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
	sn, err := mdb.seq.Next()
	if err != nil {
		return err
	}
	m, err := DecodeLog(event.(map[string]any)["message"].(string), sn)
	if err != nil {
		return err
	}
	key, val, err := m.Encode()
	if err != nil {
		return err
	}
	if len(mdb.entryChan) == cap(mdb.entryChan) {
		mdb.flush()
	}
	mdb.entryChan <- badger.NewEntry(key, val).WithTTL(48 * time.Hour)
	return nil
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

func (mdb *MsgDB) Listen(bind string) error {
	s, err := server.ListenAndServe(bind, server.V1(true), server.V2(true))
	if err != nil {
		return err
	}
	go mdb.receiveTask(s)
	return nil
}

func (mdb *MsgDB) Close() error {
	close(mdb.closeChan)
	mdb.closeWait.Wait()
	mdb.seq.Release()
	mdb.db.Close()
	return nil
}

func (mdb *MsgDB) Query(simNo string, since, until time.Time, filter func(*Msg) bool) ([]*Msg, error) {
	sinceKey, untilKey, err := EncodeKeyRange(simNo, since, until)
	if err != nil {
		return nil, err
	}
	results := make([]*Msg, 0)
	if err := mdb.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		for it.Seek(sinceKey); ; it.Next() {
			item := it.Item()
			if bytes.Compare(item.Key(), untilKey) > 0 {
				break
			}
			val, err := item.ValueCopy(make([]byte, 0, item.ValueSize()))
			if err != nil {
				log.Println(fmt.Errorf("db valueCopy: %w", err))
				continue
			}
			m, err := DecodeEntry(item.Key(), val)
			if err != nil {
				log.Println(fmt.Errorf("invalid msg: %w", err))
				continue
			}
			if filter == nil || filter(m) {
				results = append(results, m)
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return results, nil
}
