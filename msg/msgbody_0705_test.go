package msg

import (
	"testing"
	"time"
)

func TestDecodeBody_0705(t *testing.T) {
	body, err := DecodeBody_0705(mustDecodeHexString(
		"00 02",
		"01 02 03 04 56",
		"12 34 56 78 23 45 67 89 0A BC DE F1",
		"23 45 67 89 12 34 56 78 90 AB CD EF",
	))
	if err != nil {
		t.Error(err)
		return
	}
	expected := MsgBody_0705{
		Count: 2,
		Time:  mustParseInLocation("15:04:05.000", "01:02:03.456", time.Local),
		Items: []*MsgBody_0705Item{
			{
				ID:    0x12345678 & 0x1FFFFFFF,
				Flags: (0x12345678 & 0xE0000000) >> 29,
				Data:  mustDecodeHexString("23 45 67 89 0A BC DE F1"),
			},
			{
				ID:    0x23456789 & 0x1FFFFFFF,
				Flags: (0x23456789 & 0xE0000000) >> 29,
				Data:  mustDecodeHexString("12 34 56 78 90 AB CD EF"),
			},
		},
		Warnings: []string{},
	}
	if b1, b2, eq := mustMarshalEqual(body, expected); !eq {
		t.Errorf("mismatch:\n\t%s\n\t%s\n", string(b1), string(b2))
	}
}
