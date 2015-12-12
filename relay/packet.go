package relay

import (
	"bytes"
	"encoding/binary"
	"io"
)

type Packet struct {
	PacketType   PacketType
	Length       uint32
	Body, header []byte
}

func ReadPacket(r io.Reader) (*Packet, error) {
	p := &Packet{
		header: make([]byte, 5),
	}

	_, err := r.Read(p.header)
	if err != nil {
		return nil, err
	}

	p.PacketType = PacketType(p.header[4])
	err = binary.Read(bytes.NewReader(p.header[:4]), binary.LittleEndian, &p.Length)
	if err != nil {
		return nil, err
	}

	p.Body = make([]byte, p.Length)
	var tot uint32

	for tot < p.Length {
		n, err := r.Read(p.Body[tot:])
		if err != nil && err != io.EOF {
			return nil, err
		}
		tot += uint32(n)
	}

	return p, nil
}

//WriteTo sends a packet to a writer
func (p *Packet) WriteTo(w io.Writer) error {
	_, err := w.Write(p.header)
	if err != nil {
		return err
	}
	_, err = w.Write(p.Body)
	return err
}
