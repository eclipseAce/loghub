package msg

import "time"

type Body_0200 struct {
	Alarm     uint32
	Status    uint32
	Latitude  uint32
	Longitude uint32
	Altitude  uint16
	Speed     uint16
	Direction uint16
	Time      time.Time
}
