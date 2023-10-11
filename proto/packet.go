package proto

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"
)

// PacketType accounts for the possible packet types
type PacketType uint8

// Well-known packet types
//
//go:generate stringer -type=PacketType
const (
	PacketTypeKeepAlive = PacketType(iota)
	PacketTypeConnectionAccepted
	PacketTypeConnectionRefused
	PacketTypeDataUpdate
	PacketTypeMapUpdate
	PacketTypeCommand
	PacketTypeCommandResult
	PacketTypeCount
)

// Packet is the PIPProtocol wire format
type Packet struct {
	PacketType   PacketType
	Body, header []byte
	length       uint32
}

// ReadPacket returns a packet from an io.Reader.
func ReadPacket(r io.Reader) (*Packet, error) {
	br := bufio.NewReader(r)
	p := &Packet{
		header: make([]byte, 5),
	}

	_, err := br.Read(p.header)
	if err != nil {
		return nil, err
	}

	p.length = PipByteOrder.Uint32(p.header[:4])
	p.PacketType = PacketType(p.header[4])

	p.Body = make([]byte, p.length)

	_, err = io.ReadFull(r, p.Body)
	if err != nil && err != io.EOF {
		return nil, err
	}

	return p, nil
}

// WriteTo sends a packet to a writer and returns the number of bytes written.
func (p *Packet) WriteTo(w io.Writer) (int64, error) {
	b := &bytes.Buffer{}

	if p.header == nil || len(p.header) == 0 {
		err := binary.Write(b, PipByteOrder, uint32(len(p.Body)))
		if err != nil {
			return 0, err
		}

		b.Write([]byte{byte(p.PacketType)})
	} else {
		b.Write(p.header)
	}

	b.Write(p.Body)
	return b.WriteTo(w)
	// return io.Copy(w, b)
}
