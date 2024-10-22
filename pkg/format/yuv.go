package format

/*
#cgo pkg-config: libavcodec libavformat libavfilter libavutil libswscale libpostproc
#include <libswscale/swscale.h>
#include <libavcodec/avcodec.h>
#include <libavutil/imgutils.h>
#include <libavutil/avutil.h>
#include <libavfilter/avfilter.h>
#include <libavfilter/buffersink.h>
#include <libavfilter/buffersrc.h>
*/
import "C"
import (
	"cloud_gaming/pkg/ffmpeg/video"
	"cloud_gaming/pkg/utils"
	"fmt"
	"math"
	"unsafe"
)

type (
	Yuv420Fmt struct {
		*video.AVFrame
	}
)

// ConvertRGBtoYUV420 converts an RGB byte array to YUV420 format.
func ConvertRGBtoYUV420(rgbData []byte, width, height, pitch int) (IVideoFormat, error) {
	// Create a SwsContext for RGB to YUV420P conversion
	swsCtx := FmtCtx.Get(CtxKey{
		from_width:  width,
		from_height: height,
		from:        RGB,
		width:       width,
		height:      height,
		to:          YUV420,
		scalingAlgo: C.SWS_BILINEAR,
	})
	if swsCtx == nil {
		return nil, fmt.Errorf("ConvertRGBtoYUV420: could not create SwsContext for RGB to YUV conversion")
	}

	// Allocate frame for the YUV420 format
	frame := video.NewFrame()
	if frame == nil {
		return nil, fmt.Errorf("ConvertRGBtoYUV420: failed to allocate AVFrame")
	}

	frame.SetWidth(width)
	frame.SetHeight(height)
	frame.SetFormat(int(YUV420))

	// Allocate buffer for the YUV frame
	numBytes := C.av_image_get_buffer_size(int32(YUV420), C.int(width), C.int(height), 1)
	yuvBuffer := C.av_malloc(C.size_t(numBytes))
	if yuvBuffer == nil {
		return nil, fmt.Errorf("ConvertRGBtoYUV420: could not allocate YUV buffer")
	}

	// Set up the source RGB frame data pointers
	var srcData [8]*C.uint8_t
	var srcLinesize [8]C.int
	srcData[0] = (*C.uint8_t)(unsafe.Pointer(&rgbData[0]))
	srcLinesize[0] = C.int(pitch)

	// Set up the destination YUV frame data pointers
	frameData := (**C.uchar)(frame.GetData())
	frameLinesize := (*C.int)(frame.GetLinesize())
	if ret := C.av_image_fill_arrays(frameData, frameLinesize, (*C.uint8_t)(yuvBuffer), int32(YUV420), C.int(width), C.int(height), 1); ret < 0 {
		return nil, fmt.Errorf("ConvertRGBtoYUV420: set up destination YUV frame failed: %w", utils.CErrorToString(int(ret)))
	}

	// Perform the conversion from RGBA to YUV420P
	if ret := C.sws_scale(swsCtx, &srcData[0], &srcLinesize[0], 0, C.int(height), frameData, frameLinesize); ret != C.int(height) {
		return nil, fmt.Errorf("ConvertRGBtoYUV420: num of rows copied is not equal to height")
	}

	return &Yuv420Fmt{
		AVFrame: frame,
	}, nil
}

func ConvertRGBAtoYUV420(rgbaData []byte, width, height, pitch int) (IVideoFormat, error) {
	// Create a SwsContext for RGBA to YUV420P conversion
	swsCtx := FmtCtx.Get(CtxKey{
		from_width:  width,
		from_height: height,
		from:        RGBA,
		width:       width,
		height:      height,
		to:          YUV420,
		scalingAlgo: C.SWS_BILINEAR,
	})
	if swsCtx == nil {
		return nil, fmt.Errorf("ConvertRGBAtoYUV420: could not create SwsContext for RGBA to YUV420P conversion")
	}

	// Allocate frame for the YUV420 format
	frame := video.NewFrame()
	if frame == nil {
		return nil, fmt.Errorf("ConvertRGBAtoYUV420: failed to allocate AVFrame")
	}

	frame.SetWidth(width)
	frame.SetHeight(height)
	frame.SetFormat(int(YUV420))

	// Allocate buffer for the YUV frame
	numBytes := C.av_image_get_buffer_size(int32(YUV420), C.int(width), C.int(height), 1)
	yuvBuffer := C.av_malloc(C.size_t(numBytes))
	if yuvBuffer == nil {
		return nil, fmt.Errorf("ConvertRGBAtoYUV420: could not allocate YUV buffer")
	}

	// Set up the source RGBA frame data pointers
	var srcData [8]*C.uint8_t
	var srcLinesize [8]C.int
	srcData[0] = (*C.uint8_t)(unsafe.Pointer(&rgbaData[0]))
	srcLinesize[0] = C.int(pitch)

	// Set up the destination YUV frame data pointers
	frameData := (**C.uchar)(frame.GetData())
	frameLinesize := (*C.int)(frame.GetLinesize())
	if ret := C.av_image_fill_arrays(frameData, frameLinesize, (*C.uint8_t)(yuvBuffer), int32(YUV420), C.int(width), C.int(height), 1); ret < 0 {
		return nil, fmt.Errorf("ConvertRGBAtoYUV420: set up destination YUV frame failed: %w", utils.CErrorToString(int(ret)))
	}

	// Perform the conversion from RGBA to YUV420P
	if ret := C.sws_scale(swsCtx, &srcData[0], &srcLinesize[0], 0, C.int(height), frameData, frameLinesize); ret != C.int(height) {
		return nil, fmt.Errorf("ConvertRGBAtoYUV420: num of rows copied is not equal to height")
	}

	return &Yuv420Fmt{
		AVFrame: frame,
	}, nil
}

func (f *Yuv420Fmt) Resize(targetHeight, targetWidth int) (IVideoFormat, error) {
	var (
		width  = f.GetWidth()
		height = f.GetHeight()
		format = f.GetFormat()
	)

	// Close old frame
	defer f.Close()

	swsContext := FmtCtx.Get(CtxKey{
		from_width:  width,
		from_height: height,
		from:        format,
		width:       targetWidth,
		height:      targetHeight,
		to:          format,
		scalingAlgo: C.SWS_BILINEAR,
	})
	if swsContext == nil {
		return nil, fmt.Errorf("Resize YUV420: failed to create SwsContext")
	}

	// Allocate frame
	frame := video.NewFrame()
	if frame == nil {
		return nil, fmt.Errorf("Resize YUV420: failed to allocate AVFrame")
	}

	frame.SetWidth(targetWidth)
	frame.SetHeight(targetHeight)
	frame.SetFormat(int(YUV420))

	// Allocate buffer for the new frame
	numBytes := C.av_image_get_buffer_size(int32(YUV420), C.int(targetWidth), C.int(targetHeight), 1)
	yuvBuffer := C.av_malloc(C.size_t(numBytes))
	if yuvBuffer == nil {
		return nil, fmt.Errorf("Resize YUV420: could not allocate YUV buffer")
	}

	// Set up the new frame data pointers
	newframeData := (**C.uchar)(frame.GetData())
	newframeLinesize := (*C.int)(frame.GetLinesize())
	if ret := C.av_image_fill_arrays(newframeData, newframeLinesize, (*C.uint8_t)(yuvBuffer), int32(YUV420), C.int(targetWidth), C.int(targetHeight), 1); ret < 0 {
		return nil, fmt.Errorf("Resize YUV420: set up new frame failed: %w", utils.CErrorToString(int(ret)))
	}

	// Perform scaling
	oldframeData := (**C.uchar)(f.GetData())
	oldframeLinesize := (*C.int)(f.GetLinesize())
	if ret := C.sws_scale(swsContext, oldframeData, oldframeLinesize, 0, C.int(height), newframeData, newframeLinesize); ret != C.int(targetHeight) {
		return nil, fmt.Errorf("Resize YUV420: num of rows copied is not equal to height")
	}

	return &Yuv420Fmt{
		AVFrame: frame,
	}, nil
}

// RotateYUV420P rotates a YUV420P image by the specified degrees
func (f *Yuv420Fmt) Rotate(angle Angle) (IVideoFormat, error) {
	var (
		rad = fmt.Sprintf("%f", float32(angle)*math.Pi)
	)

	// Close old frame
	defer f.Close()

	// Set up filter graph
	filterGraph := C.avfilter_graph_alloc()
	if filterGraph == nil {
		return nil, fmt.Errorf("failed to allocate filter graph")
	}
	defer C.avfilter_graph_free(&filterGraph)

	// Create buffer source
	buffersrc := C.avfilter_get_by_name(C.CString("buffer"))
	buffersrcCtx := (*C.AVFilterContext)(nil)
	if ret := C.avfilter_graph_create_filter(&buffersrcCtx, buffersrc, C.CString("in"), nil, nil, filterGraph); ret < 0 {
		return nil, fmt.Errorf("failed to create buffer source: %d", ret)
	}

	// Create buffer sink
	buffersink := C.avfilter_get_by_name(C.CString("buffersink"))
	buffersinkCtx := (*C.AVFilterContext)(nil)
	if ret := C.avfilter_graph_create_filter(&buffersinkCtx, buffersink, C.CString("out"), nil, nil, filterGraph); ret < 0 {
		return nil, fmt.Errorf("failed to create buffer sink: %d", ret)
	}

	// Create rotate filter
	rotate := C.avfilter_get_by_name(C.CString("rotate"))
	rotateCtx := (*C.AVFilterContext)(nil)
	if ret := C.avfilter_graph_create_filter(&rotateCtx, rotate, C.CString("rotate"), C.CString(rad), nil, filterGraph); ret < 0 {
		return nil, fmt.Errorf("failed to create rotate filter: %d", ret)
	}

	// Link the filters
	if ret := C.avfilter_link(buffersrcCtx, 0, rotateCtx, 0); ret < 0 {
		return nil, fmt.Errorf("failed to link buffer source to rotate filter: %d", ret)
	}
	if ret := C.avfilter_link(rotateCtx, 0, buffersinkCtx, 0); ret < 0 {
		return nil, fmt.Errorf("failed to link rotate filter to buffer sink: %d", ret)
	}

	// Configure the filter graph
	if ret := C.avfilter_graph_config(filterGraph, nil); ret < 0 {
		return nil, fmt.Errorf("failed to configure filter graph: %d", ret)
	}

	// Push frame to the buffer source
	curFrameData := f.GetFrame()
	if ret := C.av_buffersrc_add_frame(buffersrcCtx, (*C.AVFrame)(unsafe.Pointer(curFrameData))); ret < 0 {
		return nil, fmt.Errorf("failed to add frame to buffer source: %d", ret)
	}

	// Retrieve the rotated frame from the buffer sink
	outputFrame := video.NewFrame()
	if ret := C.av_buffersink_get_frame(buffersinkCtx, (*C.AVFrame)(unsafe.Pointer(outputFrame))); ret < 0 {
		return nil, fmt.Errorf("failed to get frame from buffer sink: %d", ret)
	}

	return &Yuv420Fmt{
		AVFrame: outputFrame,
	}, nil
}

func (f *Yuv420Fmt) GetFrame() *video.AVFrame {
	return f.AVFrame
}

func (f *Yuv420Fmt) GetWidth() int {
	return f.AVFrame.GetWidth()
}

func (f *Yuv420Fmt) GetHeight() int {
	return f.AVFrame.GetHeight()
}

func (f *Yuv420Fmt) GetFormat() VideoFormat {
	return VideoFormat(f.AVFrame.GetFormat())
}

func (f *Yuv420Fmt) Close() {
	f.AVFrame.Close()
}
