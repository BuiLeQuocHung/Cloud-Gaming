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
		ctxMap map[SwsCtxKey]chan *C.struct_SwsContext
	}
)

func NewSwsCtxManager() *SwsCtxManager {
	return &SwsCtxManager{
		ctxMap: make(map[SwsCtxKey]chan *C.struct_SwsContext),
	}
}

func (c *SwsCtxManager) Get(k *SwsCtxKey) *C.struct_SwsContext {
	channel, ok := c.ctxMap[*k]
	if !ok {
		c.ctxMap[*k] = make(chan *C.struct_SwsContext, 100)
	}

	if len(channel) == 0 {
		c.ctxMap[*k] <- C.sws_getContext(
			C.int(k.from_width), C.int(k.from_height), int32(k.from),
			C.int(k.width), C.int(k.height), int32(k.to),
			C.int(k.scalingAlgo), nil, nil, nil)
	}

	return <-c.ctxMap[*k]
}

func (c *SwsCtxManager) Set(k *SwsCtxKey, swsCtx *C.struct_SwsContext) {
	_, ok := c.ctxMap[*k]
	if !ok {
		return
	}

	c.ctxMap[*k] <- swsCtx
}

func (c *SwsCtxManager) Reset() {
	for k, _ := range c.ctxMap {
		for len(c.ctxMap[k]) > 0 {
			C.sws_freeContext(<-c.ctxMap[k])
		}
	}

	c.ctxMap = make(map[SwsCtxKey]chan *C.struct_SwsContext)
}
