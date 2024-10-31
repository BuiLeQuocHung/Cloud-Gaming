package video

/*
#cgo pkg-config: libavutil
#include <libavutil/dict.h>
#include <stdlib.h>
*/
import "C"

type (
	Dictionary = C.AVDictionary
)

func NewDictionary(m map[string]string) *Dictionary {
	var dict *Dictionary

	for k, v := range m {
		ck := C.CString(k)
		cv := C.CString(v)

		C.av_dict_set(&dict, ck, cv, 0)
	}

	return dict
}
