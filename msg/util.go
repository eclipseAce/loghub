package msg

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"
)

func mustDecodeHexString(s ...string) []byte {
	b, err := hex.DecodeString(strings.ReplaceAll(strings.Join(s, ""), " ", ""))
	if err != nil {
		panic(err)
	}
	return b
}

func mustParseInLocation(layout, s string, l *time.Location) time.Time {
	t, err := time.ParseInLocation(layout, s, l)
	if err != nil {
		panic(err)
	}
	return t
}

func mustMarshalEqual(a, b any) (ma, mb string, eq bool) {
	blob1, err := json.Marshal(a)
	if err != nil {
		panic(err)
	}
	blob2, err := json.Marshal(b)
	if err != nil {
		panic(err)
	}
	return string(blob1), string(blob2), bytes.Equal(blob1, blob2)
}
