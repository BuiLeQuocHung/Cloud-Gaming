package video

/*
#cgo pkg-config: libavutil
#include <libavutil/frame.h>
*/
import "C"
import "unsafe"

type (
	AVFrame = C.AVFrame
)

func NewFrame() *AVFrame {
	return C.av_frame_alloc()
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

func (f *AVFrame) Close() {
	C.av_frame_free(&f)
}
