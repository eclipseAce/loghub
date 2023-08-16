package msg

import (
	"testing"
	"time"
)

func TestDecodeBody_0200(t *testing.T) {
	body, err := DecodeBody_0200(mustDecodeHexString(
		"12 34 56 78",
		"23 45 67 89",
		"34 56 78 90",
		"56 78 90 12",
		"12 34",
		"23 45",
		"34 56",
		"23 04 25 11 01 39",
		"01 02 FF FF",
		"E0",
		"E1 03 FF FF FF",
	))
	if err != nil {
		t.Error(err)
		return
	}
	expected := MsgBody_0200{
		Alarm:     0x12345678,
		Status:    0x23456789,
		Latitude:  float64(0x34567890) / 1000000,
		Longitude: float64(0x56789012) / 1000000,
		Altitude:  0x1234,
		Speed:     float64(0x2345) / 10,
		Direction: 0x3456,
		Time:      mustParseInLocation("2006-01-02 15:04:05", "2023-04-25 11:01:39", time.Local),
		ExtInfo: []*MsgBody_0200_ExtInfo{
			{ID: 0x01, Data: mustDecodeHexString("FF FF")},
			{ID: 0xE1, Data: mustDecodeHexString("FF FF FF")},
		},
		Warnings: []string{},
	}
	if b1, b2, eq := mustMarshalEqual(body, expected); !eq {
		t.Errorf("mismatch:\n\t%s\n\t%s\n", string(b1), string(b2))
	}
}

func TestError(t *testing.T) {
	msg, err := Decode(mustDecodeHexString(
		"7E 02 00 00 45 06 46 18 21 63 87 0A 3C 00 00 80 26 02 0C 00 01 02 5E 3E 99 07 16 78 66 00 00 00 00 00 00 23 08 16 14 04 11 01 04 00 BB F9 27 02 02 00 00 03 02 00 D7 11 01 00 25 04 00 00 00 00 2B 04 00 06 00 05 30 01 0E 31 01 00 E0 04 E4 02 04 1A 35 7E",
	))
	if err != nil {
		t.Error(err)
		return
	}
	body, err := DecodeBody_0200(msg.Body)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(body)
}
