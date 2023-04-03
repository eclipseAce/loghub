package msg

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"strings"
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

	b, err := io.ReadAll(
		base64.NewDecoder(base64.StdEncoding, strings.NewReader(
			"fgACAAABIYB1NpYdIEp+AAAADAACAXV9AggHCiK4ABMAAAEZIwQDFkMpAQQAAh/nAgIAAAMCAAAHAgAAFAQAAAAAFQQAAAAAFgQAAAAAFwIAABgCAAAlBAAAAAAqAgACKwQAAAAAMAFjMQEY7wEATX4=",
			//"fgIAAGEBIYB1NpR7vQAAAAAADAACAXV9AggHCiK4ABMAAAEZIwQDFkMpAQQAAh/nAgIAAAMCAAAHAgAAFAQAAAAAFQQAAAAAFgQAAAAAFwIAABgCAAAlBAAAAAAqAgACKwQAAAAAMAFjMQEY7wEA"
		)),
	)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(hex.EncodeToString(b))
}
