package video

/*
#cgo pkg-config: libswscale
#include <libswscale/swscale.h>
*/
import "C"
import "cloud_gaming/pkg/ffmpeg/video"

type (
	SwsCtxKey struct {
		from_width  int
		from_height int
		from        video.VideoFormat

		width  int
		height int
		to     video.VideoFormat

		scalingAlgo int
	}

	SwsCtxManager struct {
		ctxMap map[SwsCtxKey]*C.struct_SwsContext
	}
)

func NewSwsCtxManager() *SwsCtxManager {
	return &SwsCtxManager{
		ctxMap: make(map[SwsCtxKey]*C.struct_SwsContext),
	}
}

func (c *SwsCtxManager) Get(k SwsCtxKey) *C.struct_SwsContext {
	if ctx, ok := c.ctxMap[k]; ok {
		return ctx
	}

	c.ctxMap[k] = C.sws_getContext(
		C.int(k.from_width), C.int(k.from_height), int32(k.from),
		C.int(k.width), C.int(k.height), int32(k.to),
		C.int(k.scalingAlgo), nil, nil, nil)

	return c.ctxMap[k]
}

func (c *SwsCtxManager) Reset() {
	for _, v := range c.ctxMap {
		C.sws_freeContext(v)
	}

	c.ctxMap = make(map[SwsCtxKey]*C.struct_SwsContext)
}
