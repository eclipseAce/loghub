package msg

import (
	"encoding/json"
	"testing"
)

func TestDecodeLog_0200(t *testing.T) {
	log := `20230407135201 Tx 7e80010005012280005634589264b9070500507e`
	m, err := DecodeLog(log, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(b))
}
