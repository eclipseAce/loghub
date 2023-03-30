package t808

import (
	"encoding/hex"
	"encoding/json"
	"testing"
)

func TestBytesToPacket(t *testing.T) {
	testCases := []string{
		"7e0705001f01436833998727ba0002143327050058fec11704769004b8fd170058fef121ff0030cfffffffff327e",
	}

	for _, testCase := range testCases {
		data, err := hex.DecodeString(testCase)
		if err != nil {
			t.Error(err)
			return
		}
		pk, err := BytesToPacket(data)
		if err != nil {
			t.Error(err)
			return
		}
		text, err := json.MarshalIndent(*pk, "", "    ")
		if err != nil {
			t.Error(err)
			return
		}
		t.Log(string(text))
	}
}
