package msg

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/elastic/go-lumber/server"
)

type MsgDB struct {
	db        *badger.DB
	seq       *badger.Sequence
	counter   uint64
	entryChan chan *badger.Entry
	closeChan chan struct{}
	closeWait sync.WaitGroup
}

func OpenDB(path string, bulkSize uint) (mdb *MsgDB, err error) {
	callOnError := func(fn func() error) {
		if err != nil {
			fn()
		}
	}

	db, err := badger.Open(badger.DefaultOptions(path).WithZSTDCompressionLevel(3))
	if err != nil {
		return nil, err
	}
	defer callOnError(db.Close)

	seq, err := db.GetSequence([]byte("MSGSNSEQ"), 10000)
	if err != nil {
		return nil, err
	}
	defer callOnError(seq.Release)

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
				data := event.(map[string]any)
				eventMsg := strings.Trim(data["message"].(string), "\x00\r\n\t ")
				if err := mdb.handleEventMsg(eventMsg); err != nil {
					b, _ := json.Marshal(eventMsg)
					log.Println(fmt.Errorf("handleLogEvent: %w: %s", err, string(b)))
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
	interval := 10
	mdb.closeWait.Add(1)
	defer mdb.closeWait.Done()
	tk := time.NewTicker(time.Duration(interval * int(time.Second)))
	defer tk.Stop()
	for {
		select {
		case <-tk.C:
			log.Printf("messages rate: %.2f/s", float64(atomic.SwapUint64(&mdb.counter, 0))/float64(interval))
		case <-mdb.closeChan:
			return
		}
	}
}

func (mdb *MsgDB) handleEventMsg(eventMsg string) error {
	sn, err := mdb.seq.Next()
	if err != nil {
		return err
	}
	m, mk, err := ParseLog(eventMsg, 0, uint32(sn))
	if err != nil {
		if err == ErrEmptyMsg {
			return nil // for empty msg (just two 0x7E), ignore
		}
		return err
	}
	key, err := mk.Encode()
	if err != nil {
		return err
	}
	if len(mdb.entryChan) == cap(mdb.entryChan) {
		mdb.flush()
	}
	mdb.entryChan <- badger.NewEntry(key, m.Raw).WithTTL(72 * time.Hour)
	return nil
}

func (mdb *MsgDB) flush() {
	if err := mdb.db.Update(func(txn *badger.Txn) error {
		for i := 0; i < cap(mdb.entryChan); i++ {
			select {
			case e := <-mdb.entryChan:
				if err := txn.SetEntry(e); err != nil {
					log.Println(fmt.Errorf("flush setEntry: %w", err))
				}
			default:
				return nil
			}
		}
		return nil
	}); err != nil {
		log.Println(fmt.Errorf("flush update: %w", err))
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

type MsgItem struct {
	item *badger.Item
}

func (mi *MsgItem) Key() (*MsgKey, error) {
	return DecodeKey(mi.item.Key())
}
func (mi *MsgItem) Value() (*Msg, error) {
	val, err := mi.item.ValueCopy(make([]byte, 0, mi.item.ValueSize()))
	if err != nil {
		return nil, err
	}
	return Decode(val)
}

var ErrStopIteration = errors.New("stop iteration")

func (mdb *MsgDB) Iterate(simNo string, since time.Time, fn func(*MsgItem) error) error {
	seek, err := (&MsgKey{SimNo: simNo, Timestamp: since}).Encode()
	if err != nil {
		return err
	}
	prefix := seek[:SimNoBytes]
	return mdb.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		mi := &MsgItem{}
		for it.Seek(seek); it.ValidForPrefix(prefix); it.Next() {
			mi.item = it.Item()
			if err := fn(mi); err != nil {
				if err == ErrStopIteration {
					break
				}
				log.Println(err)
			}
		}
		return nil
	})
}
