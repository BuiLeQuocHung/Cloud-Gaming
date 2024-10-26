package video

/*
#cgo pkg-config: libavcodec
#include <libavcodec/avcodec.h>
*/
import "C"

type (
	PixelFormat int
)

const (
	RGB    PixelFormat = C.AV_PIX_FMT_RGB24
	RGBA   PixelFormat = C.AV_PIX_FMT_RGBA
	YUV420 PixelFormat = C.AV_PIX_FMT_YUV420P
)
