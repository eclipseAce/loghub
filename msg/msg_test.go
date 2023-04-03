package msg

import (
	"encoding/hex"
	"fmt"
	"log"
	"testing"
	"time"
)

func TestMsgKey(t *testing.T) {
	timestamp, _ := time.ParseInLocation("20060102150405", "20230403145355", time.Local)
	key := &MsgKey{SimNo: "12345678901", Timestamp: timestamp, SN: 0x123456}

	fmt.Printf("%v\n", key)

	keyBytes, err := key.Encode()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(hex.EncodeToString(keyBytes))

	k2, err := DecodeKey(keyBytes)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%v\n", k2)
}

func TestMsg(t *testing.T) {
	msgLog := "2023-03-31 17:01:25 GpsDataService:35 - (40666464383)收到报文类型：512,报文内容：7e02004032010000000004066646438317ad00000000000c000302324edb071a817700080000014a2303311701240104000c34290302000025040000000030011b310112347e"

	msg, err := DecodeLog(msgLog, 1)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("%v\n", msg.SimNo)
}
