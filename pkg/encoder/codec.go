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
		GetEncodedData() ([]byte, bool)
		Close() error
	}

	IAudioEncoder interface {
		Encode([]int16) ([]byte, error)
		Close() error
	}

	VideoCodec int
	AudioCodec int
)

const (
	NoVCodec VideoCodec = C.AV_CODEC_ID_NONE
	VP9      VideoCodec = C.AV_CODEC_ID_VP9
)

const (
	NoACodec AudioCodec = iota
	OPUS
)
