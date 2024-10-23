package video

import (
	"cloud_gaming/pkg/encoder"
	"cloud_gaming/pkg/ffmpeg/video"
	"cloud_gaming/pkg/libretro"
	"cloud_gaming/pkg/log"
	"sync"
	"unsafe"

	"go.uber.org/zap"
)

type (
	VideoPipeline struct {
		swsManager *SwsCtxManager
		converter  *Converter

		// pixelFmt, angle will be set by coreEnvironment once the core is loaded
		// will be consistent through core's lifetime
		pixelFmt *PixelFmt
		angle    int

		height int
		width  int

		// fps will be set once game is loaded
		fps float64

		sendVideoFrame SendVideoFrameFunc

		enc encoder.IVideoEncoder
		mu  sync.Mutex
	}

	PixelFmt struct {
		format uint32
		bpp    int //bytes per pixel
	}

	VideoFrame struct {
		Data     []byte             `json:"data"`
		Codec    encoder.VideoCodec `json:"codec"`
		Format   video.VideoFormat  `json:"format"`
		Width    int                `json:"width"`
		Height   int                `json:"height"`
		Duration float64            // in seconds
	}

	SendVideoFrameFunc func(*VideoFrame)
)

func NewVideoPipeline(sendVideoFrame SendVideoFrameFunc) (*VideoPipeline, error) {
	v := &VideoPipeline{
		swsManager:     NewSwsCtxManager(),
		converter:      NewConverter(),
		sendVideoFrame: sendVideoFrame,
		width:          256,
		height:         240,
		enc:            nil,
	}

	return v, nil
}

func (v *VideoPipeline) init() {
	go v.getEncodedDataAndSendFrame()
}

func (v *VideoPipeline) SetSystemVideoInfo(systemAVInfo *libretro.SystemAVInfo) {
	v.fps = systemAVInfo.Timing.FPS
	v.createEncoder()
	v.init()
}

func (v *VideoPipeline) createEncoder() {
	enc, err := encoder.NewVP9Encoder(v.width, v.height, int(v.fps))
	if err != nil {
		log.Debug("create encoder in video pipeline failed", zap.Error(err))
	}
	v.enc = enc
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
		log.Error("convert error", zap.Error(err))
		return
	}
	defer rgbFrame.Close()

	frame, err := v.converter.ConvertAndResize(v.swsManager, rgbFrame, v.width, v.height, video.YUV420)
	if err != nil {
		log.Error("convert and resize error", zap.Error(err))
		return
	}
	defer frame.Close()

	// if v.angle != format.ANGLE0 {
	// 	frameFmt, err = frameFmt.Rotate(v.angle)
	// 	if err != nil {
	// 		log.Error("video pipeline: rotate image error", zap.Error(err))
	// 		return
	// 	}
	// }

	v.mu.Lock()
	defer v.mu.Unlock()
	if v.enc == nil {
		log.Error("encoder is nil")
		return
	}

	err = v.enc.Encode(frame, int(v.fps))
	if err != nil {
		log.Error("encode vp9 error", zap.Error(err))
		return
	}

}

func (v *VideoPipeline) getEncodedDataAndSendFrame() {
	for {
		v.mu.Lock()
		if v.enc == nil {
			v.mu.Unlock()
			break
		}

		data, isOpen := v.enc.GetEncodedData()
		v.mu.Unlock()

		if !isOpen {
			break
		}

		if data == nil {
			continue
		}

		v.sendVideoFrame(&VideoFrame{
			Data:     data,
			Codec:    encoder.VP9,
			Format:   video.YUV420,
			Width:    v.width,
			Height:   v.height,
			Duration: 1 / v.fps * 1000,
		})
	}

	log.Debug("getEncodedDataAndSendFrame has stopped")
}

func (v *VideoPipeline) Close() error {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.enc.Close()
	v.enc = nil

	v.pixelFmt = nil
	v.angle = 0
	v.fps = 0

	v.swsManager.Reset()
	return nil
}
