package msg

import (
	"testing"
	"time"
)

func TestDecodeAndDecodeKey(t *testing.T) {
	mk := &MsgKey{
		SimNo:     "12345678901",
		Timestamp: time.Now().Truncate(time.Second),
		DS:        2,
		TX:        true,
		SN:        3,
		MsgID:     0x0200,
		PartIndex: 2,
		PartTotal: 5,
	}
	b, err := mk.Encode()
	if err != nil {
		t.Error(err)
		return
	}
	decoded, err := DecodeKey(b)
	if err != nil {
		t.Error(err)
		return
	}
	if b1, b2, eq := mustMarshalEqual(mk, decoded); !eq {
		t.Errorf("mismatch:\n\t%s\n\t%s\n", string(b1), string(b2))
	}
}
