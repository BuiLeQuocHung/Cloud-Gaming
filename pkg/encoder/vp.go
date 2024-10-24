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
	"cloud_gaming/pkg/ffmpeg/video"
	"cloud_gaming/pkg/log"
	"cloud_gaming/pkg/utils"
	"fmt"
	"sync"
	"unsafe"

	"go.uber.org/zap"
)

type (
	VP9Encoder struct {
		mu       sync.Mutex
		codecCtx *C.AVCodecContext
		codec    *C.AVCodec
		channel  chan []byte

		width  int
		height int
		fps    int
	}
)

func NewVP9Encoder(width, height, fps int) (IVideoEncoder, error) {
	codec := C.avcodec_find_encoder(C.AV_CODEC_ID_VP9)
	if codec == nil {
		return nil, fmt.Errorf("libvpx-vp9 codec not found")
	}

	codecCtx := C.avcodec_alloc_context3(codec)
	if codecCtx == nil {
		return nil, fmt.Errorf("could not allocate codec context")
	}

	codecCtx.bit_rate = C.long(400000)
	codecCtx.rc_max_rate = 600 * 1000 // 2 Mbps maximum bitrate
	codecCtx.rc_min_rate = 200 * 1000 // 500 Kbps minimum bitrate

	codecCtx.width = C.int(width)
	codecCtx.height = C.int(height)

	codecCtx.time_base = C.AVRational{num: 1, den: C.int(fps)}
	codecCtx.pix_fmt = int32(video.YUV420)
	codecCtx.gop_size = C.int(fps) / 2
	codecCtx.max_b_frames = 0

	codecCtx.thread_count = 16
	codecCtx.thread_type = C.FF_THREAD_FRAME

	codecCtx.skip_frame = C.AVDISCARD_NONREF

	crf_s := C.CString("crf")
	defer C.free(unsafe.Pointer(crf_s))
	crf_val_s := C.CString("23") // 0-63 bigger means smaller size but lower quality
	defer C.free(unsafe.Pointer(crf_val_s))
	C.av_opt_set(codecCtx.priv_data, crf_s, crf_val_s, 0)

	cpu_s := C.CString("cpu-used")
	defer C.free(unsafe.Pointer(cpu_s))
	cpu_val_s := C.CString("4") // 0-8 bigger means higher speed but lower quality and compression
	defer C.free(unsafe.Pointer(cpu_val_s))
	C.av_opt_set(codecCtx.priv_data, cpu_s, cpu_val_s, 0)

	preset_s := C.CString("preset")
	defer C.free(unsafe.Pointer(preset_s))
	preset_val_s := C.CString("faster")
	defer C.free(unsafe.Pointer(preset_val_s))
	C.av_opt_set(codecCtx.priv_data, preset_s, preset_val_s, 0)

	if ret := C.avcodec_open2(codecCtx, codec, nil); ret < 0 {
		return nil, fmt.Errorf("could not open codec: %w", utils.CErrorToString(int(ret)))
	}

	ve := &VP9Encoder{
		codecCtx: codecCtx,
		codec:    codec,
		width:    width,
		height:   height,
		channel:  make(chan []byte, 1000),
	}

	ve.init()
	return ve, nil
}

func (e *VP9Encoder) init() {
	go e.receiveEncodedPacket()
}

// Encode encodes YUV data to VP9
func (e *VP9Encoder) Encode(videoFrame *video.AVFrame, fps int) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.codecCtx == nil {
		return nil
	}

	if ret := C.avcodec_send_frame(e.codecCtx, (*C.AVFrame)(unsafe.Pointer(videoFrame))); ret < 0 {
		return fmt.Errorf("error sending frame for encoding: %w", utils.CErrorToString(int(ret)))
	}

	return nil
}

func (e *VP9Encoder) GetEncodedData() ([]byte, bool) {
	select {
	case data, ok := <-e.channel:
		if ok {
			return data, true
		} else {
			// channel is closed
			return nil, false
		}
	default:
		return nil, true
	}
}

// receiveEncodedPacket: retrieve packet from ffmpeg and store in encoder's channel
func (e *VP9Encoder) receiveEncodedPacket() {
	pkt := C.av_packet_alloc()
	if pkt == nil {
		return
	}
	defer C.av_packet_free(&pkt)

	for {
		C.av_packet_unref(pkt) // reuse packet
		e.mu.Lock()
		if e.codecCtx == nil {
			e.mu.Unlock()
			break
		}

		ret := C.avcodec_receive_packet(e.codecCtx, pkt)
		e.mu.Unlock()

		if ret == 0 {
			e.channel <- C.GoBytes(unsafe.Pointer(pkt.data), pkt.size)
		} else if ret == -C.EAGAIN {
			continue
		} else {
			log.Debug("error: receiving packet: ", zap.Error(utils.CErrorToString(int(ret))))
			break
		}
	}

	log.Debug("receiveEncodedPacket has stopped")
}

func (e *VP9Encoder) flushStream() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if ret := C.avcodec_send_frame(e.codecCtx, nil); ret < 0 {
		return fmt.Errorf("flushStream: send flush frame failed: %w", utils.CErrorToString(int(ret)))
	}

	pkt := C.av_packet_alloc()
	if pkt == nil {
		return fmt.Errorf("flushStream: could not allocate packet")
	}
	defer C.av_packet_free(&pkt)

	for {
		C.av_packet_unref(pkt)
		ret := C.avcodec_receive_packet(e.codecCtx, pkt)
		if ret == 0 {
			continue
		} else if ret == -C.EAGAIN {
			break
		} else {
			log.Debug("flushStream: flush stream success", zap.Error(utils.CErrorToString(int(ret))))
			break
		}
	}

	return nil
}

func (e *VP9Encoder) Close() error {
	if err := e.flushStream(); err != nil {
		return err
	}

	if e.codecCtx != nil {
		e.mu.Lock()
		C.avcodec_free_context(&e.codecCtx)
		e.codecCtx = nil
		e.mu.Unlock()
	}

	// drain channel
	for len(e.channel) > 0 {
		<-e.channel
	}

	return nil
}
