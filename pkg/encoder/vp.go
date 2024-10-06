package encoder

/*
#cgo pkg-config: libavutil libavcodec libavformat libswscale libdrm liblzma libswresample vpx x264
// #cgo LDFLAGS: -L../../external -lvpx -lm -ldl
#include <libswscale/swscale.h>
#include <libavcodec/avcodec.h>
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
	"cloud_gaming/pkg/format"
	"cloud_gaming/pkg/utils"
	"fmt"
	"log"
	"sync"
	"unsafe"
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

	codecCtx.bit_rate = C.long(1500000)
	codecCtx.width = C.int(width)
	codecCtx.height = C.int(height)
	codecCtx.time_base = C.AVRational{num: 1, den: C.int(fps)}
	codecCtx.pix_fmt = int32(format.YUV420)

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
func (e *VP9Encoder) Encode(videoFrame format.IVideoFormat, fps int) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.codecCtx == nil {
		return nil
	}

	resizedData := videoFrame.GetData()
	codecCtx := e.codecCtx

	// Allocate frame for the encoded output
	frame := C.av_frame_alloc()
	if frame == nil {
		return fmt.Errorf("could not allocate frame")
	}
	defer C.av_frame_free(&frame)

	// Fill the frame with YUV data
	frame.width = codecCtx.width
	frame.height = codecCtx.height
	frame.format = C.int(videoFrame.GetFormat())

	// Allocate buffer for the frame
	numBytes := C.av_image_get_buffer_size(int32(frame.format), frame.width, frame.height, 1)
	buffer := C.av_malloc(C.size_t(numBytes))
	defer C.av_free(buffer)

	// Fill the frame data
	ret := C.av_image_fill_arrays(&frame.data[0], &frame.linesize[0], (*C.uint8_t)(buffer), int32(frame.format), frame.width, frame.height, 1)
	if ret < 0 {
		return utils.CErrorToString(int(ret))
	}

	copy((*[1 << 30]byte)(unsafe.Pointer(buffer))[:numBytes], resizedData)

	if ret := C.avcodec_send_frame(e.codecCtx, frame); ret < 0 {
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
			log.Println("")
			e.mu.Unlock()
			break
		}

		ret := C.avcodec_receive_packet(e.codecCtx, pkt)
		e.mu.Unlock()

		if ret == 0 {
			// log.Println("Finally, it's working")
			e.channel <- C.GoBytes(unsafe.Pointer(pkt.data), pkt.size)
		} else if ret == -C.EAGAIN {
			// log.Println("error: C.EAGAIN")
			continue
		} else {
			log.Println("error: receiving packet: ", utils.CErrorToString(int(ret)))
			break
		}
	}

	log.Println("receiveEncodedPacket has stopped")
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
			log.Println(fmt.Errorf("flushStream: flush stream success: %w", utils.CErrorToString(int(ret))))
			return nil
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
	log.Println("drain channel")
	for len(e.channel) > 0 {
		log.Println("drain channel loop")
		<-e.channel
	}

	return nil
}
