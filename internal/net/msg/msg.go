package msg

import (
	"encoding/binary"
	"errors"
)

type Version uint8
type MsgType uint8

const (
	versionMask = 0xFE

	V1 Version = 0x1

	msgTypeMask = 0xFC

	Binary MsgType = 0x1
	TEXT   MsgType = 0x2
	JSON   MsgType = 0x3
)

type Envelope struct {
	Ver     Version
	Typ     MsgType
	Payload []byte
}

func (e *Envelope) MarshalBinary() ([]byte, error) {
	header := make([]byte, 4)
	header[0] = byte(e.Ver)
	header[1] = byte(e.Typ)
	binary.BigEndian.PutUint16(header[2:4], uint16(len(e.Payload)))

	bs := []byte{}
	bs = append(bs, header...)
	bs = append(bs, e.Payload...)

	if err := validate(bs); err != nil {
		return nil, err
	}

	return bs, nil
}

func (e *Envelope) UnmarshalBinary(bs []byte) error {
	if err := validate(bs); err != nil {
		return err
	}

	header := bs[:4]

	e.Ver = Version(header[0])
	e.Typ = MsgType(header[1])
	e.Payload = bs[4:]

	return nil
}

func validate(bs []byte) error {
	if len(bs) < 4 {
		return errors.New("missing header")
	}

	if len(bs[4:])&^0xFFFF != 0 {
		return errors.New("payload too big")
	}

	if bs[0]&versionMask != 0 {
		return errors.New("unsupported version")
	}

	if bs[1]&msgTypeMask != 0 {
		return errors.New("unsupported type")
	}

	return nil
}
