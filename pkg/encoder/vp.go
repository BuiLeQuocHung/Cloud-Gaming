package encoder

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

		isShuttingDown bool
	}
)

func NewVP9Encoder(width, height, fps int, pixFmt video.PixelFormat) (IVideoEncoder, error) {
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
		"cpu-used": "5",  // 0-8 bigger means higher speed but lower quality and compression
		"preset":   "superfast",
	})

	opts := []video.CodecCtxOption{
		video.SetBitrate(2000000),
		video.SetWidth(width),
		video.SetHeight(height),
		video.SetTimebase(*video.NewRational(1, fps)),
		video.SetPixelFormat(int(pixFmt)),
		video.SetGopSize(fps / 2),
		video.SetMaxBFrames(0),
		video.SetThreadCount(10),
		video.SetThreadType(video.ThreadFrame),
		video.SetSkipFrame(video.DiscardNonRef),
	}

	err = video.OpenContext(codecCtx, codec, dict, opts...)
	if err != nil {
		return nil, fmt.Errorf("create vp9 encoder failed: %w", err)
	}

	enc := &VP9Encoder{
		codecCtx: codecCtx,
	}

	return enc, nil
}

func (e *VP9Encoder) Encode(videoFrame *video.AVFrame) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// is shutting down
	if !e.isRunning() {
		return errors.New("vp9 encoder is shutting down")
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
	if err != nil {
		return nil, err
	}
	defer pkt.Close()

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
