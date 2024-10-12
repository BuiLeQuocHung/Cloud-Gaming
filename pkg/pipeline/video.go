package pipeline

import (
	"cloud_gaming/pkg/encoder"
	"cloud_gaming/pkg/format"
	"cloud_gaming/pkg/libretro"
	"cloud_gaming/pkg/log"
	"sync"
	"unsafe"

	"go.uber.org/zap"
)

type (
	VideoPipeline struct {
		// pixelFmt, angle will be set by coreEnvironment once the core is loaded
		// will be consistent through core's lifetime
		pixelFmt *PixelFmt
		angle    format.Angle
		height   int
		width    int

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
		Format   format.VideoFormat `json:"format"`
		Width    int                `json:"width"`
		Height   int                `json:"height"`
		Duration float64            // in seconds
	}

	SendVideoFrameFunc func(*VideoFrame)
)

func NewVideoPipeline(sendVideoFrame SendVideoFrameFunc) (*VideoPipeline, error) {
	v := &VideoPipeline{
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
	v.angle = format.Angle(int(uintptr(data)) % 4)
	v.angle = 1
}

func (v *VideoPipeline) Process(data []byte, width, height, pitch int32) {
	var (
		frameFmt format.IVideoFormat
		err      error
	)

	switch v.pixelFmt.format {
	// RGB
	case libretro.PixelFormat0RGB1555:
		frameFmt, err = format.ConvertRGBtoYUV420(data, int(width), int(height), int(pitch))
	// RGB
	case libretro.PixelFormatRGB565:
		frameFmt, err = format.ConvertRGBtoYUV420(data, int(width), int(height), int(pitch))
	// RGBA
	case libretro.PixelFormatXRGB8888:
		frameFmt, err = format.ConvertRGBAtoYUV420(data, int(width), int(height), int(pitch))
	}

	if err != nil {
		log.Error("convert error", zap.Error(err))
		return
	}

	frameFmt, err = frameFmt.Resize(v.height, v.width)
	if err != nil {
		log.Error("resize error", zap.Error(err))
		return
	}

	if v.angle != 0 {
		frameFmt, err = frameFmt.Rotate(v.angle)
		if err != nil {
			log.Error("video pipeline: rotate image error", zap.Error(err))
		}
	}

	v.mu.Lock()
	defer v.mu.Unlock()
	if v.enc == nil {
		log.Error("encoder is nil")
		return
	}

	err = v.enc.Encode(frameFmt, int(v.fps))
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
			Format:   format.YUV420,
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
	v.angle = format.ANGLE0
	v.fps = 0

	return nil
}
