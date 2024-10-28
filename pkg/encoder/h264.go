package encoder

import (
	"cloud_gaming/pkg/ffmpeg/utils"
	"cloud_gaming/pkg/ffmpeg/video"
	"errors"
	"fmt"
	"sync"
)

type (
	H264Encoder struct {
		// codec context cannot be accessed concurrently but sequentially
		mu       sync.Mutex
		codecCtx *video.CodecCtx

		width  int
		height int

		isShuttingDown bool

		totalFrames int
	}
)

func NewH264Encoder(width, height, fps int, pixFmt video.PixelFormat) (IVideoEncoder, error) {
	codec, err := video.NewCodec(video.H264)
	if err != nil {
		return nil, err
	}

	codecCtx, err := video.NewCodecCtx(codec)
	if err != nil {
		return nil, err
	}

	dict := video.NewDictionary(map[string]string{
		"crf":    "23",
		"preset": "superfast",
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
		video.SetProfile(video.MainProfile),
		video.SetLevel(40),
	}

	err = video.OpenContext(codecCtx, codec, dict, opts...)
	if err != nil {
		return nil, fmt.Errorf("create h264 encoder failed: %w", err)
	}

	enc := &H264Encoder{
		codecCtx: codecCtx,
		width:    width,
		height:   height,
	}

	return enc, nil
}

func (e *H264Encoder) Encode(videoFrame *video.AVFrame) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// is shutting down
	if !e.isRunning() {
		return nil
	}

	// only need it to be strictly increasing, value does not matter
	videoFrame.SetPTS(int64(e.totalFrames))
	e.totalFrames += 1
	return video.EncodeFrame(e.codecCtx, videoFrame)
}

func (e *H264Encoder) GetEncodedData() ([]byte, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// is shutting down
	if !e.isRunning() {
		return nil, errors.New("h264 encoder is shutting down")
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

// // calculate current frame pts
// func (e *H264Encoder) getFramePts() int64 {
// 	e.totalFrames += 1
// 	time_base := e.codecCtx.GetTimebase()
// 	return int64(float64(e.totalFrames*90000) * time_base.ToFloat())
// }

func (e *H264Encoder) stopping() {
	e.isShuttingDown = true
}

func (e *H264Encoder) isRunning() bool {
	return !e.isShuttingDown
}

func (e *H264Encoder) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.stopping()

	if err := video.Flush(e.codecCtx); err != nil {
		return err
	}

	e.codecCtx.Free()
	return nil
}
