// Code generated by "stringer -type=Sprites"; DO NOT EDIT.

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
	_ = x[SizeUpL-0]
	_ = x[SizeUpM-1]
	_ = x[SizeUpR-2]
	_ = x[SizeDnL-3]
	_ = x[SizeDnM-4]
	_ = x[SizeDnR-5]
	_ = x[SizeLfC-6]
	_ = x[SizeRtC-7]
	_ = x[SpritesN-8]
}

const _Sprites_name = "SizeUpLSizeUpMSizeUpRSizeDnLSizeDnMSizeDnRSizeLfCSizeRtCSpritesN"

var _Sprites_index = [...]uint8{0, 7, 14, 21, 28, 35, 42, 49, 56, 64}

func (i Sprites) String() string {
	if i < 0 || i >= Sprites(len(_Sprites_index)-1) {
		return "Sprites(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Sprites_name[_Sprites_index[i]:_Sprites_index[i+1]]
}

func (i *Sprites) FromString(s string) error {
	for j := 0; j < len(_Sprites_index)-1; j++ {
		if s == _Sprites_name[_Sprites_index[j]:_Sprites_index[j+1]] {
			*i = Sprites(j)
			return nil
		}
	}
	return errors.New("String: " + s + " is not a valid option for type: Sprites")
}