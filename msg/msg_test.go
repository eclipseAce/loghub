package msg

import (
	"fmt"
	"testing"
)

func TestMsg(t *testing.T) {
	msgLog := "2023-03-31 17:01:25 GpsDataService:35 - (40666464383)收到报文类型：512,报文内容：7e02004032010000000004066646438317ad00000000000c000302324edb071a817700080000014a2303311701240104000c34290302000025040000000030011b310112347e"

	msg, err := NewMsgFromLog(msgLog, func() (uint64, error) { return 1, nil })
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("%v\n", msg.SimNo)

}
