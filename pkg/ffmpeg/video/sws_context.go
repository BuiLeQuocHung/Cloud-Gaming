package video

/*
#cgo pkg-config: libswscale
#include <libswscale/swscale.h>
*/
import "C"

type (
	SwsContext = C.struct_SwsContext
)

func NewSwsCtx(from_width, from_height, to_width, to_height int, from_format, to_format VideoFormat, scalingAlgo int) *SwsContext {
	return C.sws_getContext(
		C.int(from_width), C.int(from_height), int32(from_format),
		C.int(to_width), C.int(to_height), int32(to_format),
		C.int(scalingAlgo), nil, nil, nil)
}

func (c *SwsContext) Free() {
	C.sws_freeContext(c)
}
