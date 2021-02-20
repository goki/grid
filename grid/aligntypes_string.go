// Code generated by "stringer -type=AlignTypes"; DO NOT EDIT.

package grid

import (
	"errors"
	"strconv"
)

var _ = errors.New("dummy error")

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[AlignFirst-0]
	_ = x[AlignLast-1]
	_ = x[AlignDrawing-2]
	_ = x[AlignSelectBox-3]
	_ = x[AlignTypesN-4]
}

const _AlignTypes_name = "AlignFirstAlignLastAlignDrawingAlignSelectBoxAlignTypesN"

var _AlignTypes_index = [...]uint8{0, 10, 19, 31, 45, 56}

func (i AlignTypes) String() string {
	if i < 0 || i >= AlignTypes(len(_AlignTypes_index)-1) {
		return "AlignTypes(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _AlignTypes_name[_AlignTypes_index[i]:_AlignTypes_index[i+1]]
}

func (i *AlignTypes) FromString(s string) error {
	for j := 0; j < len(_AlignTypes_index)-1; j++ {
		if s == _AlignTypes_name[_AlignTypes_index[j]:_AlignTypes_index[j+1]] {
			*i = AlignTypes(j)
			return nil
		}
	}
	return errors.New("String: " + s + " is not a valid option for type: AlignTypes")
}
