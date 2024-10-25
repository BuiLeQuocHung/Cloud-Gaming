package utils

import "C"
import "unsafe"

func PointerToSlice(p unsafe.Pointer, size int) []byte {
	return C.GoBytes(p, C.int(size))
}
