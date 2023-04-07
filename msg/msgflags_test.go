package msg

import (
	"fmt"
	"testing"
)

func TestNewMsgFlags(t *testing.T) {
	f := NewMsgFlags(0xAA, 0xBB, 0xFFFFFFFF)
	fmt.Printf("%016X", f)
}
