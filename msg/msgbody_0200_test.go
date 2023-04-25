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
