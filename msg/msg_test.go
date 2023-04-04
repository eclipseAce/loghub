package msg

import (
	"encoding/json"
	"testing"
)

func TestDecodeLog_0200(t *testing.T) {
	log := `2023-03-30 14:39:29 GpsDataService:35 - (40261394651)收到报文类型：512,报文内容：7e02000070040261394651399400000000000c00c2018be6d8071c17180006000000b422093023392801040027997303020000140400000000150400000000160400000000170200002504000000002b040000000030011f310113eb11000700d4010087a209000600f800000000ef0d000000000000492492000011035a7e`
	m, err := DecodeLog(log, 0)
	if err != nil {
		t.Fatal(err)
	}
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(b))
}
