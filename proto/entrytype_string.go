// Code generated by "stringer -type=EntryType"; DO NOT EDIT.

package proto

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[BoolEntry-0]
	_ = x[Int8Entry-1]
	_ = x[UInt8Entry-2]
	_ = x[IntEntry-3]
	_ = x[UIntEntry-4]
	_ = x[Float32Entry-5]
	_ = x[StringEntry-6]
	_ = x[ListEntry-7]
	_ = x[ModifyEntry-8]
}

const _EntryType_name = "BoolEntryInt8EntryUInt8EntryIntEntryUIntEntryFloat32EntryStringEntryListEntryModifyEntry"

var _EntryType_index = [...]uint8{0, 9, 18, 28, 36, 45, 57, 68, 77, 88}

func (i EntryType) String() string {
	if i >= EntryType(len(_EntryType_index)-1) {
		return "EntryType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _EntryType_name[_EntryType_index[i]:_EntryType_index[i+1]]
}
