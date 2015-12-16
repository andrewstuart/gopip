package proto

import (
	"bytes"
	"encoding/binary"
	"io"
)

//PacketType accounts for the possible packet types
type PacketType uint8

//Well-known packet types
const (
	KeepAlivePacket = PacketType(iota)
	ConnecctionAcceptedPacket
	ConnectionRefusedPacket
	DataUpdatePacket
	MapUpdatePacket
	CommandPacket
)

//Packet is the PIPProtocol wire format
type Packet struct {
	PacketType   PacketType
	Body, header []byte
	length       uint32
}

//ReadPacket returns a packet from an io.Reader.
func ReadPacket(r io.Reader) (*Packet, error) {
	p := &Packet{
		header: make([]byte, 5),
	}

	_, err := r.Read(p.header)
	if err != nil {
		return nil, err
	}

	p.PacketType = PacketType(p.header[4])
	err = binary.Read(bytes.NewReader(p.header[:4]), binary.LittleEndian, &p.length)
	if err != nil {
		return nil, err
	}

	p.Body = make([]byte, p.length)
	var tot uint32

	for tot < p.length {
		n, err := r.Read(p.Body[tot:])
		if err != nil && err != io.EOF {
			return nil, err
		}
		tot += uint32(n)
	}

	return p, nil
}

//WriteTo sends a packet to a writer
func (p *Packet) WriteTo(w io.Writer) (int64, error) {
	if p.header == nil {
		p.header = make([]byte, 5)
		p.header[0] = byte(p.PacketType)

		buf := &bytes.Buffer{}
		err := binary.Write(buf, binary.LittleEndian, uint32(len(p.Body)))
		if err != nil {
			return 0, err
		}
		buf.Read(p.header[1:])
	}

	n, err := w.Write(p.header)
	if err != nil {
		return int64(n), err
	}
	m, err := w.Write(p.Body)
	return int64(n) + int64(m), err
}
