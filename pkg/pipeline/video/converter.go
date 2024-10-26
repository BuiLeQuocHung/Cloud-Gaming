package video

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
func (c *Converter) ToFrame(data []byte, width, height, pitch int, format video.PixelFormat) (*video.AVFrame, error) {
	frame, err := video.NewFrameWithBufferAsArray(width, height, format, data)
	if err != nil {
		return nil, err
	}

	frame.SetLinesize([8]int{pitch})
	return frame, nil
}

func (c *Converter) ConvertAndResize(swsCtxManager *SwsCtxManager, srcFrame *video.AVFrame, targetWidth, targetHeight int, targetFormat video.PixelFormat) (*video.AVFrame, error) {
	swsCtxKey := &SwsCtxKey{
		from_width:  srcFrame.GetWidth(),
		from_height: srcFrame.GetHeight(),
		from_format: video.PixelFormat(srcFrame.GetFormat()),

		to_width:  targetWidth,
		to_height: targetHeight,
		to_format: targetFormat,

		scalingAlgo: video.SWS_BILINEAR,
	}

	swsCtx := swsCtxManager.Get(swsCtxKey)
	defer swsCtxManager.Set(swsCtxKey, swsCtx)

	return video.ScaleAndConvertFrame(swsCtx, srcFrame, targetWidth, targetHeight, targetFormat)
}
