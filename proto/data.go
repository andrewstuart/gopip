package proto

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"
)

type EntryType uint8

//go:generate stringer -type=EntryType
const (
	BoolEntry = EntryType(iota)
	Int8Entry
	UInt8Entry
	IntEntry
	UIntEntry
	Float32Entry
	StringEntry
	ListEntry
	ModifyEntry
)

// DataEntry is a type that encapsulates the
type DataEntry struct {
	Type  EntryType
	ID    uint32
	Name  string
	Value any
}

// PipByteOrder is the binary endianness of integers in the pip protocol
var PipByteOrder = binary.LittleEndian

var (
	//ErrShortEntryHeader indicates that not enough bytes were available to read a
	//data entry
	ErrShortEntryHeader = fmt.Errorf("not enough bytes for data entry header")
	//ErrShortData indicates that for the type indicated by the header, not
	//enough bytes remain for the vaulue expected
	ErrShortData = fmt.Errorf("not enough bytes for data entry value")
)

const (
	Bytes32Bit = 4
)

// UnmarshalDataEntry takes a byte slice and returns the number of bytes read,
// and a possible error
func UnmarshalDataEntry(b []byte) (*DataEntry, int, error) {
	if len(b) < 5 {
		return nil, 0, ErrShortEntryHeader
	}

	d := DataEntry{
		Type: EntryType(b[0]),
		ID:   PipByteOrder.Uint32(b[1:5]),
	}
	ct := 5

	defer func() {
		if err := recover(); err != nil {
			log.Printf("recovered %v. ct: %d, len(b): %d, entry %#v", err, ct, len(b), d)
			log.Println(hex.Dump(b))
		}
	}()

	switch d.Type {
	case BoolEntry:
		d.Value = b[ct] == 1
		ct++
	case Int8Entry:
		var v int8
		err := binary.Read(bytes.NewReader(b[ct:]), PipByteOrder, &v)
		if err != nil {
			return nil, ct, fmt.Errorf("error reading int8: %w", err)
		}
		d.Value = v
		ct++
	case UInt8Entry:
		d.Value = uint8(b[ct])
		ct++
	case IntEntry:
		var v int32
		err := binary.Read(bytes.NewReader(b[ct:]), PipByteOrder, &v)
		if err != nil {
			return nil, ct, fmt.Errorf("error reading int32: %w", err)
		}
		d.Value = v
		ct += Bytes32Bit
	case UIntEntry:
		var v uint32
		err := binary.Read(bytes.NewReader(b[ct:]), PipByteOrder, &v)
		if err != nil {
			return nil, ct, fmt.Errorf("error reading uint32: %w", err)
		}
		d.Value = v
		ct += Bytes32Bit
	case Float32Entry:
		var v float32
		err := binary.Read(bytes.NewReader(b[ct:]), PipByteOrder, &v)
		if err != nil {
			return nil, ct, fmt.Errorf("error reading float32: %w", err)
		}
		d.Value = v
		ct += Bytes32Bit
	case StringEntry:
		s := getString(b[ct:])
		ct += len(s) + 1
		d.Value = s
	case ListEntry:
		var count uint16
		err := binary.Read(bytes.NewReader(b[ct:]), PipByteOrder, &count)
		if err != nil {
			return nil, ct, fmt.Errorf("error reading list count: %w", err)
		}
		ct += 2

		arr := make([]uint32, int(count))
		rdr := bytes.NewReader(b[ct:])
		var v uint32
		for i := 0; i < int(count); i++ {
			err = binary.Read(rdr, PipByteOrder, &v)
			if err != nil {
				return nil, ct, fmt.Errorf("error reading list value %d: %w", i, err)
			}
			arr[i] = v
		}
		ct += 4 * int(count)
		d.Value = arr
	case ModifyEntry:
		var insCt uint16
		r := bytes.NewReader(b[ct:])
		err := binary.Read(r, PipByteOrder, &insCt)
		if err != nil {
			return nil, ct, err
		}
		ct += 2

		ins := make([]DictEntry, int(insCt))
		for i := 0; i < int(insCt); i++ {
			dict, n, err := DictEntryFromBytes(b[ct:])
			if err != nil {
				return nil, ct, err
			}
			ct += n
			ins[i] = dict
		}

		r = bytes.NewReader(b[ct:])

		var remLen uint16
		err = binary.Read(r, PipByteOrder, &remLen)
		if err != nil {
			return nil, ct, err
		}
		ct += 2

		rem := make([]uint32, int(remLen))
		var v uint32
		for i := 0; i < int(remLen); i++ {
			err = binary.Read(r, PipByteOrder, &v)
			if err != nil {
				return nil, ct, err
			}
			rem[i] = v
			ct += 4
		}
		d.Value = InsRemove{ins, rem}
	default:
		return nil, ct, fmt.Errorf("Unknown data type: %d", d.Type)
	}
	return &d, ct, nil
}

func UnmarshalDataEntries(bs []byte) ([]*DataEntry, error) {
	list := make([]*DataEntry, 0, 10)
	var ct int
	for ct < len(bs) {
		de, n, err := UnmarshalDataEntry(bs[ct:])
		if err != nil {
			return list, fmt.Errorf("error unmarshaling data entry %d: %w", len(list), err)
		}
		list = append(list, de)
		ct += n
	}
	return list, nil
}

type InsRemove struct {
	Insert []DictEntry
	Remove []uint32
}

type DictEntry struct {
	Ref  uint32
	Name string
}

func DictEntryFromBytes(bs []byte) (DictEntry, int, error) {
	d := DictEntry{}
	r := bytes.NewReader(bs)
	err := binary.Read(r, PipByteOrder, &d.Ref)
	if err != nil {
		return d, 0, err
	}
	ct := 4
	d.Name = getString(bs[ct:])
	ct += len(d.Name) + 1
	return d, ct, nil
}

func getString(bs []byte) string {
	i := bytes.IndexRune(bs, '\x00')
	if 0 < i && i < len(bs) {
		return string(bs[:i])
	}
	return ""
}
