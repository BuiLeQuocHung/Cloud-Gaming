package video

/*
#cgo pkg-config: libavutil libswscale
#include <libavutil/frame.h>
#include <libavutil/imgutils.h>
#include <libswscale/swscale.h>
*/
import "C"
import (
	"cloud_gaming/pkg/ffmpeg/video"
	"cloud_gaming/pkg/utils"
	"errors"
	"fmt"
	"unsafe"
)

type (
	Converter struct {
	}
)

func NewConverter() *Converter {
	return &Converter{}
}

// RGB/RGBA to frame only
func (c *Converter) ToFrame(data []byte, width, height, pitch int, format video.VideoFormat) (*video.AVFrame, error) {
	// Allocate frame
	frame := video.NewFrame()
	if frame == nil {
		return nil, errors.New("ConvertAndResize: failed to allocate frame")
	}

	frame.SetWidth(width)
	frame.SetHeight(height)
	frame.SetFormat(int(format))
	frameData := (**C.uchar)(frame.GetData())
	frameLinesize := (*C.int)(frame.GetLinesize())

	// Allocate buffer
	bufferSize := C.av_image_fill_arrays(
		frameData, frameLinesize, (*C.uint8_t)(unsafe.Pointer(&data[0])),
		C.AV_PIX_FMT_RGB24, C.int(width), C.int(height), 1)

	if bufferSize < 0 {
		frame.Close()
		return nil, fmt.Errorf("failed to fill frame arrays")
	}

	frame.SetLinesize([8]int{pitch})
	return frame, nil
}

func (c *Converter) ConvertAndResize(swsCtxManager *SwsCtxManager, srcFrame *video.AVFrame, targetWidth, targetHeight int, targetFormat video.VideoFormat) (*video.AVFrame, error) {
	curData := (**C.uchar)(srcFrame.GetData())
	curLinesize := (*C.int)(srcFrame.GetLinesize())

	// Allocate frame
	frame := video.NewFrame()
	if frame == nil {
		return frame, errors.New("ConvertAndResize: failed to allocate frame")
	}

	frame.SetWidth(targetWidth)
	frame.SetHeight(targetHeight)
	frame.SetFormat(int(targetFormat))

	// Allocate buffer for the new frame
	numBytes := C.av_image_get_buffer_size(int32(targetFormat), C.int(targetWidth), C.int(targetHeight), 1)
	buffer := C.av_malloc(C.size_t(numBytes))
	if buffer == nil {
		return nil, fmt.Errorf("ConvertAndResize: could not allocate YUV buffer")
	}

	// Set up the new frame data pointers
	frameData := (**C.uchar)(frame.GetData())
	frameLinesize := (*C.int)(frame.GetLinesize())
	if ret := C.av_image_fill_arrays(frameData, frameLinesize, (*C.uint8_t)(buffer), int32(targetFormat), C.int(targetWidth), C.int(targetHeight), 1); ret < 0 {
		frame.Close()
		return nil, fmt.Errorf("ConvertAndResize: attach buffer failed: %w", utils.CErrorToString(int(ret)))
	}

	swsCtxKey := &SwsCtxKey{
		from_width:  srcFrame.GetWidth(),
		from_height: srcFrame.GetHeight(),
		from:        video.VideoFormat(srcFrame.GetFormat()),

		width:  targetWidth,
		height: targetHeight,
		to:     targetFormat,

		scalingAlgo: C.SWS_BILINEAR,
	}

	swsCtx := swsCtxManager.Get(swsCtxKey)
	defer swsCtxManager.Set(swsCtxKey, swsCtx)

	// Converting
	if ret := C.sws_scale(swsCtx, curData, curLinesize, 0, C.int(srcFrame.GetHeight()), frameData, frameLinesize); ret != C.int(targetHeight) {
		frame.Close()
		return nil, fmt.Errorf("ConvertAndResize: num of rows copied is not equal to height: %w", utils.CErrorToString(int(ret)))
	}

	return frame, nil
}
