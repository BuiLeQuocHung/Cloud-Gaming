package encoder

/*
#cgo pkg-config: libavutil libavcodec libavformat libswscale libdrm liblzma libswresample vpx x264
#include <libswscale/swscale.h>
#include <libavcodec/avcodec.h>
#include <libavutil/opt.h>
#include <libavutil/avutil.h>
#include <libavutil/imgutils.h>
#include <libavutil/error.h>
#include <vpx/vpx_encoder.h>
#include <errno.h>


#include <stdlib.h>
#include <string.h>
#include <stdint.h>
*/
import "C"

import (
	"cloud_gaming/pkg/ffmpeg/utils"
	"cloud_gaming/pkg/ffmpeg/video"
	"errors"
	"fmt"
	"sync"
)

type (
	VP9Encoder struct {
		// codec context cannot be accessed concurrently but sequentially
		mu       sync.Mutex
		codecCtx *video.CodecCtx

		width  int
		height int
		fps    int

		isShuttingDown bool
	}
)

func NewVP9Encoder(width, height, fps int) (IVideoEncoder, error) {
	codec, err := video.NewCodec(video.VP9)
	if err != nil {
		return nil, err
	}

	codecCtx, err := video.NewCodecCtx(codec)
	if err != nil {
		return nil, err
	}

	dict := video.NewDictionary(map[string]string{
		"crf":      "23", // 0-63 bigger means smaller size but lower quality
		"cpu-used": "4",  // 0-8 bigger means higher speed but lower quality and compression
		"preset":   "faster",
	})

	opts := []video.CodecCtxOption{
		video.SetBitrate(400000),
		video.SetWidth(width),
		video.SetHeight(height),
		video.SetTimebase(*video.NewRational(1, fps)),
		video.SetPixelFormat(int(video.YUV420)),
		video.SetGopSize(fps / 2),
		video.SetMaxBFrames(0),
		video.SetThreadCount(8),
		video.SetThreadType(video.ThreadFrame),
		video.SetSkipFrame(video.DiscardNonRef),
	}

	err = video.OpenContext(codecCtx, codec, dict, opts...)
	if err != nil {
		return nil, fmt.Errorf("create vp9 encoder failed: %w", err)
	}

	enc := &VP9Encoder{
		codecCtx: codecCtx,
		width:    width,
		height:   height,
	}

	return enc, nil
}

func (e *VP9Encoder) Encode(videoFrame *video.AVFrame, fps int) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// is shutting down
	if !e.isRunning() {
		return nil
	}

	return video.EncodeFrame(e.codecCtx, videoFrame)
}

func (e *VP9Encoder) GetEncodedData() ([]byte, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// is shutting down
	if !e.isRunning() {
		return nil, errors.New("vp9 encoder is shutting down")
	}

	pkt, err := video.GetEncodedPacket(e.codecCtx)
	defer pkt.Close()

	if err != nil {
		return nil, err
	}

	if pkt == nil {
		return nil, nil
	}

	return utils.PointerToSlice(pkt.GetData(), pkt.GetSize()), nil
}

func (e *VP9Encoder) stopping() {
	e.isShuttingDown = true
}

func (e *VP9Encoder) isRunning() bool {
	return !e.isShuttingDown
}

func (e *VP9Encoder) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.stopping()

	if err := video.Flush(e.codecCtx); err != nil {
		return err
	}

	e.codecCtx.Free()
	return nil
}
