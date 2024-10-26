package video

/*
#cgo pkg-config: libavutil
#include <libavutil/frame.h>
#include <libavutil/imgutils.h>
*/
import "C"
import (
	"errors"
	"unsafe"
)

type (
	AVFrame = C.AVFrame
)

func NewFrame() *AVFrame {
	return C.av_frame_alloc()
}

func NewFrameWithBuffer(width, height int, format PixelFormat) (*AVFrame, error) {
	frame := NewFrame()
	if frame == nil {
		return nil, errors.New("create new frame failed: frame is nil")
	}

	frame.SetWidth(width)
	frame.SetHeight(height)
	frame.SetFormat(int(format))

	if ret := C.av_frame_get_buffer(frame, 0); ret < 0 {
		frame.Close()
		return nil, errors.New("allocate buffer failed")
	}

	return frame, nil
}

func NewFrameWithBufferAsArray(width, height int, format PixelFormat, data []byte) (*AVFrame, error) {
	frame := NewFrame()
	if frame == nil {
		return nil, errors.New("create new frame failed: frame is nil")
	}

	frame.SetWidth(width)
	frame.SetHeight(height)
	frame.SetFormat(int(format))

	frameData := (**C.uchar)(frame.GetData())
	frameLinesize := (*C.int)(frame.GetLinesize())

	if ret := C.av_image_fill_arrays(
		frameData, frameLinesize, (*C.uint8_t)(unsafe.Pointer(&data[0])),
		C.AV_PIX_FMT_RGB24, C.int(width), C.int(height), 1); ret < 0 {
		frame.Close()
		return nil, errors.New("attach buffer to frame failed")
	}

	return frame, nil
}

func (f *AVFrame) GetWidth() int {
	return int(f.width)
}

func (f *AVFrame) SetWidth(width int) {
	f.width = C.int(width)
}

func (f *AVFrame) GetHeight() int {
	return int(f.height)
}

func (f *AVFrame) SetHeight(height int) {
	f.height = C.int(height)
}

func (f *AVFrame) GetFormat() int {
	return int(f.format)
}

func (f *AVFrame) SetFormat(format int) {
	f.format = C.int(format)
}

func (f *AVFrame) GetData() unsafe.Pointer {
	return unsafe.Pointer(&f.data[0])
}

func (f *AVFrame) GetLinesize() unsafe.Pointer {
	return unsafe.Pointer(&f.linesize[0])
}

func (f *AVFrame) SetLinesize(linesize [8]int) {
	var arr [8]C.int = [8]C.int{}
	for i := 0; i < 8; i++ {
		arr[i] = C.int(linesize[i])
	}

	f.linesize = arr
}

func (f *AVFrame) GetPTS() int64 {
	return int64(f.pts)
}

func (f *AVFrame) SetPTS(pts int64) {
	f.pts = C.long(pts)
}

func (f *AVFrame) Close() {
	C.av_frame_free(&f)
}
