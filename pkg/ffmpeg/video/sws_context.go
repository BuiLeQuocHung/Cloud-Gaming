package video

/*
#cgo pkg-config: libswscale
#include <libswscale/swscale.h>
*/
import "C"
import (
	"errors"
	"fmt"
)

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

func ScaleAndConvertFrame(swsCtx *SwsContext, srcFrame *AVFrame, targetWidth, targetHeight int, targetFormat VideoFormat) (*AVFrame, error) {
	if swsCtx == nil {
		return nil, errors.New("ScaleAndConvertFrame: sws context is nil")
	}

	if srcFrame == nil {
		return nil, errors.New("ScaleAndConvertFrame: srcFrame is nil")
	}

	srcData := (**C.uchar)(srcFrame.GetData())
	srcLinesize := (*C.int)(srcFrame.GetLinesize())

	desFrame, err := NewFrameWithBuffer(targetWidth, targetHeight, targetFormat)
	if err != nil {
		desFrame.Close()
		return nil, errors.New("ScaleAndConvertFrame: create new frame err")
	}

	desData := (**C.uchar)(desFrame.GetData())
	desLinesize := (*C.int)(desFrame.GetLinesize())

	if ret := C.sws_scale(swsCtx, srcData, srcLinesize, 0, C.int(srcFrame.GetHeight()), desData, desLinesize); ret != C.int(targetHeight) {
		desFrame.Close()
		return nil, fmt.Errorf("ScaleAndConvertFrame: num of rows copied is not equal to height")
	}

	return desFrame, nil
}
