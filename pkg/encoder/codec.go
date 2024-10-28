package encoder

import (
	"cloud_gaming/pkg/ffmpeg/video"
)

type (
	IVideoEncoder interface {
		Encode(*video.AVFrame) error
		GetEncodedData() ([]byte, error)
		Close() error
	}

	IAudioEncoder interface {
		Encode([]int16) ([]byte, error)
		Close() error
	}
)
