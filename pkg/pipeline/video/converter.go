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
	frame, err := video.NewFrameWithBufferAsArray(width, height, format, data)
	if err != nil {
		return nil, err
	}

	frame.SetLinesize([8]int{pitch})
	return frame, nil
}

func (c *Converter) ConvertAndResize(swsCtxManager *SwsCtxManager, srcFrame *video.AVFrame, targetWidth, targetHeight int, targetFormat video.VideoFormat) (*video.AVFrame, error) {
	swsCtxKey := &SwsCtxKey{
		from_width:  srcFrame.GetWidth(),
		from_height: srcFrame.GetHeight(),
		from_format: video.VideoFormat(srcFrame.GetFormat()),

		to_width:  targetWidth,
		to_height: targetHeight,
		to_format: targetFormat,

		scalingAlgo: C.SWS_BILINEAR,
	}

	swsCtx := swsCtxManager.Get(swsCtxKey)
	defer swsCtxManager.Set(swsCtxKey, swsCtx)

	return video.ScaleAndConvertFrame(swsCtx, srcFrame, targetWidth, targetHeight, targetFormat)
}
