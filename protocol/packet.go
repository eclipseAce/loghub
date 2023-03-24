package protocol

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
)

type Packet struct {
	Id           uint16
	RsaEncrypted bool
	Versioned    bool
	Version      uint8
	IccId        string
	SerialNo     uint16
	Splitted     bool
	SplitCount   uint16
	SplitNo      uint16
	Length       uint16
	Checksum     uint8
	BodyChecksum uint8
	Body         []byte
}

func BytesToPacket(data []byte) (*Packet, error) {
	pk := &Packet{}

	buf := new(bytes.Buffer)
	for i := 1; i+1 < len(data); i++ {
		b1, b2 := data[i], data[i+1]
		switch {
		case b1 == 0x7D && b2 == 0x01:
			buf.WriteByte(0x7E)
			i++
		case b1 == 0x7D && b2 == 0x02:
			buf.WriteByte(0x7D)
			i++
		default:
			buf.WriteByte(b1)
		}
	}

	pk.BodyChecksum = 0
	unpacked := buf.Bytes()
	for _, b := range unpacked[:len(unpacked)-1] {
		pk.BodyChecksum ^= b
	}

	if err := binary.Read(buf, binary.BigEndian, &pk.Id); err != nil {
		return nil, err
	}
	var attribute uint16
	if err := binary.Read(buf, binary.BigEndian, &attribute); err != nil {
		return nil, err
	}
	pk.Length = attribute & 0x03FF
	pk.RsaEncrypted = (attribute & 0x0400) != 0
	pk.Splitted = (attribute & 0x2000) != 0
	pk.Versioned = (attribute & 0x4000) != 0

	var iccIdData []byte
	if pk.Versioned {
		if err := binary.Read(buf, binary.BigEndian, &pk.Version); err != nil {
			return nil, err
		}
		iccIdData = make([]byte, 10)
	} else {
		iccIdData = make([]byte, 6)
	}

	if err := binary.Read(buf, binary.BigEndian, iccIdData); err != nil {
		return nil, err
	}
	pk.IccId = strings.TrimLeft(hex.EncodeToString(iccIdData), "0")
	if err := binary.Read(buf, binary.BigEndian, &pk.SerialNo); err != nil {
		return nil, err
	}
	if pk.Splitted {
		if err := binary.Read(buf, binary.BigEndian, &pk.SplitCount); err != nil {
			return nil, err
		}
		if err := binary.Read(buf, binary.BigEndian, &pk.SplitNo); err != nil {
			return nil, err
		}
	}
	if buf.Len() == 0 {
		return nil, io.ErrUnexpectedEOF
	}
	// if bodyLen != buf.Len()-1 {
	// 	return nil, fmt.Errorf("unexpected body length %d in header, actually is %d", bodyLen, buf.Len()-1)
	// }
	pk.Body = make([]byte, buf.Len()-1)
	binary.Read(buf, binary.BigEndian, pk.Body)

	binary.Read(buf, binary.BigEndian, &pk.Checksum)
	// if checksum != headerChecksum {
	// 	return nil, fmt.Errorf("unexpected checksum %02X in header, actually is %02X", headerChecksum, checksum)
	// }
	return pk, nil
}

func PacketToBytes(packet *Packet) ([]byte, error) {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, packet.Id)

	bodyLen := len(packet.Body)
	if bodyLen > 0x03FF {
		return nil, fmt.Errorf("body too long, %d > %d bytes", bodyLen, 0x03FF)
	}
	var attribute uint16
	attribute |= uint16(bodyLen)
	if packet.RsaEncrypted {
		attribute |= 0x0400
	}
	if packet.Splitted {
		attribute |= 0x2000
	}
	if packet.Versioned {
		attribute |= 0x4000
	}
	binary.Write(buf, binary.BigEndian, attribute)

	var iccIdSize int
	if packet.Versioned {
		binary.Write(buf, binary.BigEndian, packet.Version)
		iccIdSize = 10
	} else {
		iccIdSize = 6
	}
	iccId := packet.IccId
	if len(iccId) > iccIdSize*2 {
		return nil, fmt.Errorf("iccId too long, %d > %d chars", len(iccId), iccIdSize*2)
	}
	if len(iccId)%2 != 0 {
		iccId = "0" + iccId
	}
	iccIdData, err := hex.DecodeString(iccId)
	if err != nil {
		return nil, fmt.Errorf("invalid iccId: %v", err)
	}
	binary.Write(buf, binary.BigEndian, iccIdData)

	binary.Write(buf, binary.BigEndian, packet.SerialNo)

	if packet.Splitted {
		binary.Write(buf, binary.BigEndian, packet.SplitCount)
		binary.Write(buf, binary.BigEndian, packet.SplitNo)
	}

	var checksum uint8
	for _, b := range buf.Bytes() {
		checksum ^= b
	}
	binary.Write(buf, binary.BigEndian, checksum)

	packed := new(bytes.Buffer)
	packed.WriteByte(0x7E)
	for _, b := range buf.Bytes() {
		switch b {
		case 0x7E:
			buf.WriteByte(0x7D)
			buf.WriteByte(0x01)
		case 0x7D:
			buf.WriteByte(0x7D)
			buf.WriteByte(0x02)
		default:
			buf.WriteByte(b)
		}
	}
	packed.WriteByte(0x7E)
	return packed.Bytes(), nil
}

func ScanPackets(data []byte, atEOF bool) (advance int, token []byte, err error) {
	start := bytes.IndexByte(data, 0x7E)
	if start < 0 {
		return len(data), nil, nil
	}
	for start+1 < len(data) && data[start+1] == 0x7E {
		start++
	}
	off := bytes.IndexByte(data[start+1:], 0x7E)
	if off < 0 {
		if atEOF {
			return len(data), nil, nil
		}
		return start, nil, nil
	}
	end := start + 1 + off
	return end + 1, data[start : end+1], nil
}
