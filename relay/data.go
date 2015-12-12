package relay

import (
	"bytes"
	"encoding/binary"
)
import "fmt"

//DataEntry is a type that encapsulates the
type DataEntry struct {
	Type uint8
	ID   uint32
}

//PipByteOrder is the binary endianness of integers in the pip protocol
const PipByteOrder = binary.LittleEndian

//ErrShortEntryHeader indicates that not enough bytes were available to read a
//data entry
var ErrShortEntryHeader = fmt.Errorf("not enough bytes for data entry header")

//UnmarshalDataEntry takes a byte slice and returns the number of bytes read,
//and a possible error
func UnmarshalDataEntry(b []byte) (ct int, err error) {
	d := DataEntry{}

	if len(b) < 5 {
		return 0, ErrShortEntryHeader
	}

	d.Type = b[0]
	err = binary.Read(bytes.NewReader(b[1:5]), PipByteOrder, &d.ID)
	if err != nil {
		return 0, err
	}

	ct = 5

	switch d.Type {
	}
}
