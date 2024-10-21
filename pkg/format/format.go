package format

/*
#cgo pkg-config: libavcodec
#include <libavcodec/avcodec.h>
*/
import "C"

import (
	"cloud_gaming/pkg/ffmpeg/video"
)

type (
	IVideoFormat interface {
		GetFrame() *video.AVFrame
		GetWidth() int
		GetHeight() int
		GetFormat() VideoFormat
		Rotate(d Angle) (IVideoFormat, error)
		Resize(int, int) (IVideoFormat, error)
		Close()
	}

	IAudioFormat interface {
		GetData() []byte
	}

	VideoFormat int
	AudioFormat int
	Angle       int
)

const (
	ANGLE0 Angle = iota
	ANGLE90
	ANGLE180
	ANGLE270
)

const (
	RGB    VideoFormat = C.AV_PIX_FMT_RGB24
	RGBA   VideoFormat = C.AV_PIX_FMT_RGBA
	YUV420 VideoFormat = C.AV_PIX_FMT_YUV420P
)

const (
	PCM AudioFormat = iota
)
