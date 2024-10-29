package video

import (
	"cloud_gaming/pkg/ffmpeg/video"
	"cloud_gaming/pkg/libretro"
	"cloud_gaming/pkg/log"
	"unsafe"

	"go.uber.org/zap"
)

type (
	VideoPipeline struct {
		swsManager *SwsCtxManager
		converter  *Converter
		enc        *Encoder

		// pixelFmt, angle will be set by coreEnvironment once the core is loaded
		// will be consistent through core's lifetime
		pixelFmt *PixelFmt
		angle    int

		height    int
		width     int
		codec     video.VideoCodec
		pixFormat video.PixelFormat

		// fps will be set once game is loaded
		fps float64

		sendVideoFrame SendVideoFrameFunc
	}

	PixelFmt struct {
		format uint32
		bpp    int //bytes per pixel
	}

	VideoFrame struct {
		Data     []byte            `json:"data"`
		Codec    video.VideoCodec  `json:"codec"`
		Format   video.PixelFormat `json:"format"`
		Width    int               `json:"width"`
		Height   int               `json:"height"`
		Duration float64           // in milliseconds
	}

	SendVideoFrameFunc func(*VideoFrame)
)

func NewVideoPipeline(sendVideoFrame SendVideoFrameFunc) (*VideoPipeline, error) {
	v := &VideoPipeline{
		swsManager:     NewSwsCtxManager(),
		converter:      NewConverter(),
		sendVideoFrame: sendVideoFrame,
		width:          256 * 1.5,
		height:         240 * 1.5,
		codec:          video.H264,
		pixFormat:      video.YUV420,
	}

	return v, nil
}

func (v *VideoPipeline) Start() {
	enc, err := NewEncoder(v.codec, v.width, v.height, v.pixFormat, v.fps)
	if err != nil {
		log.Error("create encoder failed", zap.Error(err))
		return
	}
	v.enc = enc

	go v.getEncodedDataAndSendFrame()
}

func (v *VideoPipeline) SetSystemVideoInfo(systemAVInfo *libretro.SystemAVInfo) {
	v.fps = systemAVInfo.Timing.FPS
}

func (v *VideoPipeline) SetPixelFormat(data unsafe.Pointer) {
	fmt := libretro.GetPixelFormat(data)
	log.Debug("pixel fmt: ", zap.Uint32("pixFmt", fmt))

	switch fmt {
	case libretro.PixelFormat0RGB1555:
		v.pixelFmt = &PixelFmt{
			format: libretro.PixelFormat0RGB1555,
			bpp:    2,
		}
	case libretro.PixelFormatXRGB8888:
		v.pixelFmt = &PixelFmt{
			format: libretro.PixelFormatXRGB8888,
			bpp:    4,
		}
	case libretro.PixelFormatRGB565:
		v.pixelFmt = &PixelFmt{
			format: libretro.PixelFormatRGB565,
			bpp:    2,
		}
	}
}

func (v *VideoPipeline) SetRotation(data unsafe.Pointer) {
	v.angle = int(uintptr(data)) % 4
}

func (v *VideoPipeline) Process(data []byte, width, height, pitch int32) {
	var (
		rgbFrame *video.AVFrame
		err      error
	)

	switch v.pixelFmt.format {
	// RGB
	case libretro.PixelFormat0RGB1555:
		rgbFrame, err = v.converter.ToFrame(data, int(width), int(height), int(pitch), video.RGB)
		// RGB
	case libretro.PixelFormatRGB565:
		rgbFrame, err = v.converter.ToFrame(data, int(width), int(height), int(pitch), video.RGB)
		// RGBA
	case libretro.PixelFormatXRGB8888:
		rgbFrame, err = v.converter.ToFrame(data, int(width), int(height), int(pitch), video.RGBA)
	}

	if err != nil {
		log.Error("convert failed", zap.Error(err))
		return
	}
	defer rgbFrame.Close()

	frame, err := v.converter.ConvertAndResize(v.swsManager, rgbFrame, v.width, v.height, v.pixFormat)
	if err != nil {
		log.Error("convert and resize failed", zap.Error(err))
		return
	}
	defer frame.Close()

	if v.enc == nil {
		return
	}

	err = v.enc.Encode(frame)
	if err != nil {
		log.Error("encode failed", zap.Error(err))
		return
	}
}

func (v *VideoPipeline) getEncodedDataAndSendFrame() {
	for {
		if v.enc == nil {
			break
		}

		data, err := v.enc.GetEncodedData()
		if err != nil {
			log.Debug("get encoded data failed", zap.Error(err))
			break
		}

		if data == nil {
			continue
		}

		v.sendVideoFrame(&VideoFrame{
			Data:     data,
			Codec:    v.codec,
			Format:   v.pixFormat,
			Width:    v.width,
			Height:   v.height,
			Duration: 1 / v.fps * 1000,
		})
	}

	log.Debug("getEncodedDataAndSendFrame has stopped")
}

func (v *VideoPipeline) Close() error {
	v.enc.Close()
	v.swsManager.Reset()

	return nil
}
