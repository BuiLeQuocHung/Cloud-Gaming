package video

/*
#cgo pkg-config: libavcodec libavutil
#include <libavcodec/avcodec.h>
#include <libavutil/dict.h>
*/
import "C"
import (
	"cloud_gaming/pkg/ffmpeg/utils"
	"errors"
	"fmt"
	"unsafe"
)

type (
	CodecCtx   = C.AVCodecContext
	Codec      = C.AVCodec
	Dictionary = C.AVDictionary

	VideoCodec  int
	ThreadType  int
	DiscardType int

	CodecCtxOption func(codecCtx *CodecCtx)
)

const (
	NoCodec VideoCodec = C.AV_CODEC_ID_NONE
	VP9     VideoCodec = C.AV_CODEC_ID_VP9
	H264    VideoCodec = C.AV_CODEC_ID_H264
)

const (
	ThreadFrame ThreadType = C.FF_THREAD_FRAME
	ThreadSclie ThreadType = C.FF_THREAD_SLICE
)

const (
	DiscardNone     DiscardType = C.AVDISCARD_NONE // default
	DiscardNonRef   DiscardType = C.AVDISCARD_NONREF
	DiscardDefault  DiscardType = C.AVDISCARD_DEFAULT
	DiscardBiDir    DiscardType = C.AVDISCARD_BIDIR
	DiscardNonInfra DiscardType = C.AVDISCARD_NONINTRA
	DiscardAll      DiscardType = C.AVDISCARD_ALL
)

func NewCodec(codec_type VideoCodec) (*Codec, error) {
	codec := C.avcodec_find_encoder(uint32(codec_type))
	if codec == nil {
		return nil, errors.New("codec not found")
	}
	return codec, nil
}

func NewCodecCtx(codec *Codec) (*CodecCtx, error) {
	codecCtx := C.avcodec_alloc_context3(codec)
	if codecCtx == nil {
		return nil, errors.New("allocate codec context failed")
	}

	return codecCtx, nil
}

func OpenContext(codecCtx *CodecCtx, codec *Codec, dictionary *Dictionary, options ...CodecCtxOption) error {
	for _, opt := range options {
		opt(codecCtx)
	}

	if ret := C.avcodec_open2(codecCtx, codec, &dictionary); ret < 0 {
		return fmt.Errorf("open codec context failed: %w", utils.CErrorToString(int(ret)))
	}
	return nil
}

func (c *CodecCtx) GetTimebase() Rational {
	return c.time_base
}

func SetWidth(width int) CodecCtxOption {
	return func(c *CodecCtx) {
		c.width = C.int(width)
	}
}

func SetHeight(height int) CodecCtxOption {
	return func(c *CodecCtx) {
		c.height = C.int(height)
	}
}

func SetTimebase(timebase Rational) CodecCtxOption {
	return func(c *CodecCtx) {
		c.time_base = timebase
	}
}

func SetBitrate(bitrate int) CodecCtxOption {
	return func(c *CodecCtx) {
		c.bit_rate = C.long(bitrate)
	}
}

func SetPixelFormat(pixFmt int) CodecCtxOption {
	return func(c *CodecCtx) {
		c.pix_fmt = int32(pixFmt)
	}
}

func SetMaxBFrames(maxBFrames int) CodecCtxOption {
	return func(c *CodecCtx) {
		c.max_b_frames = C.int(maxBFrames)
	}
}

func SetGopSize(gopSize int) CodecCtxOption {
	return func(c *CodecCtx) {
		c.gop_size = C.int(gopSize)
	}
}

func SetThreadCount(threadCount int) CodecCtxOption {
	return func(c *CodecCtx) {
		c.thread_count = C.int(threadCount)
	}
}

func SetThreadType(threadType ThreadType) CodecCtxOption {
	return func(c *CodecCtx) {
		c.thread_type = C.int(threadType)
	}
}

func SetSkipFrame(skipFrame DiscardType) CodecCtxOption {
	return func(c *CodecCtx) {
		c.skip_frame = int32(skipFrame)
	}
}

func SetProfile(profile ProfileType) CodecCtxOption {
	return func(c *CodecCtx) {
		c.profile = C.int(profile)
	}
}

func SetLevel(level int) CodecCtxOption {
	return func(c *CodecCtx) {
		c.level = C.int(level)
	}
}

func EncodeFrame(codecCtx *CodecCtx, frame *AVFrame) error {
	if codecCtx == nil {
		return errors.New("encode frame failed: codec context is nil")
	}

	if ret := C.avcodec_send_frame(codecCtx, frame); ret < 0 {
		return fmt.Errorf("error sending frame for encoding: %w", utils.CErrorToString(int(ret)))
	}

	return nil
}

func GetEncodedPacket(codecCtx *CodecCtx) (*Packet, error) {
	if codecCtx == nil {
		return nil, errors.New("get packet failed: codec context is nil")
	}

	pkt := NewPacket()
	ret := C.avcodec_receive_packet(codecCtx, pkt)
	if ret == 0 {
		return pkt, nil
	}
	defer pkt.Close()

	if ret == -C.EAGAIN {
		return nil, nil
	}

	return nil, utils.CErrorToString(int(ret))
}

func Flush(codecCtx *CodecCtx) error {
	// signal codec context to stop encoding
	err := EncodeFrame(codecCtx, nil)
	if err != nil {
		return err
	}

	// flush remaining frames from codecCtx
	for {
		pkt, err := GetEncodedPacket(codecCtx)
		if err != nil {
			break
		}
		pkt.Close()
	}

	return nil
}

func NewDictionary(m map[string]string) *Dictionary {
	var dict *Dictionary

	for k, v := range m {
		ck := C.CString(k)
		cv := C.CString(v)

		C.av_dict_set(&dict, ck, cv, 0)

		C.free(unsafe.Pointer(ck))
		C.free(unsafe.Pointer(cv))
	}

	return dict
}

func (c *CodecCtx) Free() {
	C.avcodec_free_context(&c)
}
