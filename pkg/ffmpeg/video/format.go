package video

/*
#cgo pkg-config: libavcodec
#include <libavcodec/avcodec.h>
*/
import "C"

type (
	VideoFormat int
)

const (
	RGB    VideoFormat = C.AV_PIX_FMT_RGB24
	RGBA   VideoFormat = C.AV_PIX_FMT_RGBA
	YUV420 VideoFormat = C.AV_PIX_FMT_YUV420P
)
