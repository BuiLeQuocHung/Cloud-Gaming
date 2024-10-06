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
	"cloud_gaming/pkg/utils"
	"fmt"
	"log"
	"math"
	"unsafe"
)

type (
	Yuv420Fmt struct {
		data   []byte
		width  int
		height int
		format VideoFormat
	}
)

// ConvertRGBtoYUV420 converts an RGB byte array to YUV420 format.
func ConvertRGBtoYUV420(rgbData []byte, width, height, pitch int) (IVideoFormat, error) {
	// Create a SwsContext for RGB to YUV420P conversion
	swsCtx := C.sws_getContext(
		C.int(width), C.int(height), int32(RGB), // Source format
		C.int(width), C.int(height), int32(YUV420), // Destination format
		C.SWS_BILINEAR, nil, nil, nil)
	if swsCtx == nil {
		return nil, fmt.Errorf("ConvertRGBtoYUV420: could not create SwsContext for RGB to YUV conversion")
	}
	defer C.sws_freeContext(swsCtx)

	// Allocate buffer for the destination YUV frame
	numBytes := C.av_image_get_buffer_size(int32(YUV420), C.int(width), C.int(height), 1)
	yuvBuffer := C.av_malloc(C.size_t(numBytes))
	if yuvBuffer == nil {
		return nil, fmt.Errorf("ConvertRGBtoYUV420: could not allocate YUV buffer")
	}
	defer C.av_free(yuvBuffer)

	// Set up the source RGB frame data pointers
	var srcData [4]*C.uint8_t
	var srcLinesize [4]C.int
	srcData[0] = (*C.uint8_t)(unsafe.Pointer(&rgbData[0]))
	srcLinesize[0] = C.int(pitch)

	// Set up the destination YUV frame data pointers
	var dstData [4]*C.uint8_t
	var dstLinesize [4]C.int
	if ret := C.av_image_fill_arrays(&dstData[0], &dstLinesize[0], (*C.uint8_t)(yuvBuffer), int32(YUV420), C.int(width), C.int(height), 1); ret < 0 {
		return nil, fmt.Errorf("ConvertRGBtoYUV420: set up destination YUV frame failed: %w", utils.CErrorToString(int(ret)))
	}

	// Perform the conversion from RGB to YUV420P
	if ret := C.sws_scale(swsCtx, &srcData[0], &srcLinesize[0], 0, C.int(height), &dstData[0], &dstLinesize[0]); ret != C.int(height) {
		return nil, fmt.Errorf("ConvertRGBtoYUV420: num of rows copied is not equal to height")
	}

	// Convert the YUV420P buffer to a Go byte slice
	yuvData := C.GoBytes(unsafe.Pointer(yuvBuffer), C.int(numBytes))

	return &Yuv420Fmt{
		data:   yuvData,
		width:  width,
		height: height,
		format: YUV420,
	}, nil
}

func ConvertRGBAtoYUV420(rgbaData []byte, width, height, pitch int) (IVideoFormat, error) {
	// Create a SwsContext for RGBA to YUV420P conversion
	swsCtx := C.sws_getContext(
		C.int(width), C.int(height), int32(RGBA), // Source format: RGBA
		C.int(width), C.int(height), int32(YUV420), // Destination format: YUV420P
		C.SWS_BILINEAR, nil, nil, nil)
	if swsCtx == nil {
		return nil, fmt.Errorf("ConvertRGBAtoYUV420: could not create SwsContext for RGBA to YUV420P conversion")
	}
	defer C.sws_freeContext(swsCtx)

	// Allocate buffer for the destination YUV frame
	numBytes := C.av_image_get_buffer_size(int32(YUV420), C.int(width), C.int(height), 1)
	yuvBuffer := C.av_malloc(C.size_t(numBytes))
	if yuvBuffer == nil {
		return nil, fmt.Errorf("ConvertRGBAtoYUV420: could not allocate YUV buffer")
	}
	defer C.av_free(yuvBuffer)

	// Set up the source RGBA frame data pointers
	var srcData [4]*C.uint8_t
	var srcLinesize [4]C.int
	srcData[0] = (*C.uint8_t)(unsafe.Pointer(&rgbaData[0]))
	srcLinesize[0] = C.int(pitch)

	// Set up the destination YUV frame data pointers
	var dstData [4]*C.uint8_t
	var dstLinesize [4]C.int
	if ret := C.av_image_fill_arrays(&dstData[0], &dstLinesize[0], (*C.uint8_t)(yuvBuffer), int32(YUV420), C.int(width), C.int(height), 1); ret < 0 {
		return nil, fmt.Errorf("ConvertRGBAtoYUV420: set up destination YUV frame failed: %w", utils.CErrorToString(int(ret)))
	}

	// Perform the conversion from RGBA to YUV420P
	if ret := C.sws_scale(swsCtx, &srcData[0], &srcLinesize[0], 0, C.int(height), &dstData[0], &dstLinesize[0]); ret != C.int(height) {
		return nil, fmt.Errorf("ConvertRGBAtoYUV420: num of rows copied is not equal to height")
	}

	// Convert the YUV420P buffer to a Go byte slice
	yuvData := C.GoBytes(unsafe.Pointer(yuvBuffer), C.int(numBytes))
	return &Yuv420Fmt{
		data:   yuvData,
		width:  width,
		height: height,
		format: YUV420,
	}, nil
}

func (f *Yuv420Fmt) Resize(targetHeight, targetWidth int) (IVideoFormat, error) {
	yuvData := f.data
	width, height := f.width, f.height
	format := f.format

	swsContext := C.sws_getContext(
		C.int(width), C.int(height), int32(format),
		C.int(targetWidth), C.int(targetHeight), int32(format),
		C.SWS_BILINEAR, nil, nil, nil,
	)
	if swsContext == nil {
		return nil, fmt.Errorf("failed to create SwsContext")
	}
	defer C.sws_freeContext(swsContext)

	// Allocate frame for resized output
	resizedFrame := C.av_frame_alloc()
	if resizedFrame == nil {
		return nil, fmt.Errorf("failed to allocate resized AVFrame")
	}
	defer C.av_frame_free(&resizedFrame)

	resizedFrame.width = C.int(targetWidth)
	resizedFrame.height = C.int(targetHeight)
	resizedFrame.format = C.int(format)

	// Allocate buffer for the resized frame
	resizedNumBytes := C.av_image_get_buffer_size(int32(format), C.int(targetWidth), C.int(targetHeight), 1)
	resizedBuffer := C.av_malloc(C.size_t(resizedNumBytes))
	if resizedBuffer == nil {
		return nil, fmt.Errorf("failed to allocate buffer for resized frame")
	}
	defer C.av_free(resizedBuffer)

	// Fill the resized frame
	if ret := C.av_image_fill_arrays(&resizedFrame.data[0], &resizedFrame.linesize[0], (*C.uint8_t)(resizedBuffer), int32(format), C.int(targetWidth), C.int(targetHeight), 1); ret < 0 {
		return nil, fmt.Errorf("ResizeYUV: set up destination YUV frame failed: %w", utils.CErrorToString(int(ret)))
	}

	// Allocate input frame
	inputFrame := C.av_frame_alloc()
	if inputFrame == nil {
		return nil, fmt.Errorf("failed to allocate input AVFrame")
	}
	defer C.av_frame_free(&inputFrame)

	inputFrame.width = C.int(width)
	inputFrame.height = C.int(height)
	inputFrame.format = C.int(format)

	inputNumBytes := C.av_image_get_buffer_size(int32(format), C.int(width), C.int(height), 1)
	inputBuffer := C.av_malloc(C.size_t(inputNumBytes))
	if inputBuffer == nil {
		return nil, fmt.Errorf("failed to allocate buffer for input frame")
	}
	defer C.av_free(inputBuffer)

	// Fill input frame with data
	if ret := C.av_image_fill_arrays(&inputFrame.data[0], &inputFrame.linesize[0], (*C.uint8_t)(inputBuffer), int32(format), C.int(width), C.int(height), 1); ret < 0 {
		return nil, fmt.Errorf("ResizeYUV: set up destination YUV frame failed: %w", utils.CErrorToString(int(ret)))
	}

	// Copy the YUV data to the input frame
	copy((*[1 << 30]byte)(unsafe.Pointer(inputBuffer))[:inputNumBytes], yuvData)

	// Resize the input frame
	if ret := C.sws_scale(swsContext, &inputFrame.data[0], &inputFrame.linesize[0], 0, C.int(height), &resizedFrame.data[0], &resizedFrame.linesize[0]); ret != C.int(targetHeight) {
		log.Println(ret, targetHeight)
		return nil, fmt.Errorf("Rotate YUV420: num of rows copied is not equal to height")
	}

	data := C.GoBytes(unsafe.Pointer(resizedBuffer), C.int(resizedNumBytes))
	return &Yuv420Fmt{
		data:   data,
		width:  targetWidth,
		height: targetHeight,
		format: format,
	}, nil
}

// RotateYUV420P rotates a YUV420P image by the specified degrees
func (f *Yuv420Fmt) Rotate(angle Angle) (IVideoFormat, error) {
	yuvData := f.data
	width := f.width
	height := f.height
	format := f.format
	rad := fmt.Sprintf("%f", float32(angle)*math.Pi)

	// Allocate AVFrame for the input YUV data
	frame := C.av_frame_alloc()
	if frame == nil {
		return nil, fmt.Errorf("failed to allocate frame")
	}
	defer C.av_frame_free(&frame)

	frame.format = C.int(format)
	frame.width = C.int(width)
	frame.height = C.int(height)

	// Allocate buffer for the frame
	if ret := C.av_frame_get_buffer(frame, 32); ret < 0 {
		return nil, fmt.Errorf("failed to get frame buffer: %d", ret)
	}

	// Copy YUV data into AVFrame
	C.memcpy(unsafe.Pointer(&frame.data[0]), unsafe.Pointer(&yuvData[0]), C.size_t(len(yuvData)))

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
	if ret := C.av_buffersrc_add_frame(buffersrcCtx, frame); ret < 0 {
		return nil, fmt.Errorf("failed to add frame to buffer source: %d", ret)
	}

	// Retrieve the rotated frame from the buffer sink
	outputFrame := C.av_frame_alloc()
	defer C.av_frame_free(&outputFrame)

	if ret := C.av_buffersink_get_frame(buffersinkCtx, outputFrame); ret < 0 {
		return nil, fmt.Errorf("failed to get frame from buffer sink: %d", ret)
	}

	// Convert the output frame data back to a byte slice
	outputYUVData := make([]byte, len(yuvData)) // For YUV420P
	C.memcpy(unsafe.Pointer(&outputYUVData[0]), unsafe.Pointer(&outputFrame.data[0]), C.size_t(len(outputYUVData)))

	newWidth, newHeight := width, height
	if angle == ANGLE90 || angle == ANGLE270 {
		newWidth, newHeight = newHeight, newWidth
	}

	return &Yuv420Fmt{
		data:   outputYUVData,
		width:  newWidth,
		height: newHeight,
		format: format,
	}, nil
}

func (f *Yuv420Fmt) GetData() []byte {
	return f.data
}

func (f *Yuv420Fmt) GetWidth() int {
	return f.width
}

func (f *Yuv420Fmt) GetHeight() int {
	return f.height
}

func (f *Yuv420Fmt) GetFormat() VideoFormat {
	return f.format
}
