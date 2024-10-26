package encoder

/*
#cgo pkg-config: libavcodec
#include <libavcodec/avcodec.h>
*/
import "C"
import (
	"cloud_gaming/pkg/ffmpeg/video"
)

type (
	IVideoEncoder interface {
		Encode(*video.AVFrame, int) error
		GetEncodedData() ([]byte, error)
		Close() error
	}

	IAudioEncoder interface {
		Encode([]int16) ([]byte, error)
		Close() error
	}
)
