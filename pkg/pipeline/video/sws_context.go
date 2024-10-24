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
		from_format video.VideoFormat

		to_width  int
		to_height int
		to_format video.VideoFormat

		scalingAlgo int
	}

	SwsCtxManager struct {
		ctxMap map[SwsCtxKey]chan *video.SwsContext
	}
)

func NewSwsCtxManager() *SwsCtxManager {
	return &SwsCtxManager{
		ctxMap: make(map[SwsCtxKey]chan *video.SwsContext),
	}
}

func (c *SwsCtxManager) Get(k *SwsCtxKey) *video.SwsContext {
	channel, ok := c.ctxMap[*k]
	if !ok {
		c.ctxMap[*k] = make(chan *video.SwsContext, 100)
	}

	if len(channel) == 0 {
		c.ctxMap[*k] <- video.NewSwsCtx(k.from_width, k.from_height, k.to_width,
			k.to_height, k.from_format, k.to_format, k.scalingAlgo)
	}

	return <-c.ctxMap[*k]
}

func (c *SwsCtxManager) Set(k *SwsCtxKey, swsCtx *video.SwsContext) {
	_, ok := c.ctxMap[*k]
	if !ok {
		return
	}

	c.ctxMap[*k] <- swsCtx
}

func (c *SwsCtxManager) Reset() {
	for _, channel := range c.ctxMap {
		for len(channel) > 0 {
			(<-channel).Free()
		}
	}

	c.ctxMap = make(map[SwsCtxKey]chan *video.SwsContext)
}
