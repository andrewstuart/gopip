package proto

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"log"
)
import "fmt"

type EntryType uint8

const (
	BoolEntry = uint8(iota)
	Int8Entry
	UInt8Entry
	IntEntry
	UIntEntry
	Float32Entry
	StringEntry
	ListEntry
	ModifyEntry
)

//DataEntry is a type that encapsulates the
type DataEntry struct {
	Type  uint8
	ID    uint32
	Name  string
	Value interface{}
}

//PipByteOrder is the binary endianness of integers in the pip protocol
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

//UnmarshalDataEntry takes a byte slice and returns the number of bytes read,
//and a possible error
func UnmarshalDataEntry(b []byte) (de *DataEntry, ct int, err error) {
	d := DataEntry{}
	de = &d

	if len(b) < 5 {
		return nil, 0, ErrShortEntryHeader
	}

	d.Type = b[0]
	err = binary.Read(bytes.NewReader(b[1:5]), PipByteOrder, &d.ID)
	if err != nil {
		return
	}

	ct = 5

	defer func() {
		if err := recover(); err != nil {
			log.Printf("recovered %v. ct: %d, len(b): %d, entry %#v", err, ct, len(b), d)
			log.Println(hex.Dump(b))
		}
	}()

	switch d.Type {
	case BoolEntry:
		d.Value = uint8(b[ct]) == 1
		ct++
	case Int8Entry:
		var v int8
		err = binary.Read(bytes.NewReader(b[ct:]), PipByteOrder, &v)
		if err != nil {
			return
		}
		d.Value = v
		ct++
	case UInt8Entry:
		d.Value = uint8(b[ct])
		ct++
	case IntEntry:
		var v int32
		err = binary.Read(bytes.NewReader(b[ct:]), PipByteOrder, &v)
		if err != nil {
			return
		}
		d.Value = v
		ct += Bytes32Bit
	case UIntEntry:
		var v uint32
		err = binary.Read(bytes.NewReader(b[ct:]), PipByteOrder, &v)
		if err != nil {
			return
		}
		d.Value = v
		ct += Bytes32Bit
	case Float32Entry:
		var v float32
		err = binary.Read(bytes.NewReader(b[ct:]), PipByteOrder, &v)
		if err != nil {
			return
		}
		d.Value = v
		ct += Bytes32Bit
	case StringEntry: //string
		s := getString(b[ct:])
		ct += len(s) + 1
		d.Value = s
	case ListEntry: //list
		var count uint16
		err = binary.Read(bytes.NewReader(b[ct:]), PipByteOrder, &count)
		if err != nil {
			return
		}
		ct += 2

		arr := make([]uint32, int(count))
		rdr := bytes.NewReader(b[ct:])
		var v uint32
		for i := 0; i < int(count); i++ {
			err = binary.Read(rdr, PipByteOrder, &v)
			if err != nil {
				return
			}
			arr[i] = v
		}
		ct += 4 * int(count)
		d.Value = arr
	case ModifyEntry:
		var insCt uint16
		r := bytes.NewReader(b[ct:])
		err = binary.Read(r, PipByteOrder, &insCt)
		if err != nil {
			return
		}
		ct += 2

		var dict DictEntry
		var n int

		ins := make([]DictEntry, int(insCt))
		for i := 0; i < int(insCt); i++ {
			dict, n, err = DictEntryFromBytes(b[ct:])
			if err != nil {
				return
			}
			ct += n
			ins[i] = dict
		}

		r = bytes.NewReader(b[ct:])

		var remLen uint16
		err = binary.Read(r, PipByteOrder, &remLen)
		if err != nil {
			return
		}
		ct += 2

		rem := make([]uint32, int(remLen))
		var v uint32
		for i := 0; i < int(remLen); i++ {
			err = binary.Read(r, PipByteOrder, &v)
			if err != nil {
				return
			}
			rem[i] = v
			ct += 4
		}
		d.Value = InsRemove{ins, rem}
	default:
		err = fmt.Errorf("Unknown data type: %d", d.Type)
		return
	}

	return
}

func UnmarshalDataEntries(bs []byte) ([]*DataEntry, error) {
	list := make([]*DataEntry, 0, 10)
	var ct int
	for ct < len(bs) {
		de, n, err := UnmarshalDataEntry(bs[ct:])
		if err != nil {
			return list, err
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
